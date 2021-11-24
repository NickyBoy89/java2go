import javalang

def generate(asts, outDir="."):
    for fileName, ast in asts.items():
        for typ in ast.types:
            if isinstance(typ, javalang.tree.ClassDeclaration):
                print(typ)

def generateClass(classDeclaration):
    print(classDeclaration.annotations)
