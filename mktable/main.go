package main

import (
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"mktools/mktable/do"
	"os"
	"path/filepath"

	"gitee.com/flwwsg/utils-go/errors"
)

const defaultOut = "docs/sql"
const configJSON = "config.json"

var ioOut io.Writer = os.Stdout

func main() {
	var config do.Config
	var saveDir string
	debug := flag.Bool("d", false, "是否打开调试")
	out := flag.String("out", defaultOut, "输出的文件夹，默认保存至当前工作目录下的"+defaultOut+"目录")
	configPath := flag.String("config", "", "json文件配置，默认为当前目录下的"+configJSON)
	help := flag.Bool("h", false, "help")
	flag.Parse()
	if *help {
		flag.Usage()
		return
	}
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	if *configPath == "" {
		*configPath = currentDir + "/" + configJSON
	}
	config = loadConfig(*configPath)
	config.Debug = *debug
	repo := do.NewRepo(&config)
	// get db
	db, err := repo.GetDB(&do.DB{Name: config.DBName})
	errors.PanicOnErr(err)
	sTables := do.SplitTable(db.Tables, config.Module, config.IgnorePattern)
	writeFile := *out != ""

	if writeFile {
		if *out != defaultOut {
			saveDir = *out
		} else {
			// 输出路径
			saveDir = currentDir + "/" + *out
		}
		_ = os.MkdirAll(saveDir, os.ModePerm)
	}
	// 根据配置获取表
	for k, v := range sTables {
		if writeFile {
			f, err := os.Create(saveDir + "/" + k + ".md")
			if err != nil {
				log.Fatal(err)
			}
			do.RenderTable(v, f)
			err = f.Close()
			if err != nil {
				log.Fatal(err)
			}
			continue
		}
		do.RenderTable(v, ioOut)
	}
}

func loadConfig(configPath string) do.Config {
	fullPath, err := filepath.Abs(configPath)
	if err != nil {
		log.Fatal(err)
	}
	data, err := ioutil.ReadFile(fullPath)
	if err != nil {
		log.Fatal(err)
	}
	config := do.Config{}
	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Fatal(err)
	}
	return config
}
