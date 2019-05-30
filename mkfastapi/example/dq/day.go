package dq

import (
	"github.com/funny/fastapi"
)

// 返回值为变量
type Day struct {
}

func (adv *Day) APIs() fastapi.APIs {
	api := fastapi.APIs{
		0: {DailyInfoIn{}, DailyInfoOut{}},
	}
	return api
}

//  秘境探险信息(请求队伍服)
type DailyInfoIn struct {
	i int // demo test
}

type DailyInfoOut struct {
	BattleData      []byte // 战斗json数据
	BattleStartTime int64  // 战斗开始时间
}
