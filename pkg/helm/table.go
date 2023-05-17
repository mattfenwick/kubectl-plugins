package helm

import (
	"fmt"
	"strings"

	"github.com/mattfenwick/collections/pkg/dict"
	"github.com/mattfenwick/collections/pkg/iterable"
	"github.com/mattfenwick/collections/pkg/slice"
	"github.com/olekukonko/tablewriter"
)

func FieldsTable(fields map[string]int) string {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetAutoWrapText(false)
	table.SetRowLine(true)
	table.SetAutoMergeCells(true)
	table.SetHeader([]string{"Field", "Count"})

	for _, field := range slice.Sort(iterable.ToSlice(dict.KeysIterator(fields))) {
		pos := fields[field]
		table.Append([]string{field, fmt.Sprintf("%d", pos)})
		fmt.Printf("field %s: %+v\n", field, pos)
	}

	table.Render()
	return tableString.String()
}
