//pkg4
//created: 2018/9/17
//author: wdj

package pkg4

import . "mktools/mkmd/example/pkg5"

//CountryMap 国家列表
type CountryMap struct {
	//已解锁的国家Id
	CountryID int
	//所有委托事件
	EntrustCityNum int
	//HeroDispatch 英雄派遣信息列表
	HeroDispatch []DispatchInfo
}

type ActInfo struct {
	TotalAct int
	Act      int
	ActTime  int64
}

//DispatchInfo 英雄派遣信息
type DispatchInfo struct {
	//英雄id
	ID string
	//已派遣次数
	Dispatched int
	//剩余可以派遣次数
	RemainDispatched int
	//info
	Info  Info5
	Info6 Info6
}
