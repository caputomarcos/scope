all: .collection.uptodate .query.uptodate

.collection.uptodate: collection/collection collection/Dockerfile
	docker build -t weaveworks/scope-collection collection/
	touch $@

.query.uptodate: query/query query/Dockerfile
	docker build -t weaveworks/scope-query query/
	touch $@

collection/collection: collection/*.go *.go
query/query: query/*.go *.go

collection/collection query/query:
	go build -ldflags "-extldflags \"-static\"" -tags netgo -o $@ ./$(@D)

clean:
	go clean ./..
	rm -f *.uptodate collection/collection query/query
