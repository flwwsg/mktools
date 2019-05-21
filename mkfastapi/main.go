package main

import (
	"mktools/mkfastapi/do"
)

func main() {
	//fpath := common.FullPackagePath("game_server")
	//dirs := common.ListDir(fpath, false, true)
	m := do.NewMaker("game_server/module/adventure")
	println(m.AsString())
}
