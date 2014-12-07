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

ADD . /tmp/sango
WORKDIR /tmp/sango

RUN if [ -f sango.tar.gz ]; then \
mkdir -p $GOPATH/src/github.com/h2so5/sango; \
tar -C $GOPATH/src/github.com/h2so5/sango -xzf sango.tar.gz; \
rm sango.tar.gz; \
fi

RUN go get -d .
RUN go install github.com/h2so5/sango/tools/jtime
RUN go build -o agent agent.go
RUN chmod 755 agent
RUN mv agent /usr/bin/agent
RUN agent test

WORKDIR /home/sango
USER sango
