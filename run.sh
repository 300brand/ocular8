#!/bin/bash
set -o errexit -o nounset

CMD=${1:-}
ROOT=$(dirname $0)
GOHANDLERS="enqueue-feeds download-entry recount-articles"

function build {
	go get -v github.com/300brand/ocular8/server
	cp -a $GOPATH/bin/server $ROOT/o8-server
	for H in $GOHANDLERS; do
		go get -v github.com/300brand/ocular8/handlers/$H
		cp -a $GOPATH/bin/$H $ROOT/handlers/$H/
	done
}

function clean {
	rm $ROOT/o8-server
	for H in $GOHANDLERS; do
		rm $ROOT/handlers/$H/$H
	done
}

function pydeps {
	for R in ${ROOT}/handlers/*/requirements2.txt; do
		python2 -m pip install --user --requirement $R
	done

	for R in ${ROOT}/handlers/*/requirements3.txt; do
		python3 -m pip install --user --requirement $R
	done

	(cd /tmp && git clone https://github.com/grangier/python-goose.git && cd python-goose && python2 setup.py install --user)
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
