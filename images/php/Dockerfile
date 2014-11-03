FROM sango/base

RUN mkdir /sango
RUN echo -n `php -v | head -n 1 | sed -r "s/ with .+$//g" | sed -r "s/\(.+\)//g" | tr -d '\n'` > /sango/version

ADD run /home/sango/
WORKDIR /home/sango
RUN chmod 755 run

USER sango
