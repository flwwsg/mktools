package main

import (
	"bytes"
	"flag"
	"log"
	"mktools/common"
	"mktools/mkfastapi/do"
	"os"
	"path"
	"sort"
	"sync"
	"text/template"
)

const mkdocTemplate = `
site_name: {{.SiteName}}
theme: readthedocs
nav:
{{- range $i, $f := .Mod}}
- '{{$f}}.md'
{{- end -}}
`
const defaultOut = "docs/docs"
const allMod = "all"
const buildTag = "debug"

type MkDoc struct {
	// mkdoc 站名
	SiteName string
	// 模块
	Mod []string
}

func main() {
	module := flag.String("module", allMod, "需要生成文档的模块(不包括apidebug, battle), 默认所有模块")
	out := flag.String("out", defaultOut, "输出的文件夹，默认保存至当前工作目录下的"+defaultOut+"目录")
	mkdoc := flag.Bool("mkdoc", true, "是否生成 mkdocs 配置文件, 默认true")
	tag := flag.String("tag", buildTag, "需要查找的编译标签(+build 开头)，对apidebug模块有效")
	help := flag.Bool("h", false, "help")
	flag.Parse()
	if *help {
		flag.Usage()
	}
	genModule := func(mod string, filePath string) bool {
		pkgPath := "game_server/module/" + mod
		m := do.NewMaker(pkgPath, *tag)
		text := m.AsString()
		if text == "" {
			return false
		}
		println("generating ", mod)
		err := common.SaveFile(filePath+"/"+mod+".md", text, true)
		if err != nil {
			panic(err)
		}
		return true
	}
	// 生成 mkdoc 配置
	genMkdoc := func(filePath string, modList []string) {
		if !*mkdoc {
			return
		}
		configFile := filePath + "/" + "mkdocs.yml"
		println("mkdocs.yml will be saved in", configFile)
		doc, err := template.New("mkdocs").Parse(mkdocTemplate)
		common.PanicOnErr(err)
		sort.Strings(modList)
		m := MkDoc{"zyq", modList}
		bf := new(bytes.Buffer)
		err = doc.Execute(bf, m)
		common.PanicOnErr(err)
		err = common.SaveFile(configFile, bf.Bytes(), true)
		common.PanicOnErr(err)
	}
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	// 输出路径
	filePath := dir + "/" + *out
	if *out != defaultOut {
		filePath = *out
	}
	println("documents will be saved in", filePath)
	_ = os.MkdirAll(filePath, os.ModePerm)
	if *module != allMod {
		genModule(*module, filePath)
		genMkdoc(path.Dir(filePath), []string{*module})
		return
	}
	// 所有模块
	fpath := common.FullPackagePath("game_server/module")
	dirs := common.ListDir(fpath, false, true)
	var existsMod sync.Map
	var wg sync.WaitGroup
	for _, d := range dirs {
		if d == "battle" {
			// 包含cgo
			continue
		}
		wg.Add(1)
		go func(m string) {
			if genModule(m, filePath) {
				existsMod.Store(m, m)
			}
			wg.Done()
		}(d)
	}
	wg.Wait()
	var mods []string
	existsMod.Range(func(key, value interface{}) bool {
		mods = append(mods, key.(string))
		return true
	})
	genMkdoc(path.Dir(filePath), mods)
}
