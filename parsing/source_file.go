package parsing

import (
	"context"
	"fmt"

	"github.com/NickyBoy89/java2go/symbol"
	sitter "github.com/tree-sitter/go-tree-sitter"
	java "github.com/tree-sitter/tree-sitter-java/bindings/go"
)

type SourceFile struct {
	Name    string
	Source  []byte
	Ast     *sitter.Node
	Symbols *symbol.FileScope
}

func (file SourceFile) String() string {
	return fmt.Sprintf("SourceFile { Name: %s, Ast: %v, Symbols: %v }", file.Name, file.Ast, file.Symbols)
}

func (file *SourceFile) ParseAST() {
	parser := sitter.NewParser()
	parser.SetLanguage(sitter.NewLanguage(java.Language()))
	tree := parser.ParseCtx(context.Background(), file.Source, nil)
	file.Ast = tree.RootNode()
}

func (file *SourceFile) ParseSymbols() *symbol.FileScope {
	symbols := symbol.ParseSymbols(file.Ast, file.Source)
	file.Symbols = symbols
	return symbols
}
