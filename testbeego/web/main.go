package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"./confreader"
	"./facedevicexiongmai"
	"./my_db"
	_ "./routers"
	"github.com/astaxie/beego"
	"github.com/kardianos/service"
)

var logger = service.ConsoleLogger

type program struct{}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *program) run() {
	err := os.Mkdir("C:/AttendanceLog/", os.ModePerm)
	if err != nil {
		fmt.Println(err)
	}

	beego.SetLogger("file", `{"filename":"C:/AttendanceLog/test.log"}`)
	beego.SetLevel(beego.LevelDebug)

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	dir = strings.Replace(dir, "\\", "/", -1)

	beego.SetViewsPath(dir + "/views")
	beego.SetStaticPath(`/manager`, dir+"/manager")

	beego.Info("开始服务")

	err = facedevicexiongmai.StartUDPSEVER()
	if err != nil {
		beego.Error("开启UDP服务失败", err)
		return
	}
	go facedevicexiongmai.UDPRecvGoroutine()

	timeSpan, err := time.ParseDuration("3s")
	if err != nil {
		beego.Error(err)
		return
	}
	for {
		err = facedevicexiongmai.StartWssCLient()
		if err != nil {
			beego.Error("开启wss客户端失败, 3s后重连", err)
			time.Sleep(timeSpan)
			continue
		}

		break
	}

	go facedevicexiongmai.WSSRecvGoroutine()

	err = facedevicexiongmai.TellDBPathToWSS()
	if err != nil {
		beego.Error(err)
		return
	}

	// var WSSTellHTTP facedevicexiongmai.CmdWSSCommonJSON
	// WSSTellHTTP.Cmd = "http_client"
	// WSSTellHTTP.Data = ""
	// byteRequst, err := json.Marshal(WSSTellHTTP)
	// if err != nil {
	// 	beego.Error(err)
	// 	return
	// }

	// byteReply, err := facedevicexiongmai.SendPackageToWSS(byteRequst)
	// if err != nil {
	// 	beego.Error(err)
	// 	return
	// }

	// var structReply facedevicexiongmai.CMDClientReplyJSON
	// err = json.Unmarshal(byteReply, &structReply)
	// if err != nil {
	// 	beego.Error(err)
	// 	return
	// }

	// if structReply.Code != 200 {
	// 	beego.Error(structReply.Message)
	// 	return
	// }

	err = confreader.ReadConfFile(my_db.GetRootPath() + "/conf/language.ini")
	if err != nil {
		beego.Error(err)
	}

	go beego.Run()

}

func (p *program) Stop(s service.Service) error {
	return nil
}

func main() {
	svcConfig := &service.Config{
		Name:        "AttendanceServer", //服务显示名称
		DisplayName: "AttendanceServer", //服务名称
		Description: "考勤服务",             //服务描述
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		logger.Error(err)
	}

	if err != nil {
		logger.Error(err)
	}

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "install":
			s.Install()
			logger.Info("服务安装成功!")
			s.Start()
			logger.Info("服务启动成功!")
			break
		case "start":
			s.Start()
			logger.Info("服务启动成功!")
			break
		case "stop":
			s.Stop()
			logger.Info("服务关闭成功!")
			break
		case "restart":
			s.Stop()
			logger.Info("服务关闭成功!")
			s.Start()
			logger.Info("服务启动成功!")
			break
		case "remove":
			s.Stop()
			logger.Info("服务关闭成功!")
			s.Uninstall()
			logger.Info("服务卸载成功!")
			break
		}
		return
	}

	err = s.Run()
	if err != nil {
		fmt.Println("服务运行失败:", err)
	}
}
