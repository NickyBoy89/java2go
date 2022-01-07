import javalang
import gogen
import sys
import coloredlogs, logging
import json

# Create a global logger
logger = logging.getLogger(__name__)

# Set the log level (DEBUG means everything)
coloredlogs.install(level='DEBUG')

outputDirectory = "out"

with open("import_mappings.json") as mappings:
    import_mappings = json.load(mappings)

def main():
    # Collect the files that are going to be parsed
    toParse = sys.argv[1:]
    asts = {}
    for fileName in toParse:
        if not fileName.endswith(".java"):
            logger.warning(f"{fileName} was not a .java file, skipping")
            continue
        with open(fileName) as inputFile:
            # Parse the file into an AST and map it to the file's name
            asts[fileName] = javalang.parse.parse(inputFile.read())
            print(gogen.from_ast(asts[fileName], import_mappings))

if __name__ == "__main__":
    main()
