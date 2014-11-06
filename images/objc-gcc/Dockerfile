FROM sango/base

RUN apt-get install -y gobjc++ libgnustep-base-dev gnustep-make gnustep gnustep-devel

ADD config.yml /home/sango/
ADD run /home/sango/
WORKDIR /home/sango
RUN chmod 755 run

USER sango
