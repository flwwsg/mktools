package common

import (
	"strings"
)

// ApiField 结构体字段
type ApiField struct {
	Name     string
	TypeName string
	Desc     string
	Required bool
	PkgPath  string
	key      string
}

// StructType 结构体
type StructType struct {
	Name     string
	ActionID string // mark action id to specify API, only request struct will be marked
	Fields   []ApiField
	Desc     string
	PkgPath  string
}

type NewType struct {
	// 显示的类型名
	TypeName string
	// 包路径
	PkgPath string
	// 键值
	Key string
	// basic value
	Value string
}

// // PkgStructs package 下的所有结构体
// type PkgStructs struct {
//	//文件路径
//	pkgPath    string
//	info       *types.Info
//	allStructs map[string]*StructType
//	scope      *types.Scope
//	fset       *token.FileSet
//	//是否存在请求、响应结构体
//	getReq  bool
//	getResp bool
// }

func (field *ApiField) SetDesc(s string) {
	desc := strings.Replace(s, field.Name, "", 1)
	desc = strings.Replace(desc, "\n", " ", -1)
	field.Desc = strings.TrimSpace(desc)
}

// // IsValidTag check tag is valid or not
// func (field *ApiField) IsValidTag(t string) bool {
//	return !strings.Contains(t, "-")
// }
//
// //ParseTag handle tag
// func (field *ApiField) ParseTag(f *ast.Field, t string) {
//	// t = "valid: \"Required, xxx\""
//	if !field.IsValidTag(t) {
//		return
//	}
//	t = t[strings.Index(t, "\"")+1 : strings.LastIndex(t, "\"")]
//	fields := strings.Split(t, ",")
//	field.Required = false
//	for _, f := range fields {
//		f = strings.TrimSpace(f)
//		if f == "" {
//			continue
//		}
//		switch f {
//		case "required":
//			field.Required = true
//		case "optional":
//		default:
//			continue
//		}
//	}
// }

func (field *ApiField) SetKey(key string) {
	field.key = key
}

func (field ApiField) GetKey() string {
	return field.key
}

func (s *StructType) SetDesc(comm string) {
	// drop struct Name
	desc := strings.Replace(comm, s.Name, "", 1)
	desc = strings.Replace(desc, "\n", " ", -1)
	s.Desc = strings.TrimSpace(desc)
}

// 包含包名的类型名
func (s StructType) FullName() string {
	return s.PkgPath + "." + s.Name
}

// print
func (nt NewType) String() string {
	return nt.Key
}
