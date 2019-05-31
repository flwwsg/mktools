package collect

import (
	"bytes"
	"html/template"
	"log"
	"sort"
	"strconv"
	"strings"
)

// APITemplate template of md
const APITemplate = `
### {{.ActionID}} {{.ActionDesc}}

#### 请求

字段|类型|描述|
---|---|---
{{range $i, $f := .ReqFields}}
{{- $f.Name}} | {{$f.TypeName | printf "%s" }}| {{printDesc $f.Desc }}
{{end}}

#### 响应

字段|类型|描述|
---|---|---
{{range $i, $f := .RespFields}}
{{- $f.Name}} | {{$f.TypeName | printf "%s" }} | {{printDesc $f.Desc }}
{{end -}}
`

// CustomTypeTemplate template for custom type
const CustomTypeTemplate = `
### 自定义数据类型
{{$length := len .}}
{{- if eq $length 0}}
#### 无
{{end -}}
{{range $i, $typ := .}}
#### {{$typ.Name}}
字段|类型|描述|
---|---|---
{{range $i, $f := $typ.Fields}}
{{- $f.Name}} | {{$f.TypeName | printf "%s" }} | {{printDesc $f.Desc }}
{{end -}}
{{end -}}
`

// SingleAPI an api
type SingleAPI struct {
	ActionID    string
	ActionDesc  string
	ReqFields   []ApiField
	RespFields  []ApiField
	CustomTypes map[string]*StructType
}

// 生成api
type ApiMaker struct {
	// 需要生成api的结构体, key = 路径+类型名
	allStruct map[string]*StructType
	// 组成接口号的结构体
	apiStruct map[string]*StructType
	// 接口， key=ActionID
	allAPI map[string]SingleAPI
	// 需要生成文档的api路径
	apiPath  string
	inStruct map[string]bool
}

func NewMaker(apiPath string) ApiMaker {
	return ApiMaker{apiPath: apiPath, allStruct: make(map[string]*StructType), allAPI: make(map[string]SingleAPI), inStruct: make(map[string]bool)}
}

// 解析需要的结构体
func (maker *ApiMaker) Parse() {
	// 确定接口需要的结构体
	maker.genRequiredStruct()
	for k, v := range maker.apiStruct {
		maker.collectTypesInStruct(v.PkgPath, k)
		for _, f := range v.Fields {
			maker.collectTypesInStruct(f.PkgPath, f.key)
		}
	}
	maker.genAPI()
}

// 确定结构体字段的类型
func (maker *ApiMaker) collectTypesInStruct(pkgPath string, key string) {
	_, ok := maker.allStruct[key]
	if ok || key == "" {
		return
	}
	maker.collectStructs(pkgPath)
	s := maker.allStruct[key]
	if s == nil {
		// not struct type
		return
	}
	for _, field := range s.Fields {
		maker.collectTypesInStruct(field.PkgPath, field.key)
	}
}

// 收集文件夹下的结构体
func (maker *ApiMaker) collectStructs(pkgPath string) {
	if maker.inStruct[pkgPath] || pkgPath == "" {
		return
	}
	pkg := NewPkgStructs(pkgPath)
	pkg.Parse()
	for k, v := range pkg.allStructs {
		maker.allStruct[k] = v
	}
	maker.inStruct[pkgPath] = true
}

// 获取请求、响应结构体
func (maker *ApiMaker) genRequiredStruct() {
	if maker.apiStruct != nil {
		return
	}
	apiStructs := make(map[string]*StructType)
	maker.collectStructs(maker.apiPath)
	for k, v := range maker.allStruct {
		if v.IsReq() || v.IsResp() {
			apiStructs[k] = v
		}
	}
	maker.apiStruct = apiStructs
}

// genAPI generating single api with ActionID
func (maker *ApiMaker) genAPI() map[string]SingleAPI {
	if len(maker.allAPI) != 0 {
		return maker.allAPI
	}
	for _, v := range maker.apiStruct {
		actionID := v.ActionID
		api := new(SingleAPI)
		api.CustomTypes = make(map[string]*StructType)
		st, ok := maker.allAPI[actionID]
		if ok {
			api.ActionDesc = st.ActionDesc
			api.CustomTypes = st.CustomTypes
			api.ReqFields = st.ReqFields
			api.RespFields = st.RespFields
		}
		// Desc of request must before Desc of response
		if v.IsResp() {
			api.RespFields = v.Fields
		}
		if v.IsReq() {
			api.ReqFields = v.Fields
		}
		api.ActionID = actionID
		maker.collectCustomTypes(api, v.Fields)
		maker.allAPI[actionID] = *api
	}
	return maker.allAPI
}

func (maker *ApiMaker) collectCustomTypes(api *SingleAPI, fields []ApiField) {
	for _, f := range fields {
		_, ok := api.CustomTypes[f.key]
		s, isStruct := maker.allStruct[f.key]
		if !ok && isStruct {
			api.CustomTypes[f.key] = s
			maker.collectCustomTypes(api, s.Fields)
		}
	}
}

// 格式化一条api
func (maker ApiMaker) formatOneSingleAPI(api SingleAPI) *bytes.Buffer {
	var printDesc = func(desc string) string {
		if desc == "" {
			return "无"
		}
		return strings.TrimSpace(desc)
	}
	var printNeed = func(need bool) string {
		if need {
			return "是"
		}
		return "否"

	}
	doc, err := template.New("request").Funcs(template.FuncMap{"printDesc": printDesc, "printNeed": printNeed}).
		Parse(APITemplate)
	if err != nil {
		log.Fatal(err)
	}
	b := new(bytes.Buffer)
	err = doc.Execute(b, api)
	if err != nil {
		panic(err)
	}
	return b
}

func (maker ApiMaker) formatCustomTypes(allStruct map[string]*StructType) *bytes.Buffer {
	var printDesc = func(desc string) string {
		if desc == "" {
			return "无"
		}
		return strings.TrimSpace(desc)
	}
	doc, err := template.New("CustomTypes").Funcs(template.FuncMap{"printDesc": printDesc}).
		Parse(CustomTypeTemplate)
	if err != nil {
		log.Fatal(err)
	}
	b := new(bytes.Buffer)
	err = doc.Execute(b, allStruct)
	if err != nil {
		panic(err)
	}
	return b
}

// 生成当前模块下接口文件
func (maker *ApiMaker) AsString() string {
	if len(maker.allAPI) == 0 {
		maker.Parse()
	}
	allAPI := maker.allAPI
	rtn := make([]string, len(allAPI)+1)
	idx := make([]int, len(allAPI))
	i := 0
	for k := range allAPI {
		n, err := strconv.Atoi(k)
		if err != nil {
			log.Fatal(err)
		}
		idx[i] = n
		i++
	}
	sort.Ints(idx)
	customTypes := make(map[string]*StructType)
	for i, aid := range idx {
		strAID := strconv.Itoa(aid)
		api := allAPI[strAID]
		b := maker.formatOneSingleAPI(api)
		rtn[i+1] = b.String()
		for k, v := range api.CustomTypes {
			customTypes[k] = v
		}
	}
	b := maker.formatCustomTypes(customTypes)
	rtn[0] = b.String()
	s := strings.Join(rtn, "")
	return s
}
