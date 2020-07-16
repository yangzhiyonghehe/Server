package mysqlitemanger

import (
	"fmt"

	"github.com/astaxie/beego"
	_ "github.com/mattn/go-sqlite3"

	"database/sql"
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

var DBPath = GetRootPath() + "/beegoServer.db"

type DeviceInfo struct {
	Id               int64
	Name             string
	DevcieId         string
	Ip               string
	Account          string
	Psw              string
	Type             int64
	CompanyID        int64
	Status           int64
	Version          string
	Firmware         string
	ActiveTime       int64
	IsAttendance     int64
	IsShowDoorStatus int64
	MultiCert        int64
}

func AddDevice(data DeviceInfo) (int64, error) {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return 0, err
	}

	strSQL := fmt.Sprintf("SELECT COUNT(*), ID FROM DEVICE_TABLE WHERE DEVICE_ID = '%s'", data.DevcieId)
	rows, err := db.Query(strSQL)
	if err != nil {
		db.Close()
		return 0, err
	}

	nId := int64(0)
	nCount := 0
	for rows.Next() {
		err := rows.Scan(&nCount, &nId)
		if err != nil {
			beego.Warning(err)
		}
	}

	if nCount > 0 {
		stmts, err := db.Prepare(`UPDATE DEVICE_TABLE 
								SET DEVICE_NAME = ?, IP = ?, ACCOUNT = ?, PSW = ?, TYPE = ?, COMPANYID = ?, 
									EFFECTIVE = ?, IS_ATTENDANCE = ? , IS_SHOW_DOOR_STATUS = ?, MULTI_CERT = ? 
								WHERE   DEVICE_ID = ?`)
		if err != nil {
			db.Close()
			return 0, err
		}

		res, err := stmts.Exec(data.Name, data.Ip, data.Account, data.Psw, data.Type, data.CompanyID, 1, data.IsAttendance, data.IsShowDoorStatus, data.MultiCert, data.DevcieId)
		if err != nil {
			db.Close()
			return 0, err
		}
		_, err = res.RowsAffected()
		if err != nil {
			db.Close()
			return 0, err
		}
		db.Close()
		return nId, nil
	}

	rows, err = db.Query("SELECT MAX(ID) FROM DEVICE_TABLE")
	if err != nil {
		db.Close()
		return 0, err
	}

	nId = 0
	for rows.Next() {
		rows.Scan(&nId)
	}

	nId++

	stm, err := db.Prepare(`INSERT INTO DEVICE_TABLE(ID,DEVICE_ID, DEVICE_NAME, IP, 
													ACCOUNT, PSW, TYPE, COMPANYID, EFFECTIVE, 
													STATUS, FIRWAREVER, IS_ATTENDANCE, IS_SHOW_DOOR_STATUS, MULTI_CERT) 
							VALUES(?,?, ?, ?, ? ,?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		db.Close()
		return 0, err
	}

	res, err := stm.Exec(nId, data.DevcieId, data.Name, data.Ip, data.Account, data.Psw, data.Type, data.CompanyID, 1, data.Status,
		data.Firmware, data.IsAttendance, data.IsShowDoorStatus, data.MultiCert)
	if err != nil {
		beego.Error("添加设备数据到表上失败：", err)
		db.Close()
		return 0, err
	}

	_, err = res.RowsAffected()
	if err != nil {
		beego.Error(err)
		db.Close()
		return 0, err
	}

	//beego.Debug(affect)
	db.Close()
	return nId, nil
}

func UpdateDevice(data DeviceInfo) error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return err
	}

	stm, err := db.Prepare(`UPDATE  DEVICE_TABLE 
							SET DEVICE_ID = ? ,DEVICE_NAME = ?, IP = ?, ACCOUNT = ?, 
							PSW = ?, STATUS = ?, IS_ATTENDANCE = ?, IS_SHOW_DOOR_STATUS = ?, MULTI_CERT = ? 
							WHERE ID = ?`)
	if err != nil {
		db.Close()
		return err
	}

	res, err := stm.Exec(data.DevcieId, data.Name, data.Ip, data.Account, data.Psw, data.Status, data.IsAttendance, data.IsShowDoorStatus, data.MultiCert, data.Id)
	if err != nil {
		beego.Error("添加设备数据到表上失败：", err)
		db.Close()
		return err
	}

	affect, err := res.RowsAffected()
	if err != nil {
		beego.Error(err)
		db.Close()
		return err
	}

	beego.Debug(affect)
	db.Close()

	return nil
}

func EnableDevice(strSerial string) error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return err
	}

	stm, err := db.Prepare("UPDATE  DEVICE_TABLE SET EFFECTIVE = 1 WHERE DEVICE_ID = ?")
	if err != nil {
		db.Close()
		return err
	}

	res, err := stm.Exec(strSerial)
	if err != nil {
		beego.Error("添加设备数据到表上失败：", err)
		db.Close()
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		beego.Error(err)
		db.Close()
		return err
	}

	//	beego.Debug(affect)
	db.Close()

	return nil
}

//UnableDevice 使设备失效
func UnableDevice(strSerial string) error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return err
	}

	stm, err := db.Prepare("UPDATE  DEVICE_TABLE SET EFFECTIVE = 0 WHERE DEVICE_ID = ?")
	if err != nil {
		db.Close()
		return err
	}

	res, err := stm.Exec(strSerial)
	if err != nil {
		beego.Error("添加设备数据到表上失败：", err)
		db.Close()
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		beego.Error(err)
		db.Close()
		return err
	}

	//beego.Debug(affect)
	db.Close()

	return nil
}

func GetDeviceList(nCompanyID int64) ([]DeviceInfo, error) {
	var structResult []DeviceInfo
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return structResult, err
	}

	strSql := fmt.Sprintf("SELECT ID,  DEVICE_ID, DEVICE_NAME, IP, ACCOUNT, PSW, TYPE, MULTI_CERT FROM DEVICE_TABLE WHERE COMPANYID = %d AND EFFECTIVE = 1 ", nCompanyID)
	rows, err := db.Query(strSql)
	if err != nil {
		db.Close()
		return structResult, err
	}

	for rows.Next() {
		var nId int64
		var nDeviceId string
		var strName string
		var strIp string
		var strAccount string
		var strPsw string
		var nType int64
		var nMultiCert int64

		err = rows.Scan(&nId, &nDeviceId, &strName, &strIp, &strAccount, &strPsw, &nType, &nMultiCert)
		if err != nil {
			beego.Error(err)
		}

		structResult = append(structResult, DeviceInfo{MultiCert: nMultiCert, Name: strName, Account: strAccount, Psw: strPsw, Id: nId, DevcieId: nDeviceId, Ip: strIp, Type: nType})
	}
	db.Close()
	return structResult, nil

}

func GetDeviceListByType(nCompanyID int64, nType int64) ([]DeviceInfo, error) {
	var structResult []DeviceInfo
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return structResult, err
	}

	strSql := fmt.Sprintf(`SELECT ID,  DEVICE_ID, DEVICE_NAME, IP, ACCOUNT, 
						  PSW, TYPE, STATUS, ACTIVE_TIME, IS_ATTENDANCE, IS_SHOW_DOOR_STATUS, MULTI_CERT 
						FROM DEVICE_TABLE 
						WHERE COMPANYID = %d AND TYPE = %d AND EFFECTIVE = 1`, nCompanyID, nType)
	rows, err := db.Query(strSql)
	if err != nil {
		db.Close()
		return structResult, err
	}

	for rows.Next() {
		var nId int64
		var nDeviceId string
		var strName string
		var strIp string
		var strAccount string
		var strPsw string
		var nType int64
		var nStatus int64
		var nActiveTime int64
		var nIsAttendance int64
		var nIsShowDoorStatus int64
		var nMultiCert int64

		err = rows.Scan(&nId, &nDeviceId, &strName, &strIp, &strAccount, &strPsw, &nType, &nStatus, &nActiveTime, &nIsAttendance, &nIsShowDoorStatus, &nMultiCert)
		if err != nil {
			beego.Warning(err)
		}

		structResult = append(structResult, DeviceInfo{MultiCert: nMultiCert, IsShowDoorStatus: nIsShowDoorStatus, IsAttendance: nIsAttendance, ActiveTime: nActiveTime, Name: strName, Account: strAccount, Psw: strPsw, Id: nId, DevcieId: nDeviceId, Ip: strIp, Type: nType, Status: nStatus})
	}
	db.Close()
	return structResult, nil

}

func GetDeviceById(nId int64) (DeviceInfo, error) {
	var structResult DeviceInfo

	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return structResult, err
	}

	strSQL := fmt.Sprintf("SELECT ID,  DEVICE_ID, DEVICE_NAME, IP, ACCOUNT, PSW, TYPE, FIRWAREVER, COMPANYID, MULTI_CERT FROM DEVICE_TABLE WHERE ID = %d AND EFFECTIVE = 1", nId)
	rows, err := db.Query(strSQL)
	if err != nil {
		db.Close()
		return structResult, err
	}

	for rows.Next() {
		var nId int64
		var nDeviceId string
		var strName string
		var strIp string
		var strAccount string
		var strPsw string
		var nType int64
		var strVersion string
		var nCompanyID int64
		var nMultiCert int64

		err = rows.Scan(&nId, &nDeviceId, &strName, &strIp, &strAccount, &strPsw, &nType, &strVersion, &nCompanyID, &nMultiCert)
		if err != nil {
			beego.Warning(err)
		}

		structResult.Id = nId
		structResult.Name = strName
		structResult.Ip = strIp
		structResult.Psw = strPsw
		structResult.Type = nType
		structResult.DevcieId = nDeviceId
		structResult.Account = strAccount
		structResult.Version = strVersion
		structResult.CompanyID = nCompanyID
		structResult.MultiCert = nMultiCert
	}
	db.Close()
	return structResult, nil
}

func GetDeviceByDeviceId(strDeviceId string) (DeviceInfo, error) {
	var structResult DeviceInfo

	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return structResult, err
	}

	strSql := fmt.Sprintf("SELECT ID,  DEVICE_ID, DEVICE_NAME, IP, ACCOUNT, PSW, TYPE, COMPANYID, IS_ATTENDANCE, MULTI_CERT FROM DEVICE_TABLE WHERE DEVICE_ID = '%s' AND EFFECTIVE = 1", strDeviceId)
	rows, err := db.Query(strSql)
	if err != nil {
		db.Close()
		return structResult, err
	}

	for rows.Next() {
		var nId int64
		var nDeviceId string
		var strName string
		var strIp string
		var strAccount string
		var strPsw string
		var nType int64
		var nCompanyID int64
		var nIsAttendance int64
		var nMultiCert int64

		err = rows.Scan(&nId, &nDeviceId, &strName, &strIp, &strAccount, &strPsw, &nType, &nCompanyID, &nIsAttendance, &nMultiCert)
		if err != nil {
			beego.Error(err)
		}

		structResult.Id = nId
		structResult.Name = strName
		structResult.Ip = strIp
		structResult.Psw = strPsw
		structResult.Type = nType
		structResult.DevcieId = nDeviceId
		structResult.Account = strAccount
		structResult.CompanyID = nCompanyID
		structResult.IsAttendance = nIsAttendance
		structResult.MultiCert = nMultiCert
	}
	db.Close()
	return structResult, nil
}

func RemoveDevice(strDeviceId string) error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return err
	}

	stm, err := db.Prepare("DELETE FROM DEVICE_TABLE WHERE DEVICE_ID = ?")
	if err != nil {
		db.Close()
		return err
	}

	res, err := stm.Exec(strDeviceId)

	affect, err := res.RowsAffected()
	if err != nil {
		beego.Error(err)
		db.Close()
		return err
	}

	beego.Debug(affect)
	db.Close()
	return nil
}

type UserIDInfo struct {
	XMUserID string
	HQUserID int64
}

func GetUserIDByXM(strID string) (int64, error) {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		//db.Close()
		return -1, err
	}

	strSQL := fmt.Sprintf("SELECT HQ_USERID FROM EMPLOYEE_ID_TABLE WHERE XM_USERID = '%s'", strID)
	rows, err := db.Query(strSQL)
	if err != nil {
		beego.Error(err)
		db.Close()
		return -1, err
	}

	nID := int64(-1)

	for rows.Next() {
		err = rows.Scan(&nID)
		if err != nil {
			beego.Warning(err)
		}
	}

	return nID, nil
}

func GetUserIDByHQ(nID int64) (string, error) {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		//db.Close()
		return "", err
	}

	strSQL := fmt.Sprintf("SELECT XM_USERID FROM EMPLOYEE_ID_TABLE WHERE HQ_USERID = %d", nID)
	rows, err := db.Query(strSQL)
	if err != nil {
		db.Close()
		return "", err
	}

	strID := ""

	for rows.Next() {
		err = rows.Scan(&strID)
		if err != nil {
			beego.Warning(err)
		}
	}

	return strID, nil
}

func InsertUserID(data UserIDInfo) error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		//db.Close()
		return err
	}

	stm, err := db.Prepare("INSERT INTO EMPLOYEE_ID_TABLE(XM_USERID, HQ_USERID) VALUES(?, ?)")
	if err != nil {
		//beego.Error(err)
		db.Close()
		return err
	}

	res, err := stm.Exec(data.XMUserID, data.HQUserID)
	if err != nil {
		//beego.Error(err)
		db.Close()
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

func DeleteByXM(strID string) error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		//db.Close()
		return err
	}

	//strSQL := fmt.Sprintf("DELETE FROM EMPLOYEE_ID_TABLE WHERE XM_USERID = %s", strID)
	stmt, err := db.Prepare("DELETE FROM EMPLOYEE_ID_TABLE WHERE XM_USERID = ?")
	if err != nil {
		db.Close()
		return err
	}

	res, err := stmt.Exec(strID)
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
	//DELETE FROM EMPLOYEE_ID_TABLE WHERE XM_USERID = ?

}
