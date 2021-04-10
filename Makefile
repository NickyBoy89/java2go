all: build run

.PHONY:
build:
	go build .

.PHONY:
run:
	./java2go

.PHONY:
compileJava:
	javac testfiles/*.java
