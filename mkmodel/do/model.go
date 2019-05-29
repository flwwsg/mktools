//do
//created: 2018/7/26
//author: wdj

package do

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"strings"

	common "gitee.com/flwwsg/utils-go/files"
)

const statTEMP = `
type ProfileCenter struct {
%s
}
`
const dynTEMP = `
var CharacterSet = []interface{}{
%s
}
`

type modelStruct struct {
	name string
	//path     string
	tokStart token.Pos
	tokEnd   token.Pos
	node     *ast.StructType
	pkgName  string
}

type modelSet struct {
	name     string
	tokStart token.Pos
	tokEnd   token.Pos
	node     *ast.ValueSpec
}

type modelElem struct {
	fieldName string
	pkgName   string
	modelName string
}

type elemList []*modelElem

func (el elemList) String() string {
	s := make([]string, len(el))
	for i, e := range el {
		s[i] = e.String()
	}
	return strings.Join(s, "\n")
}

func (e *modelElem) IsDyn() bool {
	//model of dyn has no field name
	return e.fieldName == ""
}

func (e *modelElem) String() string {
	if e.IsDyn() {
		return fmt.Sprintf("&%s.%s{},", e.pkgName, e.modelName)
	}
	return fmt.Sprintf("%s []%s.%s", e.fieldName, e.pkgName, e.modelName)
}

func (e *modelElem) eq(pkgName string, modelName string) bool {
	return e.pkgName == pkgName && e.modelName == modelName
}

func (m *modelStruct) IsDyn() bool {
	return m.pkgName == "dyn"
}

func (m *modelStruct) toElem() *modelElem {
	e := new(modelElem)
	e.pkgName = m.pkgName
	if !m.IsDyn() {
		e.fieldName = m.name
	}

	e.modelName = m.name
	return e
}

func (m *modelStruct) toElemList(src string) elemList {
	el := make(elemList, 0)
	for _, v := range m.node.Fields.List {
		elem := v.Type.(*ast.ArrayType).Elt.(*ast.SelectorExpr)
		e := new(modelElem)
		e.fieldName = v.Names[0].String()
		e.pkgName = src[elem.X.Pos()-1 : elem.X.End()-1]
		e.modelName = elem.Sel.String()
		el = append(el, e)
	}
	return el
}

func (s *modelSet) toElemList(src string) elemList {
	el := make(elemList, 0)
	for _, v := range s.node.Values[0].(*ast.CompositeLit).Elts {
		elem := v.(*ast.UnaryExpr).X.(*ast.CompositeLit).Type.(*ast.SelectorExpr)
		e := new(modelElem)
		e.pkgName = src[elem.X.Pos()-1 : elem.X.End()-1]
		e.modelName = elem.Sel.String()
		el = append(el, e)
	}
	return el
}

func RegisterModel(register string, toReg string) {
	fs := token.NewFileSet()
	src, _ := ioutil.ReadFile(register)
	strSrc := string(src)
	f, err := parser.ParseFile(fs, register, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	regModels := findStruct(toReg)
	dynSet, statSet := findDynStatSet(f)
	dynModel := dynSet.toElemList(strSrc)
	statModel := statSet.toElemList(strSrc)
	addDyn := false
	addStat := false
	for _, reg := range regModels {
		if reg.IsDyn() {
			if !common.FindInList(reg.toElem(), dynModel) {
				dynModel = append(dynModel, reg.toElem())
				addDyn = true
			}
		} else {
			if !common.FindInList(reg.toElem(), statModel) {
				statModel = append(statModel, reg.toElem())
				addStat = true
			}
		}
	}
	if !addStat && !addDyn {
		return
	}
	//backup file
	err = os.Rename(register, register+".bak")
	if err != nil {
		panic(err)
	}
	if addStat {
		statString := fmt.Sprintf(statTEMP, statModel.String())
		err = ioutil.WriteFile(register, []byte(strSrc[:statSet.tokStart-1]+statString+strSrc[statSet.tokEnd-1:]), 777)
	} else {
		dynString := fmt.Sprintf(dynTEMP, dynModel.String())
		err = ioutil.WriteFile(register, []byte(strSrc[:dynSet.tokStart-1]+dynString+strSrc[dynSet.tokEnd-1:]), 777)
	}
	if err != nil {
		panic(err)
	}
}

func findDynStatSet(n ast.Node) (*modelSet, *modelStruct) {
	dynSet := new(modelSet)
	statSet := new(modelStruct)
	var findSet = func(n ast.Node) bool {
		var structName string
		switch x := n.(type) {
		case *ast.GenDecl:
			if len(x.Specs) != 1 {
				return true
			}
			switch xs := x.Specs[0].(type) {
			case *ast.TypeSpec:
				structName = xs.Name.Name
				xxs, ok := xs.Type.(*ast.StructType)
				if !ok {
					return true
				}
				statSet.name = structName
				statSet.node = xxs
				statSet.tokStart = x.Pos()
				statSet.tokEnd = x.End()
			case *ast.ValueSpec:
				if xs.Names[0].Name != "CharacterSet" {
					return true
				}
				dynSet.name = "CharacterSet"
				dynSet.tokStart = x.Pos()
				dynSet.tokEnd = x.End()
				dynSet.node = xs
			}
		}
		return true
	}
	ast.Inspect(n, findSet)
	return dynSet, statSet
}

func FindStructs(srcPath string) []*modelStruct {
	allStruct := make([]*modelStruct, 0)
	for _, p := range common.ListDir(srcPath, true, false) {
		s := findStruct(p)
		//fmt.Printf("%q", s)
		allStruct = append(allStruct, s...)
	}
	return allStruct

}

func findStruct(srcPath string) []*modelStruct {
	allStruct := make([]*modelStruct, 0)
	if !strings.HasSuffix(srcPath, "go") {
		return allStruct
	}
	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, srcPath, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	var findStruct = func(n ast.Node) bool {
		var structName string
		var t ast.Expr
		// get type specification
		switch x := n.(type) {
		case *ast.GenDecl:
			if len(x.Specs) != 1 {
				return true
			}
			switch xs := x.Specs[0].(type) {
			case *ast.TypeSpec:
				structName = xs.Name.Name
				t = xs.Type
			}
			_, ok := t.(*ast.StructType)
			if !ok {
				return true
			}
			s := new(modelStruct)
			s.name = structName
			s.pkgName = f.Name.String()
			allStruct = append(allStruct, s)
		}
		return true
	}

	ast.Inspect(f, findStruct)
	return allStruct
}
