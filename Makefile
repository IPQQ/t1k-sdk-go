GO         = GO111MODULE=on go
GOBUILD    = $(GO) build -mod=readonly
GOTEST     = $(GO) test -v -p 1

pkgs        = ./pkg/...
examples    = custom_request_response detect_request detect_request_and_response detect_request_with_socket http_server_embedded http_server_embedded_with_response_detection

.PHONY: all
all: test build-examples

.PHONY: build-examples
build-examples:
	mkdir -p build
	for it in $(examples); do \
		$(GOBUILD) $(BUILDFLAGS) -o build/$$it ./examples/$$it ; \
	done

.PHONY: test
test:
	$(GOTEST) -failfast $(pkgs)

.PHONY: lint
lint:
# 'go list' needs to be executed before staticcheck to prepopulate the modules cache.
# Otherwise staticcheck might fail randomly for some reason not yet explained.
	$(GO) list -e -compiled -test=true -export=false -deps=true -find=false -tags= -- ./... > /dev/null
	goimports -local git.in.chaitin.net -w $$(find . -type f -name '*.go' -not -path "./vendor/*")
	golangci-lint run -v --skip-dirs vendor --deadline 10m

.PHONY: clean
clean:
	rm -rf build
