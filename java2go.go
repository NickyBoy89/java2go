package main

import (
  "log"
  "io/ioutil"
)

func main() {
  f, err := ioutil.ReadFile("testsnippets/simple.java")
  if err != nil {
    log.Fatalf("Opening source file failed with err: %v", err)
  }

  ParseFileContents(string(f))
}
