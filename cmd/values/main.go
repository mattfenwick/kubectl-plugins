package main

import (
	"fmt"
	"reflect"
	"strings"
	"text/template"
	"text/template/parse"

	"github.com/mattfenwick/collections/pkg/slice"
	"github.com/sirupsen/logrus"
)

func h(s string) func() string {
	return func() string { return s }
}

func funcMap(xs ...string) template.FuncMap {
	out := template.FuncMap{}
	for _, x := range xs {
		out[x] = h(x)
	}
	return out
}

var (
	helmBuiltins = funcMap(
		"required", "quote", "default", "trunc", "trimSuffix", "replace", "include",
		"dict", "set", "toYaml", "nindent", "sha256sum", "indent", "list", "join", "int",
		"toJson",
		"mergeOverwrite", "b64enc")
)

func main() {
	// TODO try parsing with this to see if text/template's semi-evaluation can be avoided
	// parse.Parse()

	myNewTemp := template.New("my-new")
	myNewTemp = myNewTemp.Funcs(helmBuiltins)
	myNewTemp = template.Must(myNewTemp.ParseGlob("/Users/mfenwick/gitprojects/synopsys/cnc-umbrella-chart/charts/cnc/templates/*"))
	fmt.Printf("defined templates, my-new: %+v\n", myNewTemp.DefinedTemplates())

	fields := map[string][]int{}
	sortedTemplates := slice.SortOn(func(t *template.Template) string { return t.Name() }, myNewTemp.Templates())
	for _, t := range sortedTemplates {
		fmt.Printf("starting new template: %+v\n", t)
		Traverse(t, findFields)
		fmt.Println()
		Traverse(t, collectFields(fields))
		fmt.Println()
		Traverse(t, processDebug)
	}
	for field, pos := range fields {
		fmt.Printf("field %s: %+v\n", field, pos)
	}

	// fmt.Printf("t: %s\n", json.MustMarshalToString(len(um.Root.Nodes)))
	// for _, u := range um.Root.Nodes {
	// 	fmt.Printf("%+v: %+v\n", u.Type(), u.String())
	// }
	// fmt.Printf("t.Tree: %s\n", json.MustMarshalToString(um.Tree))
	// fmt.Printf("t.Root: %s\n", json.MustMarshalToString(um.Root))
	// Traverse(myNewTemp, processDebug)
	// fmt.Println()
	// Traverse(myNewTemp, printVariables)
	// fmt.Println()
	// Traverse(myNewTemp, findIdentifiers)
	// fmt.Println()
	// Traverse(myNewTemp, findFields)
	// fmt.Println()
}

func isNil(v any) bool {
	// see https://stackoverflow.com/a/50487104/894284
	return v == nil || (reflect.ValueOf(v).Kind() == reflect.Ptr && reflect.ValueOf(v).IsNil())
}

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
	fmt.Printf(prefix+"text: %s, %d: ", NodeToString[node.Type()], node.Position())

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

var (
	NodeToString = map[parse.NodeType]string{
		parse.NodeText:       "text",
		parse.NodeAction:     "action",
		parse.NodeBool:       "bool",
		parse.NodeBreak:      "break",
		parse.NodeChain:      "chain",
		parse.NodeComment:    "comment",
		parse.NodeContinue:   "continue",
		parse.NodeDot:        "dot",
		parse.NodeField:      "field",
		parse.NodeIdentifier: "identifier",
		parse.NodeIf:         "if",
		parse.NodeList:       "list",
		parse.NodeNil:        "nil",
		parse.NodeNumber:     "number",
		parse.NodePipe:       "pipe",
		parse.NodeRange:      "range",
		parse.NodeString:     "string",
		parse.NodeTemplate:   "template",
		parse.NodeVariable:   "variable",
		parse.NodeWith:       "with",
	}
)
