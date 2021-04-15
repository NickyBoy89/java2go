all: build

.PHONY:
build:
	go build .

.PHONY:
compileJava:
	javac testfiles/*.java

.PHONY:
testParseDirectory: build
	time find $1 | xargs ./java2go --dry-run
