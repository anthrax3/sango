FROM golang:1.2

RUN useradd -m sango

ADD . /home/sango/

WORKDIR /home/sango
RUN chmod 755 run

USER sango
