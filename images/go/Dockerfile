FROM golang

RUN go get code.google.com/p/go.tools/cmd/goimports
ENV PATH $GOPATH/bin:$PATH

RUN useradd -m sango
ADD run /home/sango/
WORKDIR /home/sango
RUN chmod 755 run

USER sango
