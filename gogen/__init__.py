import javalang
from typing import List, Dict
from . import gen

def from_ast(ast: javalang.tree.CompilationUnit, import_mappings: Dict[str, str]) -> str:
    # The AST starts with several things that we need to deal with:
    # - Imports - Any external packages that are imported
    # - Package - The current package that the file belongs to
    # - Types - The class declared within the file
    result = ""
    # Generate the package name before the imports, go-style
    result += gen.gen_package(ast.package)
    if ast.imports != []:
        result += gen.gen_imports(ast.imports, import_mappings)
    # Parse the actual AST
    for ast_type in ast.types:
        result += gen.gen_class(ast_type)
    return result

def short_name(className: str) -> str:
    """
    short_name returns the golang-specific short name for a class
    for use in methods and constructors
    ex: TestClass -> ts
    """
    return (className[0] + className[-1]).lower()
