//do
//created: 2018/8/31
//author: wdj

package do

import (
	"container/list"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"io/ioutil"
	"log"
	"mktools/common"
	"os"
	"path/filepath"
	"strings"
)

const (
	tSlice = iota
	tMap
	tStruct
	tStructField
	tBuiltin
	//tUnknown
)

//CollectedType struct for collecting struct, slice, map, field of struct
type CollectedType interface {
	Type() int
}

//ValueType like []int, [3]int
type ValueType interface {
	Text() string //to client
	//Server() string // server view
	Type() int
	TypeName() string
}

type APINode struct {
	FieldName string
	Desc      string //comments description
	//TypeName  string
	ValueType ValueType
	//NodeType  int //value type slice and so on
	PkgPath string
}

//node
type StructNode struct {
	*APINode
	Fields []*StructFieldNode
}

type StructFieldNode struct {
	*APINode
	StructName string
}

type SliceNode struct {
	*APINode
}

type MapNode struct {
	*APINode
}

type BuiltinNode struct {
	*APINode
}

//collect api
type TypeTree struct {
	Value    *APINode
	Children []*TypeTree
}

type queue struct {
	l *list.List
}

func NewStructNode(n *APINode) StructNode {
	//n.NodeType = tStruct
	return StructNode{n, make([]*StructFieldNode, 0)}
}

func NewSliceNode(n *APINode) SliceNode {
	//n.NodeType = tSlice
	return SliceNode{n}
}

func NewMapNode(n *APINode) MapNode {
	//n.NodeType = tMap
	return MapNode{n}
}

func NewStructField(n *APINode, structName string) StructFieldNode {
	//n.NodeType = tStructField
	return StructFieldNode{n, structName}
}
func (t *TypeTree) AddNode(n *TypeTree) {
	t.Children = append(t.Children, n)
}

func (q queue) push(v interface{}) {
	q.l.PushBack(v)
}

func (q queue) pop() interface{} {
	if v := q.l.Front(); v != nil {
		q.l.Remove(v)
		return v.Value
	}
	return nil
}

func (sn StructNode) Type() int {
	return sn.ValueType.Type()
}

func (sn StructNode) AddField(n *StructFieldNode) {
	sn.Fields = append(sn.Fields, n)
}

func (sn SliceNode) Type() int {
	return sn.ValueType.Type()
}

func (mn MapNode) Type() int {
	return mn.ValueType.Type()
}
func (sfn StructFieldNode) Type() int {
	return sfn.ValueType.Type()
}

func (b BuiltinNode) Type() int {

	return b.ValueType.Type()
}

func LevelOrder(t *TypeTree) {
	//水平遍历
	last := t
	nextLast := new(TypeTree)
	q := queue{l: list.New()}
	q.push(t)
	lv := 1
	fmt.Println(fmt.Sprintf("\n-----level %d-----", lv))
	for q.l.Len() > 0 {
		node := q.pop().(*TypeTree)
		for _, c := range node.Children {
			q.push(c)
		}
		fmt.Printf("%q\n", node.Value)
		if len(node.Children) != 0 {
			nextLast = node.Children[len(node.Children)-1]
		}

		if last == node && q.l.Len() > 0 {
			last = nextLast
			lv++
			fmt.Println()
			fmt.Println(fmt.Sprintf("-----level %d-----", lv))
		}
	}
}

//GetDeclType get type name and type
func GetDeclType(pkgPath string) map[string]TypeTree {
	astFiles := make([]*ast.File, 0)
	src := make([]string, 0)
	files := ListPackageDir(pkgPath)
	fs := token.NewFileSet()
	imps := map[string]string{}
	//pkgName := ""
	defs := map[string]CollectedType{}
	for _, f := range files {
		if !strings.HasSuffix(f, "go") {
			continue
		}
		pf, err := parser.ParseFile(fs, f, nil, parser.ParseComments)
		if err != nil {
			log.Fatal(err)
		}
		//pkgName = pf.Name.String()
		for _, imp := range pf.Imports {
			p := imp.Path.Value[1 : len(imp.Path.Value)-1]
			if imp.Name == nil {
				name := GetPackageName(p)
				imps[name] = p
			} else {
				imps[imp.Name.String()] = p
			}
		}
		astFiles = append(astFiles, pf)
		s, _ := ioutil.ReadFile(f)
		ss := string(s)
		src = append(src, ss)
		var findDecl = func(n ast.Node) bool {
			// get type specification
			switch x := n.(type) {
			case *ast.GenDecl:
				for _, spec := range x.Specs {
					switch s := spec.(type) {
					case *ast.ImportSpec:
						//ignore import
					case *ast.ValueSpec:
					case *ast.TypeSpec:
						switch t := s.Type.(type) {
						case *ast.StructType:
							n := new(APINode)
							n.FieldName = s.Name.Name
							n.ValueType = structType{s.Name.Name}
							n.Desc = x.Doc.Text()
							sn := NewStructNode(n)
							for _, f := range t.Fields.List {
								n := new(APINode)
								if f.Comment.Text() != "" {
									n.Desc = f.Comment.Text()
								} else {
									n.Desc = f.Doc.Text()
								}
								n.FieldName = f.Names[0].Name
								vt, p := CheckPkgType(f.Type, imps, ss)
								n.ValueType = vt
								n.PkgPath = p
								snf := NewStructField(n, s.Name.Name)
								sn.AddField(&snf)
							}
							defs[pkgPath+"."+s.Name.Name] = sn
						case *ast.Ident:
							n := new(APINode)
							n.ValueType = buitinType{t.Name}
							n.Desc = x.Doc.Text()
							n.FieldName = s.Name.Name
							bn := BuiltinNode{n}
							defs[pkgPath+"."+s.Name.Name] = bn
						default:
							xs := s.Type.(*ast.StructType)
							panic(xs)
						}
					default:
						view(s, "unknown node")

					}
				}

			}
			return true
		}
		ast.Inspect(pf, findDecl)
	}
	//conf := types.Config{Importer: importer.For("source", nil)}
	//info := &types.Info{Types: make(map[ast.Expr]types.TypeAndValue), Defs: make(map[*ast.Ident]types.Object)}
	//if _, err := conf.Check(pkgName, fs, astFiles, info); err != nil {
	//	log.Fatal(err) // type error
	//}

	//for _, f := range astFiles {
	//
	//}
	view(defs, "def")

	return nil
}

//func CollectTypes(pkgPath string) {
//	astFiles := make([]*ast.File, 0)
//	files := ListPackageDir(pkgPath)
//	fs := token.NewFileSet()
//	imps := map[string]string{}
//	pkgName := ""
//	for _, f := range files {
//		if !strings.HasSuffix(f, "go") {
//			continue
//		}
//		pf, err := parser.ParseFile(fs, f, nil, parser.ParseComments)
//		if err != nil {
//			log.Fatal(err)
//		}
//		//pkgName = pf.Name.String()
//		for _, imp := range pf.Imports {
//			p := imp.Path.Value[1 : len(imp.Path.Value)-1]
//			if imp.Name == nil {
//				name := GetPackageName(p)
//				imps[name] = p
//			} else {
//				imps[imp.Name.String()] = p
//			}
//		}
//		astFiles = append(astFiles, pf)
//	}
//	conf := types.Config{Importer: importer.For("source", nil)}
//	info := &types.Info{Types: make(map[ast.Expr]types.TypeAndValue), Defs: make(map[*ast.Ident]types.Object)}
//	if _, err := conf.Check(pkgName, fs, astFiles, info); err != nil {
//		log.Fatal(err) // type error
//	}
//	//filter api
//	apiList := make([]*TypeTree, 0)
//	for id, t := range info.Defs {
//		if t == nil {
//			continue
//		}
//		if isSpecifiedStruct(t) {
//			//get api struct
//			api := new(TypeTree)
//			node := new(APINode)
//			node.FieldName = id.Name
//			node.NodeType = tStruct
//			node.ValueType = t.Type()
//			api.Value = node
//			apiList = append(apiList, api)
//		}
//	}
//	//fmt.Printf("%q\n", info.Defs)
//	//collect all current package types
//	for _, api := range apiList {
//		t := api.Value.ValueType.Underlying().(*types.Struct)
//		for i := 0; i < t.NumFields(); i++ {
//			f := t.Field(i)
//			fmt.Printf("struct field %v\n", f)
//			node := getNextNode(pkgName, f.Type(), info.Defs)
//			api.AddNode(node)
//		}
//	}
//	for _, api := range apiList {
//		LevelOrder(api)
//	}
//}

//func getNextNode(pkgName string, elem types.Type, info map[*ast.Ident]types.Object) *TypeTree {
//	thirdPkg := func(t types.Type) *TypeTree {
//		newTree := new(TypeTree)
//		node := new(APINode)
//		node.ValueType = t
//		l := strings.Split(t.String(), ".")
//		if len(l) != 2 {
//			panic(fmt.Sprintf("invalid type %q", l))
//		}
//		node.PkgPath = l[1]
//		newTree.Value = node
//		return newTree
//	}
//	for _, t := range info {
//		if t == nil || t.Type().String() != elem.String() {
//			continue
//		}
//		newTree := new(TypeTree)
//		node := new(APINode)
//		node.FieldName = t.Name()
//		node.ValueType = t.Type()
//		newTree.Value = node
//		switch x := t.Type().Underlying().(type) {
//		case *types.Slice:
//			node.NodeType = tSlice
//			nextTree := getNextNode(pkgName, x.Elem(), info)
//			if nextTree == nil {
//				//third package
//				nextTree = thirdPkg(x.Elem())
//			}
//			newTree.AddNode(nextTree)
//		case *types.Struct:
//			node.NodeType = tStruct
//			for i := 0; i < x.NumFields(); i++ {
//				f := x.Field(i)
//				nextTree := getNextNode(pkgName, f.Type(), info)
//				if nextTree == nil {
//					//third package
//					nextTree = thirdPkg(f.Type())
//				}
//				newTree.AddNode(nextTree)
//			}
//		case *types.Basic:
//			node.NodeType = tBuiltin
//		default:
//			//fmt.Printf("%v\n", x)
//			x = t.Type().Underlying().(*types.Basic)
//			fmt.Printf("%v\n", x)
//
//		}
//		return newTree
//	}
//	return nil
//}

func isSpecifiedStruct(t types.Object) bool {
	name := t.Name()
	_, ok := t.Type().Underlying().(*types.Struct)
	if !ok {
		return false
	}
	l := len(name)
	r := len("Resp")
	p := len("Params")
	if l < p || l < r {
		return false
	}
	if name[l-r:] == "Resp" {
		return true
	}
	if name[l-p:] == "Params" {
		return true
	}
	return false
}

func ListPackageDir(pkgPath string) []string {
	pkgPath = strings.Replace(pkgPath, "\\", "/", -1)
	if len(strings.Split(pkgPath, "/")) == 1 {
		//only package name
		fullPath := common.FindProjectRoot(pkgPath)
		return common.ListDir(fullPath, true, false)
	} else {
		for _, p := range common.GoPath() {
			fullPath := filepath.Join(p, "src", pkgPath)
			_, err := os.Stat(fullPath)
			if err != nil {
				continue
			}
			return common.ListDir(fullPath, true, false)
		}
		panic(fmt.Sprintf("can not find %s", pkgPath))
	}
}

func view(x interface{}, s string) {
	fmt.Printf("%s:%q\n", s, x)
}

//GetPackageName get package name with given typePath
func GetPackageName(pkgPath string) string {
	files := ListPackageDir(pkgPath)
	pkgName := ""
	fs := token.NewFileSet()
	for _, f := range files {
		if !strings.HasSuffix(f, "go") {
			continue
		}
		pf, err := parser.ParseFile(fs, f, nil, parser.PackageClauseOnly)
		if err != nil {
			continue
		}
		pkgName = pf.Name.String()
		break
	}
	return pkgName
}
