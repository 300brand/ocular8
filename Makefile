GO=/usr/bin/go

o8-server: handlers/enqueue-feeds/enqueue-feeds handlers/recount-articles/recount-articles
	$(GO) build -o $@ ./server

handlers/enqueue-feeds/enqueue-feeds:
	$(GO) build -o $@ ./handlers/enqueue-feeds

handlers/recount-articles/recount-articles:
	$(GO) build -o $@ ./handlers/recount-articles

.PHONY: run
run: clean o8-server
	./o8-server -logtostderr -assets=assets -handlers=handlers

.PHONY: clean
clean:
	rm -f o8-server handlers/enqueue-feeds/enqueue-feeds handlers/recount-articles/recount-articles
