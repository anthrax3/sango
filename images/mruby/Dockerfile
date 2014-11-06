FROM sango/base

RUN apt-get update
RUN apt-get install -y bison

RUN git clone https://github.com/mruby/mruby.git
WORKDIR mruby

RUN ./minirake

RUN chmod 755 /mruby/build/host/bin/mruby
ENV PATH /mruby/build/host/bin:$PATH

ADD config.yml /home/sango/
ADD run /home/sango/
WORKDIR /home/sango
RUN chmod 755 run

USER sango
