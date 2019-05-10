package do

import (
	"fmt"
	"go/ast"
	"strconv"
	"strings"
)

//types

type buitinType struct {
	name string
}

type arrayType struct {
	name string
	len  int
}

type structType struct {
	name string
}

//Text() string //to client
////Server() string // server view
//Type() int

func (b buitinType) Text() string {
	return b.name
}

func (b buitinType) Type() int {
	return tBuiltin
}

func (b buitinType) TypeName() string {
	return b.name
}
func (a arrayType) Text() string {
	if a.len == 0 {
		//slice
		return "[]" + a.name
	} else {
		return fmt.Sprintf("[%d]%s", a.len, a.name)
	}
}

func (a arrayType) Type() int {
	return tSlice
}

func (a arrayType) TypeName() string {
	return a.name
}

func (s structType) Text() string {
	return s.name
}
func (s structType) Type() int {
	return tStruct
}

func (s structType) TypeName() string {
	return s.name
}

//
func CheckPkgType(e ast.Expr, imps map[string]string, source string) (vt ValueType, pkgPath string) {
	splitType := func(t string) (string, string) {
		l := strings.Split(t, ".")
		if len(l) == 1 {
			//builtin
			return "", l[0]
		}
		//package path, type
		return imps[l[0]], l[1]
	}

	switch x := e.(type) {
	case *ast.Ident:
		return buitinType{x.Name}, ""
	case *ast.ArrayType:
		ts := source[x.Elt.Pos()-1 : x.Elt.End()-1]
		p, tn := splitType(ts)
		if x.Len == nil {
			return arrayType{tn, 0}, p
		} else {
			l := source[x.Len.Pos()-1 : x.Len.End()-1]
			n, _ := strconv.Atoi(l)
			return arrayType{tn, n}, p
		}
	//
	//case *ast.SelectorExpr:
	//	fmt.Printf("select :%v, %v\n", x.X, x.Sel)

	//	p, tn := splitType(x.Elem().String())
	//	//fmt.Printf("%v, %v\n", tn, p)
	//	if x.Len() == 0 {
	//		return arrayType{tn, 0}, p
	//	} else {
	//		return arrayType{tn, int(x.Len())}, p
	//	}
	//case *types.Struct:
	//	fmt.Printf("struct %v\n", x)
	default:
		xx := e.(*ast.Ident)
		panic(fmt.Sprintf("not supported type %v", xx))
	}

	return nil, ""
}
