import javalang
from typing import List
from . import gen

# ShortName returns a short representation of the classes name for the use
# in creating go methods, for example, this turns "TestClass" into "ts"
def short_name(className: str) -> str:
    return (className[0] + className[-1]).lower()

# generate takes in the asts of the files and generates a resulting file from each of them
# in the specified output directory
def generate(asts, outDir=".") -> str:
    for fileName, ast in asts.items():
        generate_file(ast)

# generate_file takes in the ast of a file and parses it
def generate_file(ast):
    if ast.imports != []:
        raise Exception("Imports not implemented")
    elif ast.package != None:
        raise Exception("Package not implemented")
    for file_type in ast.types:
        handle_type(None, file_type)

def handle_class(ctx, sourceClass: javalang.tree.ClassDeclaration) -> str:
    print(f"Found class {sourceClass.name}")
    print(ctx.gen_struct(sourceClass))
    for element in sourceClass.body:
        handle_type(ctx, element)

def handle_type(ctx, astType):
    if type(astType) == javalang.tree.ClassDeclaration:
        return handle_class(gen.GenerationContext(astType.name), astType)
    elif type(astType) == javalang.tree.MethodDeclaration:
        print(ctx.gen_method(astType))
    elif isinstance(astType, javalang.tree.ConstructorDeclaration):
        print(ctx.gen_function(astType))
    else:
        print(f"Unknown ast type: {type(astType)}")

def generateParameters(source) -> str:
    generated = ""
    for param in source:
        print(type(param))
    return generated

def generate_body(lines):
    generated = ""
    for line in lines:
        if type(line) == javalang.tree.ReturnStatement:
            generated += generate_return(line)
    return generated

def generate_return(line: javalang.tree.ReturnStatement) -> str:
    return "return" + generate_expression(line.expression)

def generate_expression(expr) -> str:
    generated = ""
    if type(expr) == javalang.tree.This:
        for selector in expr.selectors:
            generated += parse_selector(selector)
        print(expr)
    print(type(expr))
    return generated

def parse_selector(selector) -> str:
    if type(selector) == javalang.tree.MemberReference:
        return "." + selector.member
    else:
        raise Exception(f"Unknown selector type {type(selector)}")

def parse_type(varType: javalang.tree.BasicType) -> str:
    if varType.name == "int":
        return "int32"
    else:
        raise Exception(f"Unknown type {varType.name}")

def generateClass(classDeclaration):
    print(classDeclaration.annotations)
