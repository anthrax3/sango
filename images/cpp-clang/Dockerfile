FROM sango/_base

RUN apt-get install -y clang-3.5

ADD . /home/sango/

WORKDIR /home/sango
RUN go get -d .
RUN go build -o run run.go
RUN chmod 755 run

RUN ./run -t
USER sango
