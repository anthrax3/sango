FROM golang:1.2

RUN useradd -m sango

RUN mkdir /tmp/sango
ADD . /tmp/sango
WORKDIR /tmp/sango

RUN if [ ! -f agent ]; then go get -d .; go build -o agent agent.go; fi
RUN chmod 755 agent
RUN mv agent /usr/bin/agent
RUN agent test

WORKDIR /home/sango
USER sango
