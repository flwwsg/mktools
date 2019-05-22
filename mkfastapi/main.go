package main

import (
	"flag"
	"mktools/common"
	"mktools/mkfastapi/do"
)

func main() {
	module := flag.String("module", "all", "需要生成文档的模块(不包括apidebug), 默认所有模块")
	out := flag.String("out", "", "输出的文件夹，默认保存至当前目录下的docs目录")
	mkdoc := flag.Bool("mkdoc", true, "是否生成 mkdocs 配置文件, 默认true")
	help := flag.Bool("h", false, "help")
	flag.Parse()
	if *help {
		flag.Usage()
	}
	println(module, mkdoc, out)
	fpath := common.FullPackagePath("game_server/module")
	dirs := common.ListDir(fpath, false, true)
	for _, d := range dirs {
		println(d)
		if d == "apidebug" {
			continue
		}
		m := do.NewMaker("game_server/module/" + d)
		println(m.AsString())
	}
	//m := do.NewMaker("game_server/module/adventure")
	//println(m.AsString())

}
