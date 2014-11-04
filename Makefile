all: main image

main: *.go
	go build -o main *.go

image: main
	./main -norun

test: main
	go test

clean:
	@rm -rf main

.PHONY: test clean
