FROM sango/_base

ADD . /home/sango/
WORKDIR /home/sango
RUN go get -d .
RUN go build -o run run.go
RUN chmod 755 run

WORKDIR /
RUN hg clone -u default https://code.google.com/p/go
WORKDIR /go/src
RUN rm -v **/**/pprof_test.go
RUN ./all.bash
ENV PATH /go/bin:$PATH

WORKDIR /home/sango
USER sango
