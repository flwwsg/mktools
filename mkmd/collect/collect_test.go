package collect

import (
	"testing"
)

//var mkProject = "mktool"

func TestApiMaker(t *testing.T) {
	pkg := "mktools/mkmd/example/pkg1"
	mk := NewMaker(pkg)
	mk.Parse()
	println(mk.AsString())
	//for k, v := range mk.allAPI {
	//	println(k)
	//	println("custom types ")
	//	for k, t := range v.CustomTypes {
	//		println(k)
	//		fmt.Printf("%v\n", *t)
	//	}
	//}
	//fmt.Printf("%v\n", mk.allStruct)

}
