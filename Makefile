# Copyright 2018 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

.PHONY: all

CIFS_IMAGE_NAME=quay.io/nak3/cifsplugin
CIFS_IMAGE_VERSION=v0.3.0

all: cifsplugin

test:
	 go test github.com/alternative-storage/cifs-csi/pkg/... -cover
	 go vet github.com/alternative-storage/cifs-csi/pkg/...

cifsplugin:
	if [ ! -d ./vendor ]; then dep ensure; fi
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o  _output/cifsplugin ./cifs

image-cifsplugin: cifsplugin
	cp _output/cifsplugin  deploy/cifs/docker
	docker build -t $(CIFS_IMAGE_NAME):$(CIFS_IMAGE_VERSION) deploy/cifs/docker

push-image-cephfsplugin: image-cifsplugin
	docker push $(CIFS_IMAGE_NAME):$(CIFS_IMAGE_VERSION)

clean:
	go clean -r -x
	rm -f deploy/cifs/docker/cifsplugin
