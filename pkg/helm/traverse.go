package helm

import (
	"fmt"
	"strings"
	"text/template"
	"text/template/parse"

	"github.com/sirupsen/logrus"
)

type Frame struct {
	Level  int
	Parent *Frame
	String string
}

func (f *Frame) New() *Frame {
	return f.NewWithString("")
}

func (f *Frame) NewWithString(s string) *Frame {
	return &Frame{Level: f.Level + 1, Parent: f, String: s}
}

func (f *Frame) StringPrefix() string {
	return strings.Join(f.stringPrefixHelper(), ".")
}

func (f *Frame) stringPrefixHelper() []string {
	if f == nil {
		return nil
	}
	if f.String == "" {
		return f.Parent.stringPrefixHelper()
	}
	return append(f.Parent.stringPrefixHelper(), f.String)
}

func Traverse(t *template.Template, process func(parse.Node, *Frame)) {
	traverse(t.Tree.Root, process, &Frame{})
}

func todoTraverseListNode(n *parse.ListNode, process func(parse.Node, *Frame)) {
	traverse(n, process, &Frame{})
}

// TODO: mess with resolving '.'
//    see https://pkg.go.dev/text/template
//
/* TODO here are some interesting things that don't yet work:
- resolving ?variable? created by 'set':
	- define: {{- $_ :=  set . "serviceName" "stuff" }}
	- use: {{- .serviceName }}
- resolving 'index'
	- (index .Values .serviceName)
	- (index .Values "my-stuff")
- resolving variable:
    - {{- $v1Openshift := .Capabilities.APIVersions.Has "security.openshift.io/v1" }}
- with: changes meaning of .
    {{- with .Values.stuff.ingress.annotations }}
		{{- toYaml . | nindent 4 }}
	{{- end }}
- range: changes meaning of .
      tls:
		{{- range .Values.stuff.ingress.tls }}
		- hosts:
			{{- with .hosts }}
				{{- toYaml . | nindent 8 }}
			{{- end }}
			secretName: {{ required "secretName is required" .secretName }}
		{{- end }}
- template: changes meaning of .
    - {{template "name"}} => unclear
	- {{template "name" pipeline}} => see https://pkg.go.dev/text/template
- how about: array in values.yaml, default value is empty => what is the schema of each array member?
- track position as (line, col) => may need to track/count newlines?  preprocess?
- when capturing position, need to know which file we're in
- ... ??? something else ??? ...
*/
func traverse(node parse.Node, process func(parse.Node, *Frame), frame *Frame) {
	if isNil(node) {
		logrus.Infof("skipping nil")
		return
	}
	process(node, frame)

	switch v := node.(type) {
	case *parse.TextNode:
	case *parse.ActionNode:
		traverse(v.Pipe, process, frame.New())
	case *parse.BoolNode:
	case *parse.ChainNode:
		traverse(v.Node, process, frame.New())
	case *parse.CommandNode:
		for _, a := range v.Args {
			traverse(a, process, frame.New())
		}
	case *parse.DotNode:
	case *parse.FieldNode:
	case *parse.IdentifierNode:
	case *parse.IfNode:
		traverse(v.Pipe, process, frame.New())
		traverse(v.List, process, frame.New())
		traverse(v.ElseList, process, frame.New())
	case *parse.ListNode:
		for _, n := range v.Nodes {
			traverse(n, process, frame.New())
		}
	case *parse.NilNode:
	case *parse.NumberNode:
	case *parse.PipeNode:
		for _, c := range v.Cmds {
			traverse(c, process, frame.New())
		}
		for _, d := range v.Decl {
			traverse(d, process, frame.New())
		}
	case *parse.RangeNode:
		// TODO add something to frame
		traverse(v.Pipe, process, frame.New())
		traverse(v.List, process, frame.New())
		traverse(v.ElseList, process, frame.New())
	case *parse.StringNode:
	case *parse.TemplateNode:
		traverse(v.Pipe, process, frame.New())
	case *parse.VariableNode:
	case *parse.WithNode:
		// frame rule:
		//   'resolve' field from pipe
		//   use resolved field to extend frame
		//   use extended frame to travere list (not sure if elselist is necessary?)
		field, _ := resolveField(v.Pipe, frame.StringPrefix())
		traverse(v.Pipe, process, frame.New())
		traverse(v.List, process, frame.NewWithString(field))
		traverse(v.ElseList, process, frame.NewWithString(field))
	case *parse.CommentNode:
	case *parse.BreakNode:
	case *parse.ContinueNode:
	default:
		panic("unexpected node type, we're done here")
	}
}

func resolveField(node parse.Node, prefix string) (string, bool) {
	switch v := node.(type) {
	case *parse.TextNode:
		return "", true
	case *parse.ActionNode:
		return resolveField(v.Pipe, prefix)
	case *parse.BoolNode:
		return "", true
	case *parse.ChainNode:
		res, ok := resolveField(v.Node, prefix)
		if !ok {
			return "", false
		}
		return res + "." + strings.Join(v.Field, "."), true
	case *parse.CommandNode:
		if len(v.Args) == 1 {
			return resolveField(v.Args[0], prefix)
		} else if len(v.Args) == 3 {
			if a, ok := v.Args[0].(*parse.IdentifierNode); !ok || a.Ident != "index" {
				return "", false
			}
			b, ok := v.Args[1].(*parse.FieldNode)
			if !ok {
				return "", false
			}
			c, ok := v.Args[2].(*parse.StringNode)
			if !ok {
				return "", false
			}
			return strings.Join(append(append([]string{}, b.Ident...), c.Text), "."), true
		}
		return "", false
	case *parse.DotNode:
		return prefix, true
	case *parse.FieldNode:
		return strings.Join(v.Ident, "."), true
	case *parse.IdentifierNode:
		return "", true
	case *parse.IfNode:
		return "", true
	case *parse.ListNode:
		return "", true
	case *parse.NilNode:
		return "", true
	case *parse.NumberNode:
		return "", true
	case *parse.PipeNode:
		if v.IsAssign || len(v.Decl) > 0 || len(v.Cmds) != 1 {
			return "", false
		}
		return resolveField(v.Cmds[0], prefix)
	case *parse.RangeNode:
		return "", true
	case *parse.StringNode:
		return "", true
	case *parse.TemplateNode:
		return "", true
	case *parse.VariableNode:
		return "", false
	case *parse.WithNode:
		return "", true
	case *parse.CommentNode:
		return "", true
	case *parse.BreakNode:
		return "", true
	case *parse.ContinueNode:
		return "", true
	default:
		panic("unexpected node type, we're done here")
	}
}

func processDebug(node parse.Node, frame *Frame) {
	prefix := strings.Repeat(" ", frame.Level)
	fmt.Printf(prefix+"%s: %d, %d: ", NodeTypeToString(node.Type()), node.Type(), node.Position())

	switch v := node.(type) {
	case *parse.TextNode:
		fmt.Printf("%s, %d\n", v.Text, v.Pos)
	case *parse.ActionNode:
		fmt.Printf("%d\n", v.Pos)
	case *parse.BoolNode:
		fmt.Printf("%t, %d\n", v.True, v.Pos)
	case *parse.ChainNode:
		fmt.Printf("%+v, %d\n", v.Field, v.Pos)
	case *parse.CommandNode:
		fmt.Printf("%d, %d\n", len(v.Args), v.Pos)
	case *parse.DotNode:
		fmt.Printf("%d\n", v.Pos)
	case *parse.FieldNode:
		fmt.Printf("%+v, %d\n", v.Ident, v.Pos)
	case *parse.IdentifierNode:
		fmt.Printf("%s, %d\n", v.Ident, v.Pos)
	case *parse.IfNode:
		fmt.Printf("%d\n", v.Pos)
	case *parse.ListNode:
		fmt.Printf("%d, %d\n", len(v.Nodes), v.Pos)
	case *parse.NilNode:
		fmt.Printf("%d\n", v.Pos)
	case *parse.NumberNode:
		fmt.Printf("%s\n", v.Text)
	case *parse.PipeNode:
		fmt.Printf("%t, %d\n", v.IsAssign, v.Pos)
	case *parse.RangeNode:
		fmt.Printf("%d\n", v.Pos)
	case *parse.StringNode:
		fmt.Printf("%s, %d\n", v.Quoted, v.Pos)
	case *parse.TemplateNode:
		fmt.Printf("%s, %d\n", v.Name, v.Pos)
	case *parse.VariableNode:
		fmt.Printf("%+v, %d\n", v.Ident, v.Pos)
	case *parse.WithNode:
		fmt.Printf("%d\n", v.Pos)
	case *parse.CommentNode:
		fmt.Printf("%s, %d\n", v.Text, v.Pos)
	case *parse.BreakNode:
		fmt.Printf("%d\n", v.Pos)
	case *parse.ContinueNode:
		fmt.Printf("%d\n", v.Pos)
	default:
		panic("unexpected node type, we're done here")
	}
}
