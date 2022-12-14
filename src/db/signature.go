package db

import "log"

func GetTodaySignatureCountByPhone(phone string) (total int, err error) {
	err = GetDB().QueryRow("SELECT COUNT(*) AS total FROM signature WHERE phone = ? AND (create_time >=date(now()) AND create_time < DATE_ADD(date(now()),INTERVAL 1 DAY))", phone).Scan(&total)
	if err != nil {
		log.Printf("get count signature by phone fail: %v ,err: %v", phone, err)
		return total, err
	}
	return total, nil
}

func GetSignatureCountByStreet(street string) (total int, err error) {
	err = GetDB().QueryRow("SELECT COUNT(*) AS total FROM signature WHERE street = ?", street).Scan(&total)
	if err != nil {
		log.Printf("get count signature by street fail: %v ,err: %v", street, err)
		return total, err
	}
	return total, nil
}

func GetAllSignatureCount() (total int, err error) {
	err = GetDB().QueryRow("SELECT COUNT(*) AS total FROM signature").Scan(&total)
	if err != nil {
		log.Printf("get all count signature fail, err: %v", err)
		return total, err
	}
	return total, nil
}

func CreateSignature(phone, street string) (err error) {
	_, err = GetDB().Exec("INSERT INTO signature (phone, street) VALUES (?, ?)", phone, street)
	if err != nil {
		log.Printf("create signature fail, err: %v", err)
		return err
	}
	return nil
}
