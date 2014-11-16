FROM golang

RUN apt-get update
RUN apt-get install -y time

RUN useradd -m sango

ADD . /home/sango/

WORKDIR /home/sango
RUN if [ ! -f agent ]; then go get -d .; go build -o agent agent.go; fi
RUN chmod 755 agent

RUN ./agent test
USER sango
