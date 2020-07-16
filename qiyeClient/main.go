package main

import (
	"fmt"
	"os"
	"time"

	"./facedevicexiongmai"
	"./qiyewss"
	_ "./routers"
	"github.com/astaxie/beego"
	"github.com/kardianos/service"
)

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
	beego.SetLogger("file", `{"filename":"C:/AttendanceLog/qiyeClient.log"}`)
}

var logger = service.ConsoleLogger

type program struct{}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *program) run() {
	LogConfig()

	//err := facedevicexiongmai.SatrtWSClient()

	err := facedevicexiongmai.StartUDPSEVER()
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

	err = facedevicexiongmai.StartWssCLient()
	if err != nil {
		beego.Error("连接本地WSS服务失败", err)
		time.Sleep(timeSpan)
		return
	}

	go facedevicexiongmai.WSSRecvGoroutine()

	err = facedevicexiongmai.TellDBPathToWSS()
	if err != nil {
		beego.Error(err)
		return
	}

	err = qiyewss.StartWssCLient()
	if err != nil {
		beego.Error("连接企业微信服务失败:", err)
		return
	}

	go qiyewss.WSSRecvGoroutine()
	//go qiyewss.SendWSSMsgGoroutine()

	structRegistResult, err := qiyewss.Register()
	if err != nil {
		return
	}

	structActiveResult, err := qiyewss.ActiveDevice(structRegistResult.Body["active_code"].(string))
	if err != nil {
		return
	}

	_, err = qiyewss.SubScribeCorp(structActiveResult.Body["secret"].(string))
	if err != nil {
		return
	}

	go qiyewss.HeartGoRoutine()

	// structUserInfoResult, err := qiyewss.GetUserInfoByPage()
	// if err != nil {
	// 	beego.Error(err)
	// 	return
	// }

	// byteResult, err := json.Marshal(structUserInfoResult)
	// if err != nil {
	// 	beego.Error(err)
	// 	return
	// }

	// beego.Debug(string(byteResult))

	beego.Run()
}

func (p *program) Stop(s service.Service) error {
	return nil
}

func main() {
	svcConfig := &service.Config{
		Name:        "QiyeWSSServcer", //服务显示名称
		DisplayName: "QiyeWSSServcer", //服务名称
		Description: "企业微信服务",         //服务描述
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
