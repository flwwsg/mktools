package adv

import (
	"github.com/funny/fastapi"
)

//demo xxx
type Adventure struct {
}

func (adv *Adventure) APIs() fastapi.APIs {
	return fastapi.APIs{
		0: {nil, AdventureInfoOut{}},
		1: {StartAdventureIn{}, StartAdventureOut{}},
	}
}

//  秘境探险信息(请求队伍服)
type AdventureInfoIn struct {
	i int // demo test
}

type AdventureInfoOut struct {
	BattleData      []byte // 战斗json数据
	BattleStartTime int64  // 战斗开始时间
}

// 秘境选择界面 开始冒险(请求队伍服)
type StartAdventureIn struct {
	AdventureId int16 //demo
}

type StartAdventureOut struct {
	TeamProcess int8 // 队伍实时状态
	BB          *TestStruct
}

//demo
type TestStruct struct {
	BB int //
}
