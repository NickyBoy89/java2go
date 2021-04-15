package parsing

import (
  "io/ioutil"
  "testing"
  "encoding/json"

  "github.com/sergi/go-diff/diffmatchpatch"
)

func TestBasicClass(t *testing.T) {
  testFile, err := ioutil.ReadFile("../testfiles/Test.java")
  if err != nil {
    t.Fatalf("Reading file failed with err: %v", err)
  }

  comparison, err := json.MarshalIndent(ParsedClass{
    Name: "Test",
    Modifiers: []string{},
    Implements: []string{},
    ClassVariables: []ParsedVariable{
      ParsedVariable{Name: "value", DataType: "int", Modifiers: []string{}},
    },
    Methods: []ParsedMethod{
      ParsedMethod{
        Name: "Test",
        Modifiers: []string{"public"},
        ReturnType: "constructor",
        Parameters: []ParsedVariable{ParsedVariable{Name: "val", DataType: "int", Modifiers: []string{}}},
        Body: `this.value = val;`,
      },
      ParsedMethod{
        Name: "GetValue",
        Modifiers: []string{"public"},
        Parameters: []ParsedVariable{},
        ReturnType: "int",
        Body: `return this.value;`,
      },
    },
    NestedClasses: []ParsedClasses{},
    StaticBlocks: []string{},
  }, "", "  ")
  if err != nil {
    t.Fatalf("Failed to parse result with err: %v", err)
  }

  parsedResult, err := json.MarshalIndent(ParseClass(string(testFile)), "", "  ")
  if err != nil {
    t.Fatalf("Failed to parse result with err: %v", err)
  }

  if string(parsedResult) != string(comparison) {
    t.Log(string(parsedResult))
    diff := diffmatchpatch.New()
    t.Log(diff.DiffPrettyText(diff.DiffMain(string(parsedResult), string(comparison), false)))
    t.Error("Result and Original did not match")
  }
}

func TestSimpleParse(t *testing.T) {
  testFile, err := ioutil.ReadFile("../testfiles/simple.java")
  if err != nil {
    t.Fatalf("Reading file failed with err: %v", err)
  }

  comparison, err := json.MarshalIndent(ParsedClass{
    Name: "Test",
    Modifiers: []string{},
    Implements: []string{},
    ClassVariables: []ParsedVariable{
      ParsedVariable{Name: "value", DataType: "int", Modifiers: []string{}},
      ParsedVariable{Name: "value2", DataType: "int", Modifiers: []string{"public"}},
      ParsedVariable{Name: "value3", DataType: "int", Modifiers: []string{"private"}},
    },
    Methods: []ParsedMethod{
      ParsedMethod{
        Name: "Test",
        Modifiers: []string{},
        ReturnType: "constructor",
        Parameters: []ParsedVariable{ParsedVariable{Name: "value", DataType: "int", Modifiers: []string{}}},
        Body: `this.value = value;this.value2 = value + 1;this.value3 = value + 2;`,
      },
      ParsedMethod{
        Name: "getValue",
        Modifiers: []string{"public"},
        Parameters: []ParsedVariable{ParsedVariable{Name: "specified", DataType: "int", Modifiers: []string{}}},
        ReturnType: "int",
        Body: `if (specified == 1) {return this.value;} else if (specified == 2) {return this.value2;} else {return this.value3;}`,
      },
      ParsedMethod{
        Name: "hello",
        Modifiers: []string{"public", "static"},
        Parameters: []ParsedVariable{},
        ReturnType: "String",
        Body: `return "Hello World!";`,
      },
    },
    NestedClasses: []ParsedClasses{},
    StaticBlocks: []string{},
  }, "", "  ")
  if err != nil {
    t.Fatalf("Failed to parse result with err: %v", err)
  }

  parsedResult, err := json.MarshalIndent(ParseClass(string(testFile)), "", "  ")
  if err != nil {
    t.Fatalf("Failed to parse result with err: %v", err)
  }

  if string(parsedResult) != string(comparison) {
    t.Log(string(parsedResult))
    diff := diffmatchpatch.New()
    t.Log(diff.DiffPrettyText(diff.DiffMain(string(parsedResult), string(comparison), false)))
    t.Error("Result and Original did not match")
  }
}

func TestParseLinkedList(t *testing.T) {
  testFile, err := ioutil.ReadFile("../testfiles/IntLinkedList.java")
  if err != nil {
    t.Fatalf("Reading file failed with err: %v", err)
  }

  comparison, err := json.MarshalIndent(ParsedClass{
    Name: "IntLinkedList",
    Modifiers: []string{"public"},
    Implements: []string{},
    ClassVariables: []ParsedVariable{
      ParsedVariable{
        Name: "size",
        Modifiers: []string{},
        DataType: "int",
      },
      ParsedVariable{
        Name: "head",
        Modifiers: []string{},
        DataType: "Node",
      },
      ParsedVariable{
        Name: "tail",
        Modifiers: []string{},
        DataType: "Node",
      },
    },
    Methods: []ParsedMethod{
      ParsedMethod{
        Name: "IntLinkedList",
        Modifiers: []string{"public"},
        Parameters: []ParsedVariable{},
        ReturnType: "constructor",
        Body: `this.head = null;this.tail = null;this.size = 0;`,
      },
      ParsedMethod{
        Name: "size",
        Modifiers: []string{"public"},
        Parameters: []ParsedVariable{},
        ReturnType: "int",
        Body: `return this.size;`,
      },
      ParsedMethod{
        Name: "add",
        Modifiers: []string{"public"},
        Parameters: []ParsedVariable{
          ParsedVariable{
            Name: "element",
            Modifiers: []string{},
            DataType: "int",
          },
        },
        ReturnType: "void",
        Body: `if (this.head == null) { this.size++;this.head = new Node(element);this.tail = this.head;} else {this.size++;this.tail.next = new Node(element);this.tail = this.tail.next;}`,
      },
      ParsedMethod{
        Name: "get",
        Modifiers: []string{"public"},
        Parameters: []ParsedVariable{
          ParsedVariable{
            Name: "index",
            Modifiers: []string{},
            DataType: "int",
          },
        },
        ReturnType: "int",
        Body: `Node currentNode = this.head;for (int i = 1; i <= index; i++) {currentNode = currentNode.next;}return currentNode.data;`,
      },
      ParsedMethod{
        Name: "remove",
        Modifiers: []string{"public"},
        Parameters: []ParsedVariable{
          ParsedVariable{
            Name: "index",
            Modifiers: []string{},
            DataType: "int",
          },
        },
        ReturnType: "void",
        Body: `Node currentNode = this.head;Node previousNode = this.head;if (index == 0) { this.head = this.head.next;}for (int i = 1; i <= index; i++) {previousNode = currentNode;currentNode = previousNode.next;}if (currentNode.next == null) { this.tail = previousNode;} else {previousNode.next = currentNode.next;}this.size--;`,
      },
      ParsedMethod{
        Name: "toString",
        Modifiers: []string{"public"},
        Parameters: []ParsedVariable{},
        ReturnType: "String",
        Body: `String result = "{";if (this.size() > 0) {result += this.get(0);}for (int i = 1; i < this.size; i++) {result += ", " + this.get(i);}result += "}";return result;`,
      },
      ParsedMethod{
        Name: "assertArraysEqual",
        Modifiers: []string{"public", "static"},
        Parameters: []ParsedVariable{
          ParsedVariable{
            Name: "list",
            Modifiers: []string{},
            DataType: "IntLinkedList",
          },
          ParsedVariable{
            Name: "answer",
            Modifiers: []string{},
            DataType: "int[]",
          },
        },
        ReturnType: "void",
        Body: `if (list.size() != answer.length) {throw new AssertionError("Expected list of length " + answer.length + " but got " + list.size());}for (int i = 0; i < answer.length; i++) {if (list.get(i) != answer[i]) {throw new AssertionError("Expected " + answer[i] + " but got " + list.get(i) + " at index " + i);}}`,
      },
      ParsedMethod{
        Name: "test1",
        Modifiers: []string{"public", "static"},
        Parameters: []ParsedVariable{},
        ReturnType: "void",
        Body: `IntLinkedList list = new IntLinkedList();int[] answer = new int[0];assertArraysEqual(list, answer);`,
      },
      ParsedMethod{
        Name: "test2",
        Modifiers: []string{"public", "static"},
        Parameters: []ParsedVariable{},
        ReturnType: "void",
        Body: `IntLinkedList list = new IntLinkedList();for (int i = 0; i < 3; i++) {list.add(i * i);}int[] answer = {0, 1, 4};assertArraysEqual(list, answer);`,
      },
      ParsedMethod{
        Name: "test3",
        Modifiers: []string{"public", "static"},
        Parameters: []ParsedVariable{},
        ReturnType: "void",
        Body: `IntLinkedList list = new IntLinkedList();for (int i = 0; i < 5; i++) {list.add(i * i);}list.remove(1);list.remove(2);int[] answer = {0, 4, 16};IntLinkedList.assertArraysEqual(list, answer);`,
      },
      ParsedMethod{
        Name: "test4",
        Modifiers: []string{"public", "static"},
        Parameters: []ParsedVariable{},
        ReturnType: "void",
        Body: `IntLinkedList list = new IntLinkedList();for (int i = 0; i < 5; i++) {list.add(i * i);}list.remove(1);list.remove(2);list.remove(0);list.remove(0);list.remove(0);int[] answer1 = {};IntLinkedList.assertArraysEqual(list, answer1);for (int i = 0; i < 5; i++) {list.add(i * i);}list.remove(4);list.add(-1);int[] answer2 = {0, 1, 4, 9, -1};IntLinkedList.assertArraysEqual(list, answer2);`,
      },
      ParsedMethod{
        Name: "test5",
        Modifiers: []string{"public", "static"},
        Parameters: []ParsedVariable{},
        ReturnType: "void",
        Body: `IntLinkedList list = new IntLinkedList();for (int i = 0; i < 12; i++) {list.add(i * i);}int[] answer = {0, 1, 4, 9, 16, 25, 36, 49, 64, 81, 100, 121};IntLinkedList.assertArraysEqual(list, answer);`,
      },
      ParsedMethod{
        Name: "main",
        Modifiers: []string{"public", "static"},
        Parameters: []ParsedVariable{
          ParsedVariable{
            Name: "args",
            Modifiers: []string{},
            DataType: "String[]",
          },
        },
        ReturnType: "void",
        Body: `test1();test2();test3();test4();test5();System.out.println("pass");`,
      },
    },
    NestedClasses: []ParsedClasses{
      ParsedClass{
        Name: "Node",
        Modifiers: []string{},
        Implements: []string{},
        ClassVariables: []ParsedVariable{
          ParsedVariable{
            Name: "data",
            Modifiers: []string{},
            DataType: "int",
          },
          ParsedVariable{
            Name: "next",
            Modifiers: []string{},
            DataType: "Node",
          },
        },
        Methods: []ParsedMethod{
          ParsedMethod{
            Name: "Node",
            Modifiers: []string{},
            Parameters: []ParsedVariable{
              ParsedVariable{
                Name: "data",
                Modifiers: []string{},
                DataType: "int",
              },
            },
            ReturnType: "constructor",
            Body: `this.data = data;this.next = null;`,
          },
        },
        NestedClasses: []ParsedClasses{},
        StaticBlocks: []string{},
      },
    },
    StaticBlocks: []string{},
  }, "", "  ")
  if err != nil {
    t.Fatalf("Failed to parse result with err: %v", err)
  }

  parsedResult, err := json.MarshalIndent(ParseClass(string(testFile)), "", "  ")
  if err != nil {
    t.Fatalf("Failed to parse result with err: %v", err)
  }

  if string(parsedResult) != string(comparison) {
    diff := diffmatchpatch.New()
    t.Log(diff.DiffPrettyText(diff.DiffMain(string(parsedResult), string(comparison), false)))
    t.Error("Result and Original did not match")
  }
}

func TestParseSimpleInterface(t *testing.T) {
  testFile, err := ioutil.ReadFile("../testfiles/simpleinterface.java")
  if err != nil {
    t.Fatalf("Reading file failed with err: %v", err)
  }

  comparison, err := json.MarshalIndent(ParsedInterface{
    Name: "TestCoordGetter",
    Modifiers: []string{},
    Methods: []ParsedMethod{
      ParsedMethod{
        Name: "getX",
        Modifiers: []string{"public"},
        ReturnType: "int",
        Parameters: []ParsedVariable{},
      },
      ParsedMethod{
        Name: "getY",
        Modifiers: []string{"public"},
        ReturnType: "int",
        Parameters: []ParsedVariable{},
      },
      ParsedMethod{
        Name: "getZ",
        Modifiers: []string{"public"},
        ReturnType: "int",
        Parameters: []ParsedVariable{},
      },
      ParsedMethod{
        Name: "choose",
        Modifiers: []string{},
        ReturnType: "int",
        Parameters: []ParsedVariable{},
      },
    },
    DefaultMethods: []ParsedMethod{},
  }, "", "  ")
  if err != nil {
    t.Fatalf("Failed to parse result with err: %v", err)
  }

  parsedResult, err := json.MarshalIndent(ParseFile(string(testFile)), "", "  ")
  if err != nil {
    t.Fatalf("Failed to parse result with err: %v", err)
  }

  if string(parsedResult) != string(comparison) {
    t.Log(string(parsedResult))
    diff := diffmatchpatch.New()
    t.Log(diff.DiffPrettyText(diff.DiffMain(string(parsedResult), string(comparison), false)))
    t.Error("Result and Original did not match")
  }
}

func TestSimpleAnnotation(t *testing.T) {
  testFile, err := ioutil.ReadFile("../testfiles/simpleannotation.java")
  if err != nil {
    t.Fatalf("Reading file failed with err: %v", err)
  }

  comparison, err := json.MarshalIndent(ParsedClass{
    Name: "Pet",
    Modifiers: []string{"public"},
    Implements: []string{},
    ClassVariables: []ParsedVariable{
      ParsedVariable{
        Name: "name",
        DataType: "String",
        Annotation: "@Nullable",
        Modifiers: []string{"private"},
      },
    },
    Methods: []ParsedMethod{
      ParsedMethod{
        Name: "setName",
        Modifiers: []string{"public"},
        ReturnType: "void",
        Parameters: []ParsedVariable{
          ParsedVariable{
            Name: "name",
            DataType: "String",
            Annotation: "@Nullable",
            Modifiers: []string{},
          },
        },
        Body: `this.name = name;`,
      },
      ParsedMethod{
        Name: "getName",
        Modifiers: []string{"public"},
        Annotation: "@Nullable",
        Parameters: []ParsedVariable{},
        ReturnType: "String",
        Body: `return this.name;`,
      },
      ParsedMethod{
        Name: "sayHello",
        Modifiers: []string{"public"},
        Annotation: `@Deprecated
@Environment(EnvType.CLIENT)`,
        Parameters: []ParsedVariable{},
        ReturnType: "String",
        Body: `return "Hello World!";`,
      },
    },
    NestedClasses: []ParsedClasses{},
    StaticBlocks: []string{},
  }, "", "  ")
  if err != nil {
    t.Fatalf("Failed to parse result with err: %v", err)
  }

  parsedResult, err := json.MarshalIndent(ParseClass(string(testFile)), "", "  ")
  if err != nil {
    t.Fatalf("Failed to parse result with err: %v", err)
  }

  if string(parsedResult) != string(comparison) {
    t.Log(string(parsedResult))
    diff := diffmatchpatch.New()
    t.Log(diff.DiffPrettyText(diff.DiffMain(string(parsedResult), string(comparison), false)))
    t.Error("Result and Original did not match")
  }
}

func TestSimpleStatic(t *testing.T) {
  testFile, err := ioutil.ReadFile("../testfiles/simplestatic.java")
  if err != nil {
    t.Fatalf("Reading file failed with err: %v", err)
  }

  comparison, err := json.MarshalIndent(ParsedClass{
    Name: "MathUtils",
    Modifiers: []string{},
    Implements: []string{},
    ClassVariables: []ParsedVariable{
      ParsedVariable{
        Name: "PI",
        DataType: "double",
        Modifiers: []string{"public", "static", "final"},
        InitialValue: "3.14159",
      },
      ParsedVariable{
        Name: "NILSTRING",
        DataType: "String",
        Annotation: "@Nullable",
        Modifiers: []string{"public", "static", "final"},
        InitialValue: "new String()",
      },
    },
    Methods: []ParsedMethod{
      ParsedMethod{
        Name: "Min",
        Modifiers: []string{"public", "static"},
        ReturnType: "int",
        Parameters: []ParsedVariable{
          ParsedVariable{
            Name: "x1",
            DataType: "int",
            Modifiers: []string{},
          },
          ParsedVariable{
            Name: "x2",
            DataType: "int",
            Modifiers: []string{},
          },
        },
        Body: `return x1 < x2 ? x1 : x2;`,
      },
      ParsedMethod{
        Name: "GetPi",
        Modifiers: []string{"public", "static"},
        Parameters: []ParsedVariable{},
        ReturnType: "double",
        Body: `return PI;`,
      },
    },
    NestedClasses: []ParsedClasses{},
    StaticBlocks: []string{`int i = 10;
for (int j = 0; j < i; j++) {
System.out.println(j);
}`},
  }, "", "  ")
  if err != nil {
    t.Fatalf("Failed to parse result with err: %v", err)
  }

  parsedResult, err := json.MarshalIndent(ParseClass(string(testFile)), "", "  ")
  if err != nil {
    t.Fatalf("Failed to parse result with err: %v", err)
  }

  if string(parsedResult) != string(comparison) {
    t.Log(string(parsedResult))
    diff := diffmatchpatch.New()
    t.Log(diff.DiffPrettyText(diff.DiffMain(string(parsedResult), string(comparison), false)))
    t.Error("Result and Original did not match")
  }
}

func IndexThatStringsDiffer(s1, s2 string) int {
  for i := range s1 {
    if i > len(s2) || s1[i] != s2[i] {
      return i
    }
  }
  return -1
}
