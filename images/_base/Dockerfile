FROM phusion/baseimage

RUN echo "nameserver 8.8.8.8" >> /etc/resolv.conf
RUN apt-get update -m
RUN apt-get install -y git mercurial make build-essential

RUN curl -O https://storage.googleapis.com/golang/go1.3.3.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.3.3.linux-amd64.tar.gz
RUN rm go1.3.3.linux-amd64.tar.gz

ENV GOPATH /go
ENV GOROOT /usr/local/go
ENV PATH /go/bin:/usr/local/go/bin:$PATH

RUN go get github.com/tv42/base58
RUN go get github.com/vmihailenco/msgpack
RUN go get github.com/vmihailenco/bufio
RUN go get gopkg.in/yaml.v2

RUN useradd -m sango
