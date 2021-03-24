import re

# All the methods in this file handle regex parsing of the raw Java files

def parseMemberVariable(member):
    if "=" in member:
        return {
            "value": member.split("=")[1].strip(),
            "name": member.split("=")[0].split(" ")[-1][:-1].strip(),
            "type": member.split("=")[0].split(" ")[-2].strip(),
            "modifiers": member.split("=")[0].split(" ")[:-2]
        }
    else:
        return {
            "name": member.split(" ")[-1][:-1].strip(),
            "type": member.split(" ")[-2].strip(),
            "modifiers": member.split(" ")[:-2]
        }

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

def parseMethodsAndVariables(input, classname):
    inputs = input.split('\n')

    inMethod = False

    methods = []
    classFields = []

    for i in enumerate(inputs):
        methodString = re.search(".*?[^\s]\(.*\)\s\{", i[1])

        if inMethod:
            if i[0] == findIndexOfClosingBrace(inputs, i[0] + 1) - 1:
                inMethod = False
        else:
            classField = re.search(".*?;", i[1])

            if classField != None:
                classFields.append(parseMemberVariable(classField.group(0)))

        if methodString != None: # No match found
            inMethod = True
            methodName = re.search(".*\(", methodString.group(0)).group(0)[:-1].split(" ")[-1]
            # If the method's name is the same as the class name, then it is not
            # a method, but a constructor
            if methodName == classname:
                methods.append({
                    'name': methodName,
                    'isConstructor': True,
                    'args': parseMethodArgs(re.search("\(.*\)", methodString.group(0)).group(0)[1:-1]),
                    'modifiers': re.search(".*\(", methodString.group(0)).group(0).split(" ")[:-2],
                    'content': "\n".join(inputs[i[0] + 1:findIndexOfClosingBrace(inputs, i[0] + 1)])
                })
                continue

            methods.append({
                'name': methodName,
                'returnType': re.search(".*\(", methodString.group(0)).group(0).split(" ")[-2],
                'args': parseMethodArgs(re.search("\(.*\)", methodString.group(0)).group(0)[1:-1]),
                'modifiers': re.search(".*\(", methodString.group(0)).group(0).split(" ")[:-2],
                'content': "\n".join(inputs[i[0] + 1:findIndexOfClosingBrace(inputs, i[0] + 1)])
            })

    return (methods, classFields)

def findIndexOfClosingBrace(inputs, startingIndex):
    braceIndex = -1
    index = startingIndex
    for inp in inputs[startingIndex:]:
        if braceIndex == 0:
            break
        elif "{" in inp and "}" in inp:
            pass
        elif "{" in inp:
            braceIndex -= 1
        elif "}" in inp:
            braceIndex += 1
        index += 1
    return index - 1

def parseMethodArgs(argString):
    methodArgs = re.findall("[^\s]+\s[^\s]+", argString)

    parsedArgs = []
    for arg in methodArgs:
        parsedArgs.append({
            'name': arg.split(" ")[1].strip(", "),
            'type': arg.split(" ")[0]
        })

    return parsedArgs
