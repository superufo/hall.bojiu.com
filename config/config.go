package config

import (
	"common.bojiu.com/discover/kit/sd/etcdv3"

	"go.uber.org/zap"
	"hall.bojiu.com/pkg/log"
	"hall.bojiu.com/pkg/viper"
	"time"
)

var (
	Scfg serverHallCfg
)

func InitConfig() {
	Scfg = NewServerCfg()
}

type Server struct {
	Host   string
	Port   int32
	RegKey string
}

type grpc struct {
	Stream  Server
	Unitary Server
}

type serverCfg struct {
	Grpc       grpc
	EtcdServer []string
}

type serverHallCfg struct {
	Cfg    serverCfg
	Option etcdv3.ClientOptions
}

func NewServerCfg() serverHallCfg {
	cfg := serverCfg{}
	if err := viper.Vp.UnmarshalKey("ser", &cfg); err != nil {
		log.ZapLog.Error("解析配置文件失败", zap.Any("err", err))
	}
	log.ZapLog.With(zap.Stack("trace")).Info("serverCfg")

	return serverHallCfg{
		Cfg: cfg,
		Option: etcdv3.ClientOptions{
			// Path to trusted ca file
			CACert: "",
			// Path to certificate
			Cert: "",
			// Path to private key
			Key: "",
			// Username if required
			Username: "",
			// Password if required
			Password: "",
			// If DialTimeout is 0, it defaults to 3s
			DialTimeout: time.Second * 3,
			// If DialKeepAlive is 0, it defaults to 3s
			DialKeepAlive: time.Second * 3,
			// If passing `grpc.WithBlock`, dial connection will block until success.
			//DialOptions: []grpc.DialOption{grpc.WithBlock()},
		},
	}
}
