package adv

import (
	"github.com/funny/fastapi"
)

// demo xxx
type Adventure struct {
}

func (adv *Adventure) APIs() fastapi.APIs {
	return fastapi.APIs{
		0: {nil, AdventureInfoOut{}},
		1: {StartAdventureIn{}, StartAdventureOut{}},
	}
}

//  本协议说明如: 秘境探险信息(请求队伍服)
type AdventureInfoIn struct {
	I int // 接口说明
}

// 此处可以空，如果不为空，只有 AdventureInfoIn 注释为空时或者为结构体 nil 时，才会生效。
type AdventureInfoOut struct {
	// 这边也可以写，优先级低
	BattleData []byte // 优先级更高，战斗json数据
	// 战斗开始时间
	BattleStartTime int64
}

// 一样
type StartAdventureIn struct {
	AdventureId int16 // demo
}

type StartAdventureOut struct {
	TeamProcess int8 // 队伍实时状态
	BB          *TestStruct
}

// demo
type TestStruct struct {
	BB int // xxxx
}
