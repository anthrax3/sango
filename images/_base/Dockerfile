FROM phusion/baseimage

RUN apt-get update -m
RUN apt-get install -y git mercurial make ruby python php5 openjdk-7-jdk nodejs npm build-essential time valgrind

RUN curl -O https://storage.googleapis.com/golang/go1.3.3.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.3.3.linux-amd64.tar.gz
RUN rm go1.3.3.linux-amd64.tar.gz

ENV GOPATH /go
ENV GOROOT /usr/local/go
ENV PATH /usr/local/go/bin:$PATH

RUN useradd -m sango
