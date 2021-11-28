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
        return f"""type {self.name} struct {{\n{self.gen_lines(filter(lambda x: isinstance(x, javalang.tree.FieldDeclaration), source.body))}\n}}"""
    def gen_method(self, source: javalang.tree.MethodDeclaration) -> str:
        return f"""func ({self._short_name} *{self.name}) {source.name}({self.gen_params(source.parameters)}) {self.gen_type(source.return_type)} {{\n{self.gen_lines(source.body)}\n}}"""
    def gen_params(self, params: List[javalang.tree.MethodDeclaration]) -> str:
        return ""
    def gen_function(self, source) -> str:
        if isinstance(source, javalang.tree.ConstructorDeclaration):
            return f"""func {source.name}({self.gen_params(source.parameters)}) *{self.name} {{\n{self.gen_lines(source.body)}\n}}"""
        return f"""func {source.name}({self.gen_params(source.parameters)}) {self.gen_type(source.return_type)} {{\n{self.gen_lines(source.body)}\n}}"""
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
        else:
            raise Exception(f"Unknown line type: {type(line)}")
        return generated
    def gen_type(self, type) -> str:
        if type.name == "int":
            return "int32"
        else:
            raise Exception(f"Unknown type {type}")
