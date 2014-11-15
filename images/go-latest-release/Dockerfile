FROM golang

RUN apt-get update
RUN apt-get install -y time

RUN useradd -m sango

ADD . /home/sango/

WORKDIR /home/sango
RUN go get -d .
RUN go build -o agent agent.go
RUN chmod 755 agent

RUN ./agent test
USER sango
