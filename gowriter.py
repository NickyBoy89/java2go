def writeClassToFile(parsedClass, parsedClassMethods):
    gofile = ""
    generatedMethods = ""
    for method in parsedClassMethods:
        if "isConstructor" in method:
            createMethod("New" + method["name"], method["args"], "*" + method["name"], method["modifiers"], parsedClass["name"])
        else:
            createMethod(method["name"], method["args"], toGoTypes(method["returnType"]), method["modifiers"], parsedClass["name"])
# def writeClassesToFile()

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
    if "public" or "protected" in modifiers:
        return name[0].capitalize() + name[1:]
    else:
        return name[0].lower() + name[1:]

# The combine arg just makes it so that the traditional java
# int x, int y, int z becomes x, y, z int in Golang, instead of
# x int, y int, z int in Golang
def toGoArgs(args, combine=True):
    result = ""
    currentArgType = args[0]["type"]

    for arg in enumerate(args):
        if currentArgType == arg[1]["type"] and combine:
            result += f'{arg[1]["name"]}'
        else:
            result += f'{arg[1]["name"]} {arg[1]["type"]}'

        # Lart arg, don't print a comma
        if arg[0] != len(args) - 1:
            result += ", "
        else:
            result += f' {arg[1]["type"]}'
        currentArgType = arg[1]["type"]

    return result

# Note: This assumes that classNames passed in already have the correct access modifiers from toGoAccessMod
def createMethod(name, args, returnType, modifiers, className, isPointer=False):
    createdClass = ""

    # Static just means a class without any relation to a struct
    if "static" in modifiers:
        createdClass += f"func "
    else:
        createdClass += f"func ({className[0].lower()} {className}) "
        if isPointer:
            createdClass += f"func ({className[0].lower()} *{className}) "

    # Handle capitalization for Golang access modifiers
    createdClass += toGoAccessMod(name, modifiers)

    if len(args) != 0:
        createdClass += f"({toGoArgs(args)}) {returnType} {{\n}}"
    else:
        createdClass += f"() {returnType} {{\n}}"

    print(createdClass)
