FROM sango/_base

RUN apt-get install -y clang-3.5

ADD . /home/sango/

WORKDIR /home/sango
RUN if [ ! -f agent ]; then go get -d .; go build -o agent agent.go; fi
RUN chmod 755 agent

RUN ./agent test
USER sango
