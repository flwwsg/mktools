package do

import (
	"bytes"
	"fmt"
	"log"
	"mktools/common"
	"sort"
	"strconv"
	"strings"
	"text/template"
)

//生成 api 文档

// APITemplate template of md
const APITemplate = `
### {{.ActionID}} {{.ActionDesc}}

#### 请求 {{.StructInName}}

字段|类型|描述|
---|---|---
{{range $i, $f := .ReqFields}}
{{- $f.Name}} | {{$f.TypeName | printf "%s" }}| {{printDesc $f.Desc }}
{{end}}

#### 响应 {{.StructOutName}}

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
	ActionDesc    string
	StructInName  string // 请求的结体体名
	StructOutName string // 返回的结体名
	ActionID      string
	ModName       string //模块名
	ReqFields     []common.ApiField
	RespFields    []common.ApiField
	CustomTypes   map[string]*FastStructType
}

//生成api
type ApiMaker struct {
	//需要生成api的结构体, key = 路径+类型名
	allStruct map[string]*FastStructType
	//组成接口的结构体, key = package path + type
	apiStruct map[string]*FastStructType
	//接口， key=actionID
	allAPI map[string]SingleAPI
	//需要生成文档的api路径
	apiPath  string
	inStruct map[string]bool
}

func NewMaker(apiPath string) ApiMaker {
	return ApiMaker{apiPath: apiPath, allStruct: make(map[string]*FastStructType), allAPI: make(map[string]SingleAPI), inStruct: make(map[string]bool)}
}

//解析需要的结构体
func (maker *ApiMaker) Parse() {
	//确定接口需要的结构体
	maker.genRequiredStruct()
	for k, v := range maker.apiStruct {
		maker.collectTypesInStruct(v.PkgPath, k)
		for _, f := range v.Fields {
			maker.collectTypesInStruct(f.PkgPath, f.GetKey())
		}
	}
	maker.genAPI()
	for k, v := range maker.allAPI {
		fmt.Printf("%s, %v\n", k, v)
	}
}

//生成当前模块下接口文件
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
	customTypes := make(map[string]*FastStructType)
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

func (maker *ApiMaker) genRequiredStruct() {
	if maker.apiStruct != nil {
		return
	}
	maker.collectStructs(maker.apiPath)
}

//收集结构体
func (maker *ApiMaker) collectStructs(pkgPath string) {
	if maker.inStruct[pkgPath] || pkgPath == "" {
		return
	}
	pkg := NewPkgStructs(pkgPath)
	pkg.Parse()
	apiStruct := make(map[string]*FastStructType)
	for k, v := range pkg.allStructs {
		maker.allStruct[k] = v
	}
	maker.inStruct[pkgPath] = true
	if pkg.api.recv == nil {
		return
	}
	//假设所有的请求、响应结构体都在本包内
	for k, v := range pkg.api.req {
		if v != nil {
			st := maker.allStruct[v.Key]
			st.ActionID = k
			st.isReq = true
			apiStruct[v.Key] = st
			continue
		}
		st := emptyStruct()
		st.ActionID = k
		st.isReq = true
		apiStruct["defaultReq"+k] = st
	}
	for k, v := range pkg.api.resp {
		if v != nil {
			st := maker.allStruct[v.Key]
			st.ActionID = k
			st.isResp = true
			apiStruct[v.Key] = st
			continue
		}
		st := emptyStruct()
		st.ActionID = k
		st.isResp = true
		apiStruct["defaultResp"+k] = st

	}
	maker.apiStruct = apiStruct
}

//确定结构体字段的类型
func (maker *ApiMaker) collectTypesInStruct(pkgPath string, key string) {
	_, ok := maker.allStruct[key]
	if ok || key == "" {
		return
	}
	maker.collectStructs(pkgPath)
	s := maker.allStruct[key]
	if s == nil {
		//not struct type
		return
	}
	for _, field := range s.Fields {
		maker.collectTypesInStruct(field.PkgPath, field.GetKey())
	}
}

//genAPI generating single api with ActionID
func (maker *ApiMaker) genAPI() map[string]SingleAPI {
	if len(maker.allAPI) != 0 {
		return maker.allAPI
	}
	for _, v := range maker.apiStruct {
		actionID := v.ActionID
		api := new(SingleAPI)
		api.CustomTypes = make(map[string]*FastStructType)
		st, ok := maker.allAPI[actionID]
		if ok {
			api.ActionDesc = st.ActionDesc
			api.CustomTypes = st.CustomTypes
			api.ReqFields = st.ReqFields
			api.RespFields = st.RespFields
			api.StructInName = st.StructInName
			api.StructOutName = st.StructOutName
		}
		//Desc of request must before Desc of response
		if v.IsResp() {
			if api.ActionDesc == "" {
				api.ActionDesc = v.Desc
			}
			api.StructOutName = v.Name
			api.RespFields = v.Fields
			if len(api.RespFields) == 0 {
				api.RespFields = []common.ApiField{*emptyAPIField()}
			}
		}
		if v.IsReq() {
			if v.Desc != "" {
				api.ActionDesc = v.Desc
			}
			api.StructInName = v.Name
			api.ReqFields = v.Fields
			if len(api.ReqFields) == 0 {
				api.ReqFields = []common.ApiField{*emptyAPIField()}
			}
		}
		api.ActionID = actionID
		maker.collectCustomTypes(api, v.Fields)
		maker.allAPI[actionID] = *api
	}
	return maker.allAPI
}

func (maker *ApiMaker) collectCustomTypes(api *SingleAPI, fields []common.ApiField) {
	for _, f := range fields {
		_, ok := api.CustomTypes[f.GetKey()]
		s, isStruct := maker.allStruct[f.GetKey()]
		if !ok && isStruct {
			api.CustomTypes[f.GetKey()] = s
			maker.collectCustomTypes(api, s.Fields)
		}
	}
}

//格式化一条api
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

func (maker ApiMaker) formatCustomTypes(allStruct map[string]*FastStructType) *bytes.Buffer {
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

func emptyStruct() *FastStructType {
	st := new(FastStructType)
	st.Name = "无"
	st.PkgPath = ""
	st.Fields = []common.ApiField{*emptyAPIField()}
	return st
}

func emptyAPIField() *common.ApiField {
	field := new(common.ApiField)
	field.PkgPath = "无"
	field.Name = "无"
	field.TypeName = "无"
	field.Desc = "无"
	return field
}
