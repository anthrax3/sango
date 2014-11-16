FROM sango/_base

RUN mkdir /qt
WORKDIR /qt
RUN curl -L -O http://download.qt-project.org/official_releases/qt/5.3/5.3.2/single/qt-everywhere-opensource-src-5.3.2.tar.xz
RUN tar Jxf qt-everywhere-opensource-src-5.3.2.tar.xz
RUN rm qt-everywhere-opensource-src-5.3.2.tar.xz
WORKDIR /qt/qt-everywhere-opensource-src-5.3.2
RUN ./configure -opensource -confirm-license -no-xcb
RUN make -j4
RUN make install
ENV PATH /usr/local/Qt-5.3.2/bin:$PATH

ADD . /home/sango/

WORKDIR /home/sango
RUN if [ ! -f agent ]; then go get -d .; go build -o agent agent.go; fi
RUN chmod 755 agent

RUN agent test
RUN rm -rf test Makefile *.o *.pro sango

USER sango
