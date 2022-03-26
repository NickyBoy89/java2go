# Helper script for decompiling some files for testing

fernflower = fabric-fernflower-1.4.1+local.jar
quiltflower = quiltflower-1.7.0+local.jar
namedjar = 1.16.5-named.jar

decompile: quiltflower
	go run . -outDir=out -exclude-annotations="@Environment(EnvType.CLIENT)" -w quiltflower/

clean:
	# Build files
	-rm -rf quiltflower-src
	-rm -rf yarn
	-rm -rf intellij-fernflower
	# Decompiled files
	-rm -r yarn fernflower quiltflower namedSrc

# Yarn generates the deobfuscated jar files to decompile
yarn:
	git clone -b 1.16.5 git@github.com:FabricMC/yarn.git
	cd yarn && ./gradlew --no-daemon mapNamedJar

# Uses the quiltflower decompiler for decompilation
quiltflower: yarn
	git clone git@github.com:QuiltMC/quiltflower.git quiltflower-src
	cd quiltflower-src && ./gradlew --no-daemon build
	mkdir quiltflower
	java -jar "quiltflower-src/build/libs/${quiltflower}" -rsy=1 "yarn/${namedjar}" quiltflower

# Uses the CFR decompiler for decompilation
CFR: yarn
	cd yarn && ./gradlew --no-daemon decompileCFR
	mv yarn/namedSrc .

# Uses the fernflower decompiler for decompilation
fernflower: yarn
	git clone git@github.com:FabricMC/intellij-fernflower.git
	cd intellij-fernflower && ./gradlew --no-daemon build
	mkdir fernflower
	java -jar "intellij-fernflower/build/libs/${fernflower}" -rsy=1 "yarn/${namedjar}" fernflower
	unzip "fernflower/${namedjar}" -d fernflower
