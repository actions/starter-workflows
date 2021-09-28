package worker

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
	"github.com/sirupsen/logrus"
	"github.com/travis-ci/worker/context"
)

// RemoteController provides an HTTP API for controlling worker.
type RemoteController struct {
	pool       *ProcessorPool
	auth       string
	workerInfo func() workerInfo
	cancel     func()
}

// Setup installs the HTTP routes that will handle requests to the HTTP API.
func (api *RemoteController) Setup() {
	r := mux.NewRouter()

	r.HandleFunc("/healthz", api.HealthCheck).Methods("GET")
	r.HandleFunc("/ready", api.ReadyCheck).Methods("GET")

	r.HandleFunc("/worker", api.GetWorkerInfo).Methods("GET")
	r.HandleFunc("/worker", api.UpdateWorkerInfo).Methods("PATCH")
	r.HandleFunc("/worker", api.ShutdownWorker).Methods("DELETE")

	// It is preferable to use UpdateWorkerInfo to update the pool size,
	// as it does not depend on the current state of worker.
	r.HandleFunc("/pool/increment", api.IncrementPool).Methods("POST")
	r.HandleFunc("/pool/decrement", api.DecrementPool).Methods("POST")

	r.Use(api.SetContext)
	r.Use(api.CheckAuth)
	http.Handle("/", r)
}

// SetContext is a middleware function that loads some values into the request
// context. This allows these values to be shown in logging.
func (api *RemoteController) SetContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		ctx = context.FromComponent(ctx, "remote_controller")
		ctx = context.FromUUID(ctx, uuid.NewRandom().String())

		next.ServeHTTP(w, req.WithContext(ctx))
	})
}

// CheckAuth is a middleware for all HTTP API methods that ensures that the
// configured basic auth credentials were passed in the request.
func (api *RemoteController) CheckAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log := context.LoggerFromContext(req.Context())

		// skip auth for the health and ready check endpoints
		if strings.HasPrefix(req.URL.Path, "/healthz") || strings.HasPrefix(req.URL.Path, "/ready") {
			next.ServeHTTP(w, req)
			return
		}

		username, password, ok := req.BasicAuth()
		if !ok {
			log.Warn("no authentication credentials provided")

			w.Header().Set("WWW-Authenticate", "Basic realm=\"travis-ci/worker\"")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		authBytes := []byte(fmt.Sprintf("%s:%s", username, password))
		if subtle.ConstantTimeCompare(authBytes, []byte(api.auth)) != 1 {
			log.Warn("incorrect credentials provided")

			w.WriteHeader(http.StatusForbidden)
			return
		}

		// pass it on!
		next.ServeHTTP(w, req)
	})
}

// HealthCheck indicates whether worker is currently functioning in a healthy
// way. This can be used by a system like Kubernetes to determine whether to
// replace an instance of worker with a new one.
func (api *RemoteController) HealthCheck(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}

// ReadyCheck indicates whether the worker is ready to receive requests.
// This is intended to be used as a readiness check in a system like Kubernetes.
// We should not attempt to interact with the remote controller until this returns
// a successful status.
func (api *RemoteController) ReadyCheck(w http.ResponseWriter, req *http.Request) {
	if api.pool.Ready() {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprint(w, "Not Ready")
	}
}

// GetWorkerInfo writes a JSON payload with useful information about the current
// state of worker as a whole.
func (api *RemoteController) GetWorkerInfo(w http.ResponseWriter, req *http.Request) {
	log := context.LoggerFromContext(req.Context()).WithField("method", "GetWorkerInfo")

	info := api.workerInfo()
	log.Info("got worker info")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(info)
}

// UpdateWorkerInfo allows reconfiguring some parts of worker on the fly.
//
// The main use of this is adjusting the size of the processor pool without
// interrupting existing running jobs.
func (api *RemoteController) UpdateWorkerInfo(w http.ResponseWriter, req *http.Request) {
	log := context.LoggerFromContext(req.Context()).WithField("method", "UpdateWorkerInfo")

	var info workerInfo
	if err := json.NewDecoder(req.Body).Decode(&info); err != nil {
		log.WithError(err).Error("could not decode json request body")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(errorResponse{
			Message: err.Error(),
		})
		return
	}

	if info.PoolSize > 0 {
		api.pool.SetSize(info.PoolSize)
		log.WithField("pool_size", info.PoolSize).Info("updated pool size")
	}

	w.WriteHeader(http.StatusNoContent)
}

// ShutdownWorker tells the worker to shutdown.
//
// Options can be passed in the body that determine whether the shutdown is
// done gracefully or not.
func (api *RemoteController) ShutdownWorker(w http.ResponseWriter, req *http.Request) {
	log := context.LoggerFromContext(req.Context()).WithField("method", "ShutdownWorker")

	var options shutdownOptions
	if err := json.NewDecoder(req.Body).Decode(&options); err != nil {
		log.WithError(err).Error("could not decode json request body")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(errorResponse{
			Message: err.Error(),
		})
		return
	}

	if options.Graceful {
		api.pool.GracefulShutdown(options.Pause)
	} else {
		api.cancel()
	}

	log.WithFields(logrus.Fields{
		"graceful": options.Graceful,
		"pause":    options.Pause,
	}).Info("asked worker to shutdown")

	w.WriteHeader(http.StatusNoContent)
}

// IncrementPool tells the worker to spin up another processor.
func (api *RemoteController) IncrementPool(w http.ResponseWriter, req *http.Request) {
	log := context.LoggerFromContext(req.Context()).WithField("method", "IncrementPool")

	api.pool.Incr()
	log.Info("incremented pool")

	w.WriteHeader(http.StatusNoContent)
}

// DecrementPool tells the worker to gracefully shutdown a processor.
func (api *RemoteController) DecrementPool(w http.ResponseWriter, req *http.Request) {
	log := context.LoggerFromContext(req.Context()).WithField("method", "DecrementPool")

	api.pool.Decr()
	log.Info("decremented pool")

	w.WriteHeader(http.StatusNoContent)
}

type workerInfo struct {
	Version          string `json:"version"`
	Revision         string `json:"revision"`
	Generated        string `json:"generated"`
	Uptime           string `json:"uptime"`
	PoolSize         int    `json:"poolSize"`
	ExpectedPoolSize int    `json:"expectedPoolSize"`
	TotalProcessed   int    `json:"totalProcessed"`

	Processors []processorInfo `json:"processors"`
}

type processorInfo struct {
	ID        string `json:"id"`
	Processed int    `json:"processed"`
	Status    string `json:"status"`
	LastJobID uint64 `json:"lastJobId"`
}

type shutdownOptions struct {
	Graceful bool `json:"graceful"`
	Pause    bool `json:"pause"`
}

type errorResponse struct {
	Message string `json:"error"`
}
