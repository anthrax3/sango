FROM sango/_base

RUN apt-get install -y clang-3.5

ADD . /tmp/sango
WORKDIR /tmp/sango
RUN mkdir -p $GOPATH/src/github.com/h2so5/sango
RUN if [ -f sango.tar.gz ]; then tar -C $GOPATH/src/github.com/h2so5/sango -xzf sango.tar.gz; rm sango.tar.gz; fi
RUN go get -d .
RUN go build -o agent agent.go
RUN chmod 755 agent
RUN mv agent /usr/bin/agent
RUN agent test

WORKDIR /home/sango
USER sango
