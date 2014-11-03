all: main image

main: main.go src/*.go
	go build -o main main.go

image: main
	./main -norun

clean:
	@rm -rf main
