# Copyright 2022 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

VENDIR := $(abspath ../tools)/bin/vendir

.PHONY: all
all: clean sync render move validate

.PHONY: clean
clean:
	rm -rf **/build

.PHONY: sync
sync:
	$(VENDIR) sync --lock-file /dev/null

.PHONY: render
render:
	./render.sh

.PHONY: move
move:
	./move.sh

.PHONY: validate
validate: sync render
	./validate.sh
