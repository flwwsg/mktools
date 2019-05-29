//mkmodel
//created: 2018/7/26
//author: wdj

package main

import (
	"fmt"
	"mktools/mkmodel/do"
	"path/filepath"

	"gitee.com/flwwsg/utils-go/files"
)

func main() {
	root := files.FindProjectRoot("mktools")
	mPath := filepath.Join(root, "mkmodel", "model")
	dynPath := filepath.Join(mPath, "dyn")
	statPath := filepath.Join(mPath, "stat")
	// do.RegisterModel(filepath.Join(mPath, "reg.go"), filepath.Join(dynPath, "demo2.go"))
	do.RegisterModel(filepath.Join(mPath, "reg.go"), filepath.Join(statPath, "demo2.go"))
	fmt.Printf("%v\n", do.FindStructs(dynPath))
	fmt.Printf("%v\n", do.FindStructs(statPath))
}
