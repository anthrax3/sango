FROM sango/base

RUN apt-get install -y gobjc++ libgnustep-base-dev gnustep-make gnustep gnustep-devel

ADD . /home/sango/

WORKDIR /home/sango
RUN chmod 755 run

USER sango
