FROM sango/_base

RUN apt-get update
RUN apt-get install -y bison

RUN git clone https://github.com/mruby/mruby.git
WORKDIR mruby

RUN ./minirake

RUN chmod 755 /mruby/build/host/bin/mruby
ENV PATH /mruby/build/host/bin:$PATH

ADD . /home/sango/

WORKDIR /home/sango
RUN go get -d .
RUN go build -o run run.go
RUN chmod 755 run

USER sango
