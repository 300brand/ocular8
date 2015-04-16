FROM       golang:latest
MAINTAINER Jake <jtews@300brand.com>
RUN        apt-get update && \
               apt-get install -y \
                   bzr \
                   git \
                   libssl-dev \
                   python \
                   python-bson \
                   python-cryptography \
                   python-dateutil \
                   python-dev \
                   python-elasticsearch \
                   python-feedparser \
                   python-lxml \
                   python-mysql.connector \
                   python-pip \
                   python3 \
                   python3-bson \
                   python3-cryptography \
                   python3-dateutil \
                   python3-dev \
                   python3-elasticsearch \
                   python3-feedparser \
                   python3-lxml \
                   python3-mysql.connector \
                   python3-pip
RUN        python2 -m pip install python-etcd
RUN        python3 -m pip install python-etcd
RUN        cd /tmp && \
               git clone https://github.com/grangier/python-goose.git && \
               cd python-goose && \
               python2 setup.py install && \
               cd /tmp && \
               rm -rf /tmp/python-goose
RUN        go get code.google.com/p/snappy-go/snappy
RUN        go get github.com/araddon/gou
RUN        go get github.com/bitly/go-hostpool
RUN        go get github.com/bitly/go-nsq
RUN        go get github.com/bitly/go-simplejson
RUN        go get github.com/coreos/go-etcd/etcd
RUN        go get github.com/go-sql-driver/mysql
RUN        go get github.com/golang/glog
RUN        go get github.com/gorilla/context
RUN        go get github.com/gorilla/handlers
RUN        go get github.com/gorilla/mux
RUN        go get github.com/mattbaird/elastigo
RUN        go get github.com/mreiferson/go-snappystream
RUN        go get gopkg.in/mgo.v2
RUN        mkdir -p /go/src/github.com/300brand
COPY       . /go/src/github.com/300brand/ocular8
RUN        /go/src/github.com/300brand/ocular8/run.sh build
WORKDIR    /go/src/github.com/300brand/ocular8
EXPOSE     80
CMD        [ "sh", "-c", "./o8-server", "-logtostderr", "-etcd", "$ETCD"]
