package parsing

import (
	"context"
	"fmt"

	"github.com/NickyBoy89/java2go/symbol"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/java"
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

func (file *SourceFile) ParseAST() error {
	parser := sitter.NewParser()
	parser.SetLanguage(java.GetLanguage())
	tree, err := parser.ParseCtx(context.Background(), nil, file.Source)
	if err != nil {
		return err
	}

	file.Ast = tree.RootNode()
	return nil
}

func (file *SourceFile) ParseSymbols() *symbol.FileScope {
	symbols := symbol.ParseSymbols(file.Ast, file.Source)
	file.Symbols = symbols
	return symbols
}
