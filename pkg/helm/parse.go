package helm

import (
	"fmt"
	"io/ioutil"
	"text/template"
	"text/template/parse"

	"github.com/mattfenwick/collections/pkg/dict"
	"github.com/mattfenwick/collections/pkg/iterable"
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
			todoTraverseListNode(t.Root, processDebug)
		}
		fmt.Printf("found %d trees\n", len(trees))
	}

	if withTemplate {
		sortedTemplates := ParseTemplate()

		fields := map[string][]int{}
		for _, t := range sortedTemplates {
			fmt.Printf("starting new template: %+v\n", t)
			Traverse(t, findFields)
			fmt.Println()
			Traverse(t, collectFields(fields))
			fmt.Println()
			Traverse(t, processDebug)
		}

		for _, field := range slice.Sort(iterable.ToSlice(dict.KeysIterator(fields))) {
			pos := fields[field]
			fmt.Printf("field %s: %+v\n", field, pos)
		}

		fmt.Printf("%s\n", FieldsTable(dict.Map(slice.Length[int], fields)))
	}
}

func ParseTemplate() []*template.Template {
	myNewTemp := template.New("my-new")
	myNewTemp = myNewTemp.Funcs(helmBuiltins)
	myNewTemp = template.Must(myNewTemp.ParseGlob("/Users/mfenwick/gitprojects/synopsys/cnc-umbrella-chart/charts/cnc/templates/*"))
	logrus.Infof("defined templates, my-new: %+v\n", myNewTemp.DefinedTemplates())

	return slice.SortOn(func(t *template.Template) string { return t.Name() }, myNewTemp.Templates())
}

func Parse() map[string]*parse.Tree {
	bytes, err := ioutil.ReadFile("/Users/mfenwick/gitprojects/synopsys/cnc-umbrella-chart/charts/cnc/templates/ingress.yaml")
	utils.DoOrDie(errors.Wrapf(err, "unable to read file"))
	trees, err := parse.Parse("my-parse", string(bytes), "", "", helmBuiltins)
	utils.DoOrDie(errors.Wrapf(err, "unable to parse string"))
	return trees
}
