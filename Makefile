all: sangobox sango

sangobox: sangobox.go src/image.go src/agent.go
	go build -o sangobox sangobox.go

sango: sango.go src/image.go src/agent.go
	go build -o sango sango.go

test: main
	go test

clean:
	@rm -rf sangobox sango

.PHONY: test clean
