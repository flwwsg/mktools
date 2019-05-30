# mkinfo 

### 说明
真有趣服务端模块协议生成器。生成每个模块(除了battle)的协议，分别为每个模块生成 md 文件，模块adv：

```go
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
	BattleData      []byte // 同时存在时，这里注释优先级更高
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

```
生成的 md 文件为
```markdown
### 自定义数据类型

#### TestStruct
字段|类型|描述|
---|---|---
BB | int | xxxx

### 0 此处可以空，如果不为空，只有 AdventureInfoIn 注释为空时或者为结构体 nil 时，才会生效。

#### 无

字段|类型|描述|
---|---|---
无 | 无| 无


#### AdventureInfoOut

字段|类型|描述|
---|---|---
BattleData | []byte | 同时存在时，这里注释优先级更高
BattleStartTime | int64 | 战斗开始时间

### 1 一样

#### StartAdventureIn

字段|类型|描述|
---|---|---
AdventureId | int16| demo


#### StartAdventureOut

字段|类型|描述|
---|---|---
TeamProcess | int8 | 队伍实时状态
BB | TestStruct | 无
```
### 参数
- -h 显示帮助
- -module 需要生成文档的模块(不包括battle), 默认所有模块
- -out 输出的文件夹，默认保存至当前工作目录下的docs/docs目录