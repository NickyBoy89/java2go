# Helper script for decompiling some files for testing

fernflower = fabric-fernflower-1.4.1+local.jar
quiltflower = quiltflower-1.8.1+local.jar
namedjar = 1.16.5-named.jar
procyon = `curl --silent https://api.github.com/repos/mstrobel/procyon/releases/latest | jq -r .assets[0].name`

decompile: quiltflower
	go run . -outDir=out -exclude-annotations="@Environment(EnvType.CLIENT)" -w quiltflower/

clean:
	# Build files
	-rm -rf quiltflower-src
	-rm -rf yarn
	-rm -rf intellij-fernflower
	-rm ${procyon}*
	# Decompiled files
	-rm -r yarn fernflower quiltflower namedSrc procyon

# Yarn generates the deobfuscated jar files to decompile
yarn:
	git clone -b 1.16.5 git@github.com:FabricMC/yarn.git
	#git clone -b 1.16.5 git@github.com:QuiltMC/yarn.git
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

# Uses the Procyon decompiler for decompilation
procyon: yarn
	wget `curl --silent "https://api.github.com/repos/mstrobel/procyon/releases/latest" | jq -r .assets[0].browser_download_url`
	java -jar "${procyon}" -jar "yarn/${namedjar}" -o procyon

# Uses the fernflower decompiler for decompilation
fernflower: yarn
	git clone git@github.com:FabricMC/intellij-fernflower.git
	cd intellij-fernflower && ./gradlew --no-daemon build
	mkdir fernflower
	java -jar "intellij-fernflower/build/libs/${fernflower}" -rsy=1 "yarn/${namedjar}" fernflower
	unzip "fernflower/${namedjar}" -d fernflower
