package helm

import (
	"fmt"
	"text/template"

	"github.com/mattfenwick/collections/pkg/dict"
	"github.com/mattfenwick/collections/pkg/iterable"
	"github.com/mattfenwick/collections/pkg/slice"
)

func Example() {
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

	for _, field := range slice.Sort(iterable.ToSlice(dict.KeysIterator(fields))) {
		pos := fields[field]
		fmt.Printf("field %s: %+v\n", field, pos)
	}

	fmt.Printf("%s\n", FieldsTable(dict.Map(slice.Length[int], fields)))

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
