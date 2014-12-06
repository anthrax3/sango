all: sango

sango: sangobox/sangobox.go src/*.go
	go get -d .
	go build -o sango sangobox/sangobox.go

clean:
	@rm -rf sango

.PHONY: test clean
