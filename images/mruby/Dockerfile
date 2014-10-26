FROM h2so5/sango-base

RUN apt-get update
RUN apt-get install -y bison

RUN git clone https://github.com/mruby/mruby.git
WORKDIR mruby

RUN ./minirake

RUN mkdir /sango
RUN echo -n `./build/host/bin/mruby --version | tr -d '\n'` `git rev-parse HEAD | cut -b1-8 | tr -d '\n'` > /sango/version

RUN chmod 755 /mruby/build/host/bin/mruby

ADD run /home/sango/
WORKDIR /home/sango
RUN chmod 755 run

USER sango
