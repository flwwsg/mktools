// main
// created: 2018/7/26
// author: wdj

package main

import (
	"mktools/mkmd/collect"
)

// var i = flag.String("in", "", "api directory to generate md file")

func main() {
	// find file path
	mk := collect.NewMaker("mktools/mkmd/pkg4")
	print(mk.AsString())
	// flag.Parse()
	// if *i == "" {
	// 	flag.Usage()
	// 	os.Exit(1)
	// }
	// println(do.GenDoc(*i))

}
