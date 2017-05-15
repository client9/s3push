
build:
	go build .

run: build
	./s3push
fmt:
	gofmt -w -s *.go
	goimports -w *.go
clean:
	go clean ./...
	git gc --aggressive
	
