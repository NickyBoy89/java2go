import javalang
import sys
import coloredlogs, logging
import json

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
            with open(fileName[:-len(".java")]+".json", "w") as outFile:
                json.dump(parse_json(javalang.parse.parse(inputFile.read())), outFile, indent=2)
                logger.debug(f"Parsed file {fileName} to {fileName[:-len('.java')]+'.json'}")

def is_ast_type(obj):
    if isinstance(obj, javalang.ast.Node): # AST nodes
        return True
    return False

def parse_json(raw_input):
    if is_ast_type(raw_input):
        result = {"Name": str(type(raw_input)), "Contents": vars(raw_input)}
        for attrname, attrvalue in result["Contents"].items():
            if isinstance(attrvalue, list):
                for ind, item in enumerate(attrvalue):
                    attrvalue[ind] = parse_json(item)
            elif isinstance(attrvalue, dict):
                for key, val in attrvalue.items():
                    attrvalue[key] = parse_json(val)
            elif isinstance(attrvalue, set): # Since a set is not serializable, we replace it with a list
                replacement_list = []
                for item in attrvalue:
                    replacement_list.append(parse_json(item))
                result["Contents"][attrname] = replacement_list
            else:
                result["Contents"][attrname] = parse_json(attrvalue)
        return result
    return raw_input

if __name__ == "__main__":
    main()
