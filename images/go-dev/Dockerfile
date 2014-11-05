FROM sango/base

RUN hg clone -u default https://code.google.com/p/go
WORKDIR /go/src
RUN ./all.bash

ENV PATH /go/bin:$PATH

ADD run /home/sango/
WORKDIR /home/sango
RUN chmod 755 run

USER sango
