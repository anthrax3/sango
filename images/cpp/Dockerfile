FROM sango/base

#RUN git clone --depth 1 https://gitorious.org/qt/qtbase.git
#WORKDIR /qtbase
#RUN ./configure -opensource -confirm-license
#RUN make
#RUN make install

ADD run /home/sango/
WORKDIR /home/sango
RUN chmod 755 run

USER sango
