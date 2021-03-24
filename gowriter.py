
# All methods in this file handle parsing of the parsed Java data into the
# Golang format

# Controls the size of the indent for the content in the functions
INDENT_SIZE = 2

def writeClassToFile(parsedClass, parsedClassMethodsAndVariables):
    gofile = ""
    generatedMethods = ""

    generatedStruct = createStruct(parsedClass["name"], parsedClassMethodsAndVariables[1]) + "\n\n"

    for method in parsedClassMethodsAndVariables[0]:
        if "isConstructor" in method:
            generatedMethods += createMethod("New" + method["name"], method["args"], "*" + method["name"], method["modifiers"], parsedClass["name"], parseMethodContent(method["content"], parsedClass["name"], isConstructor=True), isConstructor=True) + "\n"
        else:
            generatedMethods += createMethod(method["name"], method["args"], toGoTypes(method["returnType"]), method["modifiers"], parsedClass["name"], parseMethodContent(method["content"])) + "\n"

    return (generatedStruct + generatedMethods).replace(";\n", "\n")

def toGoTypes(type):
    # Is some sort of array
    if "[]" in type:
        return "[]" + toGoTypes(type[:-2])
    # Lowering all the strings to handle the object versions of java types
    # Ex: Boolean for boolean
    if type.lower() == "double":
        return "float64"
    elif type.lower() == "float":
        return "float32"
    elif type.lower() == "long":
        return "int64"
    elif type.lower() == "string":
        return "string"
    elif type.lower() == "boolean":
        return "bool"
    elif type.lower() == "char":
        return "rune"
    elif type.lower() == "boolean":
        return "bool"
    elif type.lower() == "integer":
        return "int"
    elif type.lower() == "short":
        return "int16"
    else:
        return type

def toGoAccessMod(name, accessmodifiers):
    if "public" in accessmodifiers or "protected" in accessmodifiers:
        return name[0].capitalize() + name[1:]
    else:
        return name[0].lower() + name[1:]

# The combine arg just makes it so that the traditional java
# int x, int y, int z becomes x, y, z int in Golang, instead of
# x int, y int, z int in Golang
def toGoArgs(args, combine=True):
    result = ""

    for arg in enumerate(args):
        # Look ahead at the next arg and see if it has the same arg
        if arg[0] < len(args) - 1 and args[arg[0] + 1]["type"] == arg[1]["type"] and combine:
            result += f'{arg[1]["name"]}'
        else:
            result += f'{arg[1]["name"]} {arg[1]["type"]}'

        # Lart arg, don't print a comma
        if arg[0] != len(args) - 1:
            result += ", "

    return result

# Note: This assumes that classNames passed in already have the correct access modifiers from toGoAccessMod
def createMethod(name, args, returnType, modifiers, className, parsedContent, isPointer=False, isConstructor=False):
    createdClass = ""

    # Static just means a class without any relation to a struct
    if "static" in modifiers or isConstructor:
        createdClass += f"func "
    else:
        createdClass += f"func ({className[0].lower()} {className}) "
        if isPointer:
            createdClass += f"func ({className[0].lower()} *{className}) "

    # Handle capitalization for Golang access modifiers
    if isConstructor: # Constructor methods are always public
        createdClass += name
    else:
        createdClass += toGoAccessMod(name, modifiers)

    if len(args) != 0:
        createdClass += f"({toGoArgs(args)}) {returnType} {{\n{parsedContent}\n}}"
    else:
        createdClass += f"() {returnType} {{\n{parsedContent}\n}}"

    # print(createdClass)

    return createdClass + "\n"

# This assumes that the name of the struct already has the correct Go permissions
def createStruct(name, fields):
    struct = f"type {name} struct {{\n"
    for field in fields:
        struct += f'{INDENT_SIZE * " "}{toGoAccessMod(field["name"], field["modifiers"])} {toGoTypes(field["type"])}\n'

    return struct + "}"

# className is only needed if there is a constructor
def parseMethodContent(methodContent, className=None, isConstructor=False):

    if isConstructor:
        constrObjName = className[0].lower() + className[1:]
        content = " " * INDENT_SIZE + constrObjName + f" := new({className})\n\n"
        content += "\n".join(map(lambda x: x[2:], methodContent.replace("this", constrObjName).split("\n")))
        content += f"\n\n{' ' * INDENT_SIZE}return {constrObjName}"
        return content

    return "\n".join(map(lambda x: x[2:], methodContent.split("\n")))
