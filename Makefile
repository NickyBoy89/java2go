all: build

.PHONY:
build:
	-rm test.go
	go build .

.PHONY:
compileJava:
	javac testfiles/*.java
