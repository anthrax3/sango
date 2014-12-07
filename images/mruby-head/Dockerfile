FROM sango/_base

RUN apt-get install -y ruby bison

RUN git clone --depth=1 https://github.com/mruby/mruby.git
WORKDIR mruby

RUN ./minirake

RUN chmod 755 /mruby/build/host/bin/mruby
ENV PATH /mruby/build/host/bin:$PATH

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
