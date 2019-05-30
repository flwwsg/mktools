//travelact
//created: 2018-08-22 14:26:21
//author: wdj
package pkg4

import (
	"mktools/mkmd/example/pkg1"
	"mktools/mkmd/example/pkg5"
)

//EnterMapResp 进入世界地图
type EnterMapResp struct {
	//
	//UnlockedCountry 已解锁国家
	UnlockedCountry []CountryMap `valid:"required"`
	Beauty          [3]pkg5.Info5
	Info6           pkg5.Info6
	//CountryID 国家id
	CountryID map[string]pkg1.DemoXX

	//Act 行动力
	Act ActInfo `valid:"required"`
}

type DemoParams struct {
}

func NewInfo() {
	println("==============")
}

func init() {
	println("initializing ")
}
