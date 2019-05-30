package do

import "testing"

func TestCollect(t *testing.T) {
	pkgPath := "mktools/mkfastapi/example/adv"
	m := NewMaker(pkgPath, "")
	m.Parse()
	println(m.AsString())
}

func TestTagTag(t *testing.T) {
	m := NewMaker("mktools/mkfastapi/example/build-tag", "debug")
	m.Parse()
	if len(m.allAPI) != 1 {
		// 采集失败
		t.Error("tag debug fail", len(m.allAPI))
	}
	if m.allAPI["1"].StructInName != "StartAdventureIn" || m.allAPI["1"].StructOutName != "StartAdventureOut" {
		t.Error("invalid api ")
	}
	println(m.AsString())
	m = NewMaker("mktools/mkfastapi/example/build-tag", "release")
	m.Parse()
	if len(m.allAPI) != 1 {
		// 采集失败
		t.Error("tag debug fail", len(m.allAPI))
	}
	if m.allAPI["1"].StructInName != "StartAdventureInDot" || m.allAPI["1"].StructOutName != "StartAdventureOutDot" {
		t.Error("invalid api")
	}
	println(m.AsString())
}
