package do

import (
	"io"
	"log"
	"strings"
	"text/template"
)

const tableInfo = `
{{range $i, $table := .}}
### {{$table.Name}}
字段|类型|默认值|描述|
---|---|---|---
{{range $i, $f := $table.Columns}}
{{- $f.Name}} | {{$f.DataType }} | {{$f.Default| printDesc}} | {{ $f.Comment }}
{{end -}}
{{end -}}`

// 使用 table writer
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
