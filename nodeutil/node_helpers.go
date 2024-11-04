package nodeutil

import sitter "github.com/tree-sitter/go-tree-sitter"

// NamedChildrenOf gets all named children of a given node
func NamedChildrenOf(node *sitter.Node) []*sitter.Node {
	count := node.NamedChildCount()
	children := make([]*sitter.Node, count)
	for i := uint(0); i < count; i++ {
		children[i] = node.NamedChild(i)
	}
	return children
}

// UnnamedChildrenOf gets all the named + unnamed children of a given node
func UnnamedChildrenOf(node *sitter.Node) []*sitter.Node {
	count := node.ChildCount()
	children := make([]*sitter.Node, count)
	for i := uint(0); i < count; i++ {
		children[i] = node.Child(i)
	}
	return children
}

// Re-implemented from
// https://github.com/tree-sitter/go-tree-sitter/blob/5432ade78ad6bde797bcdaa272a5953eaba83971/src/node.c#L493
// NOTE: This does not check if the trees are equal
func Equals(fst *sitter.Node, snd *sitter.Node) bool {
	return fst.Id() == snd.Id()
}
