
build:
	go build .

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

.PHONY: ci fmt clean build run lint setup

