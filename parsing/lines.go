package parsing

import (
  "strings"
  "regexp"
  // "log"
)

// Parses the declaration of a method in the Java code
// Ex: public int GetValue {
// You can detect a declaration by the existence of a '{' character in the line,
// so methods in interfaces do not count
func ParseMethodLine(dec string) map[string]interface{} {

  methodInputsInd := regexp.MustCompile(`\(.*\)`).FindStringIndex(dec)

  declarationWords := strings.Split(dec[:methodInputsInd[0]] + dec[methodInputsInd[1]:], " ")

  // Constructors don't have any return type, so they're shorter
  if len(declarationWords) < 3 {
    return map[string]interface{}{
      "name": declarationWords[len(declarationWords) - 2],
      "declarationType": "methodConstructor",
      "returnType": "constructor",
      "accessModifiers": declarationWords[:len(declarationWords) - 2],
      "contentLines": []string{},
    }
  }

  return map[string]interface{}{
    "name": declarationWords[len(declarationWords) - 2],
    "declarationType": "method",
    "returnType": declarationWords[len(declarationWords) - 3],
    "accessModifiers": declarationWords[:len(declarationWords) - 3],
    "contentLines": []string{},
  }
}

// Parses the declaration of a class in the Java code
// Ex: protected abstract class TestClass {
// You can detect a declaration by the existence of a '{' character in the line,
// and the word "class"
func ParseClassLine(dec string) map[string]interface{} {
  declarationWords := strings.Split(dec, " ")

  return map[string]interface{}{
    "name": declarationWords[len(declarationWords) - 2],
    "declarationType": "class",
    "accessModifiers": declarationWords[:len(declarationWords) - 3],
    "methods": []map[string]interface{}{},
    "methodVariables": []map[string]interface{}{},
  }
}

// A method signature is a method that doesn't have a body section enclosed
// in brackets, but can be found if there are no brackets but there are
// parenthesies
// Ex: (In an interface) public int getX()
func ParseMethodSignatureLine(line string) map[string]interface{} {
  words := strings.Split(line, " ")

  return map[string]interface{}{
    // Strip the last two chars off (the '();' at the end of the method)
    "name": words[len(words) - 1][:len(words[len(words) - 1]) - 3],
    "declarationType": "methodSignature",
    "returnType": words[len(words) - 2],
    "accessModifiers": words[:len(words) - 2],
    "contentLines": []string{},
  }
}

// This one is identified by not having any '(' or '{', and is essentially
// just the odd one out (not a declaration or method signature)
// Ex: public int value;
func ParseMemberVariableLine(line string) map[string]interface{} {
  words := strings.Split(line, " ")

  return map[string]interface{}{
    // Strip the last char off (the semicolon)
    "name": words[len(words) - 1][:len(words[len(words) - 1]) - 1],
    "declarationType": "memberVariable",
    "returnType": words[len(words) - 2],
    "accessModifiers": words[:len(words) - 2],
  }
}
