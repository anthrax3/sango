FROM sango/_base

ADD . /home/sango/
WORKDIR /home/sango
RUN if [ ! -f agent ]; then go get -d .; go build -o agent agent.go; fi
RUN chmod 755 agent

WORKDIR /
RUN rm -rf go
RUN hg clone -u default https://code.google.com/p/go
WORKDIR /go/src
RUN rm -v **/**/pprof_test.go
RUN ./all.bash
ENV PATH /go/bin:$PATH
ENV GOROOT /go
ENV GOPATH /gosrc

WORKDIR /home/sango
RUN ./agent test
USER sango
