FROM sango/_base

RUN hg clone -u default https://code.google.com/p/go
WORKDIR /go/src
RUN rm -v **/**/pprof_test.go
RUN ./all.bash

ENV PATH /go/bin:$PATH

ADD . /home/sango/

WORKDIR /home/sango
RUN chmod 755 run

USER sango
