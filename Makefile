
.PHONY: install test

install:
	go install github.com/leafo/gifserver

test:
	go test -v github.com/leafo/gifserver/gifserver
