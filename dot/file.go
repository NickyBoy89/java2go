package dot

import (
	"fmt"
	"os"
	"strings"
)

type Dotfile struct {
	nodes map[string][]string
	*os.File
}

func New(name string) (*Dotfile, error) {
	file, err := os.Create(name)
	if err != nil {
		return nil, err
	}
	return &Dotfile{File: file, nodes: make(map[string][]string)}, nil
}

func (d *Dotfile) AddNode(name string, edges ...string) {
	d.nodes[name] = edges
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
	d.WriteString("digraph {\n")
	for name, edges := range d.nodes {
		fmt.Fprintf(d, "  \"%s\" -> {%s}\n", name, commaSeparatedString(edges))
	}
	d.WriteString("}")
}
