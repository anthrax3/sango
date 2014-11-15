FROM sango/_base

RUN apt-get install -y clang-3.5

ADD . /home/sango/

WORKDIR /home/sango
RUN go get -d .
RUN go build -o agent agent.go
RUN chmod 755 agent

RUN ./agent test
USER sango
