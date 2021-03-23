import re

def parseClasses(input):
    inputs = input.split('\n')

    classes = []

    for i in inputs:
        classArgs = re.search(".*?(class)", i)
        if classArgs == None: # No match found
            continue
        classes.append({
            'name': re.search("(?<=class)(.*)(?={)", i).group(0).strip(),
            'modifiers': classArgs.group(0)[:-len("class")].strip()
        })

    return classes

def parseMethods(input, classname):
    inputs = input.split('\n')

    methods = []

    for i in inputs:
        methodString = re.search(".*?[^\s]\(.*\)\s\{", i)
        if methodString == None: # No match found
            continue

        methodName = re.search(".*\(", methodString.group(0)).group(0)[:-1].split(" ")[-1]
        # If the method's name is the same as the class name, then it is not
        # a method, but a constructor
        if methodName == classname:
            methods.append({
                'name': methodName,
                'isConstructor': True,
                'args': parseMethodArgs(re.search("\(.*\)", methodString.group(0)).group(0)[1:-1]),
                'modifiers': re.search(".*\(", methodString.group(0)).group(0).split(" ")[:-2]
            })
            continue

        methods.append({
            'name': methodName,
            'returnType': re.search(".*\(", methodString.group(0)).group(0).split(" ")[-2],
            'args': parseMethodArgs(re.search("\(.*\)", methodString.group(0)).group(0)[1:-1]),
            'modifiers': re.search(".*\(", methodString.group(0)).group(0).split(" ")[:-2]
        })

    return methods

def parseMethodArgs(argString):
    methodArgs = re.findall("[^\s]+\s[^\s]+", argString)

    parsedArgs = []
    for arg in methodArgs:
        parsedArgs.append({
            'name': arg.split(" ")[1].strip(", "),
            'type': arg.split(" ")[0]
        })

    return parsedArgs
