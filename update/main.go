package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/astaxie/beego"
	_ "github.com/mattn/go-sqlite3"
	termbox "github.com/nsf/termbox-go"
)

type VersionInfo struct {
	TitileVer int64
	MainVer   int64
}

func GetRootPath() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return dir
	}

	dir = strings.Replace(dir, "\\", "/", -1)

	return dir
}

//GetServerVersion __
func GetServerVersion() (VersionInfo, error) {
	var structResult VersionInfo
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		//db.Close()
		return structResult, err
	}

	rows, err := db.Query("SELECT TITLE_VER, MAIN_VER FROM SERVER_VERSION")
	if err != nil {
		return structResult, err
	}

	for rows.Next() {
		err = rows.Scan(&structResult.TitileVer, &structResult.MainVer)
		if err != nil {
			beego.Warning(err)
		}
	}

	return structResult, nil
}

//UpdateVersion __
func UpdateVersion(data VersionInfo) error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		//db.Close()
		return err
	}

	_, err = db.Exec("DELETE  FROM SERVER_VERSION")
	if err != nil {
		return err
	}

	stmt, err := db.Prepare("INSERT INTO SERVER_VERSION (TITLE_VER, MAIN_VER) VALUES (?, ?)")
	if err != nil {
		return err
	}

	res, err := stmt.Exec(data.TitileVer, data.MainVer)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

var arrayCreateSQL []string = []string{
	`CREATE TABLE "SYNC_DEVICE_AND_EMPLOYEE" (
	  "EMPLOYEE_ID" INTEGER,
	  "DEVICE_ID" INTEGER,
	  "SYNC" integer
	);`}

var arraySQL []string = []string{`ALTER TABLE COMPANY_TABLE ADD LOCATION VARCHAR(20);`, `ALTER TABLE ADMIN_USER ADD COMPANYS VARCHAR(20);`,
	`ALTER TABLE VERIFY_PERSON ADD RESULT VARCHAR(20);`, `ALTER TABLE VERIFY_PERSON ADD RANGE_INDEX INTERGER;`,
	`ALTER TABLE VERIFY_PERSON ADD ATTENDANCE_TIME VARCHAR(10);`, `ALTER TABLE VERIFY_PERSON ADD DEVICE_NAME VARCHAR(10);`,
	`ALTER TABLE VERIFY_PERSON ADD CREATE_DATE VARCHAR(10);`, `ALTER TABLE EMPLOYEE_INFO ADD COMPANY_ID INTEGER;`,
	`ALTER TABLE VERIFY_PERSON ADD "COMPANY_ID"  INTERGER;`, `ALTER TABLE DEVICE_TABLE ADD ACTIVE_TIME INTERGER;`,
	`ALTER TABLE DEVICE_TABLE ADD 'IS_ATTENDANCE' INTEGER;`, `ALTER TABLE DEVICE_TABLE ADD 'IS_SHOW_DOOR_STATUS'  INTEGER;`,
	`ALTER TABLE DEVICE_TABLE ADD 'MULTI_CERT'  INTEGER;`, `ALTER TABLE COMPANY_TABLE ADD 'MULTI_CERT'  INTEGER;`}

func UpdateTable() error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		//db.Close()
		return err
	}

	for _, valueSQL := range arraySQL {
		_, err := db.Exec(valueSQL)
		if err != nil {
			beego.Warning(err)
		}
	}

	db.Close()
	return nil
}

func CreateTable() error {
	db, err := sql.Open("sqlite3", GetRootPath()+"/beegoServer.db")
	if err != nil {
		//db.Close()
		return err
	}

	for _, valueSQL := range arrayCreateSQL {
		_, err := db.Exec(valueSQL)
		if err != nil {
			beego.Warning(err)
		}

		// _, err = res.RowsAffected()
		// if err != nil {
		// 	return err
		// }
	}

	db.Close()
	return nil
}

func init() {
	if err := termbox.Init(); err != nil {
		panic(err)
	}
	termbox.SetCursor(0, 0)
	termbox.HideCursor()
}

func pause() {
	fmt.Println("请按任意键继续...")
Loop:
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			break Loop
		}
	}
}

func LogConfig() {
	err := os.Mkdir("C:/AttendanceLog/", os.ModePerm)
	if err != nil {
		fmt.Println(err)
	}

	beego.SetLevel(beego.LevelInformational)

	beego.BeeLogger.EnableFuncCallDepth(true)
	beego.BeeLogger.SetLogFuncCallDepth(4)
	beego.SetLogger("console", "")
	beego.SetLogger("file", `{"filename":"C:/AttendanceLog/Update.log"}`)

}

func main() {
	LogConfig()

	CreateTable()

	UpdateTable()

	// var structNewVersionInfo VersionInfo
	// structNewVersionInfo.MainVer = 4
	// structNewVersionInfo.TitileVer = 1

	// structVersionInfo, err := GetServerVersion()
	// if err != nil {
	// 	//fmt.Println("警告:", err)
	// 	beego.Warning(err)

	// 	err := UpdateTable()
	// 	if err != nil {
	// 		//fmt.Println("错误：", err)
	// 		beego.Error(err)
	// 		//pause()
	// 		return
	// 	}

	// 	CreateTable()
	// 	if err != nil {
	// 		//fmt.Println("错误：", err)
	// 		//pause()
	// 		beego.Error(err)
	// 		return
	// 	}

	// 	//fmt.Println("更新成功")
	// 	beego.Info("更新成功")

	// 	UpdateVersion(structNewVersionInfo)
	// 	//pause()

	// 	return
	// }

	// if structVersionInfo.MainVer < structNewVersionInfo.MainVer || structVersionInfo.TitileVer < structNewVersionInfo.TitileVer {
	// 	err := UpdateTable()
	// 	if err != nil {
	// 		//fmt.Println("错误：", err)
	// 		beego.Error(err)
	// 		//pause()
	// 		return
	// 	}

	// 	CreateTable()
	// 	if err != nil {
	// 		//fmt.Println("错误：", err)
	// 		beego.Error(err)
	// 		//pause()
	// 		return
	// 	}

	// }

	// beego.Info("更新成功")
	// UpdateVersion(structNewVersionInfo)
	//pause()
}
