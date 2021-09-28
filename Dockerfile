FROM golang:1.14 as builder
MAINTAINER Travis CI GmbH <support+travis-worker-docker-image@travis-ci.org>

COPY . /go/src/github.com/travis-ci/worker
WORKDIR /go/src/github.com/travis-ci/worker
ENV CGO_ENABLED 0
RUN make build

FROM alpine:latest
RUN apk --no-cache add ca-certificates curl bash

COPY --from=builder /go/bin/travis-worker /usr/local/bin/travis-worker
COPY --from=builder /go/src/github.com/travis-ci/worker/systemd.service /app/systemd.service
COPY --from=builder /go/src/github.com/travis-ci/worker/systemd-wrapper /app/systemd-wrapper
COPY --from=builder /go/src/github.com/travis-ci/worker/.docker-entrypoint.sh /docker-entrypoint.sh

VOLUME ["/var/tmp"]
STOPSIGNAL SIGINT

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["/usr/local/bin/travis-worker"]
