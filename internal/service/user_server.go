package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"common.bojiu.com/utils"
	"hall.bojiu.com/internal/model/entity"
	"hall.bojiu.com/pkg/log"
	"hall.bojiu.com/pkg/mysql"
	"hall.bojiu.com/pkg/redislib"
)

// UserService User 获取 user user_info 表格中信息
type UserService struct {
}

func NewUserService() *UserService {
	return &UserService{}
}

func (us UserService) GetUserByToken(token string, userID int64) (user entity.Users, err error) {
	var has bool

	if token == "" {
		err = errors.New("token为空")
		goto RET
	}

	if userID < 1 {
		err = errors.New("userID为空")
		goto RET
	}

	has, err = mysql.S1().Table(entity.TABLE_USERS).Select("*").Where("token=?  and s_id=?", token, userID).Get(&user)
	log.ZapLog.With(zap.Any("ok", has)).Info("GetUserByToken......")
	if err != nil {
		sql, _ := mysql.S1().Table(entity.TABLE_SERVER_LIST).LastSQL()
		log.ZapLog.With(zap.Namespace("database"), zap.Any("err", err), zap.Any("sql", sql)).Error("数据库查询错误")
	}

	if !has {
		err = errors.New("用户不存在")
	}

RET:
	return user, err
}

func (us UserService) SetUserToRds(user entity.Users, c *redis.Client) (bool, error) {
	key := fmt.Sprintf("user:%s", user.SId)
	ctx := context.Background()
	res := c.HMSet(
		ctx, key,
		"s_id", user.SId,
		"id", user.ID,
		"account", user.Account,
		"name", user.Name,
		"token", user.Token,
		"platform", user.Platform,
		"sex", user.Sex,
		"mac", user.Mac,
		"nickname", user.Nickname,
		"c_code", user.CCode,
		"phone", user.Phone,
		"register_time", user.RegisterTime,
		"password", user.Password,
		"agent", user.Agent,
		"status", user.Status,
		"register_ip", user.RegisterIp,
		"father_id", user.FatherId,
	)

	return res.Result()
}

func (us UserService) GetUserByPwd(account string, mac string, pwd string) (user entity.Users, err error) {
	redislib.Sclient()
	redisClient := redislib.GetClient()
	defer redisClient.Close()

	ukey := fmt.Sprintf("user:%s", user.Account)
	if err := redisClient.HGetAll(context.Background(), ukey).Scan(&user); err != nil {
		log.ZapLog.With(zap.Namespace("redis"), zap.Any("err", err), zap.Any("ukey", ukey)).Error("redis操作错误")
	}

	isOK, err := mysql.S1().Table(entity.TABLE_USERS).Select("*").Where("(account=? or mac =?) and password=?", account, mac, pwd).Get(&user)
	if err != nil {
		sql, _ := mysql.S1().Table(entity.TABLE_SERVER_LIST).LastSQL()
		log.ZapLog.With(zap.Namespace("database"), zap.Any("err", err), zap.Any("sql", sql)).Error("数据库查询错误")
	}

	//todo 判断是否为游客
	if isOK == false {
		user, _ = us.GeneralGustAccount(mac, pwd)
	}

	return user, err
}

func (us UserService) GeneralGustAccount(mac string, pwd string) (user entity.Users, err error) {
	rand.Seed(time.Now().Unix())

	user = entity.Users{
		SId:          utils.GenRandString(16),
		ID:           rand.Intn(1000),
		Account:      fmt.Sprintf("account_%s", utils.GenRandString(16)),
		Name:         fmt.Sprintf("guest_%s", utils.GenRandString(16)),
		Token:        "",
		Platform:     "",
		Sex:          int8(rand.Intn(1)),
		Mac:          mac, //utils.GenRandString(32),
		Nickname:     fmt.Sprintf("guest_%s", utils.GenRandString(16)),
		CCode:        "",
		Phone:        "",
		RegisterTime: time.Now().Unix(),
		Password:     pwd, //string(fmt.Sprintf("%x", md5.Sum([]byte{1, 2, 3, 4, 5, 6}))),
		Agent:        "",
		Status:       int8(0),
		RegisterIp:   "",
		FatherId:     "",
	}

	if affected, err := mysql.M().Table(entity.TABLE_USERS).Insert(user); err == nil {
		if affected < 1 {
			log.ZapLog.With(zap.Any("affected", affected)).Error("数据库插入数据为空")
		}
	} else {
		log.ZapLog.With(zap.Any("err", err)).Error("数据库插入错误")
	}

	return user, err
}
