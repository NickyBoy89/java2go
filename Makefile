all: build

.PHONY:
build:
	go build .

.PHONY:
compileJava:
	javac testfiles/*.java
