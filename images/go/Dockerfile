FROM golang

RUN mkdir /sango
RUN go version | sed -r "s/go version//g" | sed -r "s/(^ +| +$|)//g" | tr -d '\n' > /sango/version

RUN useradd -m sango
ADD run /home/sango/
WORKDIR /home/sango
RUN chmod 755 run

USER sango
