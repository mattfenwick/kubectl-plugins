package helm

import (
	"reflect"
	"text/template"
	"text/template/parse"

	"github.com/mattfenwick/kubectl-plugins/pkg/utils"
	"github.com/pkg/errors"
)

func funcMap(xs ...string) template.FuncMap {
	out := template.FuncMap{}
	for _, x := range xs {
		s := x
		out[s] = func() string { return s }
	}
	return out
}

var (
	helmBuiltins = funcMap(
		"required", "quote", "default", "trunc", "trimSuffix", "replace", "include",
		"dict", "set", "toYaml", "nindent", "sha256sum", "indent", "list", "join", "int",
		"sha1sum",
		"toJson",
		"mergeOverwrite", "b64enc")
)

func isNil(v any) bool {
	// see https://stackoverflow.com/a/50487104/894284
	return v == nil || (reflect.ValueOf(v).Kind() == reflect.Ptr && reflect.ValueOf(v).IsNil())
}

func NodeTypeToString(t parse.NodeType) string {
	val, ok := NodeTypeToStringMap[t]
	if !ok {
		utils.DoOrDie(errors.Errorf("invalid node type: %d", t))
	}
	return val
}

var (
	NodeTypeToStringMap = map[parse.NodeType]string{
		parse.NodeAction:     "action",
		parse.NodeBool:       "bool",
		parse.NodeBreak:      "break",
		parse.NodeChain:      "chain",
		parse.NodeCommand:    "command",
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
		parse.NodeText:       "text",
		parse.NodeVariable:   "variable",
		parse.NodeWith:       "with",
	}
)
