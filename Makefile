fernflower = fabric-fernflower-1.4.1+local.jar
quiltflower = quiltflower-1.7.0+local.jar
namedjar = 1.16.5-named.jar

all:

decompile:
	go run . -outDir=out -exclude-annotations="@Environment(EnvType.CLIENT)" -w quiltflower/

gen-sources: build run

run:
	mkdir decompiled
	# Copy over the jar file
	cp yarn/$(namedjar) .
	# Copy over fernflower
	cp intellij-fernflower/build/libs/$(fernflower) .
	# Decompile everything
	java -jar $(fernflower) -rsy=1 $(namedjar) decompiled
	cd decompiled && unzip $(namedjar)
	cp -r yarn/namedSrc .

.PHONY:clean-decom,clean,buildyarn,buildfernflower

build: buildyarn buildfernflower

# NOTE: To remove the decompiled files, manually run `rm -r decompiled`
clean:
	-rm $(fernflower) $(namedjar)
	-cd intellij-fernflower && ./gradlew --stop
	-rm -rf intellij-fernflower
	-cd yarn && ./gradlew --stop
	-rm -rf yarn
	-rm -r out

quiltflower:
	git clone git@github.com:QuiltMC/quiltflower.git quiltflower-src
	cd quiltflower-src && ./gradlew build
	mkdir quiltflower
	java -jar quiltflower-src/build/libs/$(quiltflower) -rsy=1 $(namedjar) quiltflower

yarn:
	git clone -b 1.16.5 git@github.com:FabricMC/yarn.git

buildyarn: yarn
	cd yarn && ./gradlew mapNamedJar && ./gradlew decompileCFR

intellij-fernflower:
	git clone git@github.com:FabricMC/intellij-fernflower.git

buildfernflower: intellij-fernflower
	cd intellij-fernflower && ./gradlew build
