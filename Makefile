

all: build
	
build:
	mkdir bin
	go build -o bin/publish publish/*.go
	go build -o bin/anonfundserver anonfundserver/anonfundserver.go 
	go build -o bin/gencert anonfundserver/gencert.go
	go build -o bin/fanout fanout/*.go
	./bin/gencert --host="localhost"

install-deps:
	go get -d -v github.com/NSkelsey/go-scripts/anonfundserver
	go get -d -v github.com/NSkelsey/go-scripts/fanout
	go get -d -v github.com/NSkelsey/go-scripts/publish


clean: 
	rm -rf bin
