import javalang
import gogen
import sys
import coloredlogs, logging

# Create a global logger
logger = logging.getLogger(__name__)

# Set the log level (DEBUG means everything)
coloredlogs.install(level='DEBUG')

outputDirectory = "out"

def main():
    # Collect the files that are going to be parsed
    toParse = sys.argv[1:]
    asts = {}
    for fileName in toParse:
        if not fileName.endswith(".java"):
            logger.warning(f"{fileName} was not a .java file, skipping")
            continue
        with open(fileName) as inputFile:
            asts[fileName] = javalang.parse.parse(inputFile.read())

    gogen.generate(asts, outDir=outputDirectory, write=False)

if __name__ == "__main__":
    main()
