all: main image

main:
	go build -o main main.go

image: main
	./main -norun

clean:
	@rm -rf main
