package mysqlitemanger

import (
	"errors"

	"github.com/astaxie/beego"
	_ "github.com/mattn/go-sqlite3"

	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//GetRootPath  获取当前目录
func GetRootPath() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return dir
	}

	dir = strings.Replace(dir, "\\", "/", -1)

	return dir
}

//DeviceInfo 设备信息
type DeviceInfo struct {
	AppVer      string
	DeviceID    string
	DeviceType  string
	FirmwareVer string
	PassWord    string
	Addr        string
	Status      int64
	Type        int64
	Effect      int64
}

var DBPath = ""

//AddDeviceInfo 添加设备信息到表
func AddDeviceInfo(data DeviceInfo) error {
	if len(DBPath) == 0 {
		return errors.New("数据库路径未设置")
	}

	db, err := sql.Open("sqlite3", DBPath)
	if err != nil {
		//db.Close()
		return err
	}

	strSQL := fmt.Sprintf("SELECT COUNT(*) FROM DEVICE_TABLE WHERE DEVICE_ID = '%s'", data.DeviceID)
	rows, err := db.Query(strSQL)
	if err != nil {
		db.Close()
		return err
	}
	nDeviceCount := 0
	for rows.Next() {
		err = rows.Scan(&nDeviceCount)
		if err != nil {
			return err
		}
	}

	if nDeviceCount > 0 {
		stmt, err := db.Prepare("UPDATE DEVICE_TABLE SET APP_VER = ?, DEVICE_TYPE = ?, FIRWAREVER = ?, PSW = ?,  STATUS = ?, ADDR = ? WHERE  DEVICE_ID = ?")
		if err != nil {
			db.Close()
			return err
		}

		res, err := stmt.Exec(data.AppVer, data.DeviceType, data.FirmwareVer, data.PassWord, 1, data.Addr, data.DeviceID)
		if err != nil {
			db.Close()
			return err
		}

		_, err = res.RowsAffected()
		if err != nil {
			db.Close()
			return err
		}

		return nil
	}

	rows, err = db.Query("SELECT MAX(ID) FROM DEVICE_TABLE")
	if err != nil {
		db.Close()
		return err
	}

	nId := int64(0)
	for rows.Next() {
		rows.Scan(&nId)
	}

	nId++

	stm, err := db.Prepare("INSERT INTO DEVICE_TABLE(ID,DEVICE_ID, DEVICE_NAME, IP, ACCOUNT, PSW, TYPE, EFFECTIVE, STATUS, FIRWAREVER, DEVICE_TYPE, ADDR) VALUES(?, ?, ?, ? ,?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		db.Close()
		return err
	}

	res, err := stm.Exec(nId, data.DeviceID, "", "", "", data.PassWord, data.Type, 0, 1, data.FirmwareVer, data.DeviceType, data.Addr)
	if err != nil {
		beego.Error("添加设备数据到表上失败：", err)
		db.Close()
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		db.Close()
		return err
	}

	db.Close()
	return nil
}

//UpdateDeviceStatusByAddr 根据网络地址修改设备状态
func UpdateDeviceStatusByAddr(nStatus int64, strAddr string) error {
	if len(DBPath) == 0 {
		return errors.New("数据库路径未设置")
	}
	db, err := sql.Open("sqlite3", DBPath)
	if err != nil {
		//db.Close()
		return err
	}

	strSQL := fmt.Sprintf("SELECT COUNT(*) FROM DEVICE_TABLE WHERE ADDR = '%s'", strAddr)
	rows, err := db.Query(strSQL)
	if err != nil {
		return err
	}

	nCount := int64(0)
	for rows.Next() {

		err := rows.Scan(&nCount)
		if err != nil {
			return err
		}
	}

	if nCount == 0 {
		//strErr := fmt.Sprintf("数据库没有记录该设备, 设备地址: %s", strAddr)
		return nil
	}

	stmt, err := db.Prepare("UPDATE DEVICE_TABLE SET STATUS = ? WHERE ADDR = ?")
	if err != nil {
		return err
	}

	res, err := stmt.Exec(nStatus, strAddr)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

func UpdateDeviceStatusBySerial(nStatus int64, strSerial string) error {
	if len(DBPath) == 0 {
		return errors.New("数据库路径未设置")
	}
	db, err := sql.Open("sqlite3", DBPath)
	if err != nil {
		//db.Close()
		return err
	}

	strSQL := fmt.Sprintf("SELECT COUNT(*) FROM DEVICE_TABLE WHERE DEVICE_ID = '%s'", strSerial)
	rows, err := db.Query(strSQL)
	if err != nil {
		return err
	}

	nCount := int64(0)
	for rows.Next() {

		err := rows.Scan(&nCount)
		if err != nil {
			return err
		}
	}

	if nCount == 0 {
		//strErr := fmt.Sprintf("数据库没有记录该设备, 设备地址: %s", strAddr)
		return nil
	}

	stmt, err := db.Prepare("UPDATE DEVICE_TABLE SET STATUS = ? WHERE DEVICE_ID = ?")
	if err != nil {
		return err
	}

	res, err := stmt.Exec(nStatus, strSerial)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

func UpdateDeviceActiveTime(nTime int64, strSerial string) error {
	if len(DBPath) == 0 {
		return errors.New("数据库路径未设置")
	}
	db, err := sql.Open("sqlite3", DBPath)
	if err != nil {
		//db.Close()
		return err
	}

	strSQL := fmt.Sprintf("SELECT COUNT(*) FROM DEVICE_TABLE WHERE DEVICE_ID = '%s'", strSerial)
	rows, err := db.Query(strSQL)
	if err != nil {
		return err
	}

	nCount := int64(0)
	for rows.Next() {

		err := rows.Scan(&nCount)
		if err != nil {
			return err
		}
	}

	if nCount == 0 {
		//strErr := fmt.Sprintf("数据库没有记录该设备, 设备地址: %s", strAddr)
		return nil
	}

	stmt, err := db.Prepare("UPDATE DEVICE_TABLE SET ACTIVE_TIME = ? WHERE DEVICE_ID = ?")
	if err != nil {
		return err
	}

	res, err := stmt.Exec(nTime, strSerial)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return nil

}

//GetDeviceBySerial 根据序列号获取设备信息
func GetDeviceBySerial(strSerial string) (DeviceInfo, error) {
	var structResutl DeviceInfo

	if len(DBPath) == 0 {
		return structResutl, errors.New("数据库路径未设置")
	}
	db, err := sql.Open("sqlite3", DBPath)
	if err != nil {
		//db.Close()
		return structResutl, err
	}

	strSQL := fmt.Sprintf(`SELECT DEVICE_ID, DEVICE_NAME, IP, ACCOUNT, PSW, TYPE, ID,
	 STATUS, EFFECTIVE,APP_VER, DEVICE_TYPE, FIRWAREVER, ADDR FROM DEVICE_TABLE WHERE DEVICE_ID = '%s'`, strSerial)
	rows, err := db.Query(strSQL)
	if err != nil {
		return structResutl, err
	}

	for rows.Next() {
		var strDeviceID string
		var strDeviceName string
		var strIP string
		var strAccount string
		var strPSW string
		var nType int64
		var nID int64
		var nStatus int64
		var strAppVer string
		var strDeviceType string
		var strFirmware string
		var strAddr string
		var nEffect int64

		err = rows.Scan(&strDeviceID, &strDeviceName, &strIP, &strAccount, &strPSW, &nType,
			&nID, &nStatus, &nEffect, &strAppVer, &strDeviceType, &strFirmware, &strAddr)
		if err != nil {
			beego.Warning(err)
		}

		structResutl.Addr = strAddr
		structResutl.AppVer = strAppVer
		structResutl.DeviceID = strDeviceID
		structResutl.DeviceType = strDeviceType
		structResutl.FirmwareVer = strFirmware
		structResutl.PassWord = strPSW
		structResutl.Status = nStatus
		structResutl.Effect = nEffect

	}

	return structResutl, nil
}

func GetDeviceByAddr(strAddr string) (DeviceInfo, error) {
	var structResutl DeviceInfo

	if len(DBPath) == 0 {
		return structResutl, errors.New("数据库路径未设置")
	}
	db, err := sql.Open("sqlite3", DBPath)
	if err != nil {
		//db.Close()
		return structResutl, err
	}

	strSQL := fmt.Sprintf(`SELECT DEVICE_ID, DEVICE_NAME, IP, ACCOUNT, PSW, TYPE, ID,
	 STATUS, APP_VER, DEVICE_TYPE, FIRWAREVER, ADDR, EFFECTIVE FROM DEVICE_TABLE WHERE ADDR = '%s'`, strAddr)
	rows, err := db.Query(strSQL)
	if err != nil {
		return structResutl, err
	}

	for rows.Next() {
		var strDeviceID string
		var strDeviceName string
		var strIP string
		var strAccount string
		var strPSW string
		var nType int64
		var nID int64
		var nStatus int64
		var strAppVer string
		var strDeviceType string
		var strFirmware string
		var strAddr string
		var nEffect int64

		err = rows.Scan(&strDeviceID, &strDeviceName, &strIP, &strAccount, &strPSW, &nType,
			&nID, &nStatus, &strAppVer, &strDeviceType, &strFirmware, &strAddr, &nEffect)
		if err != nil {
			beego.Warning(err)
		}

		structResutl.Addr = strAddr
		structResutl.AppVer = strAppVer
		structResutl.DeviceID = strDeviceID
		structResutl.DeviceType = strDeviceType
		structResutl.FirmwareVer = strFirmware
		structResutl.PassWord = strPSW
		structResutl.Status = nStatus
		structResutl.Effect = nEffect

	}

	return structResutl, nil

}
