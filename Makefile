.PHONY: all

CIFS_IMAGE_NAME=quay.io/nak3/cifsplugin
CIFS_IMAGE_VERSION=v0.3.0

all: cifsplugin

test:
	 go test github.com/alternative-storage/cifs-csi/pkg/... -cover
	 go vet github.com/alternative-storage/cifs-csi/pkg/...

cifsplugin:
	if [ ! -d ./vendor ]; then dep ensure -vendor-only; fi
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o  _output/cifsplugin ./cifs

image-cifsplugin: cifsplugin
	cp _output/cifsplugin  deploy/cifs/docker
	docker build -t $(CIFS_IMAGE_NAME):$(CIFS_IMAGE_VERSION) deploy/cifs/docker

push-image-cephfsplugin: image-cifsplugin
	docker push $(CIFS_IMAGE_NAME):$(CIFS_IMAGE_VERSION)

clean:
	go clean -r -x
	rm -f deploy/cifs/docker/cifsplugin
