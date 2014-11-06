FROM golang:1.2

RUN useradd -m sango

ADD config.yml /home/sango/
ADD run /home/sango/
WORKDIR /home/sango
RUN chmod 755 run

USER sango
