FROM golang:1.14.4 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

# Copy the source
COPY main.go main.go
COPY pkg/ pkg/
COPY Makefile Makefile
#COPY .git/ .git/

RUN make native

FROM kscom-docker-local.artifactory.eng.vmware.com/debian-base-amd64:v2.1.3
RUN apt-get update -y \
    && DEBIAN_FRONTEND=noninteractive apt-get install ca-certificates -y
WORKDIR /
COPY --from=builder /workspace/tkg-pinniped-post-deploy .
CMD ["/tkg-pinniped-post-deploy"]
