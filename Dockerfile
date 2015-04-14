FROM       golang:latest
MAINTAINER Jake <jtews@300brand.com>
RUN        apt-get update
RUN        apt-get install -y \
               bzr \
               git \
               python \
               python-dateutil \
               python-dev \
               python-elasticsearch \
               python-feedparser \
               python-lxml \
               python-mysql.connector \
               python-pip \
               python3 \
               python3-dateutil \
               python3-dev \
               python3-elasticsearch \
               python3-feedparser \
               python3-lxml \
               python3-mysql.connector \
               python3-pip
RUN        pip install python-etcd
RUN        cd /tmp && \
               git clone https://github.com/grangier/python-goose.git && \
               cd python-goose && \
               python2 setup.py install && \
               cd /tmp && \
               rm -rf /tmp/python-goose
RUN        mkdir -p /go/src/github.com/300brand
COPY       . /go/src/github.com/300brand/ocular8
RUN        /go/src/github.com/300brand/ocular8/run.sh build
WORKDIR    /go/src/github.com/300brand/ocular8
EXPOSE     80
CMD        [ "sh", "-c", "./o8-server", "-logtostderr", "-etcd", "http://${ETCD_PORT_4001_TCP_ADDR:-localhost}:${ETCD_PORT_4001_TCP_PORT:-4001}"]
