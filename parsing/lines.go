package parsing

// Parses the declaration of a class or method in the Java code
// Ex: protected abstract class TestClass {
// or: public int GetValue {
// You can detect a declaration by the existence of a '{' character in the line,
// so methods in interfaces do not count
func ParseDeclarationLine(dec string) interface{} {
  declarationWords := strings.Split(dec, " ")

  // Class declarations have the keyword "class"
  if sliceContains(declarationWords, "class") {
    return JavaClass{
      Name: declarationWords[len(declarationWords) - 2],
      DeclarationType: "class",
      AccessModifiers: declarationWords[:len(declarationWords) - 3],
    }
  }
  return JavaMethodItem{
    Name: declarationWords[len(declarationWords) - 2],
    DeclarationType: "method",
    ReturnType: declarationWords[len(declarationWords) - 3],
    AccessModifiers: declarationWords[:len(declarationWords) - 3],
  }
}

// A method signature is a method that doesn't have a body section enclosed
// in brackets, but can be found if there are no brackets but there are
// parenthesies
// Ex: (In an interface) public int getX()
func ParseMethodSignatureLine(line string) JavaMethodItem {
  words := strings.Split(line, " ")

  return JavaMethodItem{
    // Strip the last two chars off (the '();' at the end of the method)
    Name: words[len(words) - 1][:len(words[len(words) - 1]) - 3],
    DeclarationType: "methodSignature",
    ReturnType: words[len(words) - 2],
    AccessModifiers: words[:len(words) - 2],
  }
}

// This one is identified by not having any '(' or '{', and is essentially
// just the odd one out (not a declaration or method signature)
// Ex: public int value;
func ParseMemberVariableLine(line string) JavaMethodItem {
  words := strings.Split(line, " ")

  return JavaMethodItem{
    // Strip the last char off (the semicolon)
    Name: words[len(words) - 1][:len(words[len(words) - 1]) - 1],
    DeclarationType: "memberVariable",
    ReturnType: words[len(words) - 2],
    AccessModifiers: words[:len(words) - 2],
  }
}
