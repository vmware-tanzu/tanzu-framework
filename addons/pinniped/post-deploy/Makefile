# If you update this file, please follow:
# https://suva.sh/posts/well-documented-makefiles/

.DEFAULT_GOAL:=help

export GOPROXY ?= https://build-artifactory.eng.vmware.com/gocenter.io/
export GOSUMDB ?= off

GIT_VERSION ?= $(shell git describe --always --tags)
GIT_REF_LONG = $(shell git rev-parse --verify HEAD)
# Release version
VERSION ?= $(GIT_VERSION)

.PHONY: fmt
fmt: ## Format the code base
	gofmt -w ./pkg ./main.go
	goimports -w ./pkg ./main.go

native: ## Build binary
	CGO_ENABLED=0 go build -o tkg-pinniped-post-deploy -ldflags="-s -w -X $(GOTARGET)/pkg/buildinfo.Version=$(GIT_VERSION) -X github.com/vmware-tanzu-private/core/addons/pinniped/post-deploy/pkg/buildinfo.GitSHA=$(GIT_REF_LONG)" github.com/vmware-tanzu-private/core/addons/pinniped/post-deploy

.PHONY: build-images
build-images: ## Build tkg-pinniped-post-deploy docker images
	VERSION=$(VERSION) ./hack/scripts/build-images.sh

run:
	go run main.go \
        --supervisor-namespace=pinniped-supervisor \
        --concierge-namespace=pinniped-concierge \
        --supervisor-svc-name=pinniped-supervisor \
        --federationdomain-name=pinniped-federation-domain \
        --jwtauthenticator-name=tkg-jwt-authenticator \
        --supervisor-cert-name=pinniped-cert \
        --dex-namespace=tanzu-system-auth \
        --dex-svc-name=dexsvc \
        --dex-cert-name=dex-cert \
        --dex-configmap-name=dex