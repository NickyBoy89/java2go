all: build

.PHONY:
build:
	go build .

.PHONY:
compileJava:
	javac testfiles/*.java

.PHONY:
parseDecompiled: build
	time find decompiled/ | xargs ./java2go --dry-run
