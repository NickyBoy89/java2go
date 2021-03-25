func sliceContains(arr []string, target string) bool {
  for _, s := range arr {
    if s == target {
      return true
    }
  }
  return false
}

func prettyPrint(inp string) string {
  pretty, err := json.MarshalIndent(inp, "", "\t")
  if err != nil {
    log.Fatal(err)
  }
  return string(pretty)
}

func doBracketsMatchInString(text string) bool {

  bracketBalance := 0

  for _, char := range text {
    if char == '{' {
      bracketBalance--
    } else if char == '}' {
      bracketBalance++
    }
  }

  if bracketBalance != 0 {
    return false
  }
  return true
}
