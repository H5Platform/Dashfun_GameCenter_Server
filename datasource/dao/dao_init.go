package dao

import (
	"dashfun_gamecenter/datasource/mongoimpl"
	"dashfun_gamecenter/datasource/types"
)

var daoImpl types.DaoImpl

func init() {
	//使用mongodb实现
	daoImpl = mongoimpl.NewDaoImplMongo()
}
