import javalang
from typing import List, Dict

def gen_imports(imports: List[javalang.tree.Import], predefined_mappings: Dict[str,str]) -> str:
    result = "import (\n"
    for imp in imports:
        if imp.path not in predefined_mappings:
            raise Exception(f"Unknown equivalent import for {imp.path}")
        if predefined_mappings[imp.path] != "":
            result += f"{predefined_mappings[imp.path]}\n"
    return result + ")"

def gen_package(package_name) -> str:
    if package_name != None:
        raise Exception(type(package_name))
    return "package main\n"

def gen_class(class_source: javalang.tree.ClassDeclaration) -> str:
    generated = ""
    #class_source.implements
    #class_source.extends
    #modifiers
    #name
    struct_fields = []
    for item in class_source.body:
        match type(item):
            case javalang.tree.FieldDeclaration:
                struct_fields.append(item)
            case javalang.tree.ConstructorDeclaration:
                generated += gen_function("New" + item.name, item.parameters, "*" + item.name, item.body, constructor=True) + "\n"
            case javalang.tree.MethodDeclaration:
                generated += gen_function(item.name, item.parameters, item.return_type, item.body, isMethod=True) + "\n"
            case _:
                raise Exception(type(item))
    return generated

def gen_function(function_name: str, parameters: List[any], return_type: str, body: List[any], constructor=False, isMethod=False) -> str:
    if isMethod:
        generated = f"func (ts *TestClass) {function_name}({parameters}) {return_type} {{\n"
    else:
        generated = f"func {function_name}({parameters}) {return_type} {{\n"
    if constructor:
        generated += "ts := new(TestClass)\n"
    for line in body:
        generated += gen_line(line) + "\n"
    if constructor:
        generated += "return ts\n"
    return generated + "}"

def gen_line(line) -> str:
    generated = ""
    match type(line):
        case javalang.tree.StatementExpression:
            generated += gen_expression(line.expression)
        case javalang.tree.ReturnStatement:
            generated += "return " + gen_expression(line.expression)
        case _:
            raise Exception(type(line))
    return generated

def gen_expression(expression) -> str:
    generated = ""
    match type(expression):
        case javalang.tree.Assignment:
            return f"{gen_expression(expression.expressionl)} {expression.type} {gen_expression(expression.value)}"
        case javalang.tree.This:
            generated += "this."
            for selector in expression.selectors:
                generated += gen_expression(selector)
        case javalang.tree.MemberReference:
            return expression.member
        case _:
            raise Exception(type(expression))
    return generated


def gen_datatype(datatype) -> str:
    match type(datatype):
        case javalang.tree.BasicType:
            match datatype.name:
                case "int":
                    return "int32"
                case _:
                    raise Exception(f"Unknown basic type: {datatype.name}")
        case _:
            raise Exception(f"Unknown data type: {type(datatype)}")

def gen_struct(struct_name: str, struct_fields: Dict[str,str]) -> str:
    generated = f"type {struct_name} struct {{\n"
    for field_name, field_type in struct_fields.items():
        generated += f"{field_name} {field_type}"
    return generated + "}"
