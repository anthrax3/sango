all: main image

main: *.go
	go build -o main *.go

image: main
	./main -norun

clean:
	@rm -rf main
