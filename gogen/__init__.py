import javalang
from typing import List
from . import gen

def short_name(className: str) -> str:
    """
    short_name returns the golang-specific short name for a class
    for use in methods and constructors
    ex: TestClass -> ts
    """
    return (className[0] + className[-1]).lower()

def generate(asts, outDir=".", write=True):
    """
    generate takes in a list of asts and parses all of them as separate files,
    putting the generated file in the output directory
    """
    for fileName, ast in asts.items():
        print(f"Parsing {fileName}:")
        result = generate_file(ast)
        if write:
            with open(f"{outDir}/{fileName}.go", "w") as out:
                out.write(result)
        else:
            print(result)

def generate_file(ast) -> str:
    """
    generate_file takes in the ast of a file and parses it
    into a resulting code file
    """
    if ast.imports != []:
        print(f"File imported {ast.imports}")
    elif ast.package != None:
        raise Exception("Package not implemented")
    generated = ""
    for file_type in ast.types:
        generated += handle_type(None, file_type)
    return generated

def handle_class(ctx, sourceClass: javalang.tree.ClassDeclaration) -> str:
    """
    handle_class take the class declaration and parses its contents
    """
    print(f"Found class {sourceClass.name}")
    generated = ""
    generated += ctx.gen_struct(sourceClass)
    for element in sourceClass.body:
        generated += handle_type(ctx, element)
    return generated

def handle_type(ctx, astType) -> str:
    """
    handle_type parses parts of the ast tree into code
    """
    if type(astType) == javalang.tree.ClassDeclaration:
        return handle_class(gen.GenerationContext(astType.name), astType)
    elif type(astType) == javalang.tree.MethodDeclaration:
        return ctx.gen_method(astType)
    elif isinstance(astType, javalang.tree.ConstructorDeclaration):
        return ctx.gen_function(astType)
    elif isinstance(astType, javalang.tree.FieldDeclaration):
        # This case is already covered by the parser
        return ""
    else:
        raise Exception(f"Unknown ast type: {type(astType)}")
