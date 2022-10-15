package entity

import "fmt"

type LogUserPerRound struct {
	ID                  int64  `gorm:"column:id" xorm:"id" json:"id"`
	UserSid             string `gorm:"column:user_sid" xorm:"user_sid" json:"user_sid"`
	PerRoundSid         string `gorm:"column:per_round_sid" xorm:"per_round_sid" json:"per_round_sid"`
	GameId              int64  `gorm:"column:game_id" xorm:"game_id" json:"game_id"`
	RoomId              int64  `gorm:"column:room_id" xorm:"room_id" json:"room_id"`
	Change              int64  `gorm:"column:change" xorm:"change" json:"change"`
	EndTime             int64  `gorm:"column:end_time" xorm:"end_time" json:"end_time"`
	Bets                string `gorm:"column:bets" xorm:"bets" json:"bets"`
	Result              string `gorm:"column:result" xorm:"result" json:"result"`
	PerRoundState       int64  `gorm:"column:per_round_state" xorm:"per_round_state" json:"per_round_state"`
	Win                 int64  `gorm:"column:win" xorm:"win" json:"win"`
	BeforeMoney         int64  `gorm:"column:before_money" xorm:"before_money" json:"before_money"`
	AfterMoney          int64  `gorm:"column:after_money" xorm:"after_money" json:"after_money"`
	Platform            string `gorm:"column:platform" xorm:"platform" json:"platform"`
	Agent               string `gorm:"column:agent" xorm:"agent" json:"agent"`
	PlayerServiceCharge int64  `gorm:"column:player_service_charge" xorm:"player_service_charge" json:"player_service_charge"`
}

type LogUserPerRound0 LogUserPerRound
type LogUserPerRound1 LogUserPerRound
type LogUserPerRound2 LogUserPerRound
type LogUserPerRound3 LogUserPerRound
type LogUserPerRound4 LogUserPerRound
type LogUserPerRound5 LogUserPerRound
type LogUserPerRound6 LogUserPerRound
type LogUserPerRound7 LogUserPerRound
type LogUserPerRound8 LogUserPerRound
type LogUserPerRound9 LogUserPerRound

func GetLogUserPerRoundTable(sid int64) string {
	tableNum := int(sid % 9)
	tableName := fmt.Sprintf("log_user_per_round_%d", tableNum)

	return tableName
}
