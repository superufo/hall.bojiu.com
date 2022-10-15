package mysql

import "github.com/go-xorm/xorm"

func M() *xorm.Engine {
	return MasterDB
}

func S1() *xorm.Engine {
	return Slave1DB
}

func MLOG() *xorm.Engine {
	return MasterLogDB
}

func S1LOG() *xorm.Engine {
	return Slave1LogDB
}
