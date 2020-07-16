package main

import (
	"fmt"
	"os"

	"./wssserver"
	"github.com/astaxie/beego"
	"github.com/kardianos/service"
)

///home/qyweixin/log
//LogConfig  配置日志
func LogConfig() {
	err := os.Mkdir("C:/AttendanceLog/", os.ModePerm)
	if err != nil {
		fmt.Println(err)
	}

	beego.SetLevel(beego.LevelDebug)

	beego.BeeLogger.EnableFuncCallDepth(true)
	beego.BeeLogger.SetLogFuncCallDepth(4)
	beego.SetLogger("console", "")
	beego.SetLogger("file", `{"filename":"C:/AttendanceLog/wssServer.log"}`)
}

var logger = service.ConsoleLogger

type program struct{}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *program) run() {
	LogConfig()

	wssserver.StartServer()
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
