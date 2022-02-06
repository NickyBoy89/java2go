fernflower = fabric-fernflower-1.4.1+local.jar
namedjar = 1.16.5-named.jar

all: build run

run:
	mkdir decompiled
	# Copy over the jar file
	cp yarn/$(namedjar) .
	# Copy over fernflower
	cp intellij-fernflower/build/libs/$(fernflower) .
	java -jar $(fernflower) $(namedjar) decompiled
	cd decompiled && unzip $(namedjar)

.PHONY:clean-decom,clean,buildyarn,buildfernflower

build: buildyarn buildfernflower

clean:
	-rm $(fernflower) $(namedjar)
	-rm -r decompiled
	-cd intellij-fernflower && ./gradlew --stop
	-rm -rf intellij-fernflower
	-cd yarn && ./gradlew --stop
	-rm -rf yarn

yarn:
	git clone -b 1.16.5 git@github.com:FabricMC/yarn.git

buildyarn: yarn
	cd yarn && ./gradlew mapNamedJar

intellij-fernflower:
	git clone git@github.com:FabricMC/intellij-fernflower.git

buildfernflower: intellij-fernflower
	cd intellij-fernflower && ./gradlew build
