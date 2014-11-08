all: sangobox

sangobox: sangobox.go src/image.go src/agent.go
	go get -d .
	go build -o sangobox sangobox.go

test: main
	go test

clean:
	@rm -rf sangobox

.PHONY: test clean
