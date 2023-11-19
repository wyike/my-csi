export IMAGE_REGISTRY
export IMAGE_NAME ?= wyike/my-csi
export IMAGE_VERSION ?= dev

.PHONY: docker-image
docker-image:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s"
	docker buildx build --platform linux/amd64  -f Dockerfile -t "$(IMAGE_NAME):$(IMAGE_VERSION)"  . --load
	docker tag $(IMAGE_NAME):$(IMAGE_VERSION) $(IMAGE_REGISTRY)/$(IMAGE_NAME):$(IMAGE_VERSION)
	docker push $(IMAGE_REGISTRY)/$(IMAGE_NAME):$(IMAGE_VERSION)