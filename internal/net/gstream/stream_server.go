package gstream

import (
	Utils "common.bojiu.com/utils"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/jinzhu/copier"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	. "hall.bojiu.com/config"
	"hall.bojiu.com/enum/msg"
	"hall.bojiu.com/internal/model/entity"
	"hall.bojiu.com/internal/net/gstream/pb"
	lpb "hall.bojiu.com/internal/net/hall/pb"
	"hall.bojiu.com/pkg/log"
	"hall.bojiu.com/pkg/mysql"
	"net"
	"strings"
	"time"
)

// keepalive 参数
var (
	kaep = keepalive.EnforcementPolicy{
		MinTime:             10000000 * time.Second, // If a client pings more than once every 5 seconds, terminate the connection
		PermitWithoutStream: true,                   // Allow pings even when there are no active streams
	}

	kasp = keepalive.ServerParameters{
		MaxConnectionIdle:     10000000 * time.Second, // If a client is idle for 15 seconds, send a GOAWAY
		MaxConnectionAge:      10000000 * time.Second, // If any connection is alive for more than 30 seconds, send a GOAWAY
		MaxConnectionAgeGrace: 5 * time.Second,        // Allow 5 seconds for pending RPCs to complete before forcibly closing connections
		Time:                  5 * time.Second,        // Ping the client if it is idle for 5 seconds to ensure the connection is still active
		Timeout:               1 * time.Second,        // Wait 1 second for the ping ack before assuming the connection is dead
	}
)

type streamServer struct {
	GrpcRecvClientData chan *pb.StreamRequestData
	GrpcSendClientData chan *pb.StreamResponseData
}

func NewStreamServer() *streamServer {
	gs := &streamServer{}

	gs.GrpcRecvClientData = make(chan *pb.StreamRequestData, 100)
	gs.GrpcSendClientData = make(chan *pb.StreamResponseData, 100)

	return gs
}

// PPStream log.ZapLog.With(zap.Any("err", err)).Error("收到网关数据错误")
func (gs *streamServer) PPStream(stream pb.ForwardMsg_PPStreamServer) error {
	go func() {
		for {
			msg, err := stream.Recv()

			if err == nil {
				info := fmt.Sprintf("收到网关数据:协议号=%+v,加密字符=%+v,随机字符=%+v,protobuf=%+v", msg.GetMsg(), Utils.ToHexString(msg.GetSecret()), msg.GetSerialNum(), msg.GetData())
				log.ZapLog.Info(info)
				gs.GrpcRecvClientData <- msg
			}
		}
	}()

	// log.ZapLog.With(zap.Any("err", err)).Error("发送给网关")
	go func() {
		for {
			select {
			case sd := <-gs.GrpcSendClientData:
				//业务代码
				if err := stream.Send(sd); err == nil {
					log.ZapLog.With(zap.Any("msg", sd.GetMsg()), zap.Any("data", sd.GetData())).Info("发给网关")
				} else {
					log.ZapLog.With(zap.Any("err", err)).Info("发给网关")
				}
			}
		}
	}()

	gs.dispatch()

	return nil
}

func (gs *streamServer) dispatch() {
	//var err error
	for {
		select {
		case clientMsg := <-gs.GrpcRecvClientData:
			log.ZapLog.With(zap.Any("Msg", clientMsg.Msg)).Info("dispatch")

			if strings.Contains(msg.CMDS, fmt.Sprintf("%d", clientMsg.Msg)) == false {
				log.ZapLog.With(zap.Any("err", errors.New("不存在的消息"))).Error("不存在的消息")
			}

			// 处理登录消息
			if uint16(clientMsg.Msg) == msg.CMD_LOG {
				gs.log(clientMsg)
			}

		}
	}
}

func (gs *streamServer) log(clientMsg *pb.StreamRequestData) (err error) {
	req := lpb.MLogTos{}
	if err := proto.Unmarshal(clientMsg.Data, &req); err != nil {
		log.ZapLog.With(zap.Any("err", err)).Info("log")
		return errors.New("proto3解码错误")
	}

	tableName := entity.GetLogUserPerRoundTable(req.UserId)

	rs := make([]entity.LogUserPerRound, 0)
	err = mysql.S1LOG().Table(tableName).Select("*").Where("game_id=? and user_sid=? ", req.GameId, req.UserId).Limit(int(req.PageSize), int(req.Page)).Asc("game_id,user_sid").Find(&rs)

	if err != nil {
		log.ZapLog.With(zap.Any("err", err)).Info("log")
		return errors.New("数据库查询错误")
	}

	logs := make([]*lpb.PLogInfo, 0)
	for _, r := range rs {
		lr := lpb.PLogInfo{}
		copier.Copy(&lr, &r)

		logs = append(logs, &lr)
	}

	logsToc := lpb.MLogToc{Logs: logs}

	data, _ := proto.Marshal(&logsToc)
	sendCMsg := pb.StreamResponseData{
		ClientId: clientMsg.ClientId,
		Msg:      uint32(msg.CMD_LOG),
		Data:     data,
	}
	gs.GrpcSendClientData <- &sendCMsg

	return err
}

func Run() {
	defer func() {
		if err := recover(); err != nil {
			log.ZapLog.With(zap.Any("error", err)).Error("streamServer")
		}
	}()

	var server pb.ForwardMsgServer
	sImpl := NewStreamServer()

	server = sImpl

	// keepalive
	g := grpc.NewServer(grpc.KeepaliveEnforcementPolicy(kaep),
		//grpc.KeepaliveParams(kasp),
		grpc.MaxConcurrentStreams(1000),
		grpc.ConnectionTimeout(300*time.Second),    // 缺省 120
		grpc.MaxRecvMsgSize(2000*1024*1024),        //最大允许接收的字节数
		grpc.MaxSendMsgSize(2000*1024*1024),        //最大允许发送的字节数
		grpc.InitialWindowSize(2000*1024*1024),     //stream 滑动窗口
		grpc.InitialConnWindowSize(2000*1024*1024), // connnect 滑动窗口
		//grpc.NumStreamWorkers(100),                 //channel 对应的goroutines数目
	)

	// 2.注册逻辑到server中
	pb.RegisterForwardMsgServer(g, server)

	instance := fmt.Sprintf("%s:%d", Scfg.Cfg.Grpc.Stream.Host, Scfg.Cfg.Grpc.Stream.Port)

	log.ZapLog.With(zap.Any("addr", instance)).Info("streamServer")
	// 3.启动server
	lis, err := net.Listen("tcp", instance)
	if err != nil {
		panic("监听错误:" + err.Error())
	}

	err = g.Serve(lis)
	if err != nil {
		panic("启动错误:" + err.Error())
	}
}
