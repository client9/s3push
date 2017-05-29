
build:
	go build .   ## build

install:
	go install .  ## install into GOPATH/bin

setup: ## Install all the build and lint dependencies
	go get -u github.com/alecthomas/gometalinter
	go get -u github.com/golang/dep/...
	dep ensure
	gometalinter --install

lint: ## Run all the linters
	gometalinter --vendor --disable-all \
                --enable=deadcode \
		--enable=errcheck \
                --enable=ineffassign \
                --enable=gosimple \
                --enable=staticcheck \
                --enable=gofmt \
                --enable=goimports \
                --enable=dupl \
                --enable=misspell \
                --enable=vetshadow \
                --deadline=10m \
                ./...

ci: lint test ## Run all the tests and code checks

run: build
	./s3push
fmt:
	gofmt -w -s *.go
	goimports -w *.go
clean:
	go clean ./...
	git gc --aggressive

# Absolutely awesome: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'


.PHONY: ci fmt clean build run lint setup

