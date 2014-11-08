FROM sango/_base

RUN apt-get install -y clang-3.5

ADD . /home/sango/

WORKDIR /home/sango
RUN chmod 755 run

USER sango
