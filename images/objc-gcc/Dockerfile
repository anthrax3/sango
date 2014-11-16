FROM sango/_base

RUN apt-get install -y gobjc++ libgnustep-base-dev gnustep-make gnustep gnustep-devel

ADD . /home/sango/

WORKDIR /home/sango
RUN if [ ! -f agent ]; then go get -d .; go build -o agent agent.go; fi
RUN chmod 755 agent

RUN ./agent test
USER sango
