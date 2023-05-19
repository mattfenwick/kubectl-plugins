package helm

import (
	"fmt"
	"io/ioutil"
	"text/template"
	"text/template/parse"

	"github.com/mattfenwick/collections/pkg/dict"
	"github.com/mattfenwick/collections/pkg/slice"
	"github.com/mattfenwick/kubectl-plugins/pkg/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func Example(withSimple bool, withTemplate bool) {
	if withSimple {
		trees := Parse()
		for _, t := range trees {
			fmt.Printf("a tree: %s, %+v\n", t.Name, len(t.Root.Nodes))
			// todoTraverseListNode(t.Root, processDebug)
			fmt.Printf("%s\n", FieldsTable(dict.Map(slice.Length[int], CollectFields(t.Root))))
		}
		fmt.Printf("found %d trees\n", len(trees))
	}

	if withTemplate {
		sortedTemplates := ParseTemplate()

		// fields := map[string][]int{}
		for _, t := range sortedTemplates {
			fmt.Printf("%s\n", FieldsTable(dict.Map(slice.Length[int], CollectFields(t.Root))))

			// fmt.Printf("starting new template: %+v\n", t)
			// Traverse(t, findFields)
			// fmt.Println()
			// Traverse(t, collectFields(fields))
			// fmt.Println()
			// Traverse(t, processDebug)
		}

		// for _, field := range slice.Sort(iterable.ToSlice(dict.KeysIterator(fields))) {
		// 	pos := fields[field]
		// 	fmt.Printf("field %s: %+v\n", field, pos)
		// }

		// fmt.Printf("%s\n", FieldsTable(dict.Map(slice.Length[int], fields)))
	}
}

func CollectFields(nodes ...*parse.ListNode) map[string][]int {
	fields := map[string][]int{}
	for _, node := range nodes {
		todoTraverseListNode(node, collectFields(fields))
	}
	return fields
}

func ParseTemplate() []*template.Template {
	root := template.New("my-new")
	root = root.Funcs(helmBuiltinsMap)
	// myNewTemp = template.Must(myNewTemp.ParseGlob("./examples/*"))
	root = template.Must(root.ParseFiles("./examples/ingress.yaml"))
	fmt.Printf("hi? %t\n\n", root.Tree == nil)
	logrus.Infof("defined templates, my-new: %+v\n", root.DefinedTemplates())

	return slice.SortOn(func(t *template.Template) string { return t.Name() }, root.Templates())
}

func Parse() map[string]*parse.Tree {
	bytes, err := ioutil.ReadFile("./examples/ingress.yaml")
	utils.DoOrDie(errors.Wrapf(err, "unable to read file"))
	trees, err := parse.Parse("my-parse", string(bytes), "", "", parseBuiltinsMap)
	utils.DoOrDie(errors.Wrapf(err, "unable to parse string"))
	return trees
}
