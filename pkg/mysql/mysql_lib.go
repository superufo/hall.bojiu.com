package mysql

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"hall.bojiu.com/pkg/log"
	"hall.bojiu.com/pkg/viper"

	"time"
)

var (
	MasterDB *xorm.Engine
	err      error
	Slave1DB *xorm.Engine
	err1     error

	MasterLogDB *xorm.Engine
	Slave1LogDB *xorm.Engine
)

func MasterInit() *xorm.Engine {
	open := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8", viper.Vp.GetString("mysql-master.username"),
		viper.Vp.GetString("mysql-master.password"),
		viper.Vp.GetString("mysql-master.addr"),
		viper.Vp.GetInt64("mysql-master.port"),
		viper.Vp.GetString("mysql-master.database"))

	MasterDB, err = xorm.NewEngine("mysql", open)
	if err != nil {
		log.ZapLog.Error(fmt.Sprintf("Open mysql-master failed,err:%s\n", err.Error()))
		panic(err)
	}

	MasterDB.SetConnMaxLifetime(100 * time.Second)
	MasterDB.SetMaxOpenConns(100)
	MasterDB.SetMaxIdleConns(16)
	err = MasterDB.Ping()
	if err != nil {
		log.ZapLog.Error(fmt.Sprintf("Failed to connect to mysql-master, err:%s" + err.Error()))
		panic(err.Error())
	}

	// 显示打印语句
	if viper.Vp.GetString("active") == "dev" || viper.Vp.GetString("active") == "test" {
		MasterDB.ShowSQL(true)
	}

	log.ZapLog.Info("mysql-master connect success\r\n")
	return MasterDB
}

func Slave1Init() *xorm.Engine {
	open := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8", viper.Vp.GetString("mysql-slave1.username"),
		viper.Vp.GetString("mysql-slave1.password"),
		viper.Vp.GetString("mysql-slave1.addr"),
		viper.Vp.GetInt64("mysql-slave1.port"),
		viper.Vp.GetString("mysql-slave1.database"))

	Slave1DB, err1 = xorm.NewEngine("mysql", open)
	if err != nil {
		log.ZapLog.Error(fmt.Sprintf("Open mysql-slave1 failed,err:%s\n", err.Error()))
		panic(err)
	}

	Slave1DB.SetConnMaxLifetime(100 * time.Second)
	Slave1DB.SetMaxOpenConns(100)
	Slave1DB.SetMaxIdleConns(16)

	// 显示打印语句
	if viper.Vp.GetString("active") == "dev" || viper.Vp.GetString("active") == "test" {
		Slave1DB.ShowSQL(true)
	}

	err1 = Slave1DB.Ping()
	if err != nil {
		log.ZapLog.Error(fmt.Sprintf("Failed to connect to mysql-slave1, err:%s" + err1.Error()))
		panic(err.Error())
	}

	log.ZapLog.Info("mysql-slave1 connect success\r\n")
	return Slave1DB
}

func MasterLogDbInit() *xorm.Engine {
	open := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8", viper.Vp.GetString("mysql-log-master.username"),
		viper.Vp.GetString("mysql-log-master.password"),
		viper.Vp.GetString("mysql-log-master.addr"),
		viper.Vp.GetInt64("mysql-log-master.port"),
		viper.Vp.GetString("mysql-log-master.database"))

	MasterLogDB = DbInit(open)

	return MasterLogDB
}

func Slave1LogDbInit() *xorm.Engine {
	open := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8", viper.Vp.GetString("mysql-log-slave1.username"),
		viper.Vp.GetString("mysql-log-slave1.password"),
		viper.Vp.GetString("mysql-log-slave1.addr"),
		viper.Vp.GetInt64("mysql-log-slave1.port"),
		viper.Vp.GetString("mysql-log-slave1.database"))

	Slave1LogDB = DbInit(open)

	return Slave1LogDB
}

func DbInit(connStr string) *xorm.Engine {
	dbEngine, err := xorm.NewEngine("mysql", connStr)
	if err != nil {
		log.ZapLog.Error(fmt.Sprintf("Open %s failed,err:%s\n", connStr, err.Error()))
		panic(err)
	}

	dbEngine.SetConnMaxLifetime(100 * time.Second)
	dbEngine.SetMaxOpenConns(100)
	dbEngine.SetMaxIdleConns(16)
	err = dbEngine.Ping()
	if err != nil {
		log.ZapLog.Error(fmt.Sprintf("Failed to connect to %s, err:%s" + connStr + err.Error()))
		panic(err.Error())
	}

	// 显示打印语句
	if viper.Vp.GetString("active") == "dev" || viper.Vp.GetString("active") == "test" {
		dbEngine.ShowSQL(true)
	}

	log.ZapLog.Info(connStr + " connect success\r\n")
	return dbEngine
}
