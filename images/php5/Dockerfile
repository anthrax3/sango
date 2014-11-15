FROM sango/_base

ADD . /home/sango/

WORKDIR /home/sango
RUN go get -d .
RUN go build -o agent agent.go
RUN chmod 755 agent

RUN ./agent test
USER sango
