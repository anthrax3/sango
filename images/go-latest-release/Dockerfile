FROM golang

RUN apt-get update
RUN apt-get install -y time

RUN useradd -m sango

ADD . /home/sango/

WORKDIR /home/sango
RUN go get -d .
RUN go build -o run run.go
RUN chmod 755 run

USER sango
