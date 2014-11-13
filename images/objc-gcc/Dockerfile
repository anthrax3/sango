FROM sango/_base

RUN apt-get install -y gobjc++ libgnustep-base-dev gnustep-make gnustep gnustep-devel

ADD . /home/sango/

WORKDIR /home/sango
RUN go get -d .
RUN go build -o run run.go
RUN chmod 755 run

USER sango
