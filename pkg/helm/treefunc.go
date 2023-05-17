package helm

import (
	"fmt"
	"text/template/parse"

	"github.com/sirupsen/logrus"
)

func printVariables(node parse.Node, level int) {
	if v, ok := node.(*parse.VariableNode); ok {
		fmt.Printf("found variable: %+v, at %d\n", v.Ident, v.Pos)
	}
}

func findIdentifiers(node parse.Node, frame *Frame) {
	if v, ok := node.(*parse.IdentifierNode); ok {
		fmt.Printf("found identifier: %+v, at %d\n", v.Ident, v.Pos)
	}
}

func findFields(node parse.Node, frame *Frame) {
	if v, ok := node.(*parse.FieldNode); ok {
		fmt.Printf("found field: %+v, at %d\n", v.Ident, v.Pos)
	}
}

func collectFields(fields map[string][]int) func(parse.Node, *Frame) {
	add := func(field string, pos int) {
		if _, ok := fields[field]; !ok {
			fields[field] = []int{}
		}
		fields[field] = append(fields[field], pos)
	}
	return func(node parse.Node, frame *Frame) {
		val, ok := resolveField(node, frame.StringPrefix())
		if !ok {
			logrus.Warnf("unable to resolve field from %+v at %d", node, node.Position())
		} else if val == "" {
			logrus.Warnf("value mysteriously empty from %+v at %d", node, node.Position())
		} else {
			fmt.Printf("found field: %s, at %d\n", val, node.Position())
			add(val, int(node.Position()))
		}
	}
}
