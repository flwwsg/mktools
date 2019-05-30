// +build windows,!debug

package build_tag

import (
	"github.com/funny/fastapi"
)

// demo xxx
type AdventureDot struct {
}

func (adv *AdventureDot) APIs() fastapi.APIs {
	return fastapi.APIs{
		1: {StartAdventureInDot{}, StartAdventureOutDot{}},
	}
}

// 秘境选择界面 开始冒险(请求队伍服)
type StartAdventureInDot struct {
	AdventureId int16 //demo
}

type StartAdventureOutDot struct {
	TeamProcess int8 // 队伍实时状态
}
