# Change Log

The format is based on [Keep a Changelog](http://keepachangelog.com/).

## [Unreleased]

### Added

### Changed

### Deprecated

### Removed

### Fixed

## [6.2.4] - 2019-10-29

### Fixed
- Specifying an unsupported language routes the build to default image

## [6.2.3] - 2019-10-11

### Added
- backend/lxd: image selector

## [6.2.2] - 2019-09-30

### Fixed
- services start support on Bionic for TCI Enterprise in docker

## [6.2.1] - 2019-07-18

### Fixed
- config/provider_config: Don't unescape \*ACCOUNT_JSON variables.

## [6.2.0] - 2019-01-09

### Changed
- backend/gce: consistently retry GCP API calls with exponential backoff

### Fixed
- cli: assign job queue priority only when job queue is non-nil
- backend/gce: switch both instance and disk zone when retrying in different
  zone

## [6.1.0] - 2018-12-13

### Added
- backend/gce: retry instance insert in alternate zone on error

## [6.0.0] - 2018-12-13

### Added
- rc: remote controller HTTP API
- processor: set size of processor pool

### Changed
- ratelimit: trace redis connection pool checkout

### Removed
- http: old HTTP API, superseded by remote controller

### Fixed
- backend/gce: add missing rate limit calls
- processor: fix race conditions adding and removing processors

### Security

## [5.1.0] - 2018-11-23

### Added
- ratelimit: dynamic configuration of `max_calls` and `duration` via redis

### Fixed
- backend/gce: gets zone information from warmer in order to delete instances successfully

## [5.0.0] - 2018-11-15

### Added
- backend/gce: cache gce api calls for image and machine type, reduce total api call volume

### Removed
- backend/cloudbrain: the cloudbrain backend is not being used, so we are removing it

### Fixed
- ssh: store exit code in int32 instead of uint8

## [4.6.3] - 2018-11-13

### Changed
- processor: include state, requeue, err in "finished job" log message
- trace: include return code for start_instance, upload_script, run_script, download_trace steps
- trace: instrument CLI.Setup in order to avoid orphaned spans
- backend/gce: track rate of rate limit start calls (number of gce api calls we would make without rate limiting)

### Fixed
- trace:
    * fixed child span rendering for time.Sleep

## [4.6.2] - 2018-11-02

### Added
- trace:
    * backend/gce: calls to `time.Sleep`

### Changed
- backend/gce: track `worker.google.compute.api.client` metric for rate of calls to gce api

## [4.6.1] - 2018-10-30

### Changed
- backend/docker: additional support for additional env vars `HTTPS_PROXY`, `FTP_PROXY`, `NO_PROXY`

## [4.6.0] - 2018-10-30

### Added
- backend/docker: support for setting `HTTP_PROXY` variable in jobs

### Changed
- trace:
    * image/: selector, api_selector, env_selector for all backends
    * backend/gce: calls to google cloud API

## [4.5.2] - 2018-10-24

### Changed
- trace: expanded set of trace functions for stackdriver trace
    * ratelimit (redis and GCE API rate limiting)
    * backend/gce
    * amqp_job
- google: support loading default credentials
- processor: log image name in job finished summary

## [4.5.1] - 2018-10-17

### Added
- processor: track cumulative per-job timings

### Changed
- build: update all dependencies, build binaries via go 1.11.1
- backend/gce: propagate warmed instance name and ip correctly

## [4.5.0] - 2018-10-12

### Added
- trace: opencensus stackdriver trace for worker
- backend/gce: preliminary support for obtaining pre-warmed instances from the warmer service

## [4.4.0] - 2018-10-05

### Added
- trace: build.sh trace download support for jupiterbrain backend

### Changed
- config: refactor config propagation to pass config struct directly

## [4.3.0] - 2018-10-03

### Changed
- amqp_log_writer: propagate more job metadata with "time to first log line" event

### Fixed
- progress: omit text progress folds when progress type is not "text"

## [4.2.0] - 2018-10-01

### Changed
- config: replace `silence-metrics` option with `log-metrics`, changing log metrics from opt-out to opt-in
- backend/gce: support start attributes with `OS` value of `"windows"`

## [4.1.2] - 2018-09-13

### Fixed
- step-run-script: take into account custom job timeouts when checking for hard timeout

## [4.1.1] - 2018-09-13

### Added
- trace: build.sh trace download support for docker backend

### Fixed
- step-run-script: fix job termination in case of log maxsize or hard timeout
- gce: ensure correct ssh permissions on build vms
- gce: fix build instance on context timeout

## [4.1.0] - 2018-08-30

### Added
- backend/jupiterbrain: override for CPU count and RAM in created instances

### Changed
- trace: guard trace download step on the trace flag from the payload
- trace: propagate the trace flag in state update message
- processor: log duration of job execution
- amqp_log_writer: tag first log line with queued_at timestamp for "time to first log line" metric

### Fixed
- backend/gce: fixed host maintenance behaviour when preemptible flag is enabled

## [4.0.1] - 2018-08-22

### Added
- backend/gce: build.sh trace persisting
- backend/gce: make migration behavior on host maintenance events conditional on GPU usage

## [4.0.0] - 2018-07-23

### Added
- amqp_log_writer: support separate AMQP connection for log writing

### Changed
- build: update all dependencies, build binaries via go 1.10.3
- development: move tooling dependencies into the `deps` target
- backend/gce: specify `"TERMINATE"` on host maintenance
- processor: signature of `NewProcessor` to allow for log writer factory
    injection

### Fixed
- backend/gce: use consistent zone value

## [3.12.0] - 2018-07-18

### Added
- backend/docker: support for env-based image selection
- processor: log entries recording time delta since start of processing

## [3.11.0] - 2018-07-12

### Added
- backend/gce, backend/jupiterbrain: incremental progress reporting during
    instance startup

## [3.10.1] - 2018-07-06

### Fixed
- backend/gce: use default disk type when no zone is given via VM config

## [3.10.0] - 2018-07-03

### Added
- backend/gce: support for GPU allocation via VM config

## [3.9.0] - 2018-07-02

### Added
- support for a sharded logs queue (using the rabbitmq-sharding plugin)

## [3.8.2] - 2018-06-21

### Added
- amqp-job-queue: support for setting priority when consuming jobs via
  `x-priority` argument

## [3.8.1] - 2018-06-20

### Added
- cli: create a LogQueue that connects to a separate AMQP server, to prepare for
  splitting the build logs from the current JobQueue

### Changed
- cli: the connection to the AMQP server now uses a configurable AmqpHeartbeat
  option
- Makefile: log output from building or running the tests is now less verbose

### Fixed
- backend/docker_test: check for EOF instead of Nil for archive/tar errors

## [3.8.0] - 2018-05-31

### Added
- amqp-job-queue: pool state updates instead of creating an amqp channel per
    processor

### Fixed
- backend/gce: disable automatic restart
- backend/gce: pass context to all GCE API calls

## [3.7.0] - 2018-04-17

### Added
- logging/metrics: include more fields, alter values in various log entries
- packaging: add systemd service and wrapper script as expected by
  [`tfw`](https://github.com/travis-ci/tfw).

### Changed
- backend: handle "map string" config values delimited by either spaces or
  commas, with potentially URL-encoded parts
- config: change default dist to `trusty`

## [3.6.0] - 2018-03-02

### Added
- backend/docker: include arbitrary container labels from config

### Changed
- backend/gce: add a `no-ip` tag when allocating instance without public IP
- build: support and build using Go 1.9.4

## [3.5.0] - 2018-02-12

### Changed
- amqp-job-queue: create the logUpdates and stateUpdates AMQP channels once per
  job queue instead of per each individual job to prevent channel churn.

### Fixed
- shellcheck test failures

## [3.4.0] - 2017-11-08

### Added
- backend/docker: repo, job ID, and dist container labels

### Changed
- http-job-queue: use polling and refresh claim intervals from job delivery
  responses, if available
- backend/docker: make container inspection interval configurable via
  `INSPECT_INTERVAL`, defaulting to 500ms

### Fixed
- backend/docker: disable cpu set allocation when CPUS is 0, as documented

## [3.3.1] - 2017-11-01

### Changed
- processor: improved logging around requeue conditions

## [3.3.0] - 2017-10-30

### Added
- amqp-job: include instance name in state update sent to hub

### Changed
- backend/gce: make deterministic hostname configurable, defaulting to
  previous behavior

### Fixed
- backend: ensure generated hostnames do not contain `_` and are lowercase

## [3.2.2] - 2017-10-24

### Changed
- packages: drop ubuntu precise, add xenial

### Fixed
- processor: use processor-level context when requeueing or erroring a job

## [3.2.1] - 2017-10-24

### Added
- backend/docker: remove container by name if it already exists

### Fixed
- backend/docker: check cpu sets back in if starting the instance fails

## [3.2.0] - 2017-10-18

### Added
- backend/docker: support for bind-mounted volumes via space-delimited,
  colon-paired values in `TRAVIS_WORKER_DOCKER_BINDS`
- http-job-queue: configurable new job polling interval
- http-job: configurable job refresh claim interval

### Changed
- backend/gce: add site tag to job vms
- cli: switch graceful shutdown + pausing from `SIGWINCH` to `SIGUSR2`
- http-job:
    - account for transitional states when handling state update conflicts
    - delete self under various error conditions indicative of a requeue

### Fixed
- backend/docker: switch to container hostname with dashes instead of dots
- http-job: conditional inclusion of state-related timestamps in state updates
- step-run-script: mark job errored on unknown execution error such as poweroff

## [3.1.0] - 2017-10-06

### Changed
- backend/docker:
    - switch to official client library
    - display original image tag instead of language alias in log output
    - set container name based on repo and job ID
- cli: increase log level to "warn" for messages about receiving certain
  signals.
- processor: use the processor context when sending "Finish" state update

### Fixed
- backend/docker: do not remove links on container removal

## [3.0.2] - 2017-09-12

### Fixed
- backend/docker: properly allocate CPU sets

## [3.0.1] - 2017-09-07
### Fixed
- backend/openstack: allow reauth
- multi-source-job-queue: ignore nil jobs

## [3.0.0] - 2017-08-29
### Added
- backend/openstack: initial support for OpenStack
- log entry fields for improved correlation in various places
- context: convenience funcs for set/get JWT string

### Changed
- http-log-writer: removed buffer in favor of adding directly via log sink
- job: require context in Received and Started
- http-job-queue:
    - drop processor pool from constructor signature
    - add cancellation broadcaster to constructor signature
    - remove caching of build job channel across calls to Jobs
    - use pop + claim processor-level job-board API

### Fixed
- amqp-job:
    - check context when sending state updates
    - timeout after 10s when sending state updates
- amqp-job-queue: timeout after 1s when waiting for deliveries

## [2.11.0] - 2017-07-21
### Added
- step-transform-build-json: support for arbitrary modifications to JSON payload
- build-script-generator: include `source` param in all requests

### Changed
- logging/metrics: improved job lifecycle tracing and queue blocking timing

### Fixed
- http-log-part-sink:
    - rebuild request body with each retry
    - ensure request content type is `application/json`
- http-job: ensure request content type is `application/json`
- packaging: correctly sort tags for definition of `VERSION_TAG`

## [2.10.0] - 2017-07-12
### Added
- cli: optional local-only HTTP API listening on same port as pprof
- http-job-queue: include capacity in each request

### Changed
- http-job-queue: fetch full jobs in series for smoother HTTP traffic

### Fixed
- http-job-queue: check for context cancellation every loop

## [2.9.3] - 2017-06-27
### Changed
- cli: assign singular job queue if only one built

### Fixed
- multi-source-job-queue:
    - ensure each invocation of `Jobs` creates new `Job` channels
    - break on context done to prevent goroutine leakage

## [2.9.2] - 2017-06-16
### Added
- http-job: retry with backoff to job completion request
- http-job-queue: retry with backoff to full job fetch request

### Fixed
- http: ensure all response bodies are closed to prevent file descriptor leakage

## [2.9.1] - 2017-06-09
### Fixed
- ssh: set host key callback on ssh client config

## [2.9.0] - 2017-06-07
### Added
- multi-source-job-queue: funnels arbitrary other job queues into a single source
- "self" field in various log records for correlation
- config: initial sleep duration prior to beginning job execution
- backend/docker: optional API-based image selection

### Changed
- amqp-log-writer, http-log-writer: check context done to prevent goroutine leakage
- http-job-queue:
    - reuse cached build job channel if present
    - check context done to prevent goroutine leakage
    - attach context to all HTTP requests
    - more debug logging
- step-write-worker-info: report the job type (amqp/http/file) in instance line
- processor: check for cancellation in between various steps
- build: support and build using Go 1.8.3

### Fixed
- http-log-writer: flush buffer regularly in the background

## [2.8.2] - 2017-05-17
### Changed
- vagrant: general refresh for development purposes

### Fixed
- amqp-job: ensure `finished_at` timestamp is included with state event when available

## [2.8.1] - 2017-05-11
### Fixed
- backend/docker: ensure parsed tmpfs mount mapping does not include empty keys

## [2.8.0] - 2017-04-12
### Added
- amqp-job: include a state message counter in messages sent to hub
- backend/docker: mount a tmpfs as /run and make it executable, fixing [travis-ci/travis-ci#7062](https://github.com/travis-ci/travis-ci/issues/7062)
- backend/docker: support configurable SHM size, default to 64 MiB
- build-script-generator: include job ID in requests parameters to travis-build
- metrics: add a metric for when a job is finished, without including state name
- sentry: send the current version string to Sentry when reporting errors

### Changed
- amqp-job: send all known timestamps to hub on each state update, including queued-at
- build: support and build using Go 1.8.1

### Fixed
- amqp-canceller: fix error occurring if a job was requeued on the same worker before the previous instance had completely finished, causing cancellations to break
- amqp-job: fix a panic that could occur during shutdown due to an AMQP connection issue
- ssh: update the SSH library, pulling in the fix for [golang/go#18861](https://github.com/golang/go/issues/18861)

## [2.7.0] - 2017-02-08
### Added
- backend: add "SSH dial timeout" to all backends, with a default of 5 seconds, configurable with `SSH_DIAL_TIMEOUT` backend setting
- backend/docker: make the command to run the build script configurable with `TRAVIS_WORKER_DOCKER_EXEC_CMD` env var, default to `bash /home/travis/build.sh`
- backend/gce: make it configurable whether to give a booted instance a "public" IP with `TRAVIS_WORKER_GCE_PUBLIC_IP`, defaults to `true`
- backend/gce: make it configurable whether to connect to an instance's "public" IP with `TRAVIS_WORKER_GCE_PUBLIC_IP_CONNECT`, defaults to `true`
- log when a job is finished, including its "finishing state" (passed, failed, errored, etc.)
- log when a job is requeued

### Changed
- backend/docker: change default command run in Docker native mode from `bash -l /home/travis/build.sh` back to `bash /home/travis/build.sh`, reverting the change made in 2.6.0

## [2.6.2] - 2017-01-31
### Security
- backend/gce: remove service account from booted instances

### Added
- HTTP queue type, including implementations of `JobQueue`, `Job`, and
  `LogWriter`

### Changed
- build-script-generator: accepts `Job` instead of `*simplejson.Json`

### Fixed
- log-writer: pass timeout on creation and start timer on first write

## [2.6.1] - 2017-01-23
### Fixed
- processor: open log writer early, prevent panic

## [2.6.0] - 2017-01-17
### Security
- update to go `v1.7.4`

### Added
- cli: log processor pool total on shutdown
- amqp-job: meter for job finish state (#184)
- amqp-job: track queue time via `queued_at` field from payload
- amqp-job-queue: log job id immediately after json decode
- capture every requeue and send the error to sentry

### Changed
- image/api-selector: selection of candidates by multiple groups
- amqp-canceller: change verbosity of canceller missing job id to debug
- docker: SIGINT `STOPSIGNAL` for graceful shutdown

### Fixed
- processor: always mark job as done when Run finishes
- processor: use errors.Cause when checking error values (error job on log limit reached and similar conditions)
- backend/jupiterbrain: parse SSH key on backend init (#206)
- backend/jupiterbrain: add sleep between creating and wait-for-ip
- backend/docker: run bash with `-l` (login shell) in docker native mode
- image/api-selector: check job-board response status code on image selection
- image/api-selector: check tagsets for trailing commas before querying job-board
- amqp-job-queue: handle context cancellation when delivering build job
- ssh: request a 80x40 PTY

## [2.5.0] - 2016-10-03
### Added
- support for heartbeat URL checks a la [legacy
  worker](https://github.com/travis-ci/travis-worker/blob/4ca25dd/lib/travis/worker/application/http_heart.rb)
- max build log length is now configurable
- alpine-based Docker image
- added runtime-configurable script hooks.

### Changed
- check flags and env vars for start and stop hooks

### Fixed
- Handling `"false"` as a valid boolean false-y value
- logging for start/stop hooks is clearer now
- logging for jupiter-brain boot timeouts is more verbose
- AMQP connections get cleaned up as part of graceful shutdowns

## [2.4.0] - 2016-09-08
### Added
- backend/cloudbrain
- backend/docker: "native" script upload and execution option
- backend/gce: Support for instances without public IPs
- handle `SIGWINCH` by gracefully shutting down processor pool and pausing

### Changed
- (really) static binaries built on go 1.7.1, cross compiled natively
- step/run-script: Add a new line with a link to the docs explaining
  `travis_wait` when there's a log timeout error.
- switch to [keepachangelog](http://keepachangelog.com) format
- upgrade all dependencies
- backend/jupiterbrain: improved error logging via `pkg/errors`

### Removed
- Using [goxc](github.com/laher/goxc) for cross-compilation
- amqp-job: logging channel struct

### Fixed
- Handling `"false"` as a valid boolean false-y value

## [2.3.1] - 2016-05-30
### Changed
- backend/gce: Allow the per-job hard timeout sent from scheduler to be
  longer than the global GCE hard timeout.

### Fixed
- package/rpm: Fix the systemd service file so the Worker can start.

## [2.3.0] - 2016-04-12
### Added
- backend/gce: Add support for separate project ID for images via
  `IMAGE_PROJECT_ID` env var.
- amqp-job: Add support for custom amqp TLS certificates via `--amqp-tls-cert`
  and `--amqp-tls-cert-path` options.

### Fixed
- backend/gce: Requeue jobs on preempted instances (instances preemptively
  shutdown by gce).

## [2.2.0] - 2016-02-01
### Added
- cli: add config option (`--amqp-insecure`) to connect to AMQP without verifying TLS certificates

### Changed
- backend/gce: switched out rate limiter with Redis-backed rate limiter
- backend/gce: create a pseudo-terminal that's 80 wide by 40 high instead of 40 wide by 80 high
- backend/gce: make using preemptible instances configurable
- build: switch to Go 1.5.2 for Travis CI builds

## [2.1.0] - 2015-12-15
### Added
- backend/gce: Add support for a "premium" VM type.
- cli: Show current status and last job id via `SIGUSR1`.
- amqp-job: Track metrics for (started - received) duration.
- cli: Add more percentiles to Librato config.

### Changed
- backend/gce: Use [`multistep`](https://github.com/mitchellh/multistep)
  during instance creation and deletion.
- backend/gce: Allow configuration of script upload and startup timeouts.
- backend/gce: Various improvements to minimize API usage:
  - Exponential backoff while stopping instance.
  - Add option to skip API polling after initial instance delete call.
  - Add metrics to track API polling for instance readiness.
  - Cache the instance IP address as soon as it is available.
  - Add ticker-based rate limiting.
- backend/gce: Generate SSH key pair internally.
- config: DRY up the relationship between config struct & cli flags.

### Fixed
- backend/jupiterbrain: Ensure image name is present in instance line of
  worker summary.
- backend/docker: Pass resource limitations in `HostConfig`.
- backend/docker: Prioritize language tags over default image.

## [2.0.0] - 2015-12-09
### Added
- step/open-log-writer: Display revision URL and startup time in worker info
  header.
- Use [goxc](github.com/laher/goxc) for cross-compilation, constrained to
  `linux amd64` and `darwin amd64`.

### Changed
- backend/gce: Tweaks to minimize API calls, including:
  - Add a sleep interval for before first zone operation check.
  - Break out of instance readiness when `RUNNING` *or* `DONE`, depending
  on script upload SSH connectivity for true instance readiness.
  - Increase default script upload retries to 120.
  - Decrease default script upload sleep interval to 1s.

### Removed
- backend/gce: Remove legacy image selection method and configuration.

### Fixed
- step/run-script: Mark jobs as `errored` when log limit is exceeded,
  preventing infinite requeue.

## [1.4.0] - 2015-12-03
### Added
- backend/docker: Allow disabling the CPU and RAM allocations by setting the
	config options to 0 (this was possible in previous versions, but was not
	documented or supported until now)

### Changed
- backend/gce: Send job ID and repository slug to image selector to help in
	log correlation
- sentry: Only send fatal and panic levels to Sentry, with an option for
	sending errors as well (--sentry-hook-errors)
- image/api-selector: Send os:osx instead of the language when querying for
	an image matching an `osx_image` flag

### Fixed
- amqp-job: Send correct `received_at` and `started_at` timestamps to hub in
	the case of the job finishing before the received or started event is sent

## [1.3.0] - 2015-11-30
### Added
- backend/docker: Tests added by @jacobgreenleaf :heart_eyes_cat:

### Changed
- utils/package-\*, bin/travis-worker-install: Rework and integration into
  base CI suite.
- Switch to using `gvt` for dependency management.
- amqp-job: Send all known timestamps during state updates.
- backend: Set defaults for all `StartAttributes` fields, which are also
  exposed via CLI.

### Removed
- backend/bluebox

### Fixed
- backend/docker: Use correct `HostConfig` when creating container
- image/api-selector: Set `is_default=true` when queries consist of a single
  search dimension.

## [1.2.0] - 2015-11-10
### Added
- image/api:
  - Add image selection query with dist + group + language
  - Add last-ditch image selection query with `is_default=true`
- log-writer: Write folded worker info summary

### Changed
- utils/pkg: Official releases built with go 1.5.1.
- vendor/\*: Updated all vendored dependencies
- utils/lintall: Set 1m deadline for all linters
- backend/jupiterbrain: switch to env image selector

### Fixed
- backend/gce: Removed wonky instance group feature
- step/run-script: Do not requeue if max log length exceeded

## [1.1.1] - 2015-09-10
### Changed
- utils/pkg: updated upstart config to copy/run executable as
  `/var/tmp/run/$UPSTART_JOB`, allowing for multiple worker instances per
  host.

## [1.1.0] - 2015-09-09
### Added
- backend/gce:
  - Configurable image selector, defaulting to legacy selection
    method for backward compatibility.
  - Support for reading account JSON from filename or JSON blob.
  - Optionally add all instances to configurable instance group.
- image/\*: New image selection abstraction with env-based and api-based
  implementations.

### Changed
- vendor/\*: Upgraded all vendored dependencies to latest.
- utils/pkg:
  - Official releases built with go 1.5.
  - Packagecloud script altered to only use ruby stdlib dependencies,
    removing the need for bundler.
- backend/gce: Lots of additional test coverage.
- backend/\*: Introduction of `Setup` func for deferring mutative actions needed
  for runtime.
- config: Addition of `Unset` method on `ProviderConfig`

### Fixed
- processor: Fix graceful shutdown by using `tryClose` on shutdown channel.

## [1.0.0] - 2015-08-19
### Added
- backend/gce: Add auto implode, which will cause a VM to automatically shut
  down after a hard timeout.

### Fixed
- logger: Make the processor= field in the logs not be empty anymore
- sentry: Stringify the err field sent to Sentry, since it's usually parsed
  as a struct, making it just {} in Sentry.

## [0.7.0] - 2015-08-18
### Added
- backend/local

### Changed
- backend/jupiterbrain: Add exponential backoff on all HTTP requests
- sentry: Include stack trace in logs sent to Sentry
- step/generate-script: Add exponential backoff to script generation

### Fixed
- backend/gce: Fix a bug causing VMs for a build language ending in symbols
  (such as C++) to error while booting
- log-writer: Fix a race condition causing the log writer to be closed before
  the logs were fully flushed.
- log-writer: Minimize locking in the internals of the log writer, making
  deadlocks less likely.
- processor: Fix graceful and forceful shutdown when there are still build
  jobs waiting.

## [0.6.0] - 2015-07-23
### Added
- backend/gce
- backend/jupiterbrain: Per-image boot time and count metrics
- step/upload-script: Add a timeout for the script upload (currently 1 minute)

### Changed
- step/upload-script: Treat connection errors as recoverable errors, and requeue the job

### Fixed
- backend/jupiterbrain: Fix a goroutine/memory leak where SSH connections for
  cancelled jobs wouldn't get cleaned up
- logger: Don't print the job UUID if it's blank
- processor: Fix a panic that would sometimes happen on graceful shutdown

## [0.5.2] - 2015-07-16
### Changed
- config: Use the server hostname by default if no Librato source is given
- version: Only print the basename of the binary when showing version

### Fixed
- step/run-script: Print the log timeout and not the hard timeout in the log
  timeout error message [GH-49]

## [0.5.1] - 2015-07-14
### Added
- Runtime pool size management:  Send `SIGTTIN` and `SIGTTOU` signals to
  increase and decrease the pool size during runtime [GH-42]
- Report runtime memory metrics, including GC pause times and rates, and
  goroutine count [GH-45]
- Add more log messages so that all error messages are caught in some way

### Changed
- Many smaller internal changes to remove all lint errors

## [0.5.0] - 2015-07-09
### Added
- backend/bluebox: (See #32)
- main: Lifecycle hooks (See #33)
- config: The log timeout can be set in the configuration
- config: The log timeout and hard timeout can be set per-job in the payload
  from AMQP (See #34)

### Removed
- backend/saucelabs: (See #36)

## [0.4.4] - 2015-07-07
### Added
- backend/docker: Several new configuration settings:
  - `CPUS`: Number of CPUs available to each container (default is 2)
  - `MEMORY`: Amount of RAM available to each container (default is 4GiB)
  - `CMD`: Command to run when starting the container (default is /sbin/init)
- backend/jupiter-brain: New configuration setting: `BOOT_POLL_SLEEP`, the
  time to wait between each poll to check if a VM has booted (default is 3s)
- config: New configuration flag: `silence-metrics`, which will cause metrics
  not to be printed to the log even if no Librato credentials have been
  provided
- main: `SIGUSR1` is caught and will cause each processor in the pool to print
  its current status to the logs
- backend: Add `--help` messages for all backends

### Changed
- backend/docker: Container hostnames now begin with `travis-docker-` instead
  of `travis-go-`

### Fixed
- step/run-script: Format the timeout duration in the log timeout message as a
  duration instead of a float

## [0.4.3] - 2015-06-13
### Added
- each Travis CI build will cause three binaries to be uploaded: One for the
  commit SHA or tag, one for the branch and one for the job number.

## [0.4.2] - 2015-06-13
### Changed
- backend/docker: Improve format of instance ID in the logs for each container

## [0.4.1] - 2015-06-13
### Fixed
- config: Include the `build-api-insecure-skip-verify` when writing the
  configuration using `--echo-config`

## [0.4.0] - 2015-06-13
### Added
- config: New flag: `build-api-insecure-skip-verify`, which will skip
  verifying the TLS certificate when requesting the build script

## [0.3.0] - 2015-06-11
### Changed
- config: Hard timeout is now configurable using `HARD_TIMEOUT`
- backend/docker: Allow for running containers in privileged mode using
  `TRAVIS_WORKER_DOCKER_PRIVILEGED=true`
- main: `--help` will list configuration options
- step/run-script: The instance ID is now printed in the "Using worker" line
  at the top of the job logs
- backend/docker: Instead of just searching for images tagged with
  `travis:<language>`, also search for tags `<language>`, `travis:default` and
  `default`, in that order
- step/upload-script: Requeue job immediately if a build script has been
  uploaded, which is a possible indication of a VM being reused

## [0.2.1] - 2015-06-11
### Changed
- backend/jupiter-brain: More options available for image aliases. Now aliases
  named `<osx_image>`, `osx_image_<osx_image>`,
  `osx_image_<osx_image>_<language>`, `dist_<dist>_<language>`, `dist_<dist>`,
  `group_<group>_<language>`, `group_<group>`, `language_<language>` and
  `default_<os>` will be looked for, in that order.
- logger: The logger always prints key=value formatted logs without colors
- backend/jupiter-brain: Sleep in between requests to check if IP is available

## [0.2.0] - 2015-06-11
### Added
- backend/jupiter-brain

### Changed
- backend/docker: CPUs that can be used by containers scales according to
  number of CPUs available on host
- step/run-script: Print hostname and processor UUID at the top of the job log

## 0.1.0 - 2015-06-11
### Added
- Initial release

[Unreleased]: https://github.com/travis-ci/worker/compare/v6.2.0...HEAD
[6.2.0]: https://github.com/travis-ci/worker/compare/v6.1.0...v6.2.0
[6.1.0]: https://github.com/travis-ci/worker/compare/v6.0.0...v6.1.0
[6.0.0]: https://github.com/travis-ci/worker/compare/v5.1.0...v6.0.0
[5.1.0]: https://github.com/travis-ci/worker/compare/v5.0.0...v5.1.0
[5.0.0]: https://github.com/travis-ci/worker/compare/v4.6.3...v5.0.0
[4.6.3]: https://github.com/travis-ci/worker/compare/v4.6.2...v4.6.3
[4.6.2]: https://github.com/travis-ci/worker/compare/v4.6.1...v4.6.2
[4.6.1]: https://github.com/travis-ci/worker/compare/v4.6.0...v4.6.1
[4.6.0]: https://github.com/travis-ci/worker/compare/v4.5.2...v4.6.0
[4.5.2]: https://github.com/travis-ci/worker/compare/v4.5.1...v4.5.2
[4.5.1]: https://github.com/travis-ci/worker/compare/v4.5.0...v4.5.1
[4.5.0]: https://github.com/travis-ci/worker/compare/v4.4.0...v4.5.0
[4.4.0]: https://github.com/travis-ci/worker/compare/v4.3.0...v4.4.0
[4.3.0]: https://github.com/travis-ci/worker/compare/v4.2.0...v4.3.0
[4.2.0]: https://github.com/travis-ci/worker/compare/v4.1.2...v4.2.0
[4.1.2]: https://github.com/travis-ci/worker/compare/v4.1.1...v4.1.2
[4.1.1]: https://github.com/travis-ci/worker/compare/v4.1.0...v4.1.1
[4.1.0]: https://github.com/travis-ci/worker/compare/v4.0.1...v4.1.0
[4.0.1]: https://github.com/travis-ci/worker/compare/v4.0.0...v4.0.1
[4.0.0]: https://github.com/travis-ci/worker/compare/v3.12.0...v4.0.0
[3.12.0]: https://github.com/travis-ci/worker/compare/v3.11.0...v3.12.0
[3.11.0]: https://github.com/travis-ci/worker/compare/v3.10.1...v3.11.0
[3.10.1]: https://github.com/travis-ci/worker/compare/v3.10.0...v3.10.1
[3.10.0]: https://github.com/travis-ci/worker/compare/v3.9.0...v3.10.0
[3.9.0]: https://github.com/travis-ci/worker/compare/v3.8.2...v3.9.0
[3.8.2]: https://github.com/travis-ci/worker/compare/v3.8.1...v3.8.2
[3.8.1]: https://github.com/travis-ci/worker/compare/v3.8.0...v3.8.1
[3.8.0]: https://github.com/travis-ci/worker/compare/v3.7.0...v3.8.0
[3.7.0]: https://github.com/travis-ci/worker/compare/v3.6.0...v3.7.0
[3.6.0]: https://github.com/travis-ci/worker/compare/v3.5.0...v3.6.0
[3.5.0]: https://github.com/travis-ci/worker/compare/v3.4.0...v3.5.0
[3.4.0]: https://github.com/travis-ci/worker/compare/v3.3.1...v3.4.0
[3.3.1]: https://github.com/travis-ci/worker/compare/v3.3.0...v3.3.1
[3.3.0]: https://github.com/travis-ci/worker/compare/v3.2.2...v3.3.0
[3.2.2]: https://github.com/travis-ci/worker/compare/v3.2.1...v3.2.2
[3.2.1]: https://github.com/travis-ci/worker/compare/v3.2.0...v3.2.1
[3.2.0]: https://github.com/travis-ci/worker/compare/v3.1.0...v3.2.0
[3.1.0]: https://github.com/travis-ci/worker/compare/v3.0.2...v3.1.0
[3.0.2]: https://github.com/travis-ci/worker/compare/v3.0.1...v3.0.2
[3.0.1]: https://github.com/travis-ci/worker/compare/v3.0.0...v3.0.1
[3.0.0]: https://github.com/travis-ci/worker/compare/v2.11.0...v3.0.0
[2.11.0]: https://github.com/travis-ci/worker/compare/v2.10.0...v2.11.0
[2.10.0]: https://github.com/travis-ci/worker/compare/v2.9.3...v2.10.0
[2.9.3]: https://github.com/travis-ci/worker/compare/v2.9.2...v2.9.3
[2.9.2]: https://github.com/travis-ci/worker/compare/v2.9.1...v2.9.2
[2.9.1]: https://github.com/travis-ci/worker/compare/v2.9.0...v2.9.1
[2.9.0]: https://github.com/travis-ci/worker/compare/v2.8.2...v2.9.0
[2.8.2]: https://github.com/travis-ci/worker/compare/v2.8.1...v2.8.2
[2.8.1]: https://github.com/travis-ci/worker/compare/v2.8.0...v2.8.1
[2.8.0]: https://github.com/travis-ci/worker/compare/v2.7.0...v2.8.0
[2.7.0]: https://github.com/travis-ci/worker/compare/v2.6.2...v2.7.0
[2.6.2]: https://github.com/travis-ci/worker/compare/v2.6.1...v2.6.2
[2.6.1]: https://github.com/travis-ci/worker/compare/v2.6.0...v2.6.1
[2.6.0]: https://github.com/travis-ci/worker/compare/v2.5.0...v2.6.0
[2.5.0]: https://github.com/travis-ci/worker/compare/v2.4.0...v2.5.0
[2.4.0]: https://github.com/travis-ci/worker/compare/v2.3.1...v2.4.0
[2.3.1]: https://github.com/travis-ci/worker/compare/v2.3.0...v2.3.1
[2.3.0]: https://github.com/travis-ci/worker/compare/v2.2.0...v2.3.0
[2.2.0]: https://github.com/travis-ci/worker/compare/v2.1.0...v2.2.0
[2.1.0]: https://github.com/travis-ci/worker/compare/v2.0.0...v2.1.0
[2.0.0]: https://github.com/travis-ci/worker/compare/v1.4.0...v2.0.0
[1.4.0]: https://github.com/travis-ci/worker/compare/v1.3.0...v1.4.0
[1.3.0]: https://github.com/travis-ci/worker/compare/v1.2.0...v1.3.0
[1.2.0]: https://github.com/travis-ci/worker/compare/v1.1.1...v1.2.0
[1.1.1]: https://github.com/travis-ci/worker/compare/v1.1.0...v1.1.1
[1.1.0]: https://github.com/travis-ci/worker/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/travis-ci/worker/compare/v0.7.0...v1.0.0
[0.7.0]: https://github.com/travis-ci/worker/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/travis-ci/worker/compare/v0.5.2...v0.6.0
[0.5.2]: https://github.com/travis-ci/worker/compare/v0.5.1...v0.5.2
[0.5.1]: https://github.com/travis-ci/worker/compare/v0.5.0...v0.5.1
[0.5.0]: https://github.com/travis-ci/worker/compare/v0.4.4...v0.5.0
[0.4.4]: https://github.com/travis-ci/worker/compare/v0.4.3...v0.4.4
[0.4.3]: https://github.com/travis-ci/worker/compare/v0.4.2...v0.4.3
[0.4.2]: https://github.com/travis-ci/worker/compare/v0.4.1...v0.4.2
[0.4.1]: https://github.com/travis-ci/worker/compare/v0.4.0...v0.4.1
[0.4.0]: https://github.com/travis-ci/worker/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/travis-ci/worker/compare/v0.2.1...v0.3.0
[0.2.1]: https://github.com/travis-ci/worker/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/travis-ci/worker/compare/v0.1.0...v0.2.0
