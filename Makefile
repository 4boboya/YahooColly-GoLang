GO_CMD=go
GO_BUILD=$(GO_CMD) build
GO_CLEAN=$(GO_CMD) clean
GO_TEST=$(GO_CMD) test
GO_GET=$(GO_CMD) get
GO_VET=$(GO_CMD) vet
GO_RUN=$(GO_CMD) run
GO_MOD_DEP=$(GO_CMD) mod download
ALL_PATH=./...

BINARY_NAME=app
UNPACK_PATH=$$path

DOCKER_CMD=docker
DOCKER_BUILD=$(DOCKER_CMD) build
DOCKER_PUSH=$(DOCKER_CMD) push
DOCKER_IMAGE_NAME=crawlerstock

deps:
	$(GO_MOD_DEP)
test:
	$(GO_TEST) -v $(ALL_PATH) -cover
build:
	$(GO_BUILD) -o $(BINARY_NAME)
run:
	$(GO_RUN) main.go
clean:
	$(GO_CLEAN)
	rm -f $(BINARY_NAME)
pack:
	tar -cvzf $(BINARY_NAME)-v$(VERSION).tar.gz $(BINARY_NAME) ./
unpack:
	tar -zxf $(BINARY_NAME)-v$(VERSION).tar.gz -C $(UNPACK_PATH)
docker_build:
	@echo "開始打包 Docker Image - $(DOCKER_FULL_IMAGE)"
	$(DOCKER_BUILD) -t $(DOCKER_IMAGE_NAME) .
docker_push:
	@echo "開始 push docker image - $(DOCKER_FULL_IMAGE)"
	$(DOCKER_PUSH) $(DOCKER_IMAGE_NAME)