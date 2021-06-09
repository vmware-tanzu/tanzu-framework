# Ensure Make is run with bash shell as some syntax below is bash-specific
SHELL:=bash
.ONESHELL:
.EXPORT_ALL_VARIABLES:
.SHELLFLAGS := -eu -o pipefail -c
.DELETE_ON_ERROR:
MAKEFLAGS += --no-builtin-rules

UID := $(shell id -u)
GID := $(shell id -g)

rwildcard=$(foreach d,$(wildcard $(1:=/*)),$(call rwildcard,$d,$2) $(filter $(subst *,%,$2),$d))

# Hosts running SELinux need :z added to volume mounts
SELINUX_ENABLED := $(shell cat /sys/fs/selinux/enforce 2> /dev/null || echo 0)

ifeq ($(SELINUX_ENABLED),1)
  DOCKER_VOL_OPTS?=:z
endif

.DEFAULT_GOAL:=help

# Use GOPROXY environment variable if set
GOPROXY := $(shell go env GOPROXY)
ifeq ($(GOPROXY),)
GOPROXY := https://proxy.golang.org
endif
export GOPROXY


## --------------------------------------
## Help
## --------------------------------------

help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
