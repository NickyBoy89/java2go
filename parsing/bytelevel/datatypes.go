package bytelevel

type ParsedFile struct {
  Class ParsedClass
  StaticMethods []ParsedMethod
  StaticVariables []ParsedVariable
}

type ParsedVariable struct {
  Name string
  AccessModifiers []string
  DataType string
}

type ParsedClass struct {
  Name string
  AccessModifiers []string
  ClassVariables []ParsedVariable
  Methods []ParsedMethod
}

type ParsedMethod struct {
  Name string
  AccessModifiers []string
  Parameters []ParsedVariable
  ReturnType string
  Body string
}
