# Build Stage
FROM golang:1.20.0-alpine3.17:1.13 AS build-stage

LABEL app="build-k8s-whacky-benchmarks"
LABEL REPO="https://github.com/tehlers320/k8s-whacky-benchmarks"

ENV PROJPATH=/go/src/github.com/tehlers320/k8s-whacky-benchmarks

# Because of https://github.com/docker/docker/issues/14914
ENV PATH=$PATH:$GOROOT/bin:$GOPATH/bin

ADD . /go/src/github.com/tehlers320/k8s-whacky-benchmarks
WORKDIR /go/src/github.com/tehlers320/k8s-whacky-benchmarks

RUN make build-alpine

# Final Stage
FROM golang:1.20.0-alpine3.17

ARG GIT_COMMIT
ARG VERSION
LABEL REPO="https://github.com/tehlers320/k8s-whacky-benchmarks"
LABEL GIT_COMMIT=$GIT_COMMIT
LABEL VERSION=$VERSION

# Because of https://github.com/docker/docker/issues/14914
ENV PATH=$PATH:/opt/k8s-whacky-benchmarks/bin

WORKDIR /opt/k8s-whacky-benchmarks/bin

COPY --from=build-stage /go/src/github.com/tehlers320/k8s-whacky-benchmarks/bin/k8s-whacky-benchmarks /opt/k8s-whacky-benchmarks/bin/
RUN chmod +x /opt/k8s-whacky-benchmarks/bin/k8s-whacky-benchmarks

# Create appuser
RUN adduser -D -g '' k8s-whacky-benchmarks
USER k8s-whacky-benchmarks

ENTRYPOINT ["/usr/bin/dumb-init", "--"]

CMD ["/opt/k8s-whacky-benchmarks/bin/k8s-whacky-benchmarks"]
