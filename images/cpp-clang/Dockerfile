FROM sango/base

RUN apt-get install -y clang-3.5

ADD run /home/sango/
WORKDIR /home/sango
RUN chmod 755 run

USER sango
