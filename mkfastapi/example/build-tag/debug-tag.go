// +build debug

package build_tag

import (
	"github.com/funny/fastapi"
)

// demo xxx
type Adventure struct {
}

func (adv *Adventure) APIs() fastapi.APIs {
	return fastapi.APIs{
		1: {StartAdventureIn{}, StartAdventureOut{}},
	}
}

// 秘境选择界面 开始冒险(请求队伍服)
type StartAdventureIn struct {
	AdventureId int16 //demo
}

type StartAdventureOut struct {
	TeamProcess int8 // 队伍实时状态
	BB          *TestStruct
}

// demo
type TestStruct struct {
	BB int //
}
