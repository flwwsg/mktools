package do

import "testing"

func TestCollect(t *testing.T) {
	pkgPath := "mktools/mkfastapi/adv"
	pkg := NewPkgStructs(pkgPath)
	pkg.Parse()
	m := NewMaker(pkgPath)
	m.Parse()
	println(m.AsString())
}
