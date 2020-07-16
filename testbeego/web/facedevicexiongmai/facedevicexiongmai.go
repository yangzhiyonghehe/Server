package facedevicexiongmai

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/astaxie/beego"
	"github.com/gorilla/websocket"
)

var (
	mu              sync.Mutex
	mapIPSerial     = make(map[string]string, 0)
	mapActiveResult = make(map[string]string, 0)
	mapIPPSW        = make(map[string]string, 0)
)

//GetDeviceByIP 根据Ip获取设备序列号
func GetDeviceInfoByIP(strIP string) (map[string]string, error) {
	mapResult := make(map[string]string, 0)
	strSerial, key := mapIPSerial[strIP]
	if key {
		mapResult["serial"] = strSerial
		delete(mapIPSerial, strIP)

		strPSW, key := mapIPPSW[strIP]
		if key {
			mapResult["psw"] = strPSW
			delete(mapIPPSW, strIP)
		}

		return mapResult, nil
	}

	return mapResult, errors.New("设备信息错误或不存在该设备")
}

//GetHostIP 获取有效的本地Ip
func GetHostIP() (string, error) {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		//fmt.Println("net.Interfaces failed, err:", err.Error())
		return "", err
	}

	for i := 0; i < len(netInterfaces); i++ {
		if (netInterfaces[i].Flags & net.FlagUp) != 0 {
			addrs, _ := netInterfaces[i].Addrs()

			for _, address := range addrs {
				if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						//fmt.Println(ipnet.IP.String())
						return ipnet.IP.String(), nil
					}
				}
			}
		}
	}

	return "", errors.New("未获取到有效Ip")
}

//GetHostIPList 获取本地网卡IP列表
func GetHostIPList() ([]string, error) {
	arrayIP := make([]string, 0)
	netInterfaces, err := net.Interfaces()
	if err != nil {
		//fmt.Println("net.Interfaces failed, err:", err.Error())
		return []string{}, err
	}

	for i := 0; i < len(netInterfaces); i++ {
		if (netInterfaces[i].Flags & net.FlagUp) != 0 {
			addrs, _ := netInterfaces[i].Addrs()

			for _, address := range addrs {
				if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						//fmt.Println(ipnet.IP.String())
						arrayIP = append(arrayIP, ipnet.IP.String())
					}
				}
			}
		}
	}

	beego.Debug("网卡列表:", arrayIP)

	return arrayIP, nil
}

//ParseUDPDevice   解析设备返回的数据信息
func ParseUDPDevice(strBuffer string, strSrcAddr string) error {
	beego.Debug("接收到UDP数据:", strBuffer, "地址:", strSrcAddr)
	//设备查询命令返回
	if isOk, _ := regexp.MatchString("ActResp: yes I am online*", strBuffer); isOk {
		reg := regexp.MustCompile(`id=[0-9a-z]+`)
		//beego.Debug(reg.FindAllString(strBuffer, -1))
		arrayFindIndex := reg.FindAllString(strBuffer, -1)
		if len(arrayFindIndex) != 1 {
			return errors.New("接收到的数据错误")
		}
		arrayIndex := strings.Split(arrayFindIndex[0], "=")
		if len(arrayIndex) != 2 {
			return errors.New("接收到的数据错误")
		}

		strDeviceID := arrayIndex[1]
		//beego.Debug("deviceid:", strDeviceID)

		reg = regexp.MustCompile(`password=[0-9a-z]+`)
		//beego.Debug(reg.FindAllString(strBuffer, -1))
		arrayFindIndex = reg.FindAllString(strBuffer, -1)
		if len(arrayFindIndex) != 1 {
			return errors.New("接收到的数据错误")
		}
		arrayIndex = strings.Split(arrayFindIndex[0], "=")
		if len(arrayIndex) != 2 {
			return errors.New("接收到的数据错误")
		}

		strPSW := arrayIndex[1]
		//beego.Debug("password:", strPSW)
		mapIPPSW[strSrcAddr] = strPSW

		//beego.Debug("over from", strSrcAddr)

		mapIPSerial[strSrcAddr] = strDeviceID
	}

	//后台服务地址设置命令返回
	if isOk, _ := regexp.MatchString("ActResp: activate*", strBuffer); isOk {
		reg := regexp.MustCompile(`activate [A-Z]+`)
		arrayFindIndex := reg.FindAllString(strBuffer, -1)
		if len(arrayFindIndex) != 1 {
			return errors.New("接收到的数据错误")
		}
		arrayIndex := strings.Split(arrayFindIndex[0], " ")
		if len(arrayIndex) != 2 {
			return errors.New("接收到的数据错误")
		}

		strReult := arrayIndex[1]
		//beego.Debug("结果:", strReult)

		reg = regexp.MustCompile(`id=[0-9a-z]+`)
		arrayFindIndex = reg.FindAllString(strBuffer, -1)
		if len(arrayFindIndex) != 1 {
			return errors.New("接收到的数据错误")
		}
		arrayIndex = strings.Split(arrayFindIndex[0], "=")
		if len(arrayIndex) != 2 {
			return errors.New("接收到的数据错误")
		}

		strID := arrayIndex[1]
		//beego.Debug("deviceid:", strID)

		mapActiveResult[strID] = strReult

		// if strReult != "OK"{

		// }else{

		// }
	}

	//设备Ip配置命令返回
	if isOk, _ := regexp.MatchString("ActResp: set ip address*", strBuffer); isOk {

	}

	return nil
}

//UDPConn UDP服务句柄
var UDPConn *net.UDPConn

//StartUDPSEVER 开启广播监听
func StartUDPSEVER() error {
	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:7030")
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}

	UDPConn = conn

	return nil
}

//UDPRecvGoroutine UDP服务接收goroutine
func UDPRecvGoroutine() {
	defer UDPConn.Close()

	//beego.Info("UDP服务监听在:", addr)

	buf := make([]byte, 1024)
	for {
		n, a, err := UDPConn.ReadFromUDP(buf)
		if err != nil {
			//fmt.Println("Failed to read: ", err)
			continue
		}
		//fmt.Printf("Received %s from %s\n", string(buf[0:n]), a)

		err = ParseUDPDevice(string(buf[0:n]), a.IP.String())
		if err != nil {
			beego.Error(err)
			return
		}
	}
}

//SendUdpTagEx 根据网卡列表发送消息
func SendUdpTagEx(strRequst string, arrayHost []string) error {
	for _, value := range arrayHost {
		//mapResult := make(map[string]string, 0)
		// 这里设置发送者的IP地址，自己查看一下自己的IP自行设定
		laddr, err := net.ResolveUDPAddr("udp", value+":7030")
		if err != nil {
			return err
		}

		// 这里设置接收者的IP地址为广播地址
		raddr := net.UDPAddr{
			IP:   net.IPv4(255, 255, 255, 255),
			Port: 7030,
		}
		conn, err := net.DialUDP("udp", laddr, &raddr)
		if err != nil {
			return err
		}

		_, err = conn.Write([]byte(strRequst))
		if err != nil {
			return err
		}

		//		beego.Debug("发送UDP：", laddr)

		conn.Close()
	}

	return nil
}

//SendUpdTag 广播udp数据
func SendUpdTag(strRequst string) error {
	arrayHost, err := GetHostIPList()
	if err != nil {
		return err
	}

	SendUdpTagEx(strRequst, arrayHost)

	return nil
}

//DeviceSearch 广播查询网段内在线的设备
func DeviceSearch() error {
	//mapResult := make(map[string]string, 0)
	//beego.Debug("发送数据")
	err := SendUpdTag("ActReq: are you online?")
	if err != nil {
		return err
	}

	return nil
}

//DeviceSetServer 设置设备发送数据的服务地址
func DeviceSetServer(strSerial string, strPsw string) error {
	strHostIP, err := GetHostIP()
	if err != nil {
		return err
	}

	strCmd := fmt.Sprintf("ActReq: activate device, id=%64s server=%64s password=%64s over", strSerial, strHostIP, strPsw)

	//beego.Debug("发送数据")

	err = SendUpdTag(strCmd)
	if err != nil {
		return err
	}

	spanIndex, err := time.ParseDuration("0.2s")
	if err != nil {
		return err
	}

	for nIndex := 2000; nIndex > 0; nIndex-- {
		value, key := mapActiveResult[strSerial]
		if key && value == "OK" {
			return nil
		}

		time.Sleep(spanIndex)
	}

	return errors.New("设备激活失败")
}

//DeviceUnbingServer 解除服务地址
func DeviceUnbingServer(strSerial string) error {
	strCmd := fmt.Sprintf("ActReq: activate device, id=%64s server=%64s password=%64s over", strSerial, "0.0.0.0", "")

	err := SendUpdTag(strCmd)
	if err != nil {
		return err
	}

	spanIndex, err := time.ParseDuration("0.1s")
	if err != nil {
		return err
	}
	time.Sleep(spanIndex)
	if mapActiveResult[strSerial] != "OK" {
		spanIndex, err = time.ParseDuration("2s")
		if err != nil {
			return err
		}
		time.Sleep(spanIndex)
		if mapActiveResult[strSerial] != "OK" {
			return errors.New("激活设备失败")
		}
	}
	return nil
}

var addr = flag.String("addr", "127.0.0.1:443", "http service address")
var mapWSSCmd = make(map[string][]byte, 0)
var wssconn *websocket.Conn

//CommonCMDJSON 公用的指令数据
type CommonCMDJSON struct {
	Cmd     string `json:"cmd"`
	Message string `json:"message"`
	Code    int64  `json:"code"`
}

//StartWssCLient 开启wss客户端
func StartWssCLient() error {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "wss", Host: *addr, Path: "/websocket"}
	beego.Debug("connecting to ", u.String())

	config := tls.Config{RootCAs: nil, InsecureSkipVerify: true}
	d := websocket.Dialer{TLSClientConfig: &config}
	c, _, err := d.Dial(u.String(), nil)
	if err != nil {
		return err
	}

	wssconn = c

	return nil
}

//WSSRecvGoroutine 接收数据的goroutine
func WSSRecvGoroutine() {
	done := make(chan struct{})

	defer wssconn.Close()
	defer close(done)
	for {
		_, message, err := wssconn.ReadMessage()
		if err != nil {
			beego.Error("wss接收goroutine关闭", err)

			timeSpan, err := time.ParseDuration("3s")
			if err != nil {
				beego.Error(err)
				return
			}
			for {
				err = StartWssCLient()
				if err != nil {
					beego.Error("开启wss客户端失败, 3s后重连", err)
					time.Sleep(timeSpan)
					continue
				}

				break
			}

			err = TellDBPathToWSS()
			if err != nil {
				beego.Error(err)
				return
			}

			return
		}
		if string(message) != `{"cmd":"heartbeat"}` {
			//beego.Debug("接收到WSS服务的数据:", string(message))
		}

		var structCommoncmd CommonCMDJSON
		err = json.Unmarshal(message, &structCommoncmd)
		if err != nil {
			beego.Error("wss接收goroutine关闭", err)
			return
		}

		//beego.Debug("接收到WSS服务的数据:", string(message))

		mapWSSCmd[structCommoncmd.Cmd] = message
	}
}

//SendPackageToWSS  发送数据到wss服务
func SendPackageToWSS(byteSendPack []byte, bGetReply bool) ([]byte, error) {
	var strcutCommonCmd CommonCMDJSON
	err := json.Unmarshal(byteSendPack, &strcutCommonCmd)
	if err != nil {
		return []byte(""), err
	}

	delete(mapWSSCmd, strcutCommonCmd.Cmd+"_resp")

	err = wssconn.WriteMessage(websocket.TextMessage, []byte(byteSendPack))
	if err != nil {
		return []byte(""), err
	}

	if bGetReply {
		spanIndex, err := time.ParseDuration("0.2s")
		if err != nil {
			return []byte(""), err
		}
		// time.Sleep(spanIndex)
		// byteVlaue, key := mapWSSCmd[strcutCommonCmd.Cmd+"_resp"]
		// if !key {
		// 	spanIndex, err = time.ParseDuration("3s")
		// 	time.Sleep(spanIndex)
		// 	byteVlaue, key = mapWSSCmd[strcutCommonCmd.Cmd+"_resp"]
		// 	if !key {
		// 		return []byte(""), errors.New("设备无回应")
		// 	}

		// }

		for nIndex := 0; nIndex < 10; nIndex++ {
			byteVlaue, key := mapWSSCmd[strcutCommonCmd.Cmd+"_resp"]
			if key {
				return byteVlaue, nil
			}
			time.Sleep(spanIndex)
		}

		return []byte(""), errors.New("设备无回应")
	} else {
		return []byte(""), nil
	}

}

//GetRootPath 获取当前目录
func GetRootPath() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return dir
	}

	dir = strings.Replace(dir, "\\", "/", -1)

	return dir
}

//CmdWSSCommonJSON 向Wss服务发送获取信息的指令
type CmdWSSCommonJSON struct {
	Cmd  string `json:"cmd"`
	Data string `json:"data"`
}

//CmdWSSCommonReplyJSON WSS返回的公共解析指令
type CmdWSSCommonReplyJSON struct {
	Cmd     string `json:"cmd"`
	Code    int64  `json:"code"`
	Message string `json:"message"`
}

//TellDBPathToWSS 给WSS服务发送数据库地址
func TellDBPathToWSS() error {
	var structRequst CmdWSSCommonJSON
	structRequst.Cmd = "db_path"
	structRequst.Data = GetRootPath() + "/beegoServer.db"

	byteRequst, err := json.Marshal(structRequst)
	if err != nil {
		return err
	}

	byteReply, err := SendPackageToWSS(byteRequst, true)
	if err != nil {
		return err
	}

	var structReply CmdWSSCommonReplyJSON
	err = json.Unmarshal(byteReply, &structReply)
	if err != nil {
		return err
	}

	if structReply.Code != 200 {
		return errors.New(structReply.Message)
	}

	return nil
}

//RemoveDeviceWSS 让WSS服务关闭和该设备的相关通讯
func RemoveDeviceWSS(strSerial string) error {
	var structRequst CmdWSSCommonJSON
	structRequst.Cmd = "remove_device"
	structRequst.Data = strSerial

	byteRequst, err := json.Marshal(structRequst)
	if err != nil {
		return err
	}

	byteReply, err := SendPackageToWSS(byteRequst, true)
	if err != nil {
		return err
	}

	var structReply CmdWSSCommonReplyJSON
	err = json.Unmarshal(byteReply, &structReply)
	if err != nil {
		return err
	}

	if structReply.Code != 200 {
		return errors.New(structReply.Message)
	}

	return nil
}

//EmployeeInfoJSON 人员信息
type EmployeeInfoJSON struct {
	UserID      string `json:"user_id"`
	CardID      string `json:"card_id"`
	GroupID     string `json:"group_id"`
	Name        string `json:"name"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	FaceFeature string `json:"face_jpg"`
}

//DeviceInfoJSON 设备信息
type DeviceInfoJSON struct {
	DeviceID string `json:"device_id"`
}

//AddEmployeeInfoJSON 数据
type AddEmployeeInfoJSON struct {
	DeviceInfo  DeviceInfoJSON   `json:"device_info"`
	EmployeInfo EmployeeInfoJSON `json:"employee_info"`
}

//AddEmployeeWSSJSON 人员信息指令
type AddEmployeeWSSJSON struct {
	Cmd  string              `json:"cmd"`
	Data AddEmployeeInfoJSON `json:"data"`
}

//AddEmployeeWSS 让WSS服务添加人员信息到雄迈设备
func AddEmployeeWSS(data AddEmployeeInfoJSON) error {
	var structRequest AddEmployeeWSSJSON
	structRequest.Data = data
	structRequest.Cmd = "add_employee"

	byteRequst, err := json.Marshal(structRequest)
	if err != nil {
		return err
	}

	_, err = SendPackageToWSS(byteRequst, false)
	if err != nil {
		return err
	}

	//beego.Debug("添加人员指令:", string(byteRequst[:1000]), ", 返回结果:", string(byteReply))

	// var structReply CmdWSSCommonReplyJSON
	// err = json.Unmarshal(byteReply, &structReply)
	// if err != nil {
	// 	return err
	// }

	// if structReply.Code != 200 {
	// 	return errors.New(structReply.Message)
	// }

	return nil
}

//EmployeeDeleteInfoJSON 用户信息
type EmployeeDeleteInfoJSON struct {
	UserID string `json:"user_id"`
}

//RemoveEmployeeDataJSON  删除所需数据
type RemoveEmployeeDataJSON struct {
	EmployeeInfo EmployeeDeleteInfoJSON `json:"employee"`
	DeviceInfo   DeviceInfoJSON         `json:"device"`
}

//RemoveEmployeeJSON 删除指令
type RemoveEmployeeJSON struct {
	Cmd  string                 `json:"cmd"`
	Data RemoveEmployeeDataJSON `json:"data"`
}

//RemoveEmployeeWSS 删除设备人员数据
func RemoveEmployeeWSS(data RemoveEmployeeDataJSON) error {
	var structRequst RemoveEmployeeJSON
	structRequst.Cmd = "remove_employee"
	structRequst.Data = data

	byteRequest, err := json.Marshal(structRequst)
	if err != nil {
		return err
	}

	byteReply, err := SendPackageToWSS(byteRequest, true)
	if err != nil {
		return err
	}

	var structReply CmdWSSCommonReplyJSON
	err = json.Unmarshal(byteReply, &structReply)
	if err != nil {
		return err
	}

	if structReply.Code != 200 {
		return errors.New(structReply.Message)
	}

	return nil
}

//OpenDoorInfoJSON __
type OpenDoorInfoJSON struct {
	DeviceID string `json:"deviceID"`
	CtrlType int64  `json:"ctrl_type"`
	Duration int64  `json:"duration"`
}

//OpenDoorRequstJSON __
type OpenDoorRequstJSON struct {
	Cmd  string           `json:"cmd"`
	Data OpenDoorInfoJSON `json:"data"`
}

//OpenDoor __
func OpenDoor(data OpenDoorInfoJSON) error {
	var structRequst OpenDoorRequstJSON
	structRequst.Cmd = "open_door"
	structRequst.Data = data

	byteRequst, err := json.Marshal(structRequst)
	if err != nil {
		return err
	}

	byteReply, err := SendPackageToWSS(byteRequst, true)
	if err != nil {
		return err
	}

	var structReply CmdWSSCommonReplyJSON
	err = json.Unmarshal(byteReply, &structReply)
	if err != nil {
		return err
	}

	if structReply.Code != 200 {
		return errors.New(structReply.Message)
	}

	return nil
}

func OpenDoorDelayRouter(data OpenDoorInfoJSON, nSecond int64) {
	span, err := time.ParseDuration(fmt.Sprintf("%ds", nSecond))
	if err != nil {
		beego.Error(err)
		return
	}

	time.Sleep(span)

	OpenDoor(data)
}

type GetDoorStatusJSON struct {
	Cmd  string           `json:"cmd"`
	Data OpenDoorInfoJSON `json:"data"`
}

type DoorStatutsInfoJSON struct {
	DoorStatus    int64 `json:"door_status"`
	DoorButton    bool  `json:"door_button"`
	DoorMangnetic bool  `json:"door_magnetic"`
	AlarmInput    bool  `json:"alarm_input"`
	PreventPry    bool  `json:"prevent_pry"`
}

type DoorStatusReplyJSON struct {
	Cmd     string              `json:"cmd"`
	Code    int64               `json:"code"`
	Message string              `json:"message"`
	Data    DoorStatutsInfoJSON `json:"data"`
}

func GetDoorStatus(strDeviceID string) (int64, error) {
	var structRequst OpenDoorRequstJSON
	structRequst.Cmd = "get_door_status"
	structRequst.Data.DeviceID = strDeviceID

	byteRequst, err := json.Marshal(structRequst)
	if err != nil {
		return 0, err
	}

	byteReply, err := SendPackageToWSS(byteRequst, true)
	if err != nil {
		return 0, err
	}

	var structReply DoorStatusReplyJSON
	err = json.Unmarshal(byteReply, &structReply)
	if err != nil {
		return 0, err
	}

	if structReply.Code != 200 {
		return 0, errors.New(structReply.Message)
	}

	if structReply.Data.DoorStatus == 0 {
		return 1, nil
	}

	return 0, nil
}

//ChangePWDDataJSON _
type ChangePWDDataJSON struct {
	DeviceID string `json:"deviceID"`
	OldPwd   string `json:"old_password"`
	NewPwd   string `json:"new_password"`
}

//ChangePWDJSON _
type ChangePWDJSON struct {
	Cmd  string            `json:"cmd"`
	Data ChangePWDDataJSON `json:"data"`
}

//ChangeXMPWD __
func ChangeXMPWD(data ChangePWDDataJSON) error {
	var structRequst ChangePWDJSON
	structRequst.Cmd = "change_password"
	structRequst.Data = data

	byteRequst, err := json.Marshal(structRequst)
	if err != nil {
		return err
	}

	byteReply, err := SendPackageToWSS(byteRequst, true)
	if err != nil {
		return err
	}

	var structReply CmdWSSCommonReplyJSON
	err = json.Unmarshal(byteReply, &structReply)
	if err != nil {
		return err
	}

	if structReply.Code != 200 {
		return errors.New(structReply.Message)
	}

	return nil
}

//GetXMIOCfg __
func GetXMIOCfg(strDeviceID string) (map[string]interface{}, error) {
	mapReply := make(map[string]interface{}, 0)

	mapRequst := make(map[string]interface{}, 0)
	mapRequst["cmd"] = "get_io_cfg"
	mapRequst["data"] = map[string]string{"deivceID": strDeviceID}

	byteRequst, err := json.Marshal(mapRequst)

	byteReply, err := SendPackageToWSS(byteRequst, true)
	if err != nil {
		return mapReply, err
	}

	err = json.Unmarshal(byteReply, &mapReply)
	if err != nil {
		return mapReply, err
	}

	return mapReply, nil
}

//SetXMIOCfg __
func SetXMIOCfg(strDeviceID string, mapParam map[string]interface{}) error {
	mapReply := make(map[string]interface{}, 0)

	mapRequst := make(map[string]interface{}, 0)
	mapRequst["cmd"] = "set_io_cfg"
	mapRequst["data"] = mapParam
	mapRequst["deviceID"] = strDeviceID

	byteRequst, err := json.Marshal(mapRequst)

	byteReply, err := SendPackageToWSS(byteRequst, true)
	if err != nil {
		return err
	}

	err = json.Unmarshal(byteReply, &mapReply)
	if err != nil {
		return err
	}

	if mapReply["ret"].(float64) == 200 {
		return nil
	}

	return errors.New(mapReply["msg"].(string))
}

//RestartXMDevice __
func RestartXMDevice(strDeviceID string) error {
	mapRequst := make(map[string]interface{}, 0)

	mapRequst["cmd"] = "restart"
	mapRequst["deviceID"] = strDeviceID

	byteRequest, err := json.Marshal(mapRequst)
	if err != nil {
		return err
	}

	byteReply, err := SendPackageToWSS(byteRequest, true)
	if err != nil {
		return err
	}

	mapReply := make(map[string]interface{}, 0)
	err = json.Unmarshal(byteReply, &mapReply)

	if mapReply["ret"].(float64) == 200 {
		return nil
	}

	return errors.New(mapReply["msg"].(string))
}
