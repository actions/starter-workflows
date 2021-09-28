package worker

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	gocontext "context"

	"github.com/travis-ci/worker/config"
	"github.com/travis-ci/worker/metrics"
)

// A BuildScriptGeneratorError is sometimes used by the Generate method on a
// BuildScriptGenerator to return more metadata about an error.
type BuildScriptGeneratorError struct {
	error

	// true when this error can be recovered by retrying later
	Recover bool
}

// A BuildScriptGenerator generates a build script for a given job payload.
type BuildScriptGenerator interface {
	Generate(gocontext.Context, Job) ([]byte, error)
}

type webBuildScriptGenerator struct {
	URL               string
	aptCacheHost      string
	npmCacheHost      string
	paranoid          bool
	fixResolvConf     bool
	fixEtcHosts       bool
	cacheType         string
	cacheFetchTimeout int
	cachePushTimeout  int
	s3CacheOptions    s3BuildCacheOptions

	httpClient *http.Client
}

type s3BuildCacheOptions struct {
	scheme          string
	region          string
	bucket          string
	accessKeyID     string
	secretAccessKey string
}

// NewBuildScriptGenerator creates a generator backed by an HTTP API.
func NewBuildScriptGenerator(cfg *config.Config) BuildScriptGenerator {
	return &webBuildScriptGenerator{
		URL:               cfg.BuildAPIURI,
		aptCacheHost:      cfg.BuildAptCache,
		npmCacheHost:      cfg.BuildNpmCache,
		paranoid:          cfg.BuildParanoid,
		fixResolvConf:     cfg.BuildFixResolvConf,
		fixEtcHosts:       cfg.BuildFixEtcHosts,
		cacheType:         cfg.BuildCacheType,
		cacheFetchTimeout: int(cfg.BuildCacheFetchTimeout.Seconds()),
		cachePushTimeout:  int(cfg.BuildCachePushTimeout.Seconds()),
		s3CacheOptions: s3BuildCacheOptions{
			scheme:          cfg.BuildCacheS3Scheme,
			region:          cfg.BuildCacheS3Region,
			bucket:          cfg.BuildCacheS3Bucket,
			accessKeyID:     cfg.BuildCacheS3AccessKeyID,
			secretAccessKey: cfg.BuildCacheS3SecretAccessKey,
		},
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: cfg.BuildAPIInsecureSkipVerify,
				},
			},
		},
	}
}

func (g *webBuildScriptGenerator) Generate(ctx gocontext.Context, job Job) ([]byte, error) {
	payload := job.RawPayload()

	if g.aptCacheHost != "" {
		payload.SetPath([]string{"hosts", "apt_cache"}, g.aptCacheHost)
	}
	if g.npmCacheHost != "" {
		payload.SetPath([]string{"hosts", "npm_cache"}, g.npmCacheHost)
	}

	payload.Set("paranoid", g.paranoid)
	payload.Set("fix_resolv_conf", g.fixResolvConf)
	payload.Set("fix_etc_hosts", g.fixEtcHosts)

	if g.cacheType != "" {
		payload.SetPath([]string{"cache_options", "type"}, g.cacheType)
		payload.SetPath([]string{"cache_options", "fetch_timeout"}, g.cacheFetchTimeout)
		payload.SetPath([]string{"cache_options", "push_timeout"}, g.cachePushTimeout)
		payload.SetPath([]string{"cache_options", "s3", "scheme"}, g.s3CacheOptions.scheme)
		payload.SetPath([]string{"cache_options", "s3", "region"}, g.s3CacheOptions.region)
		payload.SetPath([]string{"cache_options", "s3", "bucket"}, g.s3CacheOptions.bucket)
		payload.SetPath([]string{"cache_options", "s3", "access_key_id"}, g.s3CacheOptions.accessKeyID)
		payload.SetPath([]string{"cache_options", "s3", "secret_access_key"}, g.s3CacheOptions.secretAccessKey)
	}

	b, err := payload.Encode()
	if err != nil {
		return nil, err
	}

	var token string
	u, err := url.Parse(g.URL)
	if err != nil {
		return nil, err
	}
	if u.User != nil {
		token = u.User.Username()
		u.User = nil
	}

	jp := job.Payload()
	if jp != nil {
		q := u.Query()
		q.Set("job_id", strconv.FormatUint(jp.Job.ID, 10))
		q.Set("source", "worker")
		u.RawQuery = q.Encode()
	}

	buf := bytes.NewBuffer(b)
	req, err := http.NewRequest("POST", u.String(), buf)
	if err != nil {
		return nil, err
	}
	if token != "" {
		req.Header.Set("Authorization", "token "+token)
	}
	req.Header.Set("User-Agent", fmt.Sprintf("worker-go v=%v rev=%v d=%v", VersionString, RevisionString, GeneratedString))
	req.Header.Set("Content-Type", "application/json")

	startRequest := time.Now()

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	metrics.TimeSince("worker.job.script.api", startRequest)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 500 {
		return nil, BuildScriptGeneratorError{error: fmt.Errorf("server error: %q", string(body)), Recover: true}
	} else if resp.StatusCode >= 400 {
		return nil, BuildScriptGeneratorError{error: fmt.Errorf("client error: %q", string(body)), Recover: false}
	}

	return body, nil
}
