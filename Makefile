all: sangobox

sangobox: main.go src/image.go src/agent.go
	go build -o sangobox main.go

test: main
	go test

clean:
	@rm -rf sangobox

.PHONY: test clean
