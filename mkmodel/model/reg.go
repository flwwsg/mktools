//model
//created: 2018/7/27
//author: wdj

package model

import (
	"mktools/mkmodel/model/dyn"
	"mktools/mkmodel/model/stat"
)

//before


type ProfileCenter struct {
DA []stat.DemoStatA
DemoStatB []stat.DemoStatB
}


//middle
//? what
//last

var CharacterSet = []interface{}{
	&dyn.DemoDynA{},
	&dyn.DemoDynB{},
}

func init() {
	println(CharacterSet)
}
