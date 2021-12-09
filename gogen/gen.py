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
        return f"""type {self.name} struct {{\n{self.gen_lines(list(filter(lambda x: isinstance(x, javalang.tree.FieldDeclaration), source.body)))}\n}}\n"""
    def gen_method(self, source: javalang.tree.MethodDeclaration) -> str:
        if source.name == "main":
            source.parameters = []
            return self.gen_function(source)
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
        assert iter(lines)
        generated = ""
        for ind, line in enumerate(lines):
            generated += self.gen_line(line)
            if ind < len(lines) - 1:
                generated += "\n"
        return generated
    def gen_line(self, line) -> str:
        generated = ""
        try:
            if line.prefix_operators != None and line.prefix_operators != []:
                generated += "".join(line.prefix_operators)
        except AttributeError:
            pass
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
            if line.else_statement != None:
                generated += f"""if {self.gen_line(line.condition)} {{
{self.gen_line(line.then_statement)}
}}
else {{
{self.gen_line(line.else_statement)}
}}
"""
            else:
                generated += f"if {self.gen_line(line.condition)} {{\n{self.gen_line(line.then_statement)}\n}}"
        elif isinstance(line, javalang.tree.BinaryOperation):
            generated += f"{self.gen_line(line.operandl)} {line.operator} {self.gen_line(line.operandr)}"
        elif isinstance(line, javalang.tree.LocalVariableDeclaration):
            # Try to get an initializer, which should be the case if the
            # variable is being created for the first time
            try:
                generated += f"{line.declarators} := {line.initializer}"
            except AttributeError:
                # No initializer, variable is referencing an
                # already created variable
                generated += f"{self.gen_lines(line.declarators)} = "
        elif isinstance(line, javalang.tree.VariableDeclaration):
            generated += f"{self.gen_lines(line.declarators)}"
        elif isinstance(line, javalang.tree.BlockStatement):
            generated += f"{self.gen_lines(line.statements)}"
        elif isinstance(line, javalang.tree.ClassCreator):
            generated += f"New{self.gen_type(line.type)}({self.gen_args(line.arguments)})"
        elif isinstance(line, javalang.tree.ForStatement):
            generated += f"for {self.gen_line(line.control.init)}; {self.gen_line(line.control.condition)}; {self.gen_lines(line.control.update)} {{\n{self.gen_line(line.body)}\n}}"
        elif isinstance(line, javalang.tree.MethodInvocation):
            if line.qualifier != "":
                generated += f"{line.qualifier}.{line.member}({self.gen_args(line.arguments)})"
            else:
                generated += f"{line.member}({self.gen_args(line.arguments)})"
        elif isinstance(line, javalang.tree.ThrowStatement):
            generated += f"panic({self.gen_line(line.expression)})"
        else:
            raise Exception(f"Unknown line type: {line}")
        try:
            if line.postfix_operators != None and line.postfix_operators != []:
                generated += "".join(line.postfix_operators)
        except AttributeError:
            pass
        return generated
    def gen_args(self, arguments):
        generated = ""
        for ind, arg in enumerate(arguments):
            generated += self.gen_line(arg)
            if ind < len(arguments) - 1:
                generated += ", "
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
