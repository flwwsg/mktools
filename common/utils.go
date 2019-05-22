//mktools
//created: 2018/7/30
//author: wdj

package common

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
)

// ListDir list only  directories or only files in given file path
func ListDir(fpath string, fullPath bool, onlyDir bool) []string {
	files, err := ioutil.ReadDir(fpath)
	dirs := make([]string, 0)
	fileName := ""
	if err != nil {
		log.Printf("list error path %s", fpath)
		log.Fatal(err)
	}
	for _, f := range files {
		if fullPath {
			fileName = path.Join(fpath, f.Name())
		} else {
			fileName = f.Name()
		}
		if f.IsDir() == onlyDir {
			// list only dir or only file
			dirs = append(dirs, fileName)
		}
	}
	return dirs
}

// FindProjectRoot get project path from GOPATH
func FindProjectRoot(name string) string {
	gopath := os.Getenv("GOPATH")
	mulPath := make([]string, 0)
	// check in windows
	if runtime.GOOS == "windows" {
		mulPath = strings.Split(gopath, ";")
	} else {
		// check in linux or mac
		mulPath = strings.Split(gopath, ":")
	}

	for _, p := range mulPath {
		src := filepath.Join(p, "src")
		dirs := ListDir(src, false, true)
		if FindInList(name, dirs) {
			return filepath.Join(src, name)
		}
	}
	panic(fmt.Sprintf("project named %s not found", name))
}

// 包路径
func FullPackagePath(pkgPath string) string {
	gopath := os.Getenv("GOPATH")
	mulPath := make([]string, 0)
	// check in windows
	if runtime.GOOS == "windows" {
		mulPath = strings.Split(gopath, ";")
	} else {
		// check in linux or mac
		mulPath = strings.Split(gopath, ":")
	}

	for _, p := range mulPath {
		src := filepath.Join(p, "src")
		fullPath := filepath.Join(src, pkgPath)
		_, err := os.Stat(fullPath)
		if err == nil || !os.IsNotExist(err) {
			return fullPath
		}
	}
	panic(fmt.Errorf("no such pacakge with path %s ", pkgPath))
}

// FindInList find item in given list
func FindInList(item interface{}, list interface{}) bool {
	switch t := reflect.TypeOf(list).Kind(); t {
	case reflect.Slice:
		val := reflect.ValueOf(list)
		for i := 0; i < val.Len(); i++ {
			if reflect.DeepEqual(item, val.Index(i).Interface()) {
				return true
			}
		}
	default:
		panic(fmt.Sprintf("%s not type of slice ", t))
	}

	return false
}

func SaveFile(fileName string, textString string, isForce bool) error {
	_, err := os.Stat(fileName)
	if os.IsExist(err) && !isForce {
		// 文件存在且不覆盖
		return os.ErrExist
	}
	return ioutil.WriteFile(fileName, []byte(textString), os.ModePerm)
}
