FROM sango/base

RUN apt-get install -y clang-format-3.3

ADD run /home/sango/
WORKDIR /home/sango
RUN chmod 755 run

USER sango
