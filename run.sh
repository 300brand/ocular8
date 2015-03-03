#!/bin/bash
set -o errexit -o nounset

CMD=${1:-}
ROOT=$(dirname $0)

function build {
	go get -v github.com/300brand/ocular8/server
	cp -a $GOPATH/bin/server $ROOT/o8-server
	HANDLERS=$(go list -f '{{.ImportPath}}^{{.Target}}' github.com/300brand/ocular8/handlers/...)
	for HANDLER in $HANDLERS; do
		IMPORTPATH=$(echo $HANDLER | cut -d^ -f1)
		TARGET=$(echo $HANDLER | cut -d^ -f2)
		go get -v $IMPORTPATH
		cp -a $TARGET $GOPATH/src/$IMPORTPATH/
	done
}

function clean {
	rm -vf $ROOT/o8-server
	DIRS=$(go list -f '{{.Dir}}' github.com/300brand/ocular8/handlers/...)
	for DIR in $DIRS; do
		BIN=$(basename $DIR)
		rm -vf $DIR/$BIN
	done
}

function pydeps {
	for R in ${ROOT}/handlers/*/requirements2.txt; do
		python2 -m pip install --user --requirement $R
	done

	for R in ${ROOT}/handlers/*/requirements3.txt; do
		python3 -m pip install --user --requirement $R
	done

	( \
		cd /tmp && \
		git clone https://github.com/grangier/python-goose.git && \
		cd python-goose && \
		python2 setup.py install --user && \
		cd /tmp && \
		rm -rf /tmp/python-goose \
	)

	( \
		cd /tmp && \
		git clone https://github.com/jplana/python-etcd.git && \
		cd python-etcd && \
		python2 setup.py install --user && \
		python3 setup.py install --user && \
		cd /tmp && \
		rm -rf /tmp/python-etcd \
	)
}

function start {
	${ROOT}/o8-server \
		-logtostderr \
		-assets=${ROOT}/assets \
		-handlers=${ROOT}/handlers
}


case "$CMD" in
	build)
		build
		;;
	clean)
		clean
		;;
	full)
		build
		start
		;;
	pydeps)
		pydeps
		;;
	start)
		start
		;;
	*)
		echo $"Usage: $0 {start|build|clean|full}"
		exit 1
esac
