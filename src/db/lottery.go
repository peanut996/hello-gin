package db

import (
	"log"
	"taurus-backend/constant"
)

// GetTodayLotteryCountByAwardType 获取今天抽中某种奖的人数
func GetTodayLotteryCountByAwardType(awardType int) (total int, err error) {
	err = GetDB().QueryRow("SELECT COUNT(*) AS total FROM lottery WHERE award_type = ? AND (create_time >=date(now()) AND create_time < DATE_ADD(date(now()),INTERVAL 1 DAY))", awardType).Scan(&total)
	if err != nil {
		log.Printf("get today lottery count by award type fail, err: %v", err)
		return 0, err
	}
	return total, nil
}

// GetTodayLotteryAllCount 获取今天抽中奖的人数
func GetTodayLotteryAllCount() (total int, err error) {
	err = GetDB().QueryRow("SELECT COUNT(*) AS total FROM lottery WHERE (award_type = '1' OR award_type = '2' OR award_type = '3') AND (create_time >=date(now()) AND create_time < DATE_ADD(date(now()),INTERVAL 1 DAY))").Scan(&total)
	if err != nil {
		log.Printf("get today lottery all count fail, err: %v", err)
		return 0, err
	}
	return total, nil
}

func GetTodayLotteryCountByPhone(phone string) (total int, err error) {
	err = GetDB().QueryRow("SELECT COUNT(*) AS total FROM lottery WHERE phone = ? AND (create_time >=date(now()) AND create_time < DATE_ADD(date(now()),INTERVAL 1 DAY))", phone).Scan(&total)
	if err != nil {
		log.Printf("get today lottery count by phone fail: %v ,err: %v", phone, err)
		return total, err
	}
	return total, nil
}

func GetWinLotteryCountByPhone(phone string) (total int, err error) {
	err = GetDB().QueryRow("SELECT COUNT(*) AS total FROM lottery WHERE phone = ? AND award_type != 0", phone).Scan(&total)
	if err != nil {
		log.Printf("get today lottery count by phone fail: %v ,err: %v", phone, err)
		return total, err
	}
	return total, nil
}

func CreateLottery(phone string, isWinLottery bool, awardType int, awardCode string) (err error) {
	_, err = GetDB().Exec("INSERT INTO lottery (phone, is_win_lottery, award_type, award_code) VALUES (?, ?, ?, ?)", phone, isWinLottery, awardType, awardCode)
	if err != nil {
		log.Printf("create lottery fail, err: %v", err)
		return err
	}
	return nil
}

func CreateAwardLottery(phone string, isWinLottery bool, awardType int) (string, error) {
	// 乐观锁
	tx, err := GetDB().Begin()
	if err != nil {
		err := tx.Rollback()
		if err != nil {
			return "", err
		}
		return "", err
	}
	awardCode := ""
	var version int
	// 需要判断
	for {
		err := tx.QueryRow("SELECT code,version FROM award WHERE type = ? AND is_used = 0 LIMIT 1", awardType).Scan(&awardCode, &version)
		if err != nil {
			_ = tx.Rollback()
			log.Printf("no award code, type:%v, err: %v", awardType, err)
			return "", err
		}

		res, err := tx.Exec("UPDATE award SET is_used = 1, version = ? WHERE code = ? AND version = ?", version+1, awardCode, version)
		if err != nil {
			_ = tx.Rollback()
			return "", err
		}
		rowCount, err := res.RowsAffected()
		if err != nil {
			_ = tx.Rollback()
			return "", err
		}
		if rowCount == 0 {
			log.Printf("award code consume failed, type: %v", awardType)
		} else {
			break
		}
	}
	_, err = tx.Exec("INSERT INTO lottery (phone, is_win_lottery, award_type, award_code) VALUES (?, ?, ?, ?)", phone, isWinLottery, awardType, awardCode)
	if err != nil {
		_ = tx.Rollback()
		return "", err
	}
	_, err = tx.Exec("INSERT INTO sms (phone, award_type, award_code, is_sms_sent, serial_no) VALUES (?, ?, ?, ?, uuid())", phone, awardType, awardCode, constant.SmsSendStatusNotSend)
	if err != nil {
		_ = tx.Rollback()
		return "", err
	}
	_ = tx.Commit()
	return awardCode, nil
}
