//main
//created: 2018/7/26
//author: wdj

package main

import "mktools/mkmd/do"

//var i = flag.String("in", "", "api directory to generate md file")

func main() {
	//find file path
	//println(do.GenDoc("./mkmd/pkg3"))
	p := "mktools/mkmd/pkg4"
	//do.CollectTypes(p)
	do.GetDeclType(p)
	//flag.Parse()
	//if *i == "" {
	//	flag.Usage()
	//	os.Exit(1)
	//}
	//println(do.GenDoc(*i))

}
