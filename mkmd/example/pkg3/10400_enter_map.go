//travelact
//created: 2018-08-22 14:26:21
//author: wdj
package pkg3

//EnterMapResp 进入世界地图
type EnterMapResp struct {

	//UnlockedCountry 已解锁国家
	UnlockedCountry []CountryMap `valid:"required"`

	//Act 行动力
	Act ActInfo `valid:"required"`
}

//CountryMap 国家列表
type CountryMap struct {
	//已解锁的国家Id
	CountryId int
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
}

type DemoParams struct {
}
