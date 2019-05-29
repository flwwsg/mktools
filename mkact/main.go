//main
//created: 2018/7/11
//author: wdj

package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"mktools/mkact/do"
	"os"
	"os/exec"
	"path"
	"strings"

	"gitee.com/flwwsg/utils-go/errors"

	"gitee.com/flwwsg/utils-go/files"
)

func main() {
	act := new(do.Act)
	toolsPath := files.FindProjectRoot("mktools")
	configPath := path.Join(toolsPath, "config", "config.json")
	conf, _ := ioutil.ReadFile(configPath)
	_ = json.Unmarshal(conf, act)
	do.UnmarshalAPI(conf, act)
	arthurPath := files.FindProjectRoot("arthur")
	fullPath := path.Join(arthurPath, "app", "actions", act.ApiType, act.PackageName())
	regPath := path.Join(arthurPath, "app", "actions", "load.go")
	_ = os.MkdirAll(fullPath, os.ModePerm)
	for _, api := range act.ApiList {
		text := api.Text(act.PackageName())
		filePath := path.Join(fullPath, api.GoFileName())
		// ignore file exist
		_, err := os.Open(filePath)
		if err == nil {
			continue
		}
		f, err := os.Create(filePath)
		if err != nil {
			panic(err)
		}

		_, err = io.Copy(f, strings.NewReader(text))
		if err != nil {
			_ = f.Close()
			panic(err)
		}
		_ = f.Close()
		// execute  gofmt
		cmd := exec.Command("gofmt", "-w", filePath)
		err = cmd.Run()
		errors.PanicOnErr(err)
	}
	// write custom types
	filePath := path.Join(fullPath, "types.go")
	f, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}
	_, err = io.Copy(f, strings.NewReader(act.TypesText()))
	if err != nil {
		_ = f.Close()
		panic(err)
	}
	_ = f.Close()
	// execute  gofmt
	cmd := exec.Command("gofmt", "-w", filePath)
	_ = cmd.Run()
	cmd = exec.Command("gen-doc.exe", "-a")
	_ = cmd.Run()
	// register act
	do.RegisterAct(regPath, *act)
	// format imports
	cmd = exec.Command("gofmt ", "-w", regPath)
	_ = cmd.Run()
}
