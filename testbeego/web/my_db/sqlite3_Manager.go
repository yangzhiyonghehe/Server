package my_db

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	_ "github.com/mattn/go-sqlite3"
)

type AdminInfo struct {
	Name      string
	Account   string
	PhoneName string
	Psw       string
	CompanyId int64
	ISRoot    int64
	Companys  string
}

func GetRootPath() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return dir
	}

	dir = strings.Replace(dir, "\\", "/", -1)

	return dir
}

func GetAdminUserList() ([]AdminInfo, error) {
	var structResult []AdminInfo

	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		beego.Error(err)
		return structResult, err
	}

	rows, err := db.Query("SELECT NAME, ACCOUNT, PHONE_NUM, PSW FROM ADMIN_USER")
	if err != nil {
		db.Close()
		beego.Error(err)
		return structResult, err
	}

	for rows.Next() {

		var strName string
		var strAccount string
		var strPhoneNum string
		var strPsw string

		err = rows.Scan(&strName, &strAccount, &strPhoneNum, &strPsw)
		if err != nil {
			beego.Error(err)
		}

		structResult = append(structResult, AdminInfo{Name: strName, Account: strAccount, PhoneName: strPhoneNum, Psw: strPsw})

	}
	db.Close()
	return structResult, nil
}

func GetAdminByAccount(strAccount string) (AdminInfo, error) {
	var structResult AdminInfo

	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return structResult, err
	}

	strSql := fmt.Sprintf("SELECT NAME, ACCOUNT, PHONE_NUM, PSW , COMPANY_ID, IS_ROOT, COMPANYS FROM ADMIN_USER WHERE ACCOUNT = '%s'", strAccount)
	rows, err := db.Query(strSql)
	if err != nil {
		db.Close()
		return structResult, err
	}

	for rows.Next() {
		var strName string
		var strAccount string
		var strPhoneNum string
		var strPsw string
		var nCompanyId int64
		var nISRoot int64
		var strCompanys string

		err = rows.Scan(&strName, &strAccount, &strPhoneNum, &strPsw, &nCompanyId, &nISRoot, &strCompanys)
		if err != nil {
			beego.Warning(err)
		}

		structResult.Name = strName
		structResult.Account = strAccount
		structResult.CompanyId = nCompanyId
		structResult.PhoneName = strPhoneNum
		structResult.Psw = strPsw
		structResult.ISRoot = nISRoot
		structResult.Companys = strCompanys
	}

	db.Close()
	return structResult, nil
}

func AddAdminUser(data AdminInfo) error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return err
	}

	strSQL := fmt.Sprintf("SELECT COUNT(*)  FROM ADMIN_USER WHERE ACCOUNT = %s", data.Account)
	rows, err := db.Query(strSQL)
	if err != nil {
		return err
	}

	nCount := 0
	for rows.Next() {

		err = rows.Scan(&nCount)
		if err != nil {
			beego.Warning(err)
		}
	}

	if nCount > 0 {
		return errors.New("已存在该用户(手机号)")
	}

	stm, err := db.Prepare("INSERT INTO ADMIN_USER(ACCOUNT, PHONE_NUM, PSW, NAME, COMPANY_ID, IS_ROOT, COMPANYS) VALUES(?, ?, ?,?, ?, ?, ?)")
	if err != nil {
		db.Close()
		return err
	}

	res, err := stm.Exec(data.Account, data.PhoneName, data.Psw, data.Name, data.CompanyId, data.ISRoot, data.Companys)
	if err != nil {
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

func UpdateAdminInfo(data AdminInfo) error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return err
	}

	stm, err := db.Prepare("UPDATE ADMIN_USER SET  COMPANY_ID = ?, COMPANYS = ?  WHERE ACCOUNT = ?")
	if err != nil {
		db.Close()
		return err
	}

	res, err := stm.Exec(data.CompanyId, data.Companys, data.Account)

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

func DeleteAdminUser(strAccount string) error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return err
	}

	stm, err := db.Prepare("DELETE FROM ADMIN_USER WHERE ACCOUNT = ? AND IS_ROOT != 1")
	if err != nil {
		db.Close()
		return err
	}

	res, err := stm.Exec(strAccount)
	if err != nil {
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

type TokenInfo struct {
	Token      string
	CreateTime string
}

func AddToken(data TokenInfo) error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return err
	}

	stm, err := db.Prepare("INSERT INTO TOKEN_TABLE(TOKEN, CREATE_TIME) VALUES(?, ?)")
	if err != nil {
		return err
	}

	res, err := stm.Exec(data.Token, data.CreateTime)

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

func GetToken(strToken string) (TokenInfo, error) {
	var structToken TokenInfo

	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return structToken, err
	}

	strSql := fmt.Sprintf("SELECT TOKEN, CREATE_TIME FROM TOKEN_TABLE WHERE TOKEN = '%s'", strToken)
	rows, err := db.Query(strSql)
	if err != nil {
		db.Close()
		return structToken, err
	}

	for rows.Next() {
		var strTokenIndex string
		var strCreateTimeIndex string
		rows.Scan(&strTokenIndex, &strCreateTimeIndex)

		structToken.CreateTime = strCreateTimeIndex
		structToken.Token = strTokenIndex
	}
	db.Close()
	return structToken, nil
}

func UpdateToken(data TokenInfo) error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return err
	}

	stmt, err := db.Prepare("UPDATE TOKEN_TABLE SET TOKEN = '?' , CREATE_TIME = '?'")
	if err != nil {
		db.Close()
		return err
	}

	res, err := stmt.Exec(data.Token, data.CreateTime)
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

func RemoveToken(strToken string) error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return err
	}

	stmt, err := db.Prepare("DELETE FROM TOKEN_TABLE WHERE TOKEN = ?")
	if err != nil {
		db.Close()
		return err
	}

	res, err := stmt.Exec(strToken)
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

type DeviceAndPersonInfo struct {
	DeviceId   int64
	EmployeeId []int64
}

type EmployeeAndDeviceInfo struct {
	EmployeeId int64
	DevcieId   []int64
}

func AddEmployeeAndDevice(data EmployeeAndDeviceInfo) error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return err
	}
	for _, nDeviceIdIndex := range data.DevcieId {
		stm, err := db.Prepare("INSERT INTO DEVICE_AND_EMPLOYEE(DEVICE_ID,EMPLOYEE_ID) VALUES(?, ?)")
		if err != nil {
			db.Close()
			return err
		}

		res, err := stm.Exec(nDeviceIdIndex, data.EmployeeId)

		affect, err := res.RowsAffected()
		if err != nil {
			db.Close()
			return err
		}

		beego.Debug(affect)
	}

	db.Close()

	return nil
}

func AddDviceAndEmployee(data DeviceAndPersonInfo) error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return err
	}
	for _, nEmployeeIdIndex := range data.EmployeeId {
		stm, err := db.Prepare("INSERT INTO DEVICE_AND_EMPLOYEE(DEVICE_ID,EMPLOYEE_ID) VALUES(?, ?)")
		if err != nil {
			db.Close()
			return err
		}

		res, err := stm.Exec(data.DeviceId, nEmployeeIdIndex)

		_, err = res.RowsAffected()
		if err != nil {
			db.Close()
			return err
		}

		//		beego.Debug(affect)
	}

	db.Close()

	return nil
}

func UpdateDeviceAndEmployee(data DeviceAndPersonInfo) error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return err
	}

	stmt, err := db.Prepare("DELETE FROM DEVICE_AND_EMPLOYEE WHERE EMPLOYEE_ID = ?")
	if err != nil {
		db.Close()
		return err
	}

	res, err := stmt.Exec(data.DeviceId)
	if err != nil {
		db.Close()
		return err
	}

	affect, err := res.RowsAffected()
	if err != nil {
		db.Close()
		return err
	}

	beego.Debug(affect)

	for _, employeeIndex := range data.EmployeeId {
		stmt_, err := db.Prepare("INSERT INTO DEVICE_AND_EMPLOYEE(DEVICE_ID,EMPLOYEE_ID) VALUES(?, ?)")
		res, err := stmt_.Exec(data.DeviceId, employeeIndex)
		if err != nil {
			db.Close()
			return err
		}
		affect_, err := res.RowsAffected()
		if err != nil {
			db.Close()
			return err
		}
		beego.Debug(affect_)
	}
	db.Close()
	return nil
}

func GetDeviceAndEmployee(nDeviceId int64) (DeviceAndPersonInfo, error) {
	var structResult DeviceAndPersonInfo
	structResult.DeviceId = nDeviceId
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return structResult, err
	}

	strSql := fmt.Sprintf("SELECT EMPLOYEE_ID FROM DEVICE_AND_EMPLOYEE WHERE DEVICE_ID = '%d'", nDeviceId)
	rows, err := db.Query(strSql)
	if err != nil {
		db.Close()
		return structResult, err
	}

	for rows.Next() {
		var nEmployee int64
		err = rows.Scan(&nEmployee)
		if err != nil {
			beego.Error(err)
		}

		structResult.EmployeeId = append(structResult.EmployeeId, nEmployee)
	}
	db.Close()
	return structResult, nil
}

func GetEmployeeAndDvice(nEmployeeId int64) (EmployeeAndDeviceInfo, error) {
	var structResult EmployeeAndDeviceInfo
	structResult.EmployeeId = nEmployeeId
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return structResult, err
	}

	strSql := fmt.Sprintf("SELECT DEVICE_ID  FROM DEVICE_AND_EMPLOYEE WHERE EMPLOYEE_ID = %d", nEmployeeId)
	rows, err := db.Query(strSql)
	if err != nil {
		db.Close()
		return structResult, err
	}

	for rows.Next() {
		var nDeviceId int64
		err = rows.Scan(&nDeviceId)
		if err != nil {
			beego.Error(err)
		}

		structResult.DevcieId = append(structResult.DevcieId, nDeviceId)
	}
	db.Close()
	return structResult, nil
}

func ISEmployeeAndDevice(nEmployeeID int64, nDeviceID int64) (bool, error) {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return false, err
	}

	strSQL := fmt.Sprintf("SELECT COUNT(*) FROM DEVICE_AND_EMPLOYEE WHERE DEVICE_ID = %d AND EMPLOYEE_ID = %d", nDeviceID, nEmployeeID)

	nCount := int64(0)
	rows, err := db.Query(strSQL)
	for rows.Next() {
		err = rows.Scan(&nCount)
		if err != nil {
			beego.Warning(err)
		}
	}

	if nCount <= 0 {
		return false, nil
	}

	return true, nil
}

func RemoveDeviceAndEmployee(nDeviceId int64) error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return err
	}

	stm, err := db.Prepare("DELETE FROM DEVICE_AND_EMPLOYEE WHERE DEVICE_ID = ?")
	if err != nil {
		return err
	}

	res, err := stm.Exec(nDeviceId)

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

func RemoveDeviceAndEmployeeByEmployeeId(nEmployeeId int64) error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return err
	}

	stm, err := db.Prepare("DELETE FROM DEVICE_AND_EMPLOYEE WHERE EMPLOYEE_ID = ?")
	if err != nil {
		return err
	}

	res, err := stm.Exec(nEmployeeId)

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

func RemoveDeviceAndEmoloyeeByAll(nEmployeeId int64, nDeviceId int64) error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return err
	}

	stm, err := db.Prepare("DELETE FROM DEVICE_AND_EMPLOYEE WHERE EMPLOYEE_ID = ? AND DEVICE_ID = ?")
	if err != nil {
		return err
	}

	res, err := stm.Exec(nEmployeeId, nDeviceId)

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

type GroupInfo struct {
	Id      string
	Name    string
	Sort    int64
	Purview int64
	Parent  int64
}

func GetGroupList() ([]GroupInfo, error) {
	var structResult []GroupInfo
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return structResult, err
	}

	rows, err := db.Query("SELECT id, name, sort FROM ORGNIZATION_INFO")
	if err != nil {
		db.Close()
		return structResult, err
	}

	for rows.Next() {
		var nId int64
		var strName string
		var nSort int64

		err = rows.Scan(&nId, &strName, &nSort)
		if err != nil {
			beego.Error(err)
		}

		structResult = append(structResult, GroupInfo{Name: strName, Id: fmt.Sprintf("org-%d", nId), Sort: nSort})
	}
	db.Close()
	return structResult, nil
}

func GetGroupListByPurview(nPurview int) ([]GroupInfo, error) {
	var structResult []GroupInfo
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return structResult, err
	}

	strSql := fmt.Sprintf("SELECT id, name, sort FROM ORGNIZATION_INFO WHERE purview = %d", nPurview)
	rows, err := db.Query(strSql)
	if err != nil {
		db.Close()
		return structResult, err
	}

	for rows.Next() {
		var nId int64
		var strName string
		var nSort int64
		err = rows.Scan(&nId, &strName, &nSort)
		if err != nil {
			beego.Error(err)
		}

		structResult = append(structResult, GroupInfo{Name: strName, Id: fmt.Sprintf("org-%d", nId), Sort: nSort})
	}

	db.Close()
	return structResult, nil
}

func GetGroupListByParent(strParent string) ([]GroupInfo, error) {
	var structResult []GroupInfo

	arrayIndex := strings.Split(strParent, "-")
	nParentId, err := strconv.ParseInt(arrayIndex[1], 0, 64)
	if err != nil {
		return structResult, err
	}

	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return structResult, err
	}

	strSql := fmt.Sprintf("SELECT  id,  name, sort FROM ORGNIZATION_INFO WHERE parent = '%d'", nParentId)
	rows, err := db.Query(strSql)
	if err != nil {
		db.Close()
		return structResult, err
	}

	for rows.Next() {
		var nId int64
		var strName string
		var nSort int64

		err = rows.Scan(&nId, &strName, &nSort)
		if err != nil {
			beego.Error(err)
		}

		structResult = append(structResult, GroupInfo{Name: strName, Id: fmt.Sprintf("org-%d", nId), Sort: nSort})
	}

	db.Close()
	return structResult, nil
}

func GetGroupById(strId string) (GroupInfo, error) {
	var structResult GroupInfo

	arrayIndex := strings.Split(strId, "-")
	nId, err := strconv.ParseInt(arrayIndex[1], 0, 64)
	if err != nil {
		return structResult, err
	}

	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return structResult, err
	}

	strSql := fmt.Sprintf("SELECT name, sort FROM ORGNIZATION_INFO WHERE id = '%d'", nId)
	rows, err := db.Query(strSql)
	if err != nil {
		db.Close()
		return structResult, err
	}

	for rows.Next() {
		var strName string
		var nSort int64

		err = rows.Scan(&strName, &nSort)
		if err != nil {
			beego.Error(err)
		}
		structResult.Id = fmt.Sprintf("org-%d", nId)
		structResult.Name = strName
	}

	db.Close()
	return structResult, nil
}

func AddGroup(data GroupInfo) (int64, error) {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return 0, err
	}

	rows, err := db.Query("SELECT MAX(id) FROM ORGNIZATION_INFO")
	if err != nil {
		return 0, err
	}

	nId := int64(0)
	for rows.Next() {
		rows.Scan(&nId)
	}

	nId++

	stm, err := db.Prepare("INSERT INTO ORGNIZATION_INFO(id, name, purview, parent, sort)  VALUES(?, ?, ?, ?,?)")
	if err != nil {
		return 0, err
	}

	res, err := stm.Exec(nId, data.Name, data.Purview, data.Parent, data.Sort)
	if err != nil {
		return 0, err
	}

	affect, err := res.RowsAffected()
	if err != nil {
		beego.Error(err)
		db.Close()
		return 0, err
	}

	beego.Debug(affect)
	db.Close()
	return nId, nil
}

func UpdateGroup(data GroupInfo) error {
	arrayIndex := strings.Split(data.Id, "-")
	nId, err := strconv.ParseInt(arrayIndex[1], 0, 64)
	if err != nil {
		return err
	}

	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return err
	}

	stm, err := db.Prepare("UPDATE ORGNIZATION_INFO SET  name = ?, sort = ? WHERE id = ?")
	if err != nil {
		return err
	}

	res, err := stm.Exec(data.Name, data.Sort, nId)
	if err != nil {
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

func RemoveGroup(nGroupId int64) error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return err
	}

	stm, err := db.Prepare("DELETE FROM ORGNIZATION_INFO WHERE id = ?")
	if err != nil {
		return err
	}

	res, err := stm.Exec(nGroupId)

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

type EmployeeInfo struct {
	Id           int64
	Name         string
	Phone        string
	Pic          string
	Sort         int64
	Status       int64
	Rights       string
	RemoteDevice string
	Incharges    string
	IsAdmin      int64
	RuleID       int64
	TagID        int64
	Working      int64
	AdminAccount string
	CompanyID    int64
}

func AddEmployee(data EmployeeInfo) (int64, error) {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return 0, err
	}

	rows, err := db.Query("SELECT MAX(id) FROM EMPLOYEE_INFO")
	if err != nil {
		db.Close()
		return 0, err
	}

	nId := int64(0)
	for rows.Next() {
		rows.Scan(&nId)
	}

	nId++

	stm, err := db.Prepare(`INSERT INTO EMPLOYEE_INFO(id, name, phone_num, status, img_data, sort, remote_devices, 
					in_charges, rights, is_admin, rule_id, tag_id, working, admin_account, COMPANY_ID) 
					VALUES(?,?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		db.Close()
		return 0, err
	}

	res, err := stm.Exec(nId, data.Name, data.Phone, data.Status, data.Pic, data.Sort, data.RemoteDevice,
		data.Incharges, data.Rights, data.IsAdmin, data.RuleID, data.TagID, data.Working, data.AdminAccount, data.CompanyID)

	affect, err := res.RowsAffected()
	if err != nil {
		beego.Error(err)
		db.Close()
		return 0, err
	}

	beego.Debug(affect)
	db.Close()
	return nId, nil

}

func UpdateEmployee(data EmployeeInfo) error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return err
	}

	stm, err := db.Prepare(`UPDATE EMPLOYEE_INFO SET name = ?, phone_num = ?, status = ?, img_data = ?, sort = ?, 
						remote_devices = ?, in_charges = ?, rights = ?, is_admin = ?, rule_id = ? , tag_id = ?,
						working = ?, admin_account = ? 
						WHERE id = ?`)
	if err != nil {
		return err
	}

	res, err := stm.Exec(data.Name, data.Phone, data.Status, data.Pic, data.Sort, data.RemoteDevice, data.Incharges,
		data.Rights, data.IsAdmin, data.RuleID, data.TagID, data.Working, data.AdminAccount, data.Id)

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

func UpdateEmployeeRuleID(nEmployeeID int64, nRuleID int64) error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return err
	}

	stm, err := db.Prepare(`UPDATE EMPLOYEE_INFO SET rule_id = ?  WHERE id = ?`)
	if err != nil {
		return err
	}

	res, err := stm.Exec(nRuleID, nEmployeeID)

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

func RemoveEmployeeRuleID(nRuleID int64) error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return err
	}

	stm, err := db.Prepare(`UPDATE EMPLOYEE_INFO SET rule_id = ? WHERE rule_id = ?`)
	if err != nil {
		return err
	}

	res, err := stm.Exec(0, nRuleID)

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

func UpdateEmployeeTagID(nEmployeeID int64, nTagID int64) error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return err
	}

	stm, err := db.Prepare(`UPDATE EMPLOYEE_INFO SET tag_id = ? WHERE id = ?`)
	if err != nil {
		return err
	}

	res, err := stm.Exec(nTagID, nEmployeeID)

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

func RemoveEmployeeTag(nTagID int64) error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return err
	}

	stm, err := db.Prepare(`UPDATE EMPLOYEE_INFO SET tag_id = ? WHERE tag_id = ?`)
	if err != nil {
		return err
	}

	res, err := stm.Exec(0, nTagID)

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

func RemoveEmployeeById(nId int64) error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return err
	}

	stm, err := db.Prepare("DELETE  FROM EMPLOYEE_INFO WHERE id = ?")
	if err != nil {
		return err
	}

	res, err := stm.Exec(nId)

	_, err = res.RowsAffected()
	if err != nil {
		beego.Error(err)
		db.Close()
		return err
	}

	db.Close()
	return nil
}

func GetEmployeeById(nEmployeeId int64) (EmployeeInfo, error) {
	var structResult EmployeeInfo
	structResult.RuleID = 0

	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return structResult, err
	}

	strSql := fmt.Sprintf(`SELECT  name, phone_num, status, img_data, sort, status, remote_devices, in_charges, rights, is_admin, rule_id, tag_id, 
						working, admin_account
						FROM EMPLOYEE_INFO WHERE id = %d`, nEmployeeId)
	rows, err := db.Query(strSql)
	if err != nil {
		db.Close()
		return structResult, err
	}

	for rows.Next() {
		var strName string
		var strPhone string
		var nStatus int64
		var strImg string
		var nSort int64
		var strRemoteDevices string
		var strInchages string
		var strRights string
		var nIsAdmin int64
		var nRuleID int64
		var nTagID int64
		var nWorking int64
		var strAdminAccount string

		err = rows.Scan(&strName, &strPhone, &nStatus, &strImg, &nSort, &nStatus, &strRemoteDevices, &strInchages,
			&strRights, &nIsAdmin, &nRuleID, &nTagID, &nWorking, &strAdminAccount)
		if err != nil {
			beego.Warning(err)
		}

		structResult.Id = nEmployeeId
		structResult.Name = strName
		structResult.Phone = strPhone
		structResult.Pic = strImg
		structResult.Sort = nSort
		structResult.Status = nStatus
		structResult.RemoteDevice = strRemoteDevices
		structResult.Rights = strRights
		structResult.Incharges = strInchages
		structResult.IsAdmin = nIsAdmin
		structResult.RuleID = nRuleID
		structResult.TagID = nTagID
		structResult.AdminAccount = strAdminAccount
		structResult.Working = nWorking
	}

	db.Close()
	return structResult, nil
}

func GetEmployeeByAdminAccount(strAccount string, nCompanyID int64) (EmployeeInfo, error) {
	var structResult EmployeeInfo

	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return structResult, err
	}

	strSql := fmt.Sprintf(`SELECT  name, phone_num, status, img_data, sort, status, remote_devices, in_charges, rights, is_admin, rule_id, tag_id, 
						working, admin_account, id
						FROM EMPLOYEE_INFO WHERE admin_account = '%s' AND COMPANY_ID = %d`, strAccount, nCompanyID)
	rows, err := db.Query(strSql)
	if err != nil {
		db.Close()
		return structResult, err
	}

	for rows.Next() {
		var strName string
		var strPhone string
		var nStatus int64
		var strImg string
		var nSort int64
		var strRemoteDevices string
		var strInchages string
		var strRights string
		var nIsAdmin int64
		var nRuleID int64
		var nTagID int64
		var nWorking int64
		var strAdminAccount string
		var nEmployeeID int64

		err = rows.Scan(&strName, &strPhone, &nStatus, &strImg, &nSort, &nStatus, &strRemoteDevices, &strInchages,
			&strRights, &nIsAdmin, &nRuleID, &nTagID, &nWorking, &strAdminAccount, &nEmployeeID)
		if err != nil {
			beego.Warning(err)
		}

		structResult.Id = nEmployeeID
		structResult.Name = strName
		structResult.Phone = strPhone
		structResult.Pic = strImg
		structResult.Sort = nSort
		structResult.Status = nStatus
		structResult.RemoteDevice = strRemoteDevices
		structResult.Rights = strRights
		structResult.Incharges = strInchages
		structResult.IsAdmin = nIsAdmin
		structResult.RuleID = nRuleID
		structResult.TagID = nTagID
		structResult.AdminAccount = strAdminAccount
		structResult.Working = nWorking
	}

	db.Close()
	return structResult, nil
}

func GetEmployeeList(nCompanyID int64, strKey string) ([]EmployeeInfo, error) {
	var ResultArray []EmployeeInfo

	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return ResultArray, err
	}

	strSQL := fmt.Sprintf(`SELECT  id, name, phone_num, status, img_data, sort, status,remote_devices, in_charges, 
	rights, is_admin,rule_id, tag_id, working, admin_account 
	FROM EMPLOYEE_INFO WHERE COMPANY_ID = %d`, nCompanyID)

	if len(strKey) > 0 {
		strSQL += ("  AND name like " + "'%" + strKey + "%'")
		// strSQL = fmt.Sprintf(`SELECT  id, name, phone_num, status, img_data, sort, status,remote_devices, in_charges,
		// rights, is_admin,rule_id, tag_id, working, admin_account
		// FROM EMPLOYEE_INFO WHERE COMPANY_ID = %d AND name like '%s' `, nCompanyID, strKey)
	}

	//	beego.Debug(strSQL)

	rows, err := db.Query(strSQL)
	if err != nil {
		db.Close()
		return ResultArray, err
	}

	for rows.Next() {
		var nEmployeeId int64
		var strName string
		var strPhone string
		var nStatus int
		var strImg string
		var nSort int64
		var strStatus int64
		var strRemoteDevices string
		var strInchages string
		var strRights string
		var nIsAdmin int64
		var nRuleID int64
		var nTagID int64
		var nWorking int64
		var strAdminAccount string

		err = rows.Scan(&nEmployeeId, &strName, &strPhone, &nStatus, &strImg, &nSort, &strStatus, &strRemoteDevices,
			&strInchages, &strRights, &nIsAdmin, &nRuleID, &nTagID, &nWorking, &strAdminAccount)
		if err != nil {
			beego.Error(err)
		}

		var structResult EmployeeInfo

		structResult.Id = nEmployeeId
		structResult.Name = strName
		structResult.Phone = strPhone
		structResult.Pic = strImg
		structResult.Sort = nSort
		structResult.Status = strStatus
		structResult.RemoteDevice = strRemoteDevices
		structResult.Rights = strRights
		structResult.Incharges = strInchages
		structResult.IsAdmin = nIsAdmin
		structResult.RuleID = nRuleID
		structResult.TagID = nTagID
		structResult.Working = nWorking
		structResult.AdminAccount = strAdminAccount

		ResultArray = append(ResultArray, structResult)
	}

	db.Close()

	return ResultArray, nil
}

func GetGroupEmployeeList(nCompanyID int64, strGroupID string, strKey string) ([]EmployeeInfo, error) {
	var ResultArray []EmployeeInfo

	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return ResultArray, err
	}

	strSQL := fmt.Sprintf(`SELECT  A.id, A.name, A.phone_num, A.status, A.img_data, A.sort, A.status,A.remote_devices, A.in_charges, 
	A.rights, A.is_admin,A.rule_id, A.tag_id, A.working, A.admin_account 
	FROM EMPLOYEE_INFO A , GROUP_AND_EMPLOYEE B WHERE A.id = B.EMPLOYEE_ID AND B.GROUP_ID = '%s' AND A.COMPANY_ID = %d`, strGroupID, nCompanyID)

	if len(strKey) > 0 {
		strSQL += (" AND A.name LIKE '%" + strKey + "%'")
		// strSQL = fmt.Sprintf(`SELECT  A.id, A.name, A.phone_num, A.status, A.img_data, A.sort, A.status,A.remote_devices, A.in_charges,
		// A.rights, A.is_admin,A.rule_id, A.tag_id, A.working, A.admin_account
		// FROM EMPLOYEE_INFO A , GROUP_AND_EMPLOYEE B WHERE A.id == B.EMPLOYEE_ID AND B.GROUP_ID == '%s' AND A.name LIKE '%%s%'`, strGroupID, strKey)
	}

	rows, err := db.Query(strSQL)
	if err != nil {
		db.Close()
		return ResultArray, err
	}

	for rows.Next() {
		var nEmployeeId int64
		var strName string
		var strPhone string
		var nStatus int
		var strImg string
		var nSort int64
		var strStatus int64
		var strRemoteDevices string
		var strInchages string
		var strRights string
		var nIsAdmin int64
		var nRuleID int64
		var nTagID int64
		var nWorking int64
		var strAdminAccount string

		err = rows.Scan(&nEmployeeId, &strName, &strPhone, &nStatus, &strImg, &nSort, &strStatus, &strRemoteDevices,
			&strInchages, &strRights, &nIsAdmin, &nRuleID, &nTagID, &nWorking, &strAdminAccount)
		if err != nil {
			beego.Error(err)
		}

		var structResult EmployeeInfo

		structResult.Id = nEmployeeId
		structResult.Name = strName
		structResult.Phone = strPhone
		structResult.Pic = strImg
		structResult.Sort = nSort
		structResult.Status = strStatus
		structResult.RemoteDevice = strRemoteDevices
		structResult.Rights = strRights
		structResult.Incharges = strInchages
		structResult.IsAdmin = nIsAdmin
		structResult.RuleID = nRuleID
		structResult.TagID = nTagID
		structResult.Working = nWorking
		structResult.AdminAccount = strAdminAccount

		ResultArray = append(ResultArray, structResult)
	}

	db.Close()

	return ResultArray, nil

}

func GetEmployeeListByRuleID(nRuleID int64) ([]int64, error) {
	var ResultArray []int64

	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return ResultArray, err
	}

	strSql := fmt.Sprintf(`SELECT  id  FROM EMPLOYEE_INFO WHERE rule_id = %d`, nRuleID)
	rows, err := db.Query(strSql)
	if err != nil {
		db.Close()
		return ResultArray, err
	}

	for rows.Next() {
		var nEmployeeId int64

		err = rows.Scan(&nEmployeeId)
		if err != nil {
			beego.Error(err)
		}

		ResultArray = append(ResultArray, nEmployeeId)
	}

	db.Close()

	return ResultArray, nil
}

func GetEmployeeIDByTagID(nTagID int64) ([]int64, error) {
	var ResultArray []int64

	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return ResultArray, err
	}

	strSql := fmt.Sprintf(`SELECT  id  FROM EMPLOYEE_INFO WHERE tag_id = %d`, nTagID)
	rows, err := db.Query(strSql)
	if err != nil {
		db.Close()
		return ResultArray, err
	}

	for rows.Next() {
		var nEmployeeId int64

		err = rows.Scan(&nEmployeeId)
		if err != nil {
			beego.Error(err)
		}

		ResultArray = append(ResultArray, nEmployeeId)
	}

	db.Close()

	return ResultArray, nil
}

type GroupAndEmployee struct {
	GroupId    string
	EmployeeId int64
}

func AddGroupAndEmployee(data GroupAndEmployee) error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return err
	}

	stm, err := db.Prepare("INSERT INTO GROUP_AND_EMPLOYEE(GROUP_ID, EMPLOYEE_ID) VALUES(?,?)")
	if err != nil {
		return err
	}

	res, err := stm.Exec(data.GroupId, data.EmployeeId)

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

func RemoveGroupAndEmployeeByEmployeeId(nId int64) error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return err
	}

	stm, err := db.Prepare("DELETE FROM GROUP_AND_EMPLOYEE WHERE EMPLOYEE_ID = ?")
	if err != nil {
		db.Close()
		return err
	}

	res, err := stm.Exec(nId)

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

func GetGroupByEmployee(nEmployee int64) ([]string, error) {
	var GroupArray []string

	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return GroupArray, err
	}

	strSql := fmt.Sprintf("SELECT GROUP_ID FROM GROUP_AND_EMPLOYEE WHERE EMPLOYEE_ID = %d", nEmployee)
	rows, err := db.Query(strSql)
	if err != nil {
		db.Close()
		return GroupArray, err
	}

	for rows.Next() {
		var strGroupId string

		err = rows.Scan(&strGroupId)
		if err != nil {
			beego.Error(err)
		}

		GroupArray = append(GroupArray, strGroupId)
	}

	db.Close()
	return GroupArray, nil
}

func GetEmployeeByGroup(strGroup string) ([]int64, error) {
	var EmployeeArray []int64

	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return EmployeeArray, err
	}

	strSql := fmt.Sprintf("SELECT EMPLOYEE_ID FROM GROUP_AND_EMPLOYEE WHERE GROUP_ID =  '%s'", strGroup)
	rows, err := db.Query(strSql)
	if err != nil {
		db.Close()
		return EmployeeArray, err
	}

	for rows.Next() {
		var nEmployeeId int64

		err = rows.Scan(&nEmployeeId)
		if err != nil {
			beego.Error(err)
		}

		EmployeeArray = append(EmployeeArray, nEmployeeId)
	}

	db.Close()
	return EmployeeArray, nil
}

type RuleInfo struct {
	Id            int64
	Days          string
	Late_time     int64
	Left_early    int64
	Name          string
	Offwork_begin string
	Offwork_end   string
	Time1         string
	Time2         string
	Work_begin    string
	Work_end      string
	OwnerAccount  string
	Type          int64
	Tags          string
	Begin         string
}

//AddRule __
func AddRule(data RuleInfo) (int64, error) {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return 0, err
	}

	rows, err := db.Query("SELECT MAX(id) FROM ATTENDACNE_RULE")
	if err != nil {
		db.Close()
		return 0, err
	}

	nID := int64(0)
	for rows.Next() {
		rows.Scan(&nID)
	}

	nID++

	stm, err := db.Prepare(`INSERT INTO ATTENDACNE_RULE(ID, NAME, ATTENDANCE_DATE, 
		ATTENDANCE_ARRIVE, ATTENDANCE_LEAVE, ATTENDANCE_ARRIVE_BEGIN, ATTENDANCE_ARRIVE_END, ATTENDANCE_LEAVE_BEGIN, 
		ATTENDANCE_LEAVE_END, ATTENDANCE_ARRIVE_LATE,ATTENDANCE_LEAVE_EARLY, OWNERACCOUNT, TYPE, TAGS, BEGIN) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		db.Close()
		return 0, err
	}

	res, err := stm.Exec(nID, data.Name, data.Days, data.Time1, data.Time2, data.Work_begin,
		data.Work_end, data.Offwork_begin, data.Offwork_end, data.Late_time, data.Left_early, data.OwnerAccount, data.Type, data.Tags, data.Begin)

	_, err = res.RowsAffected()
	if err != nil {
		beego.Error(err)
		db.Close()
		return 0, err
	}

	db.Close()
	return nID, nil
}

//UpdateRule __
func UpdateRule(data RuleInfo) error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return err
	}

	stm, err := db.Prepare(`UPDATE ATTENDACNE_RULE SET NAME = ?, ATTENDANCE_DATE = ?, ATTENDANCE_ARRIVE = ?, ATTENDANCE_LEAVE = ?, 
				ATTENDANCE_ARRIVE_BEGIN = ?, ATTENDANCE_ARRIVE_END = ?, ATTENDANCE_LEAVE_BEGIN = ?, 
				ATTENDANCE_LEAVE_END = ?, ATTENDANCE_ARRIVE_LATE = ?,ATTENDANCE_LEAVE_EARLY = ?, TYPE = ?, TAGS =? , BEGIN = ? WHERE ID = ?`)
	if err != nil {
		db.Close()
		return err
	}

	res, err := stm.Exec(data.Name, data.Days, data.Time1, data.Time2, data.Work_begin, data.Work_end,
		data.Offwork_begin, data.Offwork_end, data.Late_time, data.Left_early, data.Type, data.Tags, data.Begin, data.Id)

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

//DeleteRule __
func DeleteRule(nId int64) error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return err
	}

	stm, err := db.Prepare(`DELETE FROM ATTENDACNE_RULE WHERE ID = ?`)
	if err != nil {
		return err
	}

	res, err := stm.Exec(nId)

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

//GetRuleList __
func GetRuleList(OwnnerAccount string) ([]RuleInfo, error) {
	var structResultList []RuleInfo

	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return structResultList, err
	}

	strSQL := fmt.Sprintf(`SELECT ID, NAME, ATTENDANCE_DATE, TIME(ATTENDANCE_ARRIVE), TIME(ATTENDANCE_LEAVE), TIME(ATTENDANCE_ARRIVE_BEGIN), 
	TIME(ATTENDANCE_ARRIVE_END), TIME(ATTENDANCE_LEAVE_BEGIN), TIME(ATTENDANCE_LEAVE_END), ATTENDANCE_ARRIVE_LATE, ATTENDANCE_LEAVE_EARLY,
	TYPE, TAGS, BEGIN
	FROM ATTENDACNE_RULE WHERE  OWNERACCOUNT = '%s'`, OwnnerAccount)
	rows, err := db.Query(strSQL)
	if err != nil {
		db.Close()
		return structResultList, err
	}

	for rows.Next() {
		var nID int64
		var strName string
		var strDays string
		var strArrive string
		var strLeave string
		var strArriveBegin string
		var strArriveEnd string
		var strLeaveBegin string
		var strLeaveEnd string
		var strArriveLate int64
		var strLeaveEarly int64
		var nType int64
		var strTags string
		var strBegin string

		err = rows.Scan(&nID, &strName, &strDays, &strArrive, &strLeave, &strArriveBegin,
			&strArriveEnd, &strLeaveBegin, &strLeaveEnd, &strArriveLate, &strLeaveEarly, &nType, &strTags, &strBegin)
		if err != nil {
			beego.Error(err)
		}

		var structResult RuleInfo
		structResult.Id = nID
		structResult.Days = strDays
		structResult.Late_time = strArriveLate
		structResult.Left_early = strLeaveEarly
		structResult.Name = strName
		structResult.Offwork_begin = strLeaveBegin
		structResult.Offwork_end = strLeaveEnd
		structResult.OwnerAccount = OwnnerAccount
		structResult.Time1 = strArrive
		structResult.Time2 = strLeave
		structResult.Work_begin = strArriveBegin
		structResult.Work_end = strArriveEnd
		structResult.Type = nType
		structResult.Tags = strTags
		structResult.Begin = strBegin

		structResultList = append(structResultList, structResult)
	}

	db.Close()
	return structResultList, nil
}

//GetRuleByID __
func GetRuleByID(nID int64) (RuleInfo, error) {
	var structResult RuleInfo

	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return structResult, err
	}

	strSQL := fmt.Sprintf(`SELECT ID, NAME, ATTENDANCE_DATE, TIME(ATTENDANCE_ARRIVE), TIME(ATTENDANCE_LEAVE), TIME(ATTENDANCE_ARRIVE_BEGIN), 
	TIME(ATTENDANCE_ARRIVE_END), TIME(ATTENDANCE_LEAVE_BEGIN), TIME(ATTENDANCE_LEAVE_END), ATTENDANCE_ARRIVE_LATE, ATTENDANCE_LEAVE_EARLY, OWNERACCOUNT,
	TYPE, TAGS, BEGIN
	FROM ATTENDACNE_RULE WHERE  ID = %d`, nID)
	rows, err := db.Query(strSQL)
	if err != nil {
		db.Close()
		return structResult, err
	}

	for rows.Next() {
		var nIDIndex int64
		var strName string
		var strDays string
		var strArrive string
		var strLeave string
		var strArriveBegin string
		var strArriveEnd string
		var strLeaveBegin string
		var strLeaveEnd string
		var strArriveLate int64
		var strLeaveEarly int64
		var strOwnerAccont string
		var nType int64
		var strTags string
		var strBegin string

		err = rows.Scan(&nIDIndex, &strName, &strDays, &strArrive, &strLeave, &strArriveBegin,
			&strArriveEnd, &strLeaveBegin, &strLeaveEnd, &strArriveLate, &strLeaveEarly, &strOwnerAccont, &nType, &strTags, &strBegin)
		if err != nil {
			beego.Error(err)
		}

		structResult.Id = nIDIndex
		structResult.Days = strDays
		structResult.Late_time = strArriveLate
		structResult.Left_early = strLeaveEarly
		structResult.Name = strName
		structResult.Offwork_begin = strLeaveBegin
		structResult.Offwork_end = strLeaveEnd
		structResult.Time1 = strArrive
		structResult.Time2 = strLeave
		structResult.Work_begin = strArriveBegin
		structResult.Work_end = strArriveEnd
		structResult.OwnerAccount = strOwnerAccont
		structResult.Type = nType
		structResult.Tags = strTags
		structResult.Begin = strBegin

	}

	db.Close()
	return structResult, nil
}

//GetMaxRuleID __
func GetMaxRuleID() (int64, error) {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		//db.Close()
		return 0, err
	}

	rows, err := db.Query("SELECT MAX(ID) FROM ATTENDACNE_RULE")
	if err != nil {
		return 0, err
	}

	nID := int64(0)
	for rows.Next() {
		err = rows.Scan(&nID)
		if err != nil {
			beego.Warning(err)
		}
	}

	return nID, nil
}

type VerifyRecordInfo struct {
	Id             int64
	DeviceId       int64
	CreateTime     string
	Name           string
	VerifyStatus   int64
	VerifyPic      string
	CutomId        int64
	Stanger        int64
	Template       string
	XMDeviceID     string
	DeviceType     int64
	Result         string
	RangeIndex     int64
	AttendanceTime string
	DeviceName     string
	CreateDate     string
	CompanyID      int64
}

//AddDeviceRecord 添加抓拍记录
func AddDeviceRecord(data VerifyRecordInfo) error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		//db.Close()
		//beego.Error("添加验证内容错误", err)
		return err
	}

	rows, err := db.Query("SELECT COUNT(*) FROM VERIFY_PERSON")
	if err != nil {
		db.Close()
		//beego.Error("添加验证内容错误", err)
		return err
	}
	var count int
	for rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			db.Close()
			//beego.Error("添加验证内容错误", err)
			return err
		}
	}

	if count != 0 {
		rows, err = db.Query("SELECT MAX(ID) FROM VERIFY_PERSON")
		if err != nil {
			db.Close()
			//beego.Error("添加验证内容错误", err)
			return err
		}

		for rows.Next() {
			err = rows.Scan(&count)
			if err != nil {
				db.Close()
				//beego.Error("添加验证内容错误", err)
				return err
			}
		}
	}

	index := count + 1

	stmt, err := db.Prepare(`insert into VERIFY_PERSON(ID, DEVICE_ID, CREATE_TIME, NAME, VERIFY_STATUS,VERIFY_PIC, 
							CUSTOM_ID, STRANGER, TEMPLATE, XMDEVICE_ID, DEVICE_TYPE, RESULT, RANGE_INDEX, ATTENDANCE_TIME, 
							DEVICE_NAME,CREATE_DATE, COMPANY_ID) 
							values(?,?,?,?,?,?,?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		//	beego.Error("添加验证内容错误", err)
		db.Close()
		return err
	}
	res, err := stmt.Exec(index, data.DeviceId, data.CreateTime, data.Name, data.VerifyStatus, data.VerifyPic,
		data.CutomId, data.Stanger, data.Template, data.XMDeviceID, data.DeviceType, data.Result, data.RangeIndex, data.AttendanceTime,
		data.DeviceName, data.CreateDate, data.CompanyID)
	if err != nil {
		//beego.Error("添加验证内容错误", err)
		db.Close()
		return err
	}
	_, err = res.RowsAffected()
	if err != nil {
		//beego.Error("添加验证内容错误", err)
		db.Close()
		return err
	}
	//beego.Debug(affect)
	db.Close()

	return nil
}

func GetAllRecordListNoStranger(nCompanyID int64) ([]VerifyRecordInfo, error) {
	var nResultsArray []VerifyRecordInfo
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return nResultsArray, err
	}

	strSQL := fmt.Sprintf(`SELECT ID, DEVICE_ID, CREATE_TIME, NAME, VERIFY_STATUS, VERIFY_PIC, 
							CUSTOM_ID,  STRANGER, TEMPLATE, XMDEVICE_ID, DEVICE_TYPE, RESULT, RANGE_INDEX, ATTENDANCE_TIME,DEVICE_NAME,
							CREATE_DATE
							FROM VERIFY_PERSON 
							WHERE STRANGER = 0 AND COMPANY_ID = %d
							ORDER BY CREATE_TIME DESC`, nCompanyID)

	rows, err := db.Query(strSQL)
	if err != nil {
		db.Close()
		return nResultsArray, err
	}

	for rows.Next() {
		var nId int64
		var nDeviceId int64
		var strCreateTime string
		var strName string
		var nStatus int64
		var strPic string
		var nCustomId int64
		var nStranger int64
		var strTemplate string
		var strXMDeviceID string
		var nDevcieType int64
		var strResult string
		var nRangeIndex int64
		var strAttendanceTime string
		var strDeviceName string
		var strCreateDate string

		err = rows.Scan(&nId, &nDeviceId, &strCreateTime, &strName, &nStatus, &strPic, &nCustomId,
			&nStranger, &strTemplate, &strXMDeviceID, &nDevcieType, &strResult, &nRangeIndex, &strAttendanceTime,
			&strDeviceName, &strCreateDate)
		if err != nil {
			beego.Warning(err)
		}
		var structRecordInfo VerifyRecordInfo
		structRecordInfo.Id = nId
		structRecordInfo.Name = strName
		structRecordInfo.VerifyPic = strPic
		structRecordInfo.VerifyStatus = nStatus
		structRecordInfo.CreateTime = strCreateTime
		structRecordInfo.CutomId = nCustomId
		structRecordInfo.DeviceId = nDeviceId
		structRecordInfo.Stanger = nStranger
		structRecordInfo.Template = strTemplate
		structRecordInfo.XMDeviceID = strXMDeviceID
		structRecordInfo.DeviceType = nDevcieType
		structRecordInfo.AttendanceTime = strAttendanceTime
		structRecordInfo.DeviceName = strDeviceName
		structRecordInfo.CreateDate = strCreateDate
		structRecordInfo.Result = strResult

		nResultsArray = append(nResultsArray, structRecordInfo)
	}

	db.Close()
	return nResultsArray, nil
}

func GetAllRecordListStranger(nCompanyID int64) ([]VerifyRecordInfo, error) {
	var nResultsArray []VerifyRecordInfo
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return nResultsArray, err
	}

	strSql := fmt.Sprintf(`SELECT ID, DEVICE_ID, CREATE_TIME, NAME, VERIFY_STATUS, VERIFY_PIC, 
						CUSTOM_ID,  STRANGER, TEMPLATE, XMDEVICE_ID, DEVICE_TYPE,RESULT, RANGE_INDEX,ATTENDANCE_TIME,DEVICE_NAME,
						CREATE_DATE
					FROM VERIFY_PERSON WHERE STRANGER = 1 AND COMPANY_ID = %d
					ORDER BY CREATE_TIME DESC`, nCompanyID)

	rows, err := db.Query(strSql)
	if err != nil {
		db.Close()
		return nResultsArray, err
	}

	for rows.Next() {
		var nId int64
		var nDeviceId int64
		var strCreateTime string
		var strName string
		var nStatus int64
		var strPic string
		var nCustomId int64
		var nStranger int64
		var strTemplate string
		var strXMDeviceID string
		var nDevcieType int64
		var strResult string
		var nRangeIndex int64
		var strAttendanceTime string
		var strDeviceName string
		var strCreateDate string

		err = rows.Scan(&nId, &nDeviceId, &strCreateTime, &strName, &nStatus, &strPic, &nCustomId,
			&nStranger, &strTemplate, &strXMDeviceID, &nDevcieType, &strResult, &nRangeIndex, &strAttendanceTime, &strDeviceName,
			&strCreateDate)
		if err != nil {
			beego.Error(err)
		}
		var structRecordInfo VerifyRecordInfo
		structRecordInfo.Id = nId
		structRecordInfo.Name = strName
		structRecordInfo.VerifyPic = strPic
		structRecordInfo.VerifyStatus = nStatus
		structRecordInfo.CreateTime = strCreateTime
		structRecordInfo.CutomId = nCustomId
		structRecordInfo.DeviceId = nDeviceId
		structRecordInfo.Stanger = nStranger
		structRecordInfo.Template = strTemplate
		structRecordInfo.XMDeviceID = strXMDeviceID
		structRecordInfo.DeviceType = nDevcieType
		structRecordInfo.AttendanceTime = strAttendanceTime
		structRecordInfo.DeviceName = strDeviceName
		structRecordInfo.CreateDate = strCreateDate
		nResultsArray = append(nResultsArray, structRecordInfo)
	}

	db.Close()
	return nResultsArray, nil
}

func GetVerifyRecordList(nCompanyID int64, bStranger bool, nPage int64, nLimit int64) ([]VerifyRecordInfo, error) {
	var nResultsArray []VerifyRecordInfo
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return nResultsArray, err
	}

	strSql := ""
	if !bStranger {
		strSql = fmt.Sprintf(`SELECT ID, DEVICE_ID, CREATE_TIME, 
		NAME, VERIFY_STATUS, VERIFY_PIC, CUSTOM_ID,
		STRANGER, TEMPLATE, XMDEVICE_ID, DEVICE_TYPE,RESULT, RANGE_INDEX,ATTENDANCE_TIME,DEVICE_NAME,
		CREATE_DATE
		FROM VERIFY_PERSON WHERE STRANGER = 0 AND COMPANY_ID = %d
		ORDER BY CREATE_TIME DESC LIMIT %d,%d`, nCompanyID, nLimit*(nPage-1), nLimit*nPage)
	} else {
		strSql = fmt.Sprintf(`SELECT ID, DEVICE_ID, CREATE_TIME, 
		NAME, VERIFY_STATUS, VERIFY_PIC, CUSTOM_ID,
		STRANGER, TEMPLATE, XMDEVICE_ID, DEVICE_TYPE,RESULT, RANGE_INDEX ,ATTENDANCE_TIME,DEVICE_NAME,
		CREATE_DATE
		FROM VERIFY_PERSON WHERE COMPANY_ID = %d
		ORDER BY CREATE_TIME DESC LIMIT %d,%d`, nCompanyID, nLimit*(nPage-1), nLimit*nPage)
	}

	beego.Debug("搜索sql：", strSql)

	rows, err := db.Query(strSql)
	if err != nil {
		db.Close()
		return nResultsArray, err
	}

	for rows.Next() {
		var nId int64
		var nDeviceId int64
		var strCreateTime string
		var strName string
		var nStatus int64
		var strPic string
		var nCustomId int64
		var nStranger int64
		var strTemplate string
		var strXMDeviceID string
		var nDevcieType int64
		var strResult string
		var nRangeIndex int64
		var strAttendance string
		var strDeviceName string
		var strCreateDate string

		err = rows.Scan(&nId, &nDeviceId, &strCreateTime, &strName, &nStatus, &strPic, &nCustomId,
			&nStranger, &strTemplate, &strXMDeviceID, &nDevcieType, &strResult, &nRangeIndex, &strAttendance, &strDeviceName,
			&strCreateDate)
		if err != nil {
			beego.Warning(err)
		}
		var structRecordInfo VerifyRecordInfo
		structRecordInfo.Id = nId
		structRecordInfo.Name = strName
		structRecordInfo.VerifyPic = strPic
		structRecordInfo.VerifyStatus = nStatus
		structRecordInfo.CreateTime = strCreateTime
		structRecordInfo.CutomId = nCustomId
		structRecordInfo.DeviceId = nDeviceId
		structRecordInfo.Stanger = nStranger
		structRecordInfo.Template = strTemplate
		structRecordInfo.XMDeviceID = strXMDeviceID
		structRecordInfo.DeviceType = nDevcieType
		structRecordInfo.Result = strResult
		structRecordInfo.RangeIndex = nRangeIndex
		structRecordInfo.AttendanceTime = strAttendance
		structRecordInfo.DeviceName = strDeviceName
		structRecordInfo.CreateDate = strCreateDate

		nResultsArray = append(nResultsArray, structRecordInfo)
	}

	db.Close()
	return nResultsArray, nil

}

func GetVerifyRecordCountByDate(nCompanyID int64, dateBegin string, dateEnd string) (int64, error) {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return 0, err
	}

	nCount := int64(0)

	strSql := fmt.Sprintf(`SELECT COUNT(*) FROM VERIFY_PERSON
				WHERE CREATE_TIME <= '%s' AND CREATE_TIME >= '%s' AND COMPANY_ID = %d`, dateEnd, dateBegin, nCompanyID)

	rows, err := db.Query(strSql)
	if err != nil {
		db.Close()
		return nCount, err
	}
	for rows.Next() {
		rows.Scan(&nCount)
	}
	db.Close()
	return nCount, err
}

func GetTotalVerifyRecordList(nCompanyID int64, bStranger bool) (int64, error) {
	nCount := int64(0)

	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return nCount, err
	}

	strSql := ""
	if !bStranger {
		strSql = fmt.Sprintf(`SELECT COUNT(*) FROM VERIFY_PERSON WHERE STRANGER = 0 AND COMPANY_ID = %d`, nCompanyID)
	} else {
		strSql = fmt.Sprintf(`SELECT COUNT(*) FROM VERIFY_PERSON WHERE COMPANY_ID = %d`, nCompanyID)
	}

	rows, err := db.Query(strSql)
	if err != nil {
		db.Close()
		return nCount, err
	}

	for rows.Next() {
		err = rows.Scan(&nCount)
		if err != nil {
			beego.Error(err)
		}
	}

	db.Close()
	return nCount, nil
}

type SearchStruct struct {
	TimeBegin     string
	TimeEnd       string
	SearchContent string
}

func GetVerifyRecordBySearchStruct(nCompanyID int64, data SearchStruct, bStrange bool, nPage int64, nLimit int64) ([]VerifyRecordInfo, error) {

	var nResultsArray []VerifyRecordInfo
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return nResultsArray, err
	}

	strSql := ""
	if !bStrange {
		if data.SearchContent == "" {
			strSql = fmt.Sprintf(`SELECT ID, DEVICE_ID, CREATE_TIME, NAME, VERIFY_STATUS, VERIFY_PIC, CUSTOM_ID, 
								   STRANGER, TEMPLATE, XMDEVICE_ID, DEVICE_TYPE,RESULT, RANGE_INDEX,ATTENDANCE_TIME, DEVICE_NAME ,
								   CREATE_DATE
								FROM VERIFY_PERSON 
								WHERE CREATE_TIME >= '%s' AND CREATE_TIME <= '%s' AND STRANGER = 0 AND COMPANY_ID = %d
								ORDER BY CREATE_TIME DESC LIMIT %d, %d`,
				data.TimeBegin, data.TimeEnd, nCompanyID, nLimit*(nPage-1), nLimit*nPage)
		} else if data.TimeBegin == "" {
			strSql = fmt.Sprintf(`SELECT ID, DEVICE_ID, CREATE_TIME, NAME, VERIFY_STATUS, VERIFY_PIC, CUSTOM_ID,
								STRANGER, TEMPLATE, XMDEVICE_ID, DEVICE_TYPE ,RESULT, RANGE_INDEX ,ATTENDANCE_TIME,DEVICE_NAME,
								CREATE_DATE
								FROM VERIFY_PERSON WHERE NAME LIKE '%s' AND STRANGER = 0 AND COMPANY_ID = %d
								ORDER BY CREATE_TIME DESC LIMIT %d, %d`,
				data.SearchContent, nCompanyID, nLimit*(nPage-1), nLimit*nPage)
		} else {
			strSql = fmt.Sprintf(`SELECT ID, DEVICE_ID, CREATE_TIME, NAME, VERIFY_STATUS, VERIFY_PIC, CUSTOM_ID,
								STRANGER, TEMPLATE, XMDEVICE_ID, DEVICE_TYPE  ,RESULT, RANGE_INDEX ,ATTENDANCE_TIME,DEVICE_NAME,
								CREATE_DATE
								FROM VERIFY_PERSON 
								WHERE CREATE_TIME >= '%s' AND CREATE_TIME <= '%s' AND NAME LIKE '%s' AND STRANGER = 0 AND COMPANY_ID = %d
								ORDER BY CREATE_TIME DESC LIMIT %d, %d`,
				data.TimeBegin, data.TimeEnd, data.SearchContent, nCompanyID, nLimit*(nPage-1), nLimit*nPage)
		}
	} else {
		if data.SearchContent == "" {
			strSql = fmt.Sprintf(`SELECT ID, DEVICE_ID, CREATE_TIME, NAME, VERIFY_STATUS, VERIFY_PIC, CUSTOM_ID, 
								STRANGER, TEMPLATE, XMDEVICE_ID, DEVICE_TYPE ,RESULT, RANGE_INDEX ,ATTENDANCE_TIME,DEVICE_NAME,
								CREATE_DATE
								FROM VERIFY_PERSON 
								WHERE CREATE_TIME >= '%s' AND CREATE_TIME <= '%s' AND COMPANY_ID = %d
								ORDER BY CREATE_TIME DESC LIMIT %d, %d`,
				data.TimeBegin, data.TimeEnd, nCompanyID, nLimit*(nPage-1), nLimit*nPage)
		} else if data.TimeBegin == "" {
			strSql = fmt.Sprintf(`SELECT ID, DEVICE_ID, CREATE_TIME, NAME, VERIFY_STATUS, VERIFY_PIC, CUSTOM_ID, 
							STRANGER, TEMPLATE, XMDEVICE_ID, DEVICE_TYPE ,RESULT, RANGE_INDEX ,ATTENDANCE_TIME,DEVICE_NAME,
							CREATE_DATE
							FROM VERIFY_PERSON 	
							WHERE NAME LIKE '%s' AND COMPANY_ID = %d
							ORDER BY CREATE_TIME DESC LIMIT %d, %d`,
				data.SearchContent, nCompanyID, nLimit*(nPage-1), nLimit*nPage)
		} else {
			strSql = fmt.Sprintf(`SELECT ID, DEVICE_ID, CREATE_TIME, NAME, VERIFY_STATUS, VERIFY_PIC, CUSTOM_ID,
								STRANGER, TEMPLATE, XMDEVICE_ID, DEVICE_TYPE ,RESULT, RANGE_INDEX ,ATTENDANCE_TIME,DEVICE_NAME,
								CREATE_DATE			
								FROM VERIFY_PERSON 
								WHERE CREATE_TIME >= '%s' AND CREATE_TIME <= '%s' AND NAME LIKE '%s'  AND COMPANY_ID = %d
								ORDER BY CREATE_TIME DESC LIMIT %d, %d`,
				data.TimeBegin, data.TimeEnd, data.SearchContent, nCompanyID, nLimit*(nPage-1), nLimit*nPage)
		}
	}

	beego.Debug("sql:", strSql)
	rows, err := db.Query(strSql)
	if err != nil {
		db.Close()
		return nResultsArray, err
	}

	for rows.Next() {
		var nId int64
		var nDeviceId int64
		var strCreateTime string
		var strName string
		var nStatus int64
		var strPic string
		var nCustomId int64
		var nStranger int64
		var strTemplate string
		var strXMDeviceID string
		var nDevcieType int64
		var strResult string
		var nRangeIndex int64
		var strAttendanceTime string
		var strDeviceName string
		var strCreateDate string
		err = rows.Scan(&nId, &nDeviceId, &strCreateTime, &strName, &nStatus, &strPic, &nCustomId,
			&nStranger, &strTemplate, &strXMDeviceID, &nDevcieType, &strResult, &nRangeIndex, &strAttendanceTime, &strDeviceName,
			&strCreateDate)
		if err != nil {
			beego.Error(err)
		}
		var structRecordInfo VerifyRecordInfo
		structRecordInfo.Id = nId
		structRecordInfo.Name = strName
		structRecordInfo.VerifyPic = strPic
		structRecordInfo.VerifyStatus = nStatus
		structRecordInfo.CreateTime = strCreateTime
		structRecordInfo.CutomId = nCustomId
		structRecordInfo.DeviceId = nDeviceId
		structRecordInfo.Stanger = nStranger
		structRecordInfo.Template = strTemplate
		structRecordInfo.XMDeviceID = strXMDeviceID
		structRecordInfo.DeviceType = nDevcieType
		structRecordInfo.Result = strResult
		structRecordInfo.RangeIndex = nRangeIndex
		structRecordInfo.AttendanceTime = strAttendanceTime
		structRecordInfo.DeviceName = strDeviceName
		structRecordInfo.CreateDate = strCreateDate
		nResultsArray = append(nResultsArray, structRecordInfo)
	}

	db.Close()
	return nResultsArray, nil

}

func GetTotalVerifyRecordBySearchStruct(nCompanyID int64, data SearchStruct, bStrange bool) (int64, error) {
	nCount := int64(0)

	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return nCount, err
	}

	strSql := ""
	if !bStrange {
		if data.SearchContent == "" {
			strSql = fmt.Sprintf(`SELECT COUNT(*) FROM VERIFY_PERSON 
			WHERE CREATE_TIME >= '%s' AND CREATE_TIME <= '%s' AND STRANGER = 0  AND COMPANY_ID = %d`, data.TimeBegin, data.TimeEnd, nCompanyID)
		} else if data.TimeBegin == "" {
			strSql = fmt.Sprintf(`SELECT COUNT(*) FROM VERIFY_PERSON 
			WHERE NAME LIKE '%s' AND STRANGER = 0 AND COMPANY_ID = %d`, data.SearchContent, nCompanyID)
		} else {
			strSql = fmt.Sprintf(`SELECT COUNT(*) FROM VERIFY_PERSON 
			WHERE CREATE_TIME >= '%s' AND CREATE_TIME <= '%s' AND NAME LIKE '%s' AND STRANGER = 0  AND COMPANY_ID = %d`,
				data.TimeBegin, data.TimeEnd, data.SearchContent, nCompanyID)
		}
	} else {
		if data.SearchContent == "" {
			strSql = fmt.Sprintf(`SELECT COUNT(*) FROM VERIFY_PERSON 
			WHERE CREATE_TIME >= '%s' AND CREATE_TIME <= '%s'  AND COMPANY_ID = %d`,
				data.TimeBegin, data.TimeEnd, nCompanyID)
		} else if data.TimeBegin == "" {
			strSql = fmt.Sprintf(`SELECT COUNT(*) FROM VERIFY_PERSON 
			WHERE NAME LIKE '%s'  AND COMPANY_ID = %d`, data.SearchContent, nCompanyID)
		} else {
			strSql = fmt.Sprintf(`SELECT COUNT(*) FROM VERIFY_PERSON 
			WHERE CREATE_TIME >= '%s' AND CREATE_TIME <= '%s' AND NAME LIKE '%s' AND COMPANY_ID = %d`,
				data.TimeBegin, data.TimeEnd, data.SearchContent, nCompanyID)
		}
	}

	rows, err := db.Query(strSql)
	if err != nil {
		db.Close()
		return nCount, err
	}

	for rows.Next() {
		err = rows.Scan(&nCount)
		if err != nil {
			beego.Error(err)
		}
	}

	db.Close()
	return nCount, nil

}

func RemoveAttendRecordByCustomId(nCustomId int64) error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return err
	}

	stmt, err := db.Prepare("UPDATE VERIFY_PERSON SET CUSTOM_ID = -1 , STRANGER = 1, NAME = '陌生人' WHERE CUSTOM_ID = ?")
	if err != nil {
		logs.Error(err.Error())
		db.Close()
		return err
	}
	res, err := stmt.Exec(nCustomId)
	if err != nil {
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

func GetEarlyDateFromRecord(nCompanyID int64) (string, error) {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		//db.Close()
		return "", err
	}

	strSQL := fmt.Sprintf("SELECT MIN(CREATE_TIME) FROM VERIFY_PERSON WHERE COMPANY_ID =%d", nCompanyID)
	rows, err := db.Query(strSQL)
	if err != nil {
		return "", err
	}

	strEarlyDay := ""

	for rows.Next() {
		err = rows.Scan(&strEarlyDay)
		if err != nil {
			beego.Warning(err)
		}
	}

	return strEarlyDay, nil
}

func GetRecordCountByContent(strXMDeviceID string, nVerifyStatus int64, strCreateTime string, strUserName string) (int64, error) {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		//db.Close()
		return -1, err
	}

	strSQL := fmt.Sprintf(`SELECT COUNT(*) FROM VERIFY_PERSON 
						WHERE XMDEVICE_ID = '%s' AND VERIFY_STATUS = %d AND CREATE_TIME >= '%s' AND NAME <> '%s' ;`,
		strXMDeviceID, nVerifyStatus, strCreateTime, strUserName)

	rows, err := db.Query(strSQL)
	if err != nil {
		return -1, err
	}

	nCount := int64(0)
	for rows.Next() {
		err = rows.Scan(&nCount)
		if err != nil {
			beego.Warning(err)
		}
	}

	return nCount, nil
}

type CompanyInfo struct {
	Id       int64
	Address  string
	Name     string
	Phone    string
	Location string
	MultiCer int64
}

func GetCompanyList() ([]CompanyInfo, error) {
	var arrayResult = make([]CompanyInfo, 0)

	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return arrayResult, err
	}

	rows, err := db.Query(`SELECT ID, ADDRESS, NAME, PHONE, LOCATION, MULTI_CERT FROM  COMPANY_TABLE`)
	if err != nil {
		return arrayResult, err
	}

	for rows.Next() {
		var structCompanyInfo CompanyInfo

		err = rows.Scan(&structCompanyInfo.Id, &structCompanyInfo.Address, &structCompanyInfo.Name,
			&structCompanyInfo.Phone, &structCompanyInfo.Location, &structCompanyInfo.MultiCer)
		if err != nil {
			beego.Warning(err)
		}

		arrayResult = append(arrayResult, structCompanyInfo)
	}

	return arrayResult, nil
}

func GetCompnyById(nId int64) (CompanyInfo, error) {
	var structResult CompanyInfo

	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return structResult, err
	}

	strSql := fmt.Sprintf("SELECT ID, ADDRESS, NAME, PHONE, LOCATION, MULTI_CERT FROM  COMPANY_TABLE WHERE ID = %d", nId)
	rows, err := db.Query(strSql)
	if err != nil {
		db.Close()
		return structResult, err
	}

	for rows.Next() {
		var nId int64
		var strAddress string
		var strName string
		var strPhone string
		var strLocation string
		var nMultiCer int64

		err := rows.Scan(&nId, &strAddress, &strName, &strPhone, &strLocation, &nMultiCer)
		if err != nil {
			beego.Warning(err)
		}

		structResult.Id = nId
		structResult.Name = strName
		structResult.Phone = strPhone
		structResult.Address = strAddress
		structResult.Location = strLocation
		structResult.MultiCer = nMultiCer
	}
	db.Close()
	return structResult, nil
}

func InsertCompanyInfo(data CompanyInfo) (int64, error) {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return 0, err
	}

	nCount := int64(0)

	rows, err := db.Query("SELECT COUNT(*) FROM  COMPANY_TABLE")
	if err != nil {
		db.Close()
		return nCount, err
	}

	for rows.Next() {
		err := rows.Scan(&nCount)
		if err != nil {
			db.Close()
			return nCount, err
		}
	}

	if nCount != 0 {
		rows_, err := db.Query("SELECT MAX(ID) FROM COMPANY_TABLE")
		if err != nil {
			db.Close()
			return nCount, err
		}

		for rows_.Next() {
			err := rows_.Scan(&nCount)
			if err != nil {
				db.Close()
				return nCount, err
			}
		}
	}

	nCount++

	stmt, err := db.Prepare("INSERT INTO COMPANY_TABLE(ID, ADDRESS, NAME, PHONE, LOCATION, MULTI_CERT) VALUES(?, ?, ? ,?, ?, ?)")
	if err != nil {
		db.Close()
		return nCount, err
	}
	stmt.Exec(nCount, data.Address, data.Name, data.Phone, data.Location, data.MultiCer)

	db.Close()
	return nCount, nil

}

func UpdateCompanyInfo(data CompanyInfo) error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return err
	}

	stmt, err := db.Prepare("UPDATE COMPANY_TABLE SET ADDRESS = ?, NAME = ?, PHONE = ?, LOCATION = ?, MULTI_CERT = ? WHERE ID = ?")
	if err != nil {
		db.Close()
		return err
	}
	stmt.Exec(data.Address, data.Name, data.Phone, data.Location, data.MultiCer, data.Id)

	db.Close()
	return nil
}

func RemoveCompanyByID(nID int64) error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		//db.Close()
		return err
	}

	stmt, err := db.Prepare("DELETE FROM COMPANY_TABLE WHERE ID = ?")
	if err != nil {
		return err
	}

	res, err := stmt.Exec(nID)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

//RangeInfo __
type RangeInfo struct {
	ID           int64
	LateTime     int64
	LeftEarly    int64
	OffWorkBegin string
	OffWorkEnd   string
	OffWorkCheck int64
	Time1        string
	Time2        string
	WorkBegin    string
	WorkEnd      string
	WorkCheck    int64
	Begin        string
}

//AddRangeInfo __
func AddRangeInfo(data RangeInfo) (int64, error) {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return 0, err
	}

	nID := int64(0)

	rows, err := db.Query("SELECT COUNT(*) FROM RANGE_TABLE")
	if err != nil {
		return -1, err
	}

	nCount := 0
	for rows.Next() {
		err := rows.Scan(&nCount)
		if err != nil {
			beego.Warning(err)
		}
	}

	if nCount > 0 {
		rows, err = db.Query("SELECT MAX(ID) FROM RANGE_TABLE")
		if err != nil {
			return 0, err
		}

		for rows.Next() {
			err = rows.Scan(&nID)
			if err != nil {
				beego.Warning(err)
			}
		}
	}

	nID++

	stmt, err := db.Prepare(`INSERT INTO RANGE_TABLE(ID, LATE_TIME,  LEFT_EARLY, OFFWORK_BEGIN, OFFWORK_END, 
						OFFWORK_CHECK, TIME1, TIME2, WORK_BEGIN, WORK_END, WORK_CHECK, BEGIN) 
						VALUES(?, ?, ?, ?, ?, ?, ?, ?, ? ,? ,?, ?)`)

	if err != nil {
		return -1, err
	}

	res, err := stmt.Exec(nID, data.LateTime, data.LeftEarly, data.OffWorkBegin, data.OffWorkEnd, data.OffWorkCheck, data.Time1,
		data.Time2, data.WorkBegin, data.WorkEnd, data.WorkCheck, data.Begin)

	if err != nil {
		return -1, err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return -1, err
	}

	return nID, nil
}

//GetRangeByID __
func GetRangeByID(nID int64) (RangeInfo, error) {
	var structResult RangeInfo

	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return structResult, err
	}

	strSQL := fmt.Sprintf(`SELECT ID, LATE_TIME, LEFT_EARLY, OFFWORK_BEGIN, OFFWORK_END, OFFWORK_CHECK, TIME1, TIME2, WORK_BEGIN, WORK_END, WORK_CHECK, BEGIN FROM RANGE_TABLE WHERE ID = %d`, nID)
	rows, err := db.Query(strSQL)
	if err != nil {
		return structResult, err
	}

	for rows.Next() {
		err := rows.Scan(&structResult.ID, &structResult.LateTime, &structResult.LeftEarly,
			&structResult.OffWorkBegin, &structResult.OffWorkEnd, &structResult.OffWorkCheck, &structResult.Time1, &structResult.Time2,
			&structResult.WorkBegin, &structResult.WorkEnd, &structResult.WorkCheck, &structResult.Begin)

		if err != nil {
			beego.Warning(err)
		}
	}

	return structResult, nil
}

//UpdateRangeInfo __
func UpdateRangeInfo(data RangeInfo) error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return err
	}

	stmt, err := db.Prepare(`UPDATE RANGE_TABLE SET LATE_TIME = ?, LEFT_EARLY = ?, OFFWORK_BEGIN = ?, OFFWORK_END = ?, OFFWORK_CHECK = ?, 
				TIME1 = ?, TIME2 = ?. WORK_BEGIN = ?, WORK_END = ?, WORK_CHECK = ?, BEGIN = ? WHERE ID = ?`)

	if err != nil {
		return err
	}

	res, err := stmt.Exec(data.LateTime, data.LeftEarly, data.OffWorkBegin, data.OffWorkEnd, data.OffWorkCheck, data.Time1, data.Time2, data.WorkBegin, data.WorkEnd,
		data.WorkCheck, data.Begin, data.ID)

	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

//RemoveRangeInfo __
func RemoveRangeInfo(nID int64) error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return err
	}

	stmt, err := db.Prepare("DELETE FROM RANGE_TABLE WHERE ID = ?")
	if err != nil {
		return err
	}

	res, err := stmt.Exec(nID)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

//TagInfo __
type TagInfo struct {
	ID       int64
	Index    int64
	Selected int64
	Active   int64
	Range1   int64
	Range2   int64
	Range3   int64
	Color    string
	Name     string
	Persons  string
	Type     int64
}

//AddTagInfo __
func AddTagInfo(data TagInfo) (int64, error) {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return -1, err
	}

	nID := int64(0)

	rows, err := db.Query("SELECT COUNT(*) FROM  TAG_TABLE")
	if err != nil {
		return -1, err
	}

	nCount := 0
	for rows.Next() {
		err = rows.Scan(&nCount)
		if err != nil {
			beego.Warning(err)
		}
	}

	if nCount > 0 {
		rows, err = db.Query("SELECT MAX(ID) FROM TAG_TABLE")
		if err != nil {
			return -1, err
		}

		for rows.Next() {
			err = rows.Scan(&nID)
			if err != nil {
				beego.Warning(err)
			}
		}
	}

	nID++

	stmt, err := db.Prepare(`INSERT INTO TAG_TABLE(ID, "INDEX", SELECTED, ACTIVE, RANGER1, RANGER2, RANGER3, COLOR, NAME, PERSONS, "TYPE") VALUES(?,?,?,?,?,?,?,?,?, ?, ?)`)
	if err != nil {
		return -1, err
	}

	res, err := stmt.Exec(nID, data.Index, data.Selected, data.Active, data.Range1, data.Range2, data.Range3, data.Color, data.Name, data.Persons, data.Type)
	if err != nil {
		return -1, err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return -1, err
	}

	return nID, nil
}

//GetTagByID __
func GetTagByID(nID int64) (TagInfo, error) {
	var structResult TagInfo

	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return structResult, err
	}

	strSQL := fmt.Sprintf(`SELECT ID, "INDEX", SELECTED, ACTIVE, RANGER1, RANGER2, RANGER3, COLOR, NAME, PERSONS, TYPE FROM TAG_TABLE  WHERE ID = %d`, nID)
	rows, err := db.Query(strSQL)
	if err != nil {
		return structResult, err
	}

	for rows.Next() {
		err = rows.Scan(&structResult.ID, &structResult.Index, &structResult.Selected, &structResult.Active, &structResult.Range1, &structResult.Range2,
			&structResult.Range3, &structResult.Color, &structResult.Name, &structResult.Persons, &structResult.Type)
		if err != nil {
			beego.Warning(err)
		}
	}

	return structResult, nil
}

//UpdateTagInfo __
func UpdateTagInfo(data TagInfo) error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return err
	}

	stmt, err := db.Prepare(`UPDATE TAG_TABLE SET "INDEX" = ?, SELECTED = ?, ACTIVE = ?,
				 RANGER1 = ?, RANGER2 = ? ,RANGER3 = ?, COLOR = ?, NAME = ?, PERSONS = ?, TYPE = ? WHERE ID = ?`)

	if err != nil {
		return err
	}

	res, err := stmt.Exec(data.Index, data.Selected, data.Active, data.Range2, data.Range2, data.Range3, data.Color, data.Name, data.Persons, data.Type, data.ID)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

//RemoveTag __
func RemoveTag(nID int64) error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return err
	}

	stmt, err := db.Prepare("DELETE FROM TAG_TABLE WHERE ID = ?")
	if err != nil {
		return err
	}

	res, err := stmt.Exec(nID)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

//AddRangeAndEmployee __
func AddRangeAndEmployee(nRange int64, nEmployee int64) error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return err
	}

	stmt, err := db.Prepare("INSERT INTO RANGE_AND_EMPLOYEE(RANGE_ID, EMPLOYEE_ID) VALUES(?, ?)")
	if err != nil {
		return err
	}

	res, err := stmt.Exec(nRange, nEmployee)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

//GetRangeAndEmployeeByEmployeeID __
func GetRangeAndEmployeeByEmployeeID(nEmployeeID int64) ([]int64, error) {
	arrayResult := make([]int64, 0)

	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return arrayResult, err
	}

	strSQL := fmt.Sprintf("SELECT RANGE_ID FROM RANGE_AND_EMPLOYEE WHERE EMPLOYEE_ID = %d", nEmployeeID)
	rows, err := db.Query(strSQL)
	if err != nil {
		return arrayResult, err
	}

	for rows.Next() {
		var nRangeID int64
		err = rows.Scan(&nRangeID)
		if err != nil {
			beego.Warning(err)
		}
		arrayResult = append(arrayResult, nRangeID)
	}

	return arrayResult, nil
}

//DeleteRangeAndEmployeeByEmployeeID __
func DeleteRangeAndEmployeeByEmployeeID(nEmployeeID int64) error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return err
	}

	stmt, err := db.Prepare("DELETE FROM RANGE_AND_EMPLOYEE WHERE EMPLOYEE_ID = ?")
	if err != nil {
		return err
	}

	res, err := stmt.Exec(nEmployeeID)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

//DeleteRangeAndEmployeeByRangeID __
func DeleteRangeAndEmployeeByRangeID(nRangeID int64) error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return err
	}

	stmt, err := db.Prepare("DELETE FROM RANGE_AND_EMPLOYEE WHERE RANGE_ID = ?")
	if err != nil {
		return err
	}

	res, err := stmt.Exec(nRangeID)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

type GroupDeviceInfo struct {
	GroupName string
	GroupID   int64
	ID        int64
	Name      string
}

func AddGroupDevice(data GroupDeviceInfo) (int64, error) {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return 0, err
	}

	//strSQL := fmt.Sprintf("")
	rows, err := db.Query("SELECT COUNT(*) FROM GROUP_DEVICE")
	if err != nil {
		db.Close()
		return 0, err
	}

	nID := int64(0)

	for rows.Next() {
		err = rows.Scan(&nID)
		if err != nil {
			beego.Warning(err)
		}
	}

	nGroupID := nID

	if nID > 0 {
		rows, err = db.Query("SELECT MAX(ID), MAX(GROUP_ID) FROM GROUP_DEVICE")
		if err != nil {
			db.Close()
			return 0, err
		}
		for rows.Next() {
			err = rows.Scan(&nID, &nGroupID)
			if err != nil {
				beego.Warning(err)
			}
		}
	}

	nID++
	nGroupID++

	stmt, err := db.Prepare("INSERT INTO GROUP_DEVICE( ID, NAME, GROUP_ID, GROUP_NAME) VALUES( ?, ?, ?, ?)")
	if err != nil {
		return 0, err
	}

	var res sql.Result
	if data.ID <= 0 && data.GroupID <= 0 {
		res, err = stmt.Exec(nID, data.Name, nGroupID, data.GroupName)
	} else if data.ID <= 0 && data.GroupID > 0 {
		res, err = stmt.Exec(nID, data.Name, data.GroupID, data.GroupName)
		nGroupID = data.GroupID
	} else if data.ID > 0 && data.GroupID <= 0 {
		res, err = stmt.Exec(data.ID, data.Name, nGroupID, data.GroupName)
	} else if data.ID > 0 && data.GroupID > 0 {
		res, err = stmt.Exec(data.ID, data.Name, data.GroupID, data.GroupName)
		nGroupID = data.GroupID
	} else {
		beego.Warning("错误的传入数据:", data)
	}

	if res != nil {
		_, err = res.RowsAffected()
	}

	if err != nil {
		return 0, err
	}

	return nGroupID, nil
}

func GetGroupDevice() (map[int64][]GroupDeviceInfo, error) {
	var mapResult map[int64][]GroupDeviceInfo
	mapResult = make(map[int64][]GroupDeviceInfo, 0)
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return mapResult, err
	}

	rows, err := db.Query("SELECT GROUP_NAME, GROUP_ID, ID, NAME FROM GROUP_DEVICE")
	if err != nil {
		return mapResult, err
	}

	for rows.Next() {
		var structDevice GroupDeviceInfo
		err = rows.Scan(&structDevice.GroupName, &structDevice.GroupID, &structDevice.ID, &structDevice.Name)
		if err != nil {
			beego.Warning(err)
		}

		mapResult[structDevice.GroupID] = append(mapResult[structDevice.GroupID], structDevice)

	}

	return mapResult, nil
}

func RemoveGroupDevice(nDeviceID int64) error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return err
	}

	stmt, err := db.Prepare("DELETE FROM GROUP_DEVICE WHERE ID = ?")
	if err != nil {
		return err
	}

	res, err := stmt.Exec(nDeviceID)
	if err != nil {
		return err
	}
	_, err = res.RowsAffected()
	if err != nil {
		return err
	}
	return nil
}

func RemoveGroupDeviceByGroupID(nGroupID int64) error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return err
	}

	stmt, err := db.Prepare("DELETE FROM GROUP_DEVICE WHERE GROUP_ID = ?")
	if err != nil {
		return err
	}

	res, err := stmt.Exec(nGroupID)
	if err != nil {
		return err
	}
	_, err = res.RowsAffected()
	if err != nil {
		return err
	}
	return nil
}

type DayRangeInfo struct {
	ID           int64
	Father       int64
	FatherName   string
	LateTime     int64
	LeftEarly    int64
	OffworkBegin string
	OffworkEnd   string
	OffworkCheck int64
	Time1        string
	Time2        string
	WorkBegin    string
	WorkEnd      string
	WorkCheck    int64
	Days         string
}

func AddDayRange(data DayRangeInfo) (int64, int64, error) {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		//db.Close()
		return 0, 0, err
	}

	rows, err := db.Query("SELECT MAX(ID) FROM RULE_DAY_RANGE")
	if err != nil {

		return 0, 0, err
	}

	data.ID = int64(0)
	for rows.Next() {
		err = rows.Scan(&data.ID)
		if err != nil {
			beego.Warning(err)
		}
	}

	data.ID++

	if data.Father <= 0 {
		rows, err = db.Query("SELECT MAX(FATHER) FROM RULE_DAY_RANGE")
		if err != nil {

			return 0, 0, err
		}
		data.Father = int64(0)

		for rows.Next() {
			err = rows.Scan(&data.Father)
			if err != nil {
				beego.Warning(err)
			}
		}

		data.Father++
	}

	stmt, err := db.Prepare(`INSERT INTO RULE_DAY_RANGE (ID, FATHER, LATE_TIME, LEFTE_EARLY, OFFWORK_BEGIN, 
							OFFWORK_END, OFFWORK_CHECK, TIME1, TIME2, WORK_BEGIN, WORK_END, WORK_CHECK, DAYS, FATHER_NAME) 
							VALUES (?, ?, ?, ?, ?, ? ,?, ?, ?, ?, ?, ?, ?, ?)`)

	if err != nil {

		return 0, 0, err
	}

	res, err := stmt.Exec(data.ID, data.Father, data.LateTime, data.LeftEarly, data.OffworkBegin, data.OffworkEnd, data.OffworkCheck,
		data.Time1, data.Time2, data.WorkBegin, data.WorkEnd, data.WorkCheck, data.Days, data.FatherName)

	if err != nil {

		return 0, 0, err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return 0, 0, err
	}

	return data.ID, data.Father, nil
}

func GetDayRangeList() (map[int64][]DayRangeInfo, error) {
	var mapResult = make(map[int64][]DayRangeInfo, 0)

	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		db.Close()
		return mapResult, err
	}

	rows, err := db.Query(`SELECT ID, FATHER, LATE_TIME, LEFTE_EARLY,
	 		OFFWORK_BEGIN, OFFWORK_END, OFFWORK_CHECK, TIME1, 
			 TIME2, WORK_BEGIN, WORK_END, WORK_CHECK, DAYS, FATHER_NAME FROM RULE_DAY_RANGE`)

	if err != nil {
		beego.Error(err)
		return mapResult, err
	}

	for rows.Next() {
		var structDayRange DayRangeInfo
		rows.Scan(&structDayRange.ID, &structDayRange.Father, &structDayRange.LateTime, &structDayRange.LeftEarly,
			&structDayRange.OffworkBegin, &structDayRange.OffworkEnd, &structDayRange.OffworkCheck, &structDayRange.Time1,
			&structDayRange.Time2, &structDayRange.WorkBegin, &structDayRange.WorkEnd, &structDayRange.WorkCheck, &structDayRange.Days, &structDayRange.FatherName)

		_, bHas := mapResult[structDayRange.Father]
		if !bHas {
			mapResult[structDayRange.Father] = make([]DayRangeInfo, 0)
		}

		mapResult[structDayRange.Father] = append(mapResult[structDayRange.Father], structDayRange)
	}

	return mapResult, nil
}

//DeleteRangeByFather __
func DeleteRangeByFather(nFather int64) error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		//db.Close()
		return err
	}

	stmt, err := db.Prepare("DELETE FROM RULE_DAY_RANGE WHERE FATHER = ?")
	if err != nil {
		return err
	}

	res, err := stmt.Exec(nFather)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

//GetDayRangeByFather __
func GetDayRangeByFather(nFather int64) ([]DayRangeInfo, error) {
	arrayResult := make([]DayRangeInfo, 0)
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		//db.Close()
		return arrayResult, err
	}

	strSQL := fmt.Sprintf(`SELECT ID, FATHER, LATE_TIME, LEFTE_EARLY,
	OFFWORK_BEGIN, OFFWORK_END, OFFWORK_CHECK, TIME1, 
	TIME2, WORK_BEGIN, WORK_END, WORK_CHECK, DAYS,FATHER ,FATHER_NAME FROM RULE_DAY_RANGE
	WHERE FATHER = %d`, nFather)

	rows, err := db.Query(strSQL)
	if err != nil {
		beego.Error(err)
		return arrayResult, err
	}

	for rows.Next() {
		var structRange DayRangeInfo
		err = rows.Scan(&structRange.ID, &structRange.Father, &structRange.LateTime, &structRange.LeftEarly,
			&structRange.OffworkBegin, &structRange.OffworkEnd, &structRange.OffworkCheck, &structRange.Time1,
			&structRange.Time2, &structRange.WorkBegin, &structRange.WorkEnd, &structRange.WorkCheck, &structRange.Days,
			&structRange.Father, &structRange.FatherName)

		if err != nil {
			beego.Warning(err)
		}

		arrayResult = append(arrayResult, structRange)
	}

	return arrayResult, nil
}

func InserSyncOpt(nDeviceID int64, nEmployeeID int64, nOpt int64) error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		//db.Close()
		return err
	}

	//strSQL := fmt.Sprintf("INSERT INTO SYNC_DEVICE_AND_EMPLOYEE (DEVICE_ID, EMPLOYEE_ID) VALUES(%d, %d);", nDeviceID, nEmployeeID)
	stmt, err := db.Prepare("INSERT INTO SYNC_DEVICE_AND_EMPLOYEE (DEVICE_ID, EMPLOYEE_ID, SYNC) VALUES(?, ?, ?)")
	if err != nil {
		return err
	}

	res, err := stmt.Exec(nDeviceID, nEmployeeID, nOpt)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

func GetSycOptByContent(nDeviceID int64, nEmployeeID int64) (int64, error) {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		//db.Close()
		return 0, err
	}

	strSQL := fmt.Sprintf("SELECT SYNC FROM SYNC_DEVICE_AND_EMPLOYEE WHERE EMPLOYEE_ID = %d AND DEVICE_ID = %d", nEmployeeID, nDeviceID)
	rows, err := db.Query(strSQL)
	if err != nil {
		return 0, err
	}

	nSync := int64(0)
	for rows.Next() {
		err = rows.Scan(&nSync)
		if err != nil {
			beego.Warning(err)
		}
	}

	return nSync, nil
}

func RemoveOptByContent(nDeviceID int64, nEmployeeID int64) error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		//db.Close()
		return err
	}

	stmt, err := db.Prepare("DELETE FROM SYNC_DEVICE_AND_EMPLOYEE WHERE DEVICE_ID = ? AND EMPLOYEE_ID = ?")
	if err != nil {
		return err
	}

	res, err := stmt.Exec(nDeviceID, nEmployeeID)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()

	if err != nil {
		return err
	}

	return nil
}
