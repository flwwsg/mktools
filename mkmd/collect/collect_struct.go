package collect

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"log"
	"path/filepath"
	"regexp"
	"strings"

	"gitee.com/flwwsg/utils-go/files"
)

const TokenTag = "valid"

// ApiField 结构体字段
type ApiField struct {
	Name     string
	TypeName string
	Desc     string
	Required bool
	PkgPath  string
	key      string
}

type NewType struct {
	//显示的类型名
	TypeName string
	//包路径
	PkgPath string
	//键值
	Key string
}

// StructType 结构体
type StructType struct {
	Name     string
	ActionID string //mark action id to specify API, only request struct will be marked
	Fields   []ApiField
	Desc     string
	PkgPath  string
}

//PkgStructs package 下的所有结构体
type PkgStructs struct {
	//文件路径
	pkgPath    string
	info       *types.Info
	allStructs map[string]*StructType
	scope      *types.Scope
	fset       *token.FileSet
	//是否存在请求、响应结构体
	getReq  bool
	getResp bool
}

func NewPkgStructs(srcPath string) PkgStructs {
	return PkgStructs{pkgPath: srcPath, allStructs: make(map[string]*StructType), fset: token.NewFileSet()}
}

func (field *ApiField) SetDesc(s string) {
	desc := strings.Replace(s, field.Name, "", 1)
	desc = strings.Replace(desc, "\n", " ", -1)
	field.Desc = strings.TrimSpace(desc)
}

// IsValidTag check tag is valid or not
func (field *ApiField) IsValidTag(t string) bool {
	return !strings.Contains(t, "-")
}

// ParseTag handle tag
func (field *ApiField) ParseTag(f *ast.Field, t string) {
	// t = "valid: \"Required, xxx\""
	if !field.IsValidTag(t) {
		return
	}
	t = t[strings.Index(t, "\"")+1 : strings.LastIndex(t, "\"")]
	fields := strings.Split(t, ",")
	field.Required = false
	for _, f := range fields {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		switch f {
		case "required":
			field.Required = true
		case "optional":
		default:
			continue
		}
	}
}

func (s *StructType) SetDesc(comm string) {
	//drop struct Name
	desc := strings.Replace(comm, s.Name, "", 1)
	desc = strings.Replace(desc, "\n", " ", -1)
	s.Desc = strings.TrimSpace(desc)
}

//包含包名的类型名
func (s StructType) FullName() string {
	return s.PkgPath + "." + s.Name
}

//响应类
func (s *StructType) IsResp() bool {
	//response struct Name is like DemoLoginResp
	l := len(s.Name)
	if l < 4 {
		return false
	}
	if s.Name[l-4:] == "Resp" {
		return true
	}
	return false
}

//请求类
func (s *StructType) IsReq() bool {
	//request struct Name is like DemoLoginParams
	l := len(s.Name)
	if l < 6 {
		return false
	}
	if s.Name[l-6:] == "Params" {
		return true
	}
	return false
}

//判断结构体类型
func (s *StructType) IsTypeOf(typeName string) bool {
	index := strings.Index(typeName, s.Name)
	if index != -1 {
		return typeName[index:] == s.Name
	}
	return false
}

//Parse 使用go/types收集
func (ps *PkgStructs) Parse() {
	fullPath := files.FullPackagePath(ps.pkgPath)
	files := files.ListDir(fullPath, true, false)
	allFiles := make([]*ast.File, 0)

	for i := range files {
		fileName := files[i]
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

func (ps *PkgStructs) parseByFile(filePath string, f ast.Node) {
	if !strings.HasSuffix(filePath, "go") {
		//非go文件
		return
	}
	actionID, _ := FindActionID(filePath)
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
				s := new(StructType)
				s.Name = structName
				s.Fields = ps.genField(x, filePath)
				s.SetDesc(structDec)
				s.PkgPath = ps.pkgPath
				s.ActionID = actionID
				//请求、响应
				if s.IsReq() && !ps.getReq {
					ps.getReq = true
				}
				if s.IsResp() && !ps.getResp {
					ps.getResp = true
				}
				//路径+类型名
				key := ps.pkgPath + "." + s.Name
				ps.allStructs[key] = s
			}
		}
		return true
	}
	ast.Inspect(f, findStruct)
	if actionID == "" {
		return
	}
	if !ps.getReq {
		req := defaultReq(actionID)
		key := ps.pkgPath + "." + req.Name
		ps.allStructs[key] = req
	}
	if !ps.getResp {
		resp := defaultResp(actionID)
		key := ps.pkgPath + "." + resp.Name
		ps.allStructs[key] = resp
	}
}

func (ps *PkgStructs) genField(node *ast.StructType, srcPath string) []ApiField {
	field := make([]ApiField, 0)
	for i := range node.Fields.List {
		f := node.Fields.List[i]
		if !ast.IsExported(f.Names[0].Name) {
			continue
		}
		newField := new(ApiField)
		//ignore invalid tag
		if f.Tag != nil && strings.Contains(f.Tag.Value, TokenTag) {
			tags := ps.getTag(f.Tag.Value, TokenTag)
			newField.ParseTag(f, tags)
		}
		newField.Name = f.Names[0].Name
		if f.Comment.Text() != "" {
			newField.SetDesc(f.Comment.Text())
		} else {
			newField.SetDesc(f.Doc.Text())
		}
		nt := ps.checkTypes(f.Type)
		newField.TypeName = nt.TypeName
		newField.PkgPath = nt.PkgPath
		newField.key = nt.Key
		field = append(field, *newField)
	}
	return field
}

//根据token获取标签
func (ps PkgStructs) getTag(t string, tk string) string {
	// tag = "`valid:"ass:xxx; sss""
	tagStart := strings.Index(t, tk)
	firstQ := strings.Index(t[tagStart:], `"`)
	tagEnd := strings.Index(t[tagStart+firstQ+1:], `"`)
	if tagEnd != -1 && tagStart != -1 {
		return t[tagStart : tagStart+firstQ+tagEnd+2]
	}
	return ""
}

//检查类型
func (ps *PkgStructs) checkTypes(typeToCheck ast.Expr) NewType {
	switch t := typeToCheck.(type) {
	case *ast.Ident:
		obj := ps.info.ObjectOf(t)
		newType := new(NewType)
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
		newType := new(NewType)
		if t.Len == nil {
			//slice
			newType.TypeName = "[]" + elemType.TypeName
		} else {
			var v string
			_, ok := t.Len.(*ast.Ident)
			if ok {
				// 常量
				v = t.Len.(*ast.Ident).String()
			} else {
				v = t.Len.(*ast.BasicLit).Value
			}
			newType.TypeName = fmt.Sprintf("[%s]"+elemType.TypeName, v)
		}
		newType.PkgPath = elemType.PkgPath
		newType.Key = elemType.PkgPath + "." + elemType.TypeName
		return *newType
	case *ast.MapType:
		k := ps.checkTypes(t.Key)
		v := ps.checkTypes(t.Value)
		newType := new(NewType)
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

//FindActionID if find ActionID, return ActionID and identifier(bool)
func FindActionID(s string) (string, bool) {
	s = filepath.Base(s)
	t := strings.Split(s, "_")
	if t[len(t)-1] == "test" || t[len(t)-1] == "test.go" {
		return "", false
	}
	re := regexp.MustCompile("[0-9]+")
	res := re.FindAllString(s, -1)
	if len(res) == 1 {
		return res[0], true
	}
	return "", false
}

func emptyField() ApiField {
	field := new(ApiField)
	field.Name = "无"
	field.TypeName = "无"
	field.Desc = ""
	field.Required = false
	return *field
}

//default request
func defaultReq(aid string) *StructType {
	s := new(StructType)
	s.ActionID = aid
	s.Name = "Default" + aid + "Params"
	s.Fields = []ApiField{emptyField()}
	return s
}

func defaultResp(aid string) *StructType {
	s := new(StructType)
	s.ActionID = aid
	s.Name = "Default" + aid + "Resp"
	s.Fields = []ApiField{emptyField()}
	return s
}
