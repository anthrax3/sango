all: main image

main: main.go sango/image.go sango/agent.go
	go build -o main main.go

image: main
	./main -norun

test: main
	go test

clean:
	@rm -rf main

.PHONY: test clean
