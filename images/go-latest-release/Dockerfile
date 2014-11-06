FROM golang

RUN apt-get install -y time

RUN useradd -m sango
ADD run /home/sango/
WORKDIR /home/sango
RUN chmod 755 run

USER sango
