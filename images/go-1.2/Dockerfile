FROM golang:1.2

RUN useradd -m sango

ADD . /home/sango/

WORKDIR /home/sango
RUN go get -d .
RUN go build -o run run.go
RUN chmod 755 run

RUN ./run -t
USER sango
