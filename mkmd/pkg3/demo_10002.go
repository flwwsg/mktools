package pkg3

import (
	//demo4 "mktools/mkmd/pkg2/demo2"
	. "mktools/mkmd/pkg2/demo3"
)

/*
Created on 2018-08-16 15:56:41
author: Auto Generate
*/

////获取商品列表请求参数
type GetStoreListParams struct {
}

//获取商品列表响应
type GetStoreListResp struct {
	Items   []Item //商品列表
	NewItem Item
}

type Item struct {
	ItemNo       NoType //道具编号 = int16
	Price        int    //价格
	VipLvRequire int16  //vip等级要求
	BuyLimit     int    //购买限制
	ItemOrder    int    //商品排序
	No           NotS
}

////Nots
type NotS int
