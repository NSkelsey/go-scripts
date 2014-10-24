
all:
	mkdir bin
	go build -o bin/publish publish/*.go
	go build -o bin/anonfundserver anonfundserver/anonfundserver.go 
	go build -o bin/gencert anonfundserver/gencert.go
	go build -o bin/fanout fanout/*.go
	./bin/gencert --host="localhost"

clean: 
	rm -rf bin
