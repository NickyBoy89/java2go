package goparser

import (
	"fmt"
	"unicode"
	"strings"

	"gitlab.nicholasnovak.io/snapdragon/java2go/parsing"
	"gitlab.nicholasnovak.io/snapdragon/java2go/parsetools"
	"gitlab.nicholasnovak.io/snapdragon/java2go/codeparser"
)

const indentNum = 2

// Stores a list of generated errors for each package
var currentErrors = make(map[string][]string)

// Stores a list of strings to generate to the current file
var toGenerate []string

var lastDoWhile codeparser.LineTyper

// NewClass is set to true if the class is a nested class
func ParseFile(sourceFile parsing.ParsedClasses, newClass bool, filename string, ignoredAnnotations []string) string {
	var generated string
	if newClass {
		generated += fmt.Sprintf("package %s\n\n", filename)
	}
	switch sourceFile.GetType() {
	case "class":
		generated += ParseClass(sourceFile.(parsing.ParsedClass), filename, ignoredAnnotations) // Parse the class into one struct
	case "interface":
		generated += ParseInterface(sourceFile.(parsing.ParsedInterface), filename, ignoredAnnotations)
	case "enum":
		generated += ParseEnum(sourceFile.(parsing.ParsedEnum), filename, ignoredAnnotations)
	default:
		panic("Unknown class type: " + sourceFile.GetType())
	}

	// Replace generated methods
	for found, newFunc := range specializedFunctions {
		generated = strings.ReplaceAll(generated, found, newFunc)
	}

	return generated
}

// Parse a given class
func ParseClass(source parsing.ParsedClass, filename string, ignoredAnnotations []string) string {
	var generated string

	// Create a context for the class, so that the methods have some frame of reference
	classContext := new(ClassContext)
	// Set the name of the class context
	classContext.Name = ToPublic(source.Name)
	// Extract the method names from the class itself, before it has been parsed
	classContext.Methods = source.MethodContext()
	classContext.Package = filename
	classContext.IgnoredAnnotations = ignoredAnnotations

	// If the line below is commented out, then every struct will be declared as public

	// if !IsPublic(source.Modifiers) {
	// 	classContext.Name = ToPrivate(source.Name)
	// }


	// Add the implements as a comment
	generated += fmt.Sprintf("//%v\n", source.Implements)

	// Add the extends as a comment
	generated += fmt.Sprintf("//%v\n", source.Extends)

	// If the class has any static blocks, then generated them
	for _, staticBlock := range source.StaticBlocks {
		generated += CreateStaticBlock(staticBlock, classContext, 0)
		generated += "\n\n"
	}

	// Parse the class itself as a struct
	// If the class is static, don't generate a struct for it
	if !parsetools.Contains("static", source.Modifiers) {
		generated += CreateStruct(classContext, source.ClassVariables)
	}

	generated += "\n\n" // Add some spacing after the initial struct

	// Parse the methods of the class
	for _, method := range source.Methods {
		generated += CreateMethod(classContext, method)
		generated += "\n\n" // Add some spacing in between the methods
	}

	// If there are any things to generate, do them now
	for _, genStr := range toGenerate {
		generated += "\n" + genStr + "\n"
	}
	toGenerate = []string{}

	// Parse the nested classes
	for _, nested := range source.NestedClasses {
		generated += ParseFile(nested, false, filename, ignoredAnnotations)
	}

	return generated
}

func ParseEnum(source parsing.ParsedEnum, filename string, ignoredAnnotations []string) string {
	var generated string

	classContext := new(ClassContext)
	classContext.Name = ToPublic(source.Name)
	classContext.Methods = source.MethodContext()
	classContext.Package = filename
	classContext.IgnoredAnnotations = ignoredAnnotations

	// Parse the enum fields
	var parsedEnums []string
	for _, field := range source.EnumFields {
		parsedEnums = append(parsedEnums, CreateEnumField(field, classContext))
	}

	// Add the implements as a comment
	generated += fmt.Sprintf("//%v\n", source.Implements)

  // If the class has any static blocks, then generated them
  for _, staticBlock := range source.StaticBlocks {
    generated += CreateStaticBlock(staticBlock, classContext, 0)
    generated += "\n\n"
  }

	if !parsetools.Contains("static", source.Modifiers) {
		generated += CreateStruct(classContext, source.ClassVariables)
	}
	generated += "\n\n"

	// Every enum in java has an inplicit method "values", which returns an array of all the enum's fields
	valuesMethod := parsing.ParsedMethod{
		// Name is classname + Values (ex: CompassValues)
		Name: classContext.Name + "Values",
		Modifiers: []string{"public", "static"},
		Parameters: []parsing.ParsedVariable{},
		// Returns an array of that type of object
		ReturnType: "[]" + ToReferenceType(ReplaceWord(classContext.Name)),
		Body: []codeparser.LineTyper{
			codeparser.LineType{
				Name: "ReturnStatement",
				Words: map[string]interface{}{
					"Expression": []codeparser.LineType{
						codeparser.LineType{
							Name: "ImplicitArrayAssignment",
							Words: map[string]interface{}{
								"ArrayType": "[]" + classContext.Name,
								"Elements": EnumFieldsToLineType(source.EnumFields, classContext),
							},
						},
					},
				},
			},
		},
	}
	generated += CreateMethod(classContext, valuesMethod) + "\n\n"
	classContext.Methods[classContext.Name + "Values"] = []string{}

	// If there are any things to generate, do them now
	for _, genStr := range toGenerate {
		generated += "\n" + genStr + "\n"
	}
	toGenerate = []string{}

	// Populate the generated enums into a var block
	generated += CreateVarBlock(parsedEnums)

	generated += "\n"

	for _, method := range source.Methods {
		generated += CreateMethod(classContext, method)
		generated += "\n\n" // Add some spacing in between the methods
	}

	// Parse the nested classes
	for _, nested := range source.NestedClasses {
		generated += ParseFile(nested, false, filename, ignoredAnnotations)
	}

	return generated
}

func ParseInterface(source parsing.ParsedInterface, filename string, ignoredAnnotations []string) string {
	var generated string

	// Create a context for the class, so that the methods have some frame of reference
	classContext := new(ClassContext)
	// Set the name of the class context
	classContext.Name = ToPublic(source.Name)
	// Extract the method names from the class itself, before it has been parsed
	classContext.Methods = source.MethodContext()
	classContext.Package = filename
	classContext.IgnoredAnnotations = ignoredAnnotations

	// If the line below is commented out, then every struct will be declared as public

	// if !IsPublic(source.Modifiers) {
	// 	classContext.Name = ToPrivate(source.Name)
	// }

	generated += CreateInterface(classContext, source.Methods)

	generated += "\n\n" // Add some spacing after the initial struct

	// Parse the nested classes
	for _, nested := range source.NestedClasses {
		generated += ParseFile(nested, false, filename, ignoredAnnotations)
	}

	return generated
}

func CreateStruct(classContext *ClassContext, fields []parsing.ParsedVariable) string {
	result := fmt.Sprintf("type %s struct {", classContext.Name)
	for _, field := range fields { // Struct fields
		// If the struct field's annotation is on the global IgnoredAnnotations, then skip creating it
		for _, ignoredAnno := range classContext.IgnoredAnnotations {
		  if strings.Contains(field.Annotation, ignoredAnno) {
		    continue
		  }
		}
		// Generate the field's annotation as a comment
		if field.Annotation != "" {
			result += fmt.Sprintf("\n%s//%s", strings.Repeat(" ", indentNum), field.Annotation)
		}
		// Write out a field (ex: value int)
		if IsPublic(field.Modifiers) {
			result += fmt.Sprintf("\n%s%s %s", strings.Repeat(" ", indentNum), ToPublic(field.Name), FormatVariable(field.DataType))
		} else {
			result += fmt.Sprintf("\n%s%s %s", strings.Repeat(" ", indentNum), ToPrivate(field.Name), FormatVariable(field.DataType))
		}
	}
	return result + "\n}"
}

func CreateInterface(classContext *ClassContext, methods []parsing.ParsedMethod) string {
	result := fmt.Sprintf("type %s interface {", classContext.Name)
		for _, method := range methods {
			var methodParams string // Yes, allocating this outside of the loop would make this slightly faster
			for pi, param := range method.Parameters {
				methodParams += FormatVariable(param.DataType)
				if pi != len(method.Parameters) - 1 {
					methodParams += ", "
				}
			}
			if IsPublic(method.Modifiers) {
				result += fmt.Sprintf("\n%s%s(%s) %s", strings.Repeat(" ", indentNum), ToPublic(method.Name), methodParams, FormatVariable(method.ReturnType))
			} else {
				result += fmt.Sprintf("\n%s%s(%s) %s", strings.Repeat(" ", indentNum), ToPrivate(method.Name), methodParams, FormatVariable(method.ReturnType))
			}
		}
	return result + "\n}"
}

func ToPublic(name string) string {
	return string(unicode.ToUpper(rune(name[0]))) + name[1:]
}

func ToPrivate(name string) string {
	return string(unicode.ToLower(rune(name[0]))) + name[1:]
}

// Tests if an object is public, given its modifiers
func IsPublic(modifiers []string) bool {
	if parsetools.Contains("public", modifiers) || parsetools.Contains("protected", modifiers) {
		return true
	}
	return false
}

// Generates a two-character shorthand of a struct's name for a method (ex: IntLinkedList -> it)
func AsShorthand(name string) string {
	return string(unicode.ToLower(rune(name[0]))) + string(unicode.ToLower(rune(name[len(name) - 1])))
}

// Creates a string representation of a method from the parsed method and the name of the class that it came from
// For a static method (standalone class) pass in an empty class name
func CreateMethod(classContext *ClassContext, methodSource parsing.ParsedMethod) string {
	var result string

	// If the method's annotation is on the global IgnoredAnnotations, then skip creating it
	for _, ignoredAnno := range classContext.IgnoredAnnotations {
    if strings.Contains(methodSource.Annotation, ignoredAnno) {
      return ""
    }
  }

	// Generate the method's annotation as a comment
	if methodSource.Annotation != "" {
		result += "//" + methodSource.Annotation + "\n"
	}

	if parsetools.Contains("static", methodSource.Modifiers) { // Method is static, so not associated with any class
		// Special methods, ex: main, init
		switch methodSource.Name {
		case "main":
			result += fmt.Sprintf("func %s(", methodSource.Name)
		default:
			if IsPublic(methodSource.Modifiers) {
				result += fmt.Sprintf("func %s(", ToPublic(methodSource.Name))
			} else {
				result += fmt.Sprintf("func %s(", ToPrivate(methodSource.Name))
			}
		}
	} else if methodSource.ReturnType == "constructor" { // Constructor methods just get handled as generator functions
		// if IsPublic(methodSource.Modifiers) {
			result += fmt.Sprintf("func New%s(", classContext.Name) // If public, the constructor function is public as well
		// } else {
			// result += fmt.Sprintf("func new%s(", classContext.Name) // Private constructor
		// }
	} else {
		if IsPublic(methodSource.Modifiers) {
			result += fmt.Sprintf(
				"func (%s *%s) %s(",
				AsShorthand(classContext.Name),
				classContext.Name,
				ToPublic(methodSource.Name),
			)
		} else {
			result += fmt.Sprintf(
				"func (%s *%s) %s(",
				AsShorthand(classContext.Name),
				classContext.Name,
				ToPrivate(methodSource.Name),
			)
		}
	}

	if methodSource.Name != "main" {
		for pi, param := range methodSource.Parameters { // Parameters
			result += param.Name + " " + FormatVariable(param.DataType)
			if pi < len(methodSource.Parameters) - 1 {
				result += ", "
			}
		}
	}

	if methodSource.ReturnType == "constructor" {
		result += fmt.Sprintf(
			") *%s {\n%s%s := new(%s)\n%s\n%sreturn %s\n}",
			classContext.Name,
			strings.Repeat(" ", indentNum),
			AsShorthand(classContext.Name),
			classContext.Name,
			CreateBody(methodSource.Body, classContext, 2),
			strings.Repeat(" ", indentNum),
			AsShorthand(classContext.Name),
		)
		return result
	}
	result += fmt.Sprintf(") %v {\n%s\n}", ReplaceWord(methodSource.ReturnType), CreateBody(methodSource.Body, classContext, 2))
	return result
}

func CreateStaticBlock(lines []codeparser.LineTyper, classContext *ClassContext, indentation int) string {
	block := fmt.Sprintf("func init() {\n")
	block += CreateBody(lines, classContext, indentation + 2)
	block += fmt.Sprintf("\n%s}", strings.Repeat(" ", indentation))
	return block
}

// Parses the lines of the body
func CreateBody(body []codeparser.LineTyper, classContext *ClassContext, indentation int) string {
	var result string
	for _, line := range body {
		// fmt.Printf("Going through line of type: %s\n", line.GetName())
		result += CreateLine(line, classContext, indentation, true)
	}
	return result
}

func CreateLine(line codeparser.LineTyper, classContext *ClassContext, indentation int, indent bool) string {
	var result string
	if indent {
		result += "\n"
	}

	// result += fmt.Sprintf("//%s\n", line.GetName())
	switch line.GetName() {
	case "StringLiteral":
		result += strings.Repeat(" ", indentation) + fmt.Sprintf("%s", line.(codeparser.LineType).Words["String"])
	case "CreateAndAssignVariable":
		var body string
		for _, line := range line.(codeparser.LineType).Words["Expression"].([]codeparser.LineType) {
			body += CreateLine(line, classContext	, 0, false)
		}
		// Commented-out block is for variables to be declared explicitly
		// result += strings.Repeat(" ", indentation) + fmt.Sprintf(
		// 	"var %s %s = %s",
		// 	line.(codeparser.LineType).Words["VariableName"],
		// 	JavaToGoArray(ReplaceWord(line.(codeparser.LineType).Words["VariableType"].(string))),
		// 	body,
		// )
		result += strings.Repeat(" ", indentation) + fmt.Sprintf(
			"%s := %s",
			CreateLine(line.(codeparser.LineType).Words["VariableName"].(codeparser.LineType), classContext, 0, false),
			body,
		)
	case "AssignVariable":
		var body string
		for _, line := range line.(codeparser.LineType).Words["Expression"].([]codeparser.LineType) {
			body += ReplaceWord(CreateLine(line, classContext, 0, false))
		}
		result += strings.Repeat(" ", indentation) + fmt.Sprintf(
			"%s = %s",
			CreateLine(line.(codeparser.LineType).Words["VariableName"].([]codeparser.LineType)[0], classContext, 0, false),
			body,
		)
	case "CompoundAssignment":
		var body string
		for _, line := range line.(codeparser.LineType).Words["Expression"].([]codeparser.LineType) {
			body += ReplaceWord(CreateLine(line, classContext, 0, false))
		}
		result += strings.Repeat(" ", indentation) + fmt.Sprintf(
			"%s %s= %s",
			line.(codeparser.LineType).Words["VariableName"],
			line.(codeparser.LineType).Words["Operator"],
			body,
		)
	case "FunctionCall":
		functionName := line.(codeparser.LineType).Words["FunctionName"].(string)
		// Special case in enums, the "values" method returns all of the enum's fields
		if functionName == "values" {
			functionName = classContext.Name + "Values"
		}

		// Populate the parameters of the function
		var body string
		for li, expressionLine := range line.(codeparser.LineType).Words["Parameters"].([][]codeparser.LineType) {
			for _, expLine := range expressionLine {
				body += CreateLine(expLine, classContext, 0, false)
			}
			if li != len(line.(codeparser.LineType).Words["Parameters"].([][]codeparser.LineType)) - 1 { // For the commas, don't add one to the last element
				body += ", "
			}
		}

		if classContext.ContainsMethod(ToPrivate(functionName)) {
			functionName = ToPrivate(functionName)
		} else if classContext.ContainsMethod(ToPublic(functionName)) {
			functionName = ToPublic(functionName)
		// Pretty much every call to an exception in Java ends with the word "Exception"
		} else {
			// panic("Unknown non-package function " + functionName + "")
		}

		// Handle Exceptions and Errors
		if parsetools.EndsWith(functionName, "Exception") {
			// If there is no other exception of the same name in the current package, generate it
			if !parsetools.Contains(functionName, currentErrors[classContext.Package]) {
				currentErrors[classContext.Package] = append(currentErrors[classContext.Package], functionName)
				toGenerate = append(toGenerate, CreateError(functionName, body))
			}
		// Assumes errors are capitalized and end with the word "Error"
		} else if parsetools.EndsWith(functionName, "Error") && unicode.IsUpper(rune(functionName[0])) {
			// If there is no other error of the same name in the current package, generate it
			if !parsetools.Contains(functionName, currentErrors[classContext.Package]) {
				currentErrors[classContext.Package] = append(currentErrors[classContext.Package], functionName)
				toGenerate = append(toGenerate, CreateError(functionName, body))
			}
		}

		// Check if the function is an Error or Exception
		if parsetools.Contains(functionName, currentErrors[classContext.Package]) {
			result += functionName
		} else {
			result += strings.Repeat(" ", indentation) + fmt.Sprintf(
				"%s(%s)",
				functionName,
				body,
			)
		}
	case "ConstructArray":
		result += strings.Repeat(" ", indentation) + fmt.Sprintf("make([]%s, %s)", line.(codeparser.LineType).Words["ArrayType"], line.(codeparser.LineType).Words["InitialSize"])
	case "AccessArrayElement":
		result += strings.Repeat(" ", indentation) + fmt.Sprintf("%s[%s]", line.(codeparser.LineType).Words["ArrayName"], line.(codeparser.LineType).Words["Index"])
	case "ReturnStatement":
		var body string
		for _, line := range line.(codeparser.LineType).Words["Expression"].([]codeparser.LineType) {
			body += CreateLine(line, classContext, 0, false)
		}
		result += strings.Repeat(" ", indentation) + fmt.Sprintf("return %s", body)
	case "IfStatement":
		var body string
		for _, line := range line.(codeparser.LineBlock).Words["Condition"].([]codeparser.LineType) {
			body += CreateLine(line, classContext, 0, false) + " "
		}
		result += strings.Repeat(" ", indentation) + fmt.Sprintf(
			"if %s {%s\n%s}",
			body,
			CreateBody(line.(codeparser.LineBlock).Lines, classContext, indentation + 2),
			strings.Repeat(" ", indentation),
		)
	case "ElseIfStatement":
		var body string
		for _, line := range line.(codeparser.LineBlock).Words["Condition"].([]codeparser.LineType) {
			body += CreateLine(line, classContext, 0, false) + " "
		}
		result += strings.Repeat(" ", indentation) + fmt.Sprintf(
			"else if %s {%s\n%s}",
			body,
			CreateBody(line.(codeparser.LineBlock).Lines, classContext, indentation + 2),
			strings.Repeat(" ", indentation),
		)
	case "ElseBlock":
		// This is an equals only, to cut out the newline
		result = fmt.Sprintf(" else {%s\n%s}\n", CreateBody(line.(codeparser.LineBlock).Lines, classContext, indentation + 2), strings.Repeat(" ", indentation))
	case "TryBlock":
		result += fmt.Sprintf("//Try block\n %s", CreateBody(line.(codeparser.LineBlock).Lines, classContext, indentation + 2), strings.Repeat(" ", indentation))
	case "CatchBlock":
		result += fmt.Sprintf("//Catch block\n %s", CreateBody(line.(codeparser.LineBlock).Lines, classContext, indentation + 2), strings.Repeat(" ", indentation))
	case "FinallyBlock":
		result += fmt.Sprintf("//Finally block\n %s", CreateBody(line.(codeparser.LineBlock).Lines, classContext, indentation + 2), strings.Repeat(" ", indentation))
	case "SynchronizedBlock":
		result += fmt.Sprintf("//Synchronized block\n %s", CreateBody(line.(codeparser.LineBlock).Lines, classContext, indentation + 2), strings.Repeat(" ", indentation))
	case "ForLoop":
		var body string
		for _, line := range line.(codeparser.LineBlock).Words["Conditional"].([]codeparser.LineType) {
			body += CreateLine(line, classContext, 0, false) + " "
		}
		result += strings.Repeat(" ", indentation) + fmt.Sprintf(
			"for %s; %s; %s {%s\n%s}",
			CreateLine(line.(codeparser.LineBlock).Words["Initializer"].(codeparser.LineTyper), classContext, 0, false),
			body,
			CreateLine(line.(codeparser.LineBlock).Words["Incrementer"].(codeparser.LineTyper), classContext, 0, false),
			CreateBody(line.(codeparser.LineBlock).Lines, classContext, indentation + 2),
			strings.Repeat(" ", indentation),
		)
	case "EnhancedForLoop":
		var body string
		for _, line := range line.(codeparser.LineBlock).Words["Iterable"].([]codeparser.LineType) {
			body += CreateLine(line, classContext, 0, false) + " "
		}
		result += strings.Repeat(" ", indentation) + fmt.Sprintf(
			"for _, %s := range %s {%s\n%s}",
			line.(codeparser.LineBlock).Words["DeclarationName"],
			body,
			CreateBody(line.(codeparser.LineBlock).Lines, classContext, indentation + 2),
			strings.Repeat(" ", indentation),
		)
	case "NewConstructor":
		result += strings.Repeat(" ", indentation) + fmt.Sprintf("New%s", CreateLine(line.(codeparser.LineType).Words["Expression"].(codeparser.LineType), classContext, 0, false))
	case "ThrowException":
		result += strings.Repeat(" ", indentation) + fmt.Sprintf("panic(%s)\n", CreateLine(line.(codeparser.LineType).Words["Expression"].([]codeparser.LineType)[0], classContext, 0, false))
	case "ImplicitArrayAssignment":
		var body string
		for li, expressionLine := range line.(codeparser.LineType).Words["Elements"].([][]codeparser.LineType) {
			for _, expLine := range expressionLine {
				body += CreateLine(expLine, classContext, 0, false)
			}
			if li != len(line.(codeparser.LineType).Words["Elements"].([][]codeparser.LineType)) - 1 { // For the commas, don't add one to the last element
				body += ", "
			}
		}
		result += strings.Repeat(" ", indentation) + fmt.Sprintf("%s{%s}", ToReferenceType(JavaToGoArray(ReplaceWord(line.(codeparser.LineType).Words["ArrayType"].(string)))), body)
	// The expression types, don't have a newline associated with them
	case "LocalVariableOrExpression":
		result += strings.Repeat(" ", indentation) + fmt.Sprintf("%s", ReplaceWord(line.(codeparser.LineType).Words["Expression"].(string)))
	case "RemoteVariableOrExpression":
		packageName := line.(codeparser.LineType).Words["RemotePackage"]

		// The expression field will only ever contain one entry (ex: value in this.value)
		// Let's just assert that
		// if len(line.(codeparser.LineType).Words["Expression"].([]codeparser.LineType)) != 1 {
		// 	panic("Remote expression does not have one expression, this should not be the case")
		// }

		expression := CreateLine(line.(codeparser.LineType).Words["Expression"].([]codeparser.LineType)[0], classContext, 0, false)

		switch packageName {
		case "this": // If package name is reserved word "this", then treat it as referring to the struct method's shorthand
			packageName = AsShorthand(classContext.Name)
		case classContext.Name: // If the package name is the current package
			result += strings.Repeat(" ", indentation) + fmt.Sprintf("%s", expression)
			expression = ""
		}

		switch expression {
		case "":
		case "length": // Getting the "length" field of a variable will instead call the len() go builtin function
			result += strings.Repeat(" ", indentation) + fmt.Sprintf(
				"len(%s)",
				packageName,
			)
		default:
			result += strings.Repeat(" ", indentation) + fmt.Sprintf(
				"%s.%s",
				packageName,
				expression,
			)
		}
	case "GenericExpression":
		var body string
		for _, line := range line.(codeparser.LineType).Words["Expression"].([]codeparser.LineType) {
			body += CreateLine(line, classContext, 0, false)
		}
		result += strings.Repeat(" ", indentation) + fmt.Sprintf("%s", body)
	case "TernaryOperator":
		var trueExp, falseExp string
		for _, exp := range line.(codeparser.LineType).Words["TrueExpression"].([]codeparser.LineType) {
			trueExp += CreateLine(exp, classContext, 0, false)
		}
		for _, exp2 := range line.(codeparser.LineType).Words["FalseExpression"].([]codeparser.LineType) {
			falseExp += CreateLine(exp2, classContext, 0, false)
		}
		result += strings.Repeat(" ", indentation) + fmt.Sprintf("if %s {\n%s\n}\n else {\n%s\n}", "", trueExp, falseExp)
	case "ComparisonOperator":
		result += "=="
	case "NotEquals":
		result += "!="
	case "LambdaExpression":
		var lambdaParams string
		for li, param := range line.(codeparser.LineType).Words["Parameters"].([][]codeparser.LineType) {
			for _, expLine := range param {
				lambdaParams += CreateLine(expLine, classContext, 0, false)
			}
			if li != len(line.(codeparser.LineType).Words["Parameters"].([][]codeparser.LineType)) - 1 { // For the commas, don't add one to the last element
				lambdaParams += ", "
			}
		}
		result += strings.Repeat(" ", indentation) + fmt.Sprintf("func (%s) {\n%s\n}", lambdaParams, CreateBody(line.(codeparser.LineType).Words["Lines"].([]codeparser.LineTyper), classContext, indentation + 2))
	case "ImplicitObjectCreation":
		fmt.Printf("Implicit object: %s\n", line.(codeparser.LineBlock).Words["MethodLine"])
	case "ParenthesiedExpression":
		var body string
		for _, line := range line.(codeparser.LineType).Words["Expression"].([]codeparser.LineType) {
			body += CreateLine(line, classContext, 0, false)
		}
		result += fmt.Sprintf("(%s)", body)
	case "WhileStatement":
		var body string
		for _, line := range line.(codeparser.LineBlock).Words["Condition"].([]codeparser.LineType) {
			body += CreateLine(line, classContext, 0, false) + " "
		}
		result += strings.Repeat(" ", indentation) + fmt.Sprintf(
			"for %s {%s\n%s}",
			body,
			CreateBody(line.(codeparser.LineBlock).Lines, classContext, indentation + 2),
			strings.Repeat(" ", indentation),
		)
	case "DoWhileCondition": // The while condition for a do-while, comes after the do block
		var body string
		for _, line := range line.(codeparser.LineBlock).Words["Condition"].([]codeparser.LineType) {
			body += CreateLine(line, classContext, 0, false) + " "
		}
		result += strings.Repeat(" ", indentation) + fmt.Sprintf(
			"for {%s\n%s\n%sif %s {\n%sbreak\n%s}}",
			CreateBody(lastDoWhile.(codeparser.LineBlock).Lines, classContext, indentation + 2),
			strings.Repeat(" ", indentation),
			strings.Repeat(" ", indentation),
			body,
			strings.Repeat(" ", indentation + 2),
			strings.Repeat(" ", indentation),
		)
		lastDoWhile = nil
	case "DoWhileStatement":
		lastDoWhile = line
	case "TypeAssertion":
		// Skip type assertion
		// result += fmt.Sprintf("(%s)", line.(codeparser.LineType).Words["AssertedType"])
	case "SwitchExpression":
		result += strings.Repeat(" ", indentation) + fmt.Sprintf(
			"switch %s {%s\n%s}",
			line.(codeparser.LineBlock).Words["SwitchExpression"],
			CreateBody(line.(codeparser.LineBlock).Lines, classContext, indentation + 2),
			strings.Repeat(" ", indentation),
		)
	case "SwitchCase":
		result += strings.Repeat(" ", indentation) + fmt.Sprintf(
			"case %s:\n%s",
			line.(codeparser.LineBlock).Words["Case"],
			CreateBody(line.(codeparser.LineBlock).Lines, classContext, indentation + 2),
		)
	case "DefaultCase":
		result += strings.Repeat(" ", indentation) + fmt.Sprintf(
			"default :%s",
			CreateBody(line.(codeparser.LineBlock).Lines, classContext, indentation + 2),
		)
	case "RuneLiteral":
		result += fmt.Sprintf("'%s'", line.(codeparser.LineType).Words["Rune"])
	case "ContentLabel":
		result += strings.Repeat(" ", indentation) + fmt.Sprintf("%s:", line.(codeparser.LineType).Words["LabelName"])
	case "MethodReference":
		result += strings.Repeat(" ", indentation) + fmt.Sprintf(
			"%s.%s",
			line.(codeparser.LineType).Words["MethodClass"],
			line.(codeparser.LineType).Words["MethodName"],
		)
	case "BlankStatement":
		result += strings.Repeat(" ", indentation) + fmt.Sprintf("%s", CreateBody(line.(codeparser.LineBlock).Lines, classContext, indentation + 2))
	case "LabeledBlock":
		result += strings.Repeat(" ", indentation) + fmt.Sprintf("%s\n%s", line.(codeparser.LineBlock).Words["Label"], fmt.Sprintf("%s", CreateBody(line.(codeparser.LineBlock).Lines, classContext, indentation + 2)))
	default:
		panic("Unknown line type: [" + line.GetName() + "]")
	}

	return result
}

func CreateVarBlock(vars []string) string {
	generated := "var (\n"

	for _, inputVar := range vars {
		generated += fmt.Sprintf("  %s\n", inputVar)
	}

	generated += "\n)\n"

	return generated
}

func CreateEnumField(field parsing.EnumField, ctx *ClassContext) string {
	var fieldParams string
	for pi, fieldParam := range field.Parameters {
		// Assumes that the only thing in the parameters that we care about are their names
		fieldParams += fmt.Sprintf("%s", fieldParam.Name)
		if pi != len(field.Parameters) - 1 {
			fieldParams += ", "
		}
	}
	// Parses the fields as classname_fieldname, ex: Compass_NORTH = NewCompass()
	return fmt.Sprintf("%s_%s = New%s(%s)", ctx.Name, field.Name, ctx.Name, fieldParams)
}

func EnumFieldsToLineType(enumFields []parsing.EnumField, ctx *ClassContext) [][]codeparser.LineType {
	outputLines := [][]codeparser.LineType{}
	for _, field := range enumFields {
		outputLines = append(outputLines, []codeparser.LineType{
			codeparser.LineType{
				Name: "LocalVariableOrExpression",
				Words: map[string]interface{}{
					"Expression": ctx.Name + "_" + field.Name,
				},
			},
		})
	}

	return outputLines
}

func CreateError(name, content string) string {
	return fmt.Sprintf("var %s = errors.New(\"%s: \" + %s)", name, name, content)
}
