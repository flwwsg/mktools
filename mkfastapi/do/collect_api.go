package do

// 采集真有趣接口, 生成api文档

import (
	"fmt"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"log"
	"mktools/common"
	"strings"

	"gitee.com/flwwsg/utils-go/files"

	"go/ast"
)

const apiStructKey = "github.com/funny/fastapi.APIs"
const apiFuncName = "APIs"

type FastField struct {
	common.ApiField
}

type FastStructType struct {
	common.StructType
	// 所属 api 的模块
	APIPkg string
	isReq  bool
	isResp bool
	// isProvider bool // 是否实现了 fastapi.Provider 接口
}

type apiFunc struct {
	// 接收函数
	recv *common.NewType
	// 请求
	req map[string]*common.NewType
	// 响应
	resp map[string]*common.NewType
}

// 包下的结构体
type FastPkgStructs struct {
	// 文件路径
	pkgPath    string
	info       *types.Info
	allStructs map[string]*FastStructType
	scope      *types.Scope
	fset       *token.FileSet
	// api 接口的struct
	api      apiFunc
	buildTag string
}

func (st FastStructType) IsReq() bool {
	return st.isReq
}

func (st FastStructType) IsResp() bool {
	return st.isResp
}

// 不需要标签
func (api FastField) IsValidTag(t string) bool {
	return false
}

func NewPkgStructs(srcPath string, tag string) FastPkgStructs {
	return FastPkgStructs{pkgPath: srcPath, allStructs: make(map[string]*FastStructType), fset: token.NewFileSet(), buildTag: tag}
}

// Parse 使用go/types收集
func (ps *FastPkgStructs) Parse() {
	fullPath := files.FullPackagePath(ps.pkgPath)
	listFiles := files.ListDir(fullPath, true, false)
	allFiles := make([]*ast.File, 0)

	for i := range listFiles {
		fileName := listFiles[i]
		if !strings.HasSuffix(fileName, "go") || strings.HasSuffix(fileName, "_test.go") || strings.HasSuffix(fileName, "fastapi.go") || strings.HasSuffix(fileName, "fastbin.go") {
			// 非go文件， 测试文件
			continue
		}
		f, err := parser.ParseFile(ps.fset, fileName, nil, parser.ParseComments)
		if err != nil {
			log.Fatal(err)
		}
		if !ps.checkBuildTag(f.Comments) {
			continue
		}
		allFiles = append(allFiles, f)
	}
	info := types.Info{Types: make(map[ast.Expr]types.TypeAndValue), Defs: make(map[*ast.Ident]types.Object),
		Uses: make(map[*ast.Ident]types.Object), Selections: make(map[*ast.SelectorExpr]*types.Selection)}
	typeConf := types.Config{Importer: importer.ForCompiler(token.NewFileSet(), "source", nil)}
	// 需要预编译
	// typeConf := types.Config{Importer: importer.Default()}
	pkg, err := typeConf.Check(ps.pkgPath, ps.fset, allFiles, &info)
	if err != nil {
		log.Fatal(err) // type error
	}
	ps.info = &info
	ps.scope = pkg.Scope()
	for i := range allFiles {
		f := allFiles[i]
		filePath := listFiles[i]
		ps.parseByFile(filePath, f)
	}
}

// 判断编译标签是否需要编译, 仅支持 and 操作
// A build constraint is evaluated as the OR of space-separated options;
// each option evaluates as the AND of its comma-separated terms;
// and each term is an alphanumeric word or, preceded by !, its negation.
func (ps FastPkgStructs) checkBuildTag(comments []*ast.CommentGroup) bool {
	// 没有标签
	if len(comments) < 1 || ps.buildTag == "" {
		return true
	}
	c := comments[0].Text()
	c = strings.TrimSpace(c)
	tmp := strings.Split(c, " ")
	if tmp[0] != "+build" {
		return true
	}
	for _, t := range tmp[1:] {
		tags := strings.Split(t, ",")
		for _, tt := range tags {
			if tt == ps.buildTag {
				return true
			}
			// 检查 !tag
			if string(tt[0]) == "!" && string(tt[1:]) != ps.buildTag {
				return true
			}
		}

	}
	return false
}

func (ps *FastPkgStructs) parseByFile(filePath string, f ast.Node) {
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
				// 路径+类型名
				key := ps.pkgPath + "." + s.Name
				ps.allStructs[key] = s
			}
		case *ast.FuncDecl:
			// collecting function
			// 查找函数类似
			// 	func (adv *Adventure) APIs() fastapi.APIs {
			// 	return fastapi.APIs{
			// 	0: {AdventureInfoIn{}, AdventureInfoOut{}},
			// 	1: {StartAdventureIn{}, StartAdventureOut{}},
			// }
			// }
			if x.Type.Results == nil || len(x.Type.Params.List) != 0 || len(x.Type.Results.List) != 1 || x.Name.String() != apiFuncName || ps.checkTypes(x.Type.Results.List[0].Type).Key != apiStructKey {
				// 没有返回，非空请求参数，多个返回值，函数名不为api，返回参数类型不为github.com/funny/fastapi.APIs
				return true
			}
			api := new(apiFunc)
			api.recv = ps.checkTypes(x.Recv.List[0].Type)
			api.req = make(map[string]*common.NewType)
			api.resp = make(map[string]*common.NewType)
			for _, v := range x.Body.List {
				switch xv := v.(type) {
				case *ast.ReturnStmt:
					// 返回声明
					ps.checkResults(api, xv.Results[0])
					return true
				default:
					// fmt.Printf("unsupported type %v", xv)
					// return true
					continue
				}

			}
		}
		return true
	}
	ast.Inspect(f, findStruct)
}

func (ps *FastPkgStructs) checkResults(api *apiFunc, vv ast.Expr) {
	switch vvv := vv.(type) {
	case *ast.CompositeLit:
		// key value
		for _, v := range vvv.Elts {
			// 类型不对，会panic， 不需要检测
			kv := v.(*ast.KeyValueExpr)
			tk := ps.checkTypes(kv.Key)
			tv := kv.Value.(*ast.CompositeLit)
			var reqStruct *common.NewType
			var respStruct *common.NewType
			_, ok := tv.Elts[0].(*ast.Ident)
			if !ok {
				// not nil
				req := tv.Elts[0].(*ast.CompositeLit)
				reqStruct = ps.checkTypes(req.Type)
			}
			_, ok = tv.Elts[1].(*ast.Ident)
			if !ok {
				// not nil
				resp := tv.Elts[1].(*ast.CompositeLit)
				respStruct = ps.checkTypes(resp.Type)
			}
			api.req[tk.Value] = reqStruct
			api.resp[tk.Value] = respStruct
		}
		ps.api = *api
	case *ast.Ident:
		// api := fastapi.APIs{}
		t := vvv.Obj.Decl.(*ast.AssignStmt)
		tr := t.Rhs[0].(ast.Expr)
		ps.checkResults(api, tr)
	default:
		_ = vv.(*ast.CompositeLit)

	}

}
func (ps *FastPkgStructs) genField(node *ast.StructType, srcPath string) []common.ApiField {
	var field []common.ApiField
	for i := range node.Fields.List {
		f := node.Fields.List[i]
		if f.Names == nil || !ast.IsExported(f.Names[0].Name) {
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
		if nt == nil {
			continue
		}
		newField.TypeName = nt.TypeName
		newField.PkgPath = nt.PkgPath
		newField.SetKey(nt.Key)
		field = append(field, *newField)
	}
	return field
}

// 检查类型
func (ps *FastPkgStructs) checkTypes(typeToCheck ast.Expr) *common.NewType {
	switch t := typeToCheck.(type) {
	case *ast.Ident:
		obj := ps.info.ObjectOf(t)
		newType := new(common.NewType)
		if obj == nil {
			// 基本类型
			newType.TypeName = t.Name
		} else {
			newType.TypeName = obj.Name()
			if obj.Pkg() != nil {
				newType.PkgPath = obj.Pkg().Path()
				newType.Key = newType.PkgPath + "." + newType.TypeName
			}
		}
		return newType
	case *ast.SelectorExpr:
		// 其它包里面的类型， 如x.t
		return ps.checkTypes(t.Sel)
	case *ast.ArrayType:
		// 列表
		elemType := ps.checkTypes(t.Elt)
		newType := new(common.NewType)
		if t.Len == nil {
			// slice
			newType.TypeName = "[]" + elemType.TypeName
		} else {
			v := t.Len.(*ast.BasicLit).Value
			newType.TypeName = fmt.Sprintf("[%s]"+elemType.TypeName, v)
		}
		newType.PkgPath = elemType.PkgPath
		if elemType.PkgPath != "" {
			// 自定义结构体
			newType.Key = elemType.PkgPath + "." + elemType.TypeName
		}
		return newType
	case *ast.MapType:
		k := ps.checkTypes(t.Key)
		v := ps.checkTypes(t.Value)
		newType := new(common.NewType)
		newType.TypeName = fmt.Sprintf("map[%s]%s", k.TypeName, v.TypeName)
		newType.PkgPath = v.PkgPath
		newType.Key = v.Key
		return newType
	case *ast.StarExpr:
		// 引用类型
		return ps.checkTypes(t.X)
	case *ast.BasicLit:
		newType := new(common.NewType)
		newType.TypeName = t.Kind.String()
		newType.PkgPath = ""
		newType.Key = newType.TypeName
		newType.Value = t.Value
		return newType
	case *ast.StructType:
		// t := map[string]struct {}
		newType := new(common.NewType)
		newType.TypeName = "struct"
		newType.PkgPath = ""
		newType.Key = ""
		newType.Value = "struct{}"
		return newType
	case *ast.FuncType:
		return nil
	case *ast.InterfaceType:
		// interface
		newType := new(common.NewType)
		newType.TypeName = "interface"
		newType.PkgPath = ""
		newType.Key = ""
		newType.Value = "interface{}"
		return newType
	default:
		t = typeToCheck.(*ast.Ident)
		panic(fmt.Errorf("需要新增类型 %v", t))
	}
}
