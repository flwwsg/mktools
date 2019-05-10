/*
Author: Minsi Ruan
Data: 2018/6/25 10:01
*/

package do

import (
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

type field struct {
	Name      string
	FieldType string
	Required  bool
	Desc      string
	req       bool
}

type api struct {
	FileName string
	ActName  string
	Desc     string
	Fields   []*field
	ActionID string
	ApiType  string
}

type Act struct {
	ApiName string
	ApiType string
	ApiList []*api
}

//UnmarshalAPI get api from json
func UnmarshalAPI(b []byte, a *Act) {
	x := make(map[string]interface{})
	err := json.Unmarshal(b, &x)
	if err != nil {
		panic(err)
	}
	apiType := x["apiType"].(string)
	delete(x, "apiName")
	delete(x, "apiType")
	for k, v := range x {
		//get api
		api := new(api)
		api.ActionID = k
		api.ApiType = apiType
		//get field
		for kk, vv := range v.(map[string]interface{}) {
			switch kk {
			case "desc":
				api.Desc = vv.(string)
			case "actName":
				api.ActName = strings.Title(vv.(string))
			case "req", "resp":
				for _, req := range vv.([]interface{}) {
					field := new(field)
					r := req.([]interface{})
					field.Name = strings.Title(r[0].(string))
					field.FieldType = r[1].(string)
					if r[2].(string) == "required" {
						field.Required = true
					}
					field.Desc = r[3].(string)
					if kk == "req" {
						field.req = true
					}
					api.Fields = append(api.Fields, field)
				}
			case "fileName":
				api.FileName = vv.(string)
			}

		}
		a.ApiList = append(a.ApiList, api)
	}
}

func (act *Act) PackageName() string {
	return act.ApiName + "act"
}
func (act *Act) ImportName() string {
	return "_"
}

func (act *Act) ImportPath() string {
	return fmt.Sprintf("arthur/app/actions/%s/%s", act.ApiType, act.PackageName())
}

func (a *api) Text(pkgName string) string {
	//generating request struct
	pkgText := a.packageText(pkgName)
	req := a.structText(true)
	resp := a.structText(false)
	fText := a.funcText()
	iText := a.initText(req)
	return strings.Join([]string{pkgText, req, resp, fText, iText}, "\n")
}
func (a *api) GoFileName() string {
	return a.ActionID + "_" + a.FileName + ".go"

}

func (a *api) ReqName() string {
	return a.ActName + "Params"
}
func (a *api) RespName() string {
	return a.ActName + "Resp"
}

func (a *api) packageText(pkgName string) string {
	t := `
//%s 
//created: %s
//author: wdj
package %s
import (
	"arthur/app/actions/%s"

	"gitlab.dianchu.cc/goutil/dcapi.v2"
)
`
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		panic(err)
	}
	dt := time.Now()
	return fmt.Sprintf(t, pkgName, dt.In(loc).Format("2006-01-02 15:04:05"), pkgName, a.ApiType)
}

//generating struct text
func (a *api) structText(req bool) string {
	template := `
//%s %s
type %s struct {
%s
}`
	fieldText := make([]string, 0)
	structName := ""
	for _, field := range a.Fields {
		if field.req == req {
			fieldText = append(fieldText, field.Text())
		}
	}
	if len(fieldText) == 0 {
		return ""
	}
	if req {
		structName = a.ReqName()
	} else {
		structName = a.RespName()
	}

	return fmt.Sprintf(template, structName, a.Desc, structName, strings.Join(fieldText, "\n"))
}

func (a *api) funcText() string {
	t := `
func %s(ctx %s.Context) dcapi.Response{
	return dcapi.Success(nil)
}
`
	return fmt.Sprintf(t, a.ActName, a.ApiType)
}

func (a *api) initText(params string) string {
	t := `
func init(){
	%s.RegisterAct(%s, %s, %s)
}
`
	p := "nil"
	if params != "" {
		p = a.ReqName() + "{}"
	}
	return fmt.Sprintf(t, a.ApiType, a.ActionID, p, a.ActName)
}

func (f *field) Text() string {
	t := `
//%s %s
%s %s %s`
	rTag := ""
	if f.Required {
		rTag = "`valid:\"required\"`"
	}
	return fmt.Sprintf(t, f.Name, f.Desc, f.Name, f.FieldType, rTag)
}

func RegisterAct(regPath string, act Act) {
	//register import  path in load.go
	fs := token.NewFileSet()
	src, err := ioutil.ReadFile(regPath)
	if err != nil {
		log.Fatal(err)
	}
	f, err := parser.ParseFile(fs, regPath, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	importName := act.ImportName()
	importPath := act.ImportPath()
	exists := false
	strSrc := string(src)
	for _, i := range f.Imports {
		//drop "" in path
		if i.Path.Value[1:len(i.Path.Value)-1] == importPath && i.Name.String() == importName {
			exists = true
			break
		}
	}
	// need not add import
	if exists {
		return
	}
	var source string
	if len(f.Imports) == 0 {
		// first import
		source = strSrc[:f.Name.End()-1] + "\nimport " + importName + importPath + strSrc[f.Name.End()-1:]
	} else {
		last := f.Imports[len(f.Imports)-1]
		source = strSrc[:last.End()-1] + "\n\t" + importName + " \"" + importPath + "\"" + strSrc[last.End()-1:]
	}
	// backup file ignore error
	os.Rename(regPath, regPath+".bak")
	err = ioutil.WriteFile(regPath, []byte(source), 777)
	if err != nil {
		log.Fatal(err)
	}

}
