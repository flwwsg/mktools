package do

//采集真有趣接口, 生成api文档

import (
	"fmt"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"log"
	"mktools/common"
	"strings"

	"go/ast"
)

//var provider fastapi.Provider

type FastApiField struct {
	common.ApiField
}

type FastStructType struct {
	common.StructType
	//isProvider bool // 是否实现了 fastapi.Provider 接口
}

//包下的结构体
type FastPkgStructs struct {
	//文件路径
	pkgPath    string
	info       *types.Info
	allStructs map[string]*FastStructType
	scope      *types.Scope
	fset       *token.FileSet
}

//不需要标签
func (api FastApiField) IsValidTag(t string) bool {
	return false
}

func NewPkgStructs(srcPath string) FastPkgStructs {
	return FastPkgStructs{pkgPath: srcPath, allStructs: make(map[string]*FastStructType), fset: token.NewFileSet()}
}

func (ps *FastPkgStructs) GetRequiredStruct() {
	//var allNamed []*types.Named
	//for _, name := range ps.scope.Names() {
	//	if obj, ok := ps.scope.Lookup(name).(*types.TypeName); ok {
	//		allNamed = append(allNamed, obj.Type().(*types.Named))
	//	}
	//}
	//for _, T := range allNamed {
	//	for i := 0; i < T.NumMethods(); i++ {
	//		m := T.Method(i)
	//		println(m.String())
	//		println(m.FullName())
	//		println(m.Name())
	//		println(m.Pkg().Name())
	//	}
	//}
}

//Parse 使用go/types收集
func (ps *FastPkgStructs) Parse() {
	fullPath := common.FullPackagePath(ps.pkgPath)
	files := common.ListDir(fullPath, true, false)
	allFiles := make([]*ast.File, 0)

	for i := range files {
		fileName := files[i]
		if !strings.HasSuffix(fileName, "go") {
			//非go文件
			return
		}
		f, err := parser.ParseFile(ps.fset, fileName, nil, parser.ParseComments)
		if err != nil {
			log.Fatal(err)
		}
		allFiles = append(allFiles, f)
	}
	info := types.Info{Types: make(map[ast.Expr]types.TypeAndValue), Defs: make(map[*ast.Ident]types.Object),
		Uses: make(map[*ast.Ident]types.Object), Selections: make(map[*ast.SelectorExpr]*types.Selection)}
	typeConf := types.Config{Importer: importer.ForCompiler(token.NewFileSet(), "source", nil)}

	pkg, err := typeConf.Check(ps.pkgPath, ps.fset, allFiles, &info)
	if err != nil {
		log.Fatal(err) // type error
	}
	ps.info = &info
	ps.scope = pkg.Scope()
	for i := range allFiles {
		f := allFiles[i]
		filePath := files[i]
		ps.parseByFile(filePath, f)
	}
}

func (ps *FastPkgStructs) parseByFile(filePath string, f ast.Node) {
	if !strings.HasSuffix(filePath, "go") {
		//非go文件
		return
	}
	var findStruct = func(n ast.Node) bool {
		var structName string
		var t ast.Expr
		var structDec string
		// get type specification
		switch x := n.(type) {
		case *ast.GenDecl:
			if len(x.Specs) != 1 {
				return true
			}
			structDec = x.Doc.Text()
			switch xs := x.Specs[0].(type) {
			case *ast.TypeSpec:
				structName = xs.Name.Name
				t = xs.Type
				x, ok := t.(*ast.StructType)
				if !ok {
					return true
				}
				s := new(FastStructType)
				s.Name = structName
				s.PkgPath = ps.pkgPath
				s.Fields = ps.genField(x, filePath)
				s.SetDesc(structDec)
				s.PkgPath = ps.pkgPath
				//路径+类型名
				key := ps.pkgPath + "." + s.Name
				ps.allStructs[key] = s
				// TODO 查找实现 api 的接口
			}
		}
		return true
	}
	ast.Inspect(f, findStruct)
	//for i, v := range ps.allStructs {
	//	fmt.Printf("%v, %v\n", i, v)
	//}
}

func (ps *FastPkgStructs) genField(node *ast.StructType, srcPath string) []common.ApiField {
	field := make([]common.ApiField, 0)
	for i := range node.Fields.List {
		f := node.Fields.List[i]
		if !ast.IsExported(f.Names[0].Name) {
			continue
		}
		newField := new(common.ApiField)
		newField.Name = f.Names[0].Name
		if f.Comment.Text() != "" {
			newField.SetDesc(f.Comment.Text())
		} else {
			newField.SetDesc(f.Doc.Text())
		}
		nt := ps.checkTypes(f.Type)
		newField.TypeName = nt.TypeName
		newField.PkgPath = nt.PkgPath
		newField.SetKey(nt.Key)
		field = append(field, *newField)
	}
	return field
}

//检查类型
func (ps *FastPkgStructs) checkTypes(typeToCheck ast.Expr) common.NewType {
	switch t := typeToCheck.(type) {
	case *ast.Ident:
		obj := ps.info.ObjectOf(t)
		newType := new(common.NewType)
		if obj == nil {
			//基本类型
			newType.TypeName = t.Name
		} else {
			newType.TypeName = obj.Name()
			if obj.Pkg() != nil {
				newType.PkgPath = obj.Pkg().Path()
				newType.Key = newType.PkgPath + "." + newType.TypeName
			}
		}
		return *newType
	case *ast.SelectorExpr:
		//其它包里面的类型， 如x.t
		return ps.checkTypes(t.Sel)
	case *ast.ArrayType:
		//列表
		elemType := ps.checkTypes(t.Elt)
		newType := new(common.NewType)
		if t.Len == nil {
			//slice
			newType.TypeName = "[]" + elemType.TypeName
		} else {
			v := t.Len.(*ast.BasicLit).Value
			newType.TypeName = fmt.Sprintf("[%s]"+elemType.TypeName, v)
		}
		newType.PkgPath = elemType.PkgPath
		newType.Key = elemType.PkgPath + "." + elemType.TypeName
		return *newType
	case *ast.MapType:
		k := ps.checkTypes(t.Key)
		v := ps.checkTypes(t.Value)
		newType := new(common.NewType)
		newType.TypeName = fmt.Sprintf("map[%s]%s", k.TypeName, v.TypeName)
		newType.PkgPath = v.PkgPath
		newType.Key = v.Key
		return *newType
	case *ast.StarExpr:
		//引用类型
		return ps.checkTypes(t.X)
	default:
		t = typeToCheck.(*ast.Ident)
		panic(fmt.Errorf("%v", t))

	}
}

//判断接口是否实现
func init() {
	fullPath := common.FullPackagePath("github.com/funny/fastapi")
	files := common.ListDir(fullPath, true, false)
	allFiles := make([]*ast.File, 0)
	fset := token.NewFileSet()
	for i := range files {
		fileName := files[i]
		if !strings.HasSuffix(fileName, "go") {
			//非go文件
			continue
		}
		f, err := parser.ParseFile(fset, fileName, nil, parser.ParseComments)
		if err != nil {
			log.Fatal(err)
		}
		allFiles = append(allFiles, f)
	}

	//implemented types.Interface
	typeConf := types.Config{Importer: importer.Default()}
	pkg, _ := typeConf.Check("", fset, allFiles, nil)
	var allNamed []*types.Named
	for _, name := range pkg.Scope().Names() {
		if obj, ok := pkg.Scope().Lookup(name).(*types.TypeName); ok {
			allNamed = append(allNamed, obj.Type().(*types.Named))
		}
	}

	// Test assignability of all distinct pairs of
	// named types (T, U) where U is an interface.
	for _, T := range allNamed {
		println(T.String())
		//for _, U := range allNamed {
		//	if T == U || !types.IsInterface(U) {
		//		continue
		//	}
		//	if types.AssignableTo(T, U) {
		//		fmt.Printf("%s satisfies %s\n", T, U)
		//	} else if !types.IsInterface(T) &&
		//		types.AssignableTo(types.NewPointer(T), U) {
		//		fmt.Printf("%s satisfies %s\n", types.NewPointer(T), U)
		//	}
		//}
	}
}
