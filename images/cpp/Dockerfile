FROM h2so5/sango-base

RUN mkdir /sango
RUN gcc -v 2>&1 | tail -n 1 | sed -r "s/\(.+\)//g" | sed -r "s/(^ +| +$|)//g" | tr -d '\n' > /sango/version

#RUN git clone --depth 1 https://gitorious.org/qt/qtbase.git
#WORKDIR /qtbase
#RUN ./configure -opensource -confirm-license
#RUN make
#RUN make install

ADD run /home/sango/
WORKDIR /home/sango
RUN chmod 755 run

USER sango
