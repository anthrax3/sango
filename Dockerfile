FROM jpetazzo/dind
 
RUN apt-get update
RUN apt-get install -y mercurial make
 
RUN wget https://storage.googleapis.com/golang/go1.3.3.linux-amd64.tar.gz
 
RUN tar -C /usr/local -xzf go1.3.3.linux-amd64.tar.gz
 
ENV GOPATH /go
ENV PATH /usr/local/go/bin:$GOPATH/bin:$PATH
 
ADD . /sango
WORKDIR /sango
 
RUN go get -d .
RUN make

ENV LOG file
RUN chmod +x ./start.sh
 
ENTRYPOINT ["/sango/start.sh"]
 
EXPOSE 3000