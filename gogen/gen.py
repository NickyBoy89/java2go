import javalang
from typing import List

class GenerationContext:
    name: str # The name of the class
    _short_name: str # The generated short name of the class (ex: "TestClass" -> "ts")
    initialized: List[javalang.tree.VariableDeclarator] = []# A list of variables that are initialized on the creation of the class
    def __init__(self, name: str):
        self.name = name
        self._short_name = (name[0] + name[-1]).lower()
    def gen_struct(self, source: javalang.tree.ClassDeclaration) -> str:
        return f"""type {self.name} struct {{\n{self.gen_lines(filter(lambda x: isinstance(x, javalang.tree.FieldDeclaration), source.body))}\n}}\n"""
    def gen_method(self, source: javalang.tree.MethodDeclaration) -> str:
        return f"""func ({self._short_name} *{self.name}) {source.name}({self.gen_params(source.parameters)}) {self.gen_type(source.return_type)} {{\n{self.gen_lines(source.body)}\n}}\n"""
    def gen_params(self, params) -> str:
        generated = ""
        for ind, param in enumerate(params):
            generated += self.gen_param(param)
            if ind < len(params) - 1:
                generated += ", "
        return generated
    def gen_param(self, param) -> str:
        if isinstance(param, javalang.tree.FormalParameter):
            return f"{param.name} {self.gen_type(param.type)}"
        else:
            raise Exception(f"Unknown parameter type: {type(param)}")
    def gen_function(self, source) -> str:
        if isinstance(source, javalang.tree.ConstructorDeclaration):
            return f"""func New{source.name}({self.gen_params(source.parameters)}) *{self.name} {{
{self._short_name} := new({self.name})
{self.gen_lines(source.body)}
return {self._short_name}
}}
"""
        return f"""func {source.name}({self.gen_params(source.parameters)}) {self.gen_type(source.return_type)} {{\n{self.gen_lines(source.body)}\n}}\n"""
    def gen_lines(self, lines: List) -> str:
        generated = ""
        for line in lines:
            generated += self.gen_line(line)
        return generated
    def gen_line(self, line) -> str:
        generated = ""
        if isinstance(line, javalang.tree.ReturnStatement):
            generated += f"return {self.gen_line(line.expression)}"
        elif isinstance(line, javalang.tree.FieldDeclaration):
            generated += f"{self.gen_lines(line.declarators)} {self.gen_type(line.type)}"
        elif isinstance(line, javalang.tree.ConstructorDeclaration):
            pass
        elif isinstance(line, javalang.tree.MethodDeclaration):
            pass
        elif isinstance(line, javalang.tree.This):
            generated += f"{self._short_name}.{self.gen_lines(line.selectors)}"
        elif isinstance(line, javalang.tree.MemberReference):
            generated += f"{line.member}"
        elif isinstance(line, javalang.tree.VariableDeclarator):
            self.initialized.append(line)
            generated += f"{line.name}"
        elif isinstance(line, javalang.tree.StatementExpression):
            generated += self.gen_line(line.expression)
        elif isinstance(line, javalang.tree.Assignment):
            generated += f"{self.gen_line(line.expressionl)} = {self.gen_line(line.value)}"
        elif isinstance(line, javalang.tree.Literal):
            if line.value == "null":
                generated += "nil"
            else:
                generated += line.value
        elif isinstance(line, javalang.tree.IfStatement):
            print(line)
            if line.else_statement != None:
                pass
            generated += f"""if {self.gen_line(line.condition)} {{\n{""}\n}}"""
        elif isinstance(line, javalang.tree.BinaryOperation):
            generated += f"{self.gen_line(line.operandl)} {line.operator} {self.gen_line(line.operandr)}"
        elif isinstance(line, javalang.tree.LocalVariableDeclaration):
            generated += f"{line.declarators} := {line.initializer}"
        else:
            raise Exception(f"Unknown line type: {line}")
        return generated
    def gen_type(self, type) -> str:
        # Void return types
        if type == None:
            return ""
        if isinstance(type, javalang.tree.ReferenceType):
            return "*" + type.name
        elif type.name == "int":
            return "int32"
        else:
            raise Exception(f"Unknown type {type}")
