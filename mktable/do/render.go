package do

import (
	"io"
	"log"
	"strings"
	"text/template"

	"gitee.com/flwwsg/utils-go/files"
)

const tableInfo = `
{{range $i, $table := .}}
### {{$table.Name}} {{$table.Comment}}
字段|类型|默认值|描述|
---|---|---|---
{{range $i, $f := $table.Columns}}
{{- $f.Name}} | {{$f.DataType }} | {{$f.Default| printDesc}} | {{ $f.Comment }}
{{end -}}
{{end -}}`

// RenderTable 使用 table writer
func RenderTable(tables []Table, out io.Writer) {
	var printDesc = func(desc string) string {
		if desc == "" {
			return "无"
		}
		return strings.TrimSpace(desc)
	}
	doc, err := template.New("tables").Funcs(template.FuncMap{"printDesc": printDesc}).Parse(tableInfo)
	if err != nil {
		log.Fatal(err)
	}
	err = doc.Execute(out, tables)
	if err != nil {
		panic(err)
	}
}

// SplitTable 根据条件区分模块
func SplitTable(tables []Table, module map[string][]string, ignorePattern []string) (moduleTable map[string][]Table) {
	moduleTable = make(map[string][]Table)
	for k, v := range module {
		for i := range tables {
			name := tables[i].Name
			if !files.RexTest(name, ignorePattern...) && files.RexTest(name, v...) {
				_, ok := moduleTable[k]
				if ok {
					moduleTable[k] = append(moduleTable[k], tables[i])
				} else {
					moduleTable[k] = []Table{tables[i]}
				}
			}
		}

	}
	return
}
