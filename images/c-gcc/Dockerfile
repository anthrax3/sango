FROM sango/_base

ADD . /tmp/sango
WORKDIR /tmp/sango

RUN if [ -f sango.tar.gz ]; then \
mkdir -p $GOPATH/src/github.com/h2so5/sango; \
tar -C $GOPATH/src/github.com/h2so5/sango -xzf sango.tar.gz; \
rm sango.tar.gz; \
fi

RUN go get -d .
RUN go install github.com/h2so5/sango/tools/jtime
RUN go build -o agent agent.go
RUN chmod 755 agent
RUN mv agent /usr/bin/agent
RUN agent test

WORKDIR /home/sango
USER sango
