//main
//created: 2018/7/11
//author: wdj

package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"mktool/common"
	"mktool/mkact/do"
	"os"
	"os/exec"
	"path"
	"strings"
)

func main() {
	act := new(do.Act)
	toolsPath := common.FindProjectRoot("mktool")
	configPath := path.Join(toolsPath, "config", "config.json")
	conf, _ := ioutil.ReadFile(configPath)
	json.Unmarshal(conf, act)
	do.UnmarshalAPI(conf, act)
	arthurPath := common.FindProjectRoot("arthur")
	fullPath := path.Join(arthurPath, "app", "actions", act.ApiType, act.PackageName())
	regPath := path.Join(arthurPath, "app", "actions", "load.go")
	os.MkdirAll(fullPath, 777)
	for _, api := range act.ApiList {
		text := api.Text(act.PackageName())
		filePath := path.Join(fullPath, api.GoFileName())
		//ignore file exist
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
			f.Close()
			panic(err)
		}
		f.Close()
		//execute  gofmt
		cmd := exec.Command("gofmt", "-w", filePath)
		cmd.Run()
	}
	//write custom types
	filePath := path.Join(fullPath, "types.go")
	f, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}
	_, err = io.Copy(f, strings.NewReader(act.TypesText()))
	if err != nil {
		f.Close()
		panic(err)
	}
	f.Close()
	//execute  gofmt
	cmd := exec.Command("gofmt", "-w", filePath)
	cmd.Run()
	cmd = exec.Command("gen-doc.exe", "-a")
	cmd.Run()
	// register act
	do.RegisterAct(regPath, *act)
	// format imports
	cmd = exec.Command("gofmt ", "-w", regPath)
	cmd.Run()
}
