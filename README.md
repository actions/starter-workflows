# Worker [![Build Status](https://travis-ci.org/travis-ci/worker.svg?branch=master)](https://travis-ci.org/travis-ci/worker)

Worker is the component of Travis CI that will run a CI job on some form of
compute instance.

It's responsible for getting the bash script from
[travis-build](https://github.com/travis-ci/travis-build), spinning up the
compute instance (VM, Docker container, LXD container, or maybe something different),
uploading the bash script, running it, and streaming the logs back to
[travis-logs](https://github.com/travis-ci/travis-logs). It also sends state
updates to [travis-hub](https://github.com/travis-ci/travis-hub).

## Installing

### from binary

Find the version you wish to install on the [GitHub Releases
page](https://github.com/travis-ci/worker/releases) and download either the
`darwin-amd64` binary for macOS or the `linux-amd64` binary for Linux. No other
operating systems or architectures have pre-built binaries at this time.

### from package

Use the [`./bin/travis-worker-install`](./bin/travis-worker-install) script,
or take a look at the [packagecloud
instructions](https://packagecloud.io/travisci/worker/install).

### from snap

Using a linux distribution which supports [Snaps](https://snapcraft.io/store)
you can run: `sudo snap install travis-worker --edge`

### from source

1. install [Go](http://golang.org) `v1.7+`
1. clone this down into your `$GOPATH`
  * `mkdir -p $GOPATH/src/github.com/travis-ci`
  * `git clone https://github.com/travis-ci/worker $GOPATH/src/github.com/travis-ci/worker`
  * `cd $GOPATH/src/github.com/travis-ci/worker`
1. install [gometalinter](https://github.com/alecthomas/gometalinter):
  * `go get -u github.com/alecthomas/gometalinter`
  * `gometalinter --install`
1. install [shellcheck](https://github.com/koalaman/shellcheck)
1. `make`


## Configuring Travis Worker

Travis Worker is configured with environment variables or command line flags via
the [urfave/cli](https://github.com/urfave/cli) library.  A list of
the non-dynamic flags and environment variables may be found by invoking the
built-in help system:

``` bash
travis-worker --help
```

### Environment-based image selection configuration

Some backend providers support image selection based on environment variables.
The required format uses keys that are prefixed with the provider-specific
prefix:

- `TRAVIS_WORKER_{UPPERCASE_PROVIDER}_IMAGE_{UPPERCASE_NAME}`: contains an image name
  string to be used by the backend provider

The following example is for use with the Docker backend:

``` bash
# matches on `dist: trusty`
export TRAVIS_WORKER_DOCKER_IMAGE_DIST_TRUSTY=travisci/ci-connie:packer-1420290255-fafafaf

# matches on `dist: bionic`
export TRAVIS_WORKER_DOCKER_IMAGE_DIST_BIONIC=registry.business.com/fancy/ubuntu:bionic

# resolves for `language: ruby`
export TRAVIS_WORKER_DOCKER_IMAGE_RUBY=registry.business.com/travisci/ci-ruby:whatever

# resolves for `group: edge` + `language: python`
export TRAVIS_WORKER_DOCKER_IMAGE_GROUP_EDGE_PYTHON=travisci/ci-garnet:packer-1530230255-fafafaf

# used when no dist, language, or group matches
export TRAVIS_WORKER_DOCKER_IMAGE_DEFAULT=travisci/ci-garnet:packer-1410230255-fafafaf
```


## Development: Running Travis Worker locally

This section is for anyone wishing to contribute code to Worker. The code
itself _should_ have godoc-compatible docs (which can be viewed on godoc.org:
<https://godoc.org/github.com/travis-ci/worker>), this is mainly a higher-level
overview of the code.

### Environment

Ensure you've defined the necessary environment variables (see `.example.env`).

### Pull Docker images

```
docker pull travisci/ci-amethyst:packer-1504724461
docker tag travisci/ci-amethyst:packer-1504724461 travis:default
```

### Configuration

For configuration, there are some things like the job-board (`TRAVIS_WORKER_JOB_BOARD_URL`)
and travis-build (`TRAVIS_WORKER_BUILD_API_URI`) URLs that need to be set. These
can be set to the staging values.

```
export TRAVIS_WORKER_JOB_BOARD_URL='https://travis-worker:API_KEY@job-board-staging.travis-ci.com'
export TRAVIS_WORKER_BUILD_API_URI='https://x:API_KEY@build-staging.travis-ci.org/script'
```

`TRAVIS_WORKER_BUILD_API_URI` can be found in the env of the job board app, e.g.:
`heroku config:get JOB_BOARD_BUILD_API_ORG_URL -a job-board-staging`.

#### Images

TODO

#### Configuring the requested provider/backend

Each provider requires its own configuration, which must be provided via
environment variables namespaced by `TRAVIS_WORKER_{PROVIDER}_`.

##### Docker

The backend should be configured to be Docker, e.g.:

``` bash
export TRAVIS_WORKER_PROVIDER_NAME='docker'
export TRAVIS_WORKER_DOCKER_ENDPOINT=unix:///var/run/docker.sock        # or "tcp://localhost:4243"
export TRAVIS_WORKER_DOCKER_PRIVILEGED="false"                          # optional
export TRAVIS_WORKER_DOCKER_CERT_PATH="/etc/secret-docker-cert-stuff"   # optional
```

### Queue configuration

#### File-based queue

For the queue configuration, there is a file-based queue implementation so you
don't have to mess around with RabbitMQ.

You can generate a payload via the `generate-job-payload.rb` script on travis-scheduler:

`heroku run -a travis-scheduler-staging script/generate-job-payload.rb <job id> > payload.json`

Place the file in the `$TRAVIS_WORKER_QUEUE_NAME/10-created.d/` directory, where
it will be picked up by the worker.

See `example-payload.json` for an example payload.

#### AMQP-based queue

```
export TRAVIS_WORKER_QUEUE_TYPE='amqp'
export TRAVIS_WORKER_AMQP_URI='amqp://guest:guest@localhost'
```

The web interface is accessible at http://localhost:15672/

To verify your messages are being published, try:

`rabbitmqadmin get queue=reporting.jobs.builds`

Note: You will first need to install `rabbitmqadmin`. See http://localhost:15672/cli

See `script/publish-example-payload` for a script to enqueue `example-payload.json`.

### Building and running

Run `make build` after making any changes. `make` also executes the test suite.

0. `make`
0. `${GOPATH%%:*}/bin/travis-worker`

or in Docker (FIXME):

0. `docker build -t travis-worker .` # or `docker pull travisci/worker`
0. `docker run --env-file ENV_FILE -ti travis-worker` # or `travisci/worker`

### Testing

Run `make test`. To run backend tests matching `Docker`, for example, run
`go test -v ./backend -test.run Docker`.

### Verifying and exporting configuration

To inspect the parsed configuration in a format that can be used as a base
environment variable configuration, use the `--echo-config` flag, which will
exit immediately after writing to stdout:

``` bash
travis-worker --echo-config
```


## Stopping Travis Worker

Travis Worker has two shutdown modes: Graceful and immediate. The graceful
shutdown will tell the worker to not start any additional jobs but finish the
jobs it is currently running before it shuts down. The immediate shutdown will
make the worker stop the jobs it's working on, requeue them, and clean up any
open resources (shut down VMs, cleanly close connections, etc.)

To start a graceful shutdown, send an INT signal to the worker (for example
using `kill -INT`). To start an immediate shutdown, send a TERM signal to the
worker (for example using `kill -TERM`).

## Go dependency management

Travis Worker is built via the standard `go` commands and dependencies managed
by using Go Modules.

## Release process

Since we want to easily keep track of worker changes, we often associate them with a version number.
To find out the current version, check the [changelog](https://github.com/travis-ci/worker/blob/master/CHANGELOG.md) or run `travis-worker --version`.
We typically use [semantic versioning](https://semver.org/) to determine how to increase this number.

Once you've decided what the next version number should be, update the [changelog](https://github.com/travis-ci/worker/blob/master/CHANGELOG.md) making sure you include all relevant changes that happened since the previous version was tagged. You can see these by running `git diff vX.X.X...HEAD`, where `v.X.X.X` is the name of the previous version.

Once the changelog has been updated and merged to `master`, the merge commit needs to be signed and manually tagged with the version number. To do this, run:

```
git tag --sign -a vX.X.X -m "Worker version vX.X.X"
git push origin vX.X.X
```

The Travis build corresponding to this push should build and upload a worker image with the new tag to [Dockerhub](https://hub.docker.com/r/travisci/worker/tags/).

The next step is to create a new [Github release tag](https://github.com/travis-ci/worker/releases/new) with the appropriate information from the changelog.

## License and Copyright Information

See [LICENSE file](./LICENSE).

© 2018 Travis CI GmbH
