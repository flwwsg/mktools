package main

import (
	"flag"
	"log"
	"mktools/common"
	"mktools/mkfastapi/do"
	"os"
)

func main() {
	module := flag.String("module", "all", "需要生成文档的模块(不包括apidebug, battle), 默认所有模块")
	out := flag.String("out", "docs", "输出的文件夹，默认保存至当前工作目录下的docs目录")
	mkdoc := flag.Bool("mkdoc", true, "是否生成 mkdocs 配置文件, 默认true")
	help := flag.Bool("h", false, "help")
	flag.Parse()
	if *help {
		flag.Usage()
	}
	println(module, mkdoc, out)
	// cmd := exec.Command("go", "install", "game_server/module/..")
	// current path
	genModule := func(mod string, filePath string) {
		m := do.NewMaker("game_server/module/" + mod)
		text := m.AsString()
		if text == "" {
			return
		}
		println(filePath + "/" + mod + ".md")
		err := common.SaveFile(filePath+"/"+mod+".md", text, true)
		if err != nil {
			panic(err)
		}
	}
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	// 输出路径
	filePath := dir + "/" + *out
	if *out != "docs" {
		filePath = *out
	}
	_ = os.MkdirAll(filePath, os.ModePerm)
	if *module != "all" {
		genModule(filePath, *module)
		return
	}
	// 所有模块
	fpath := common.FullPackagePath("game_server/module")
	dirs := common.ListDir(fpath, false, true)
	for _, d := range dirs {
		if d == "apidebug" || d == "battle" {
			// 包含cgo
			continue
		}
		genModule(d, filePath)
	}
}
