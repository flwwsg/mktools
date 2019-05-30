package pkg4

import demo4 "mktools/mkmd/example/pkg2/demo2"

//import . "mktools/mkmd/pkg2/demo3"

/*
Created on 2018-08-16 15:56:41
author: Auto Generate
*/

////获取商品列表请求参数
//type GetStoreListParams struct {
//}

//获取商品列表响应
type GetStoreListResp struct {
	No  [3]demo4.NotS
	No2 []Item
	//No NoType
	//Items []Item //商品列表
	//Item2 []Item2
	//No    NotS
}

type Item struct {
	//ItemNo NoType //道具编号 = int16
	Price        int //价格
	VipLvRequire int //vip等级要求
	BuyLimit     int //购买限制
	ItemOrder    int //商品排序
	//No           demo4.NotS //third party
}

type Item2 struct {
	P2 int
	V2 int
	B2 int
}

//Nots sss
type NotS int
