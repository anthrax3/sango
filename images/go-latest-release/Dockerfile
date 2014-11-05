FROM golang

RUN useradd -m sango
ADD run /home/sango/
WORKDIR /home/sango
RUN chmod 755 run

USER sango
