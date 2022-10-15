package main

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	. "hall.bojiu.com/config"
	"hall.bojiu.com/internal/net/gstream"
	"hall.bojiu.com/pkg/log"
	"hall.bojiu.com/pkg/mysql"
	"hall.bojiu.com/pkg/redislib"
	"hall.bojiu.com/pkg/viper"
	//_ "net/http/pprof"
	//"hall.bojiu.com/internal/net/hall"

	. "common.bojiu.com/discover/kit/sd/etcdv3"
)

func main() {
	// 初始化配置文件
	viper.InitVp()

	// 初始化日志文件
	log.ZapLog = log.InitLogger()

	// 初始化redis
	redislib.Sclient()

	// 初始化数据库 获取 mysql.M()  mysql.S()
	MasterDB := mysql.MasterInit()
	defer MasterDB.Close()
	Slave1DB := mysql.Slave1Init()
	defer Slave1DB.Close()

	MasterLogDB := mysql.MasterLogDbInit()
	defer MasterLogDB.Close()
	Slave1LogDB := mysql.Slave1LogDbInit()
	defer Slave1LogDB.Close()

	// config
	InitConfig()

	//with里面输出json的结构化数据，便利后期做数据分析
	level := viper.Vp.GetString("log.level")
	log.ZapLog.With(zap.Namespace("日志级别"),
		zap.Any("level", level),
	).Info("main")

	//go func() {
	//	http.ListenAndServe("127.0.0.1:6060", nil)
	//}()

	/*******服务注册 start*******/
	instance := fmt.Sprintf("%s:%d", Scfg.Cfg.Grpc.Unitary.Host, Scfg.Cfg.Grpc.Unitary.Port)
	eClient, err := NewClient(context.Background(), Scfg.Cfg.EtcdServer, Scfg.Option)
	if err != nil {
		log.ZapLog.With(zap.Error(err), zap.Stack("trace")).Info("error")
	}

	registrar := NewRegistrar(eClient, Service{
		Key:   Scfg.Cfg.Grpc.Unitary.RegKey,
		Value: instance,
	}, log.ZapLog)

	registrar.Register()
	defer registrar.Deregister()
	v, _ := eClient.GetEntries(Scfg.Cfg.Grpc.Unitary.RegKey)
	log.ZapLog.With(zap.Any("regKey", v)).Info("main")

	streamInstance := fmt.Sprintf("%s:%d", Scfg.Cfg.Grpc.Stream.Host, Scfg.Cfg.Grpc.Stream.Port)
	streamRegistrar := NewRegistrar(eClient, Service{
		Key:   Scfg.Cfg.Grpc.Stream.RegKey,
		Value: streamInstance,
	}, log.ZapLog)

	streamRegistrar.Register()
	defer streamRegistrar.Deregister()
	v, _ = eClient.GetEntries(Scfg.Cfg.Grpc.Stream.RegKey)
	log.ZapLog.With(zap.Any("regKey", v)).Info("main")
	/*******服务注册 end *******/
	// hall 一元流服务器
	//hall.Run()

	// grpc 流服务器
	gstream.Run()
}
