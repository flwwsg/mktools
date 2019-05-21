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

const apiStructKey = "github.com/funny/fastapi.APIs"
const apiFuncName = "APIs"

type FastField struct {
	common.ApiField
}

type FastStructType struct {
	common.StructType
	//所属 api 的模块
	APIPkg string
	isReq  bool
	isResp bool
	//isProvider bool // 是否实现了 fastapi.Provider 接口
}

type apiFunc struct {
	//接收函数
	recv *common.NewType
	//请求
	req map[string]*common.NewType
	//响应
	resp map[string]*common.NewType
}

//包下的结构体
type FastPkgStructs struct {
	//文件路径
	pkgPath    string
	info       *types.Info
	allStructs map[string]*FastStructType
	scope      *types.Scope
	fset       *token.FileSet
	//api 接口的struct
	api apiFunc
}

func (st FastStructType) IsReq() bool {
	return st.isReq
}

func (st FastStructType) IsResp() bool {
	return st.isResp
}

//不需要标签
func (api FastField) IsValidTag(t string) bool {
	return false
}

func NewPkgStructs(srcPath string) FastPkgStructs {
	return FastPkgStructs{pkgPath: srcPath, allStructs: make(map[string]*FastStructType), fset: token.NewFileSet()}
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
			}
		case *ast.FuncDecl:
			//collecting function
			// 查找函数类似
			//	func (adv *Adventure) APIs() fastapi.APIs {
			//	return fastapi.APIs{
			//	0: {AdventureInfoIn{}, AdventureInfoOut{}},
			//	1: {StartAdventureIn{}, StartAdventureOut{}},
			//}
			//}
			if x.Type.Results == nil || len(x.Type.Params.List) != 0 || len(x.Type.Results.List) != 1 || x.Name.String() != apiFuncName || ps.checkTypes(x.Type.Results.List[0].Type).Key != apiStructKey {
				//没有返回，非空请求参数，多个返回值，函数名不为api，返回参数类型不为github.com/funny/fastapi.APIs
				return true
			}
			api := new(apiFunc)
			api.recv = ps.checkTypes(x.Recv.List[0].Type)
			api.req = make(map[string]*common.NewType)
			api.resp = make(map[string]*common.NewType)
			for _, v := range x.Body.List {
				switch xv := v.(type) {
				case *ast.ReturnStmt:
					//返回声明
					for _, vv := range xv.Results {
						switch vvv := vv.(type) {
						case *ast.CompositeLit:
							//key value
							for _, v := range vvv.Elts {
								//类型不对，会panic， 不需要检测
								kv := v.(*ast.KeyValueExpr)
								tk := ps.checkTypes(kv.Key)
								tv := kv.Value.(*ast.CompositeLit)
								var reqStruct *common.NewType
								var respStruct *common.NewType
								_, ok := tv.Elts[0].(*ast.Ident)
								if !ok {
									//not nil
									req := tv.Elts[0].(*ast.CompositeLit)
									reqStruct = ps.checkTypes(req.Type)
								}
								_, ok = tv.Elts[1].(*ast.Ident)
								if !ok {
									//not nil
									resp := tv.Elts[1].(*ast.CompositeLit)
									respStruct = ps.checkTypes(resp.Type)
								}
								api.req[tk.Value] = reqStruct
								api.resp[tk.Value] = respStruct
							}
							ps.api = *api
							return true
						default:
							panic("unsupported type")
						}
					}
				default:
					panic("unsupported type")
				}

			}
		}
		return true
	}
	ast.Inspect(f, findStruct)
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

//检查类型
func (ps *FastPkgStructs) checkTypes(typeToCheck ast.Expr) *common.NewType {
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
		return newType
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
		if elemType.PkgPath != "" {
			//自定义结构体
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
		//引用类型
		return ps.checkTypes(t.X)
	case *ast.BasicLit:
		newType := new(common.NewType)
		newType.TypeName = t.Kind.String()
		newType.PkgPath = ""
		newType.Key = newType.TypeName
		newType.Value = t.Value
		return newType
	case *ast.StructType:
		//t := map[string]struct {}
		newType := new(common.NewType)
		newType.TypeName = "struct"
		newType.PkgPath = ""
		newType.Key = ""
		newType.Value = "struct{}"
		return newType
	case *ast.FuncType:
		return nil
	default:
		t = typeToCheck.(*ast.Ident)
		panic(fmt.Errorf("需要新增类型 %v", t))
	}
}
