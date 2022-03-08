package dot

import (
	"fmt"
	"os"
	"strings"
)

type Dotfile struct {
	SubGraph
	*os.File
}

func (d Dotfile) Name() string {
	return d.SubGraph.Name()
}

func (d *Dotfile) DeleteSubgraph(name string) {
	delete(d.SubGraph.subgraphs, name)
}

type GraphPrinter interface {
	AsDot() (string, []Edge)
	Name() string
}

// GraphItem is any item in the graph
type GraphItem interface {
	GraphPrinter
	HasSubgraph(name string) bool         // Whether the item has the given subgraph
	Subgraph(name string) GraphItem       // Returns the named subgraph, adding it if necessary
	AddNode(name string, edges ...string) // Adds a node to the graph
}

type Edge struct {
	From string
	To   []string
}

type SubGraph struct {
	name      string
	nodes     map[string]Node
	subgraphs map[string]GraphItem
}

// Methods for GraphItem
func (g SubGraph) HasSubgraph(name string) bool {
	_, has := g.subgraphs[name]
	return has
}

func (g *SubGraph) Subgraph(name string) GraphItem {
	if g.subgraphs == nil {
		g.subgraphs = make(map[string]GraphItem)
	}
	if _, in := g.subgraphs[name]; !in {
		g.subgraphs[name] = &SubGraph{name: name}
	}
	return g.subgraphs[name]
}

func (g *SubGraph) AddNode(name string, edges ...string) {
	if g.nodes == nil {
		g.nodes = make(map[string]Node)
	}
	g.nodes[name] = Node{name: name, edges: edges}
}

func (g SubGraph) Name() string {
	return g.name
}

func (g SubGraph) AsDot() (string, []Edge) {
	totalEdges := []Edge{}
	total := fmt.Sprintf("subgraph cluster_%s {\n", g.name)
	total += fmt.Sprintf("label=\"%s\"\n", g.name)
	for _, item := range g.nodes {
		totalEdges = append(totalEdges, Edge{From: item.name, To: item.edges})
		total += item.Name() + "\n"
	}
	for _, item := range g.subgraphs {
		sub, edges := item.AsDot()
		total += sub + "\n"
		totalEdges = append(totalEdges, edges...)
	}
	return total + "}", totalEdges
}

type Node struct {
	name  string
	edges []string
}

func (n Node) AsDot() (string, []Edge) {
	return "", []Edge{Edge{From: n.name, To: n.edges}}
}

func (n Node) Name() string {
	return n.name
}

func New(name string) (*Dotfile, error) {
	file, err := os.Create(name)
	if err != nil {
		return nil, err
	}
	return &Dotfile{File: file}, nil
}

func (d *Dotfile) AddNode(name string, edges ...string) {
	if d.nodes == nil {
		d.nodes = make(map[string]Node)
	}
	d.nodes[name] = Node{name: name, edges: edges}
}

func (d *Dotfile) HasNode(name string) bool {
	_, in := d.nodes[name]
	return in
}

func (d *Dotfile) AddEdge(node string, edge string) {
	temp := d.nodes[node]
	// If the node doesn't exist, create it
	if temp.name == "" {
		temp = Node{name: node}
	}
	temp.edges = append(temp.edges, edge)
	d.nodes[node] = temp
}

func (d *Dotfile) HasEdge(node string, edge string) bool {
	for _, e := range d.nodes[node].edges {
		if e == edge {
			return true
		}
	}
	return false
}

func commaSeparatedString(list []string) string {
	var total strings.Builder
	for ind, item := range list {
		total.WriteString("\"" + item + "\"")
		if ind < len(list)-1 {
			total.WriteString(", ")
		}
	}
	return total.String()
}

func (d *Dotfile) WriteToFile() {
	totalEdges := []Edge{}
	d.WriteString("digraph {\n")
	//d.WriteString("compound=true\n")

	// First, write out all the subgraphs
	for _, graph := range d.subgraphs {
		sub, edges := graph.AsDot()
		totalEdges = append(totalEdges, edges...)
		d.WriteString(sub + "\n")
	}

	// Then, go through the nodes
	for name, node := range d.nodes {
		fmt.Fprintf(d, "  \"%s\" -> {%s}\n", name, commaSeparatedString(node.edges))
	}

	// Finally, connect all the edges from everything else
	for _, edge := range totalEdges {
		// Skip creating edges that don't point anywhere
		if len(edge.To) == 0 {
			continue
		}
		// Also skip empty nodes
		if edge.From == "" {
			continue
		}
		fmt.Fprintf(d, "\"%s\" -> {%s}\n", edge.From, commaSeparatedString(edge.To))
	}
	d.WriteString("}")
}
