package echo

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/astaxie/beego"

	"../LinkList"
	"../mysqlitemanger"
)

var MapAddrAndConn = make(map[string]bool, 0)

const TIME_LAYOUT = "2006-01-02 15:04:05"

//MapIPSendChan 发送缓冲区
var MapIPSendChan = make(map[string]chan []byte, 20)

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

//LoginReplyJSON 登录返回
type LoginReplyJSON struct {
	Cmd string `json:"cmd"`
	Ret int64  `json:"ret"`
	Msg string `json:"msg"`
}

//LoginDataInfoJSON 登录信息
type LoginDataInfoJSON struct {
	DeviceID    string `json:"device_id"`
	Password    string `json:"password"`
	DeviceType  string `json:"device_type"`
	AppVer      string `json:"app_ver"`
	FirmwareVer string `json:"firmware_ver"`
}

//LoginRequstJSON 登录接收
type LoginRequstJSON struct {
	Cmd  string            `json:"cmd"`
	Data LoginDataInfoJSON `json:"data"`
}

//Login  登录处理
func Login(byteJSON []byte, strAddr string) ([]byte, error) {
	var loginRqust LoginRequstJSON
	err := json.Unmarshal(byteJSON, &loginRqust)
	if err != nil {
		return []byte(""), err
	}

	var structReply LoginReplyJSON

	structDeviceLoginInfo, err := mysqlitemanger.GetDeviceBySerial(loginRqust.Data.DeviceID)
	if err != nil {
		structReply.Cmd = "login_resp"
		structReply.Msg = err.Error()
		structReply.Ret = 500
		byteReply, err := json.Marshal(structReply)
		if err != nil {
			return []byte(""), err
		}

		return byteReply, nil
	}

	if structDeviceLoginInfo.Effect <= 0 {
		structReply.Cmd = "login_resp"
		structReply.Msg = "已删除设备"
		structReply.Ret = 500
		byteReply, err := json.Marshal(structReply)
		if err != nil {
			return []byte(""), err
		}

		// err = mysqlitemanger.UpdateDeviceStatusBySerial(0, loginRqust.Data.DeviceID)
		// if err != nil {
		// 	return byteReply, err
		// }

		//		beego.Debug("已删除设备:", strAddr)

		return byteReply, nil
	}

	if loginRqust.Data.Password != structDeviceLoginInfo.PassWord {

		structReply.Cmd = "login_resp"
		structReply.Msg = "密码错误"
		structReply.Ret = 500
		byteReply, err := json.Marshal(structReply)
		if err != nil {
			return []byte(""), err
		}

		// err = mysqlitemanger.UpdateDeviceStatusBySerial(0, loginRqust.Data.DeviceID)
		// if err != nil {
		// 	return byteReply, err
		// }

		beego.Debug("密码错误:", strAddr)

		return byteReply, nil
	}

	var structDeviceInfo mysqlitemanger.DeviceInfo
	structDeviceInfo.AppVer = loginRqust.Data.AppVer
	structDeviceInfo.DeviceID = loginRqust.Data.DeviceID
	structDeviceInfo.DeviceType = loginRqust.Data.DeviceType
	structDeviceInfo.FirmwareVer = loginRqust.Data.FirmwareVer
	structDeviceInfo.PassWord = loginRqust.Data.Password

	//	arrayIndex := strings.Split(strAddr, ":")
	structDeviceInfo.Addr = strAddr
	structDeviceInfo.Type = 1
	err = mysqlitemanger.AddDeviceInfo(structDeviceInfo)
	if err != nil {
		structReply.Cmd = "login_resp"
		structReply.Msg = err.Error()
		structReply.Ret = 500
		byteReply, err := json.Marshal(structReply)
		if err != nil {
			return byteReply, err
		}

		return byteReply, nil
	}

	err = mysqlitemanger.UpdateDeviceActiveTime(time.Now().Unix(), structDeviceInfo.DeviceID)
	if err != nil {
		return []byte(""), err
	}

	structReply.Cmd = "login_resp"
	structReply.Msg = "xxx"
	structReply.Ret = 200
	byteReply, err := json.Marshal(structReply)
	if err != nil {
		return []byte(""), err
	}

	return byteReply, nil
}

//HeartRequstJSON 心跳包
type HeartRequstJSON struct {
	Cmd string `json:"cmd"`
}

//GetHeartRequst 获取心跳包
func GetHeartRequst() ([]byte, error) {
	var structRequst HeartRequstJSON
	structRequst.Cmd = "heartbeat"

	byteRequst, err := json.Marshal(structRequst)
	if err != nil {
		return []byte(""), err
	}

	return byteRequst, nil
}

//GetSysLogJSON 获取日志指令
type GetSysLogJSON struct {
	Cmd string `json:"cmd"`
}

//GetSysLogCmd 获取设备日志
func GetSysLogCmd() ([]byte, error) {
	var structRequst GetSysLogJSON
	structRequst.Cmd = "get_log"
	byteRequst, err := json.Marshal(structRequst)
	if err != nil {
		return []byte(""), err
	}

	return byteRequst, nil
}

//SetServerInfoJSON 设置数据服务地址的信息
type SetServerInfoJSON struct {
	URL string `json:"url"`
}

//SetDataServerJSON 设置数据服务地址
type SetDataServerJSON struct {
	Cmd  string            `json:"cmd"`
	Data SetServerInfoJSON `json:"data"`
}

//GetSetDataServerCmd 获取设置数据服务地址的指令
func GetSetDataServerCmd() ([]byte, error) {
	var structRquest SetDataServerJSON

	strHostIP, err := GetHostIP()
	if err != nil {
		return []byte(""), err
	}

	structRquest.Cmd = "set_upload_url"
	//structRquest.Data.URL = "http://iot.ezhangtong.com/api/device/xm/upload"

	structRquest.Data.URL = "http://" + strHostIP + "/api/device/xm/upload"
	//structRquest.Data.URL = "http://zhikong.qidun.cn/api/device/xm/upload"

	byteRequst, err := json.Marshal(structRquest)
	if err != nil {
		return []byte(""), err
	}

	return byteRequst, nil
}

//ValidTimeJSON 时间段数据
type ValidTimeJSON struct {
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	OpType    int64  `json:"op_type"`
}

//GroupInfoJSON 组数据
type GroupInfoJSON struct {
	GrupID     string        `json:"group_id"`
	ValidTime1 ValidTimeJSON `json:"valid_time1"`
	ValidTime2 ValidTimeJSON `json:"valid_time2"`
	ValidTime3 ValidTimeJSON `json:"valid_time3"`
	ValidTime4 ValidTimeJSON `json:"valid_time4"`
	ValidTime5 ValidTimeJSON `json:"valid_time5"`
}

//AddGroupJSON 添加组数据
type AddGroupJSON struct {
	Cmd  string        `json:"cmd"`
	Data GroupInfoJSON `json:"data"`
}

//GetAddGroupCmd  获取添加组数据的指令
func GetAddGroupCmd() ([]byte, error) {
	var structRequest AddGroupJSON
	structRequest.Cmd = "add_group"
	structRequest.Data.GrupID = "default"
	structRequest.Data.ValidTime1.StartTime = "00:00"
	structRequest.Data.ValidTime1.EndTime = "24:00"
	structRequest.Data.ValidTime1.OpType = 2

	structRequest.Data.ValidTime2.StartTime = "00:00"
	structRequest.Data.ValidTime2.EndTime = "00:00"
	structRequest.Data.ValidTime2.OpType = 5

	structRequest.Data.ValidTime3.StartTime = "00:00"
	structRequest.Data.ValidTime3.EndTime = "00:00"
	structRequest.Data.ValidTime3.OpType = 5

	structRequest.Data.ValidTime4.StartTime = "00:00"
	structRequest.Data.ValidTime4.EndTime = "00:00"
	structRequest.Data.ValidTime4.OpType = 5

	structRequest.Data.ValidTime5.StartTime = "00:00"
	structRequest.Data.ValidTime5.EndTime = "00:00"
	structRequest.Data.ValidTime5.OpType = 5

	byteReply, err := json.Marshal(structRequest)
	if err != nil {
		return []byte(""), err
	}

	return byteReply, nil
}

func GetAddForbidGroupCmd() ([]byte, error) {
	var structRequest AddGroupJSON
	structRequest.Cmd = "add_group"
	structRequest.Data.GrupID = "Forbid"
	structRequest.Data.ValidTime1.StartTime = "00:00"
	structRequest.Data.ValidTime1.EndTime = "24:00"
	structRequest.Data.ValidTime1.OpType = 5

	structRequest.Data.ValidTime2.StartTime = "00:00"
	structRequest.Data.ValidTime2.EndTime = "00:00"
	structRequest.Data.ValidTime2.OpType = 5

	structRequest.Data.ValidTime3.StartTime = "00:00"
	structRequest.Data.ValidTime3.EndTime = "00:00"
	structRequest.Data.ValidTime3.OpType = 5

	structRequest.Data.ValidTime4.StartTime = "00:00"
	structRequest.Data.ValidTime4.EndTime = "00:00"
	structRequest.Data.ValidTime4.OpType = 5

	structRequest.Data.ValidTime5.StartTime = "00:00"
	structRequest.Data.ValidTime5.EndTime = "00:00"
	structRequest.Data.ValidTime5.OpType = 5

	byteReply, err := json.Marshal(structRequest)
	if err != nil {
		return []byte(""), err
	}

	return byteReply, nil
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

//AddEmployeeJSON 添加人员信息指令
type AddEmployeeJSON struct {
	Cmd  string           `json:"cmd"`
	Data EmployeeInfoJSON `json:"data"`
}

//GetAddEmployeeCmd 获取添加人员信息
func GetAddEmployeeCmd(data EmployeeInfoJSON) ([]byte, error) {
	var structRequst AddEmployeeJSON
	structRequst.Cmd = "add_user"
	data.FaceFeature = strings.ToUpper(data.FaceFeature)
	structRequst.Data = data

	byteRequst, err := json.Marshal(structRequst)
	if err != nil {
		return []byte(""), err
	}

	return byteRequst, nil
}

//EmployeeDeleteInfoJSON 用户信息
type EmployeeDeleteInfoJSON struct {
	UserID string `json:"user_id"`
}

//EmployeeRemoveJSON 删除人员指令
type EmployeeRemoveJSON struct {
	Cmd  string                 `json:"cmd"`
	Data EmployeeDeleteInfoJSON `json:"data"`
}

//GetRemoveEmoloyeeCmd 获取删除人员指令
func GetRemoveEmoloyeeCmd(data EmployeeDeleteInfoJSON) ([]byte, error) {
	var structRequst EmployeeRemoveJSON
	structRequst.Cmd = "del_user"
	structRequst.Data = data

	byteRequst, err := json.Marshal(structRequst)
	if err != nil {
		return []byte(""), err
	}

	return byteRequst, nil
}

//TimeInfoJSON 时间数据
type TimeInfoJSON struct {
	Time string `json:"time"`
}

//SetTimeJSON 设置时间指令
type SetTimeJSON struct {
	Cmd  string       `json:"cmd"`
	Data TimeInfoJSON `json:"data"`
}

//GetSetTimeCmd 获取设置时间的指令
func GetSetTimeCmd() ([]byte, error) {
	var structRequst SetTimeJSON

	structRequst.Cmd = "set_time"
	structRequst.Data.Time = time.Now().Format(TIME_LAYOUT)
	byteRequst, err := json.Marshal(structRequst)
	if err != nil {
		return []byte(""), err
	}

	return byteRequst, nil
}

//RemoteInfoJSON __
type RemoteInfoJSON struct {
	CtrlType int64 `json:"ctrl_type"`
	Duration int64 `json:"duration"`
}

//RemoteJSON __
type RemoteJSON struct {
	Cmd  string         `json:"cmd"`
	Data RemoteInfoJSON `json:"data"`
}

//GetRemoteCmd __
func GetRemoteCmd(data OpenDoorInfoJSON) ([]byte, error) {
	var structRequst RemoteJSON
	structRequst.Cmd = "remote_ctrl"
	structRequst.Data.CtrlType = data.CtrlType
	structRequst.Data.Duration = data.Duration

	byteRequst, err := json.Marshal(structRequst)
	if err != nil {
		return []byte(""), err
	}

	return byteRequst, nil
}

type ChangePwdInfoJSON struct {
	OldPwd string `json:"old_password"`
	NewPwd string `json:"new_password"`
}

type ChangePwdJSON struct {
	Cmd  string            `json:"cmd"`
	Data ChangePwdInfoJSON `json:"data"`
}

func GetChangePwdCMD(data ChangePWDDataJSON) ([]byte, error) {
	var structRequst ChangePwdJSON
	structRequst.Cmd = "change_password"
	structRequst.Data.NewPwd = data.NewPwd
	structRequst.Data.OldPwd = data.OldPwd

	byteRequst, err := json.Marshal(structRequst)
	if err != nil {
		return []byte(""), err
	}
	return byteRequst, nil
}

//HeartDeviceInfo 心跳包的记录
type HeartDeviceInfo struct {
	SendTime time.Time
	Addr     string
}

var listHeart = LinkList.NewLinkList()

var mutexHeart sync.Mutex

//PushHeart 心跳数据插入队列
func PushHeart(data HeartDeviceInfo) {
	listHeart.TailInsert(data)
	//beego.Debug("listHeart.TailInsert:", data)
}

//HeartResp 返回心跳包解析处理
func HeartResp(byteJSON []byte, strAddr string) error {
	structDeviceInfo, err := mysqlitemanger.GetDeviceByAddr(strAddr)
	if err != nil {
		return err
	}

	err = mysqlitemanger.UpdateDeviceActiveTime(time.Now().Unix(), structDeviceInfo.DeviceID)
	if err != nil {
		return err
	}
	//node := listHeart.GetHead()
	// if node == nil {
	// 	return nil
	// }
	// for node != nil {
	// 	if node.Value.(HeartDeviceInfo).Addr == strIP {
	// 		err := mysqlitemanger.UpdateDeviceStatusByAddr(1, strIP)
	// 		if err != nil {
	// 			return err
	// 		}

	// 		node = listHeart.Remove(node)
	// 	} else {
	// 		node = node.Next
	// 	}
	// }

	// for index, heartInfo := range arrayHeart {
	// 	if strIP == heartInfo.Addr {
	// 		err := mysqlitemanger.UpdateDeviceStatusByAddr(1, heartInfo.Addr)
	// 		if err != nil {
	// 			return err
	// 		}
	// 		beego.Debug("心跳列表", arrayHeart, "index:", index)

	// 		mutexHeart.Lock()
	// 		arrayHeart = append(arrayHeart[:index], arrayHeart[index+1:]...)
	// 		mutexHeart.Unlock()

	// 	}
	// }
	return nil
}

//CheckHearArray 1s检查一次心跳队列
func CheckHearArray() {
	spanIndex, err := time.ParseDuration("1s")
	if err != nil {
		beego.Error("心跳包检查goroutine错误", err)
		return
	}
	for {
		node := listHeart.GetHead()
		if node == nil {
			time.Sleep(spanIndex)
			continue
		}

		listHeart.Print()
		for node != nil {
			timeIndex := node.Value.(HeartDeviceInfo).SendTime
			timeNow := time.Now()
			heartSpan, err := time.ParseDuration("10s")
			if err != nil {
				beego.Error("心跳包检查goroutine错误", err)
				return
			}
			if timeNow.Sub(timeIndex) > heartSpan {
				mysqlitemanger.UpdateDeviceStatusByAddr(0, node.Value.(HeartDeviceInfo).Addr)
				node = listHeart.Remove(node)
			} else {
				node = node.Next
			}
		}
		// for index, heartInfo := range arrayHeart {
		// 	timeIndex := heartInfo.SendTime
		// 	timeNow := time.Now()
		// 	heartSpan, err := time.ParseDuration("10s")
		// 	if err != nil {
		// 		beego.Error("心跳包检查goroutine错误", err)
		// 		return
		// 	}
		// 	if timeNow.Sub(timeIndex) > heartSpan {
		// 		mysqlitemanger.UpdateDeviceStatusByAddr(0, heartInfo.Addr)
		// 		mutexHeart.Lock()
		// 		arrayHeart = append(arrayHeart[:index], arrayHeart[index+1:]...)
		// 		mutexHeart.Unlock()
		// 		break
		// 	}

		// }
		time.Sleep(spanIndex)
	}
}

//SyncDeviceInfo 1秒更新一次数据
func SyncDeviceInfo() {
	spanIndex, err := time.ParseDuration("1s")
	if err != nil {
		beego.Error(err)
		return
	}
	for {

		time.Sleep(spanIndex)
	}

}

//CmdWSSCommonJSON 向Wss服务发送获取信息的指令
type CmdWSSCommonJSON struct {
	Cmd  string `json:"cmd"`
	Data string `json:"data"`
}

//CMDClientReplyJSON 客户端识别返回
type CMDClientReplyJSON struct {
	Cmd     string `json:"cmd"`
	Code    int64  `json:"code"`
	Message string `json:"message"`
}

//HTTPClientAddr HTTP客户端的地址
var HTTPClientAddr string

//DealHTTPClient 处理HTTP客户端识别指令
func DealHTTPClient(byteJSON []byte, strAddr string) ([]byte, error) {
	var structRequst CmdWSSCommonJSON
	err := json.Unmarshal(byteJSON, &structRequst)
	if err != nil {
		return []byte(""), err
	}

	HTTPClientAddr = strAddr

	var structReply CMDClientReplyJSON
	structReply.Cmd = "http_client_resp"
	structReply.Code = 200
	structReply.Message = ""

	byteReply, err := json.Marshal(structReply)
	if err != nil {
		return []byte(""), err
	}

	return byteReply, nil
}

//SetDBPathRequstJSON 设置数据库路径接收消息
type SetDBPathRequstJSON struct {
	Cmd  string `json:"cmd"`
	Data string `json:"data"`
}

//SetDBPathReplyJSON 设置数据库路径返回消息
type SetDBPathReplyJSON struct {
	Cmd     string `json:"cmd"`
	Code    int64  `json:"code"`
	Message string `json:"message"`
}

//SetDBPath 设置数据库的路径
func SetDBPath(byteJSON []byte, strAddr string) ([]byte, error) {
	var structRequest SetDBPathRequstJSON
	err := json.Unmarshal(byteJSON, &structRequest)
	if err != nil {
		return []byte(""), err
	}

	if mysqlitemanger.DBPath == string("") {
		mysqlitemanger.DBPath = structRequest.Data
	}

	var structReply SetDBPathReplyJSON
	structReply.Cmd = "db_path_resp"
	structReply.Code = 200
	structReply.Message = "设置成功"

	byteReply, err := json.Marshal(structReply)
	if err != nil {
		return []byte(""), err
	}

	return byteReply, nil
}

//RemoveDeviceRequst 删除设备接收信息
type RemoveDeviceRequst struct {
	Cmd  string `json:"cmd"`
	Data string `json:"data"`
}

//RemoveDeviceReply 删除设备返回信息
type RemoveDeviceReply struct {
	Cmd     string `json:"cmd"`
	Code    int64  `json:"code"`
	Message string `json:"message"`
}

//DeviceRemoveArray 删除设备列表
var DeviceRemoveArray = make([]string, 0)

//RemoveDevice  删除设备
func RemoveDevice(byteJSON []byte, strAddr string) ([]byte, error) {
	var structRequest RemoveDeviceRequst
	err := json.Unmarshal(byteJSON, &structRequest)
	if err != nil {
		return []byte(""), err
	}

	DeviceRemoveArray = append(DeviceRemoveArray, structRequest.Data)

	var StructRemoveInfo EmployeeDeleteInfoJSON
	StructRemoveInfo.UserID = "ALL"
	byteRequest, err := GetRemoveEmoloyeeCmd(StructRemoveInfo)
	if err != nil {
		return []byte(""), err
	}

	structDevcieInfo, err := mysqlitemanger.GetDeviceBySerial(structRequest.Data)
	if err != nil {
		return []byte(""), err
	}

	MapIPSendChan[structDevcieInfo.Addr] <- byteRequest

	var structReply RemoveDeviceReply
	structReply.Code = 200
	structReply.Cmd = "remove_device_resp"
	structReply.Message = ""

	byteReply, err := json.Marshal(structReply)
	if err != nil {
		return []byte(""), err
	}

	return byteReply, nil
}

//SetUploadURLReplyJSON 更新URL返回数据
type SetUploadURLReplyJSON struct {
	Cmd string `json:"cmd"`
	Msg string `json:"msg"`
	Ret int64  `json:"ret"`
}

//UploadURLResp 更新URL返回处理
func UploadURLResp(byteJSON []byte, strAddr string) ([]byte, error) {
	var structReply SetUploadURLReplyJSON
	err := json.Unmarshal(byteJSON, &structReply)
	if err != nil {
		beego.Error(err)
		return []byte(""), err
	}

	if structReply.Ret != 200 {
		strError := "设置数据服务URL失败," + structReply.Msg
		return []byte(""), errors.New(strError)
	}

	return []byte(""), nil
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

//AddEmployee 添加人员
func AddEmployee(byteJSON []byte, strAddr string) ([]byte, error) {
	var structRequest AddEmployeeWSSJSON
	err := json.Unmarshal(byteJSON, &structRequest)
	if err != nil {
		return []byte(""), err
	}

	byteRequst, err := GetAddEmployeeCmd(structRequest.Data.EmployeInfo)
	if err != nil {
		return []byte(""), err
	}

	structDevcieInfo, err := mysqlitemanger.GetDeviceBySerial(structRequest.Data.DeviceInfo.DeviceID)
	if err != nil {
		return []byte(""), err
	}

	//	beego.Debug("添加人员, deviceID:", structRequest.Data.DeviceInfo.DeviceID)
	//beego.Debug("添加人员指令:", string(byteRequst))

	MapIPSendChan[structDevcieInfo.Addr] <- byteRequst

	structResult := <-addResult

	var structReply CMDClientReplyJSON

	structReply.Cmd = "add_employee_resp"
	structReply.Code = structResult.Ret
	structReply.Message = structResult.Msg

	byteReply, err := json.Marshal(structReply)
	if err != nil {
		return []byte(""), err
	}

	beego.Debug("structReply:", structResult, "Name:", structRequest.Data.EmployeInfo.Name, "ID:", structRequest.Data.EmployeInfo.UserID)

	return byteReply, nil
}

//CommonWssReply  WSS服务统一的返回数据
type CommonWssReply struct {
	Cmd string `json:"cmd"`
	Msg string `json:"msg"`
	Ret int64  `json:"ret"`
}

var addResult = make(chan CommonWssReply, 20)

var AddResultMutex sync.Mutex

//AddUserResp 根据返回的数据更新用户同步状态
func AddUserResp(byteJSON []byte, strAddr string) ([]byte, error) {
	var structRequst CommonWssReply
	err := json.Unmarshal(byteJSON, &structRequst)
	if err != nil {
		return []byte(""), err
	}

	//AddResultMutex.Lock()
	addResult <- structRequst
	//AddResultMutex.Unlock()

	return []byte(""), nil
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

//RemoveEmoployeeRequstJSON  删除指令
type RemoveEmoployeeRequstJSON struct {
	Cmd  string                 `json:"cmd"`
	Data EmployeeDeleteInfoJSON `json:"data"`
}

//RemoveEmployee 删除人员
func RemoveEmployee(byteJSON []byte, strAddr string) ([]byte, error) {
	var structRequest RemoveEmployeeJSON
	err := json.Unmarshal(byteJSON, &structRequest)
	if err != nil {
		return []byte(""), err
	}

	byteRequst, err := GetRemoveEmoloyeeCmd(structRequest.Data.EmployeeInfo)
	if err != nil {
		return []byte(""), err
	}

	structDevcieInfo, err := mysqlitemanger.GetDeviceBySerial(structRequest.Data.DeviceInfo.DeviceID)
	if err != nil {
		return []byte(""), err
	}

	//beego.Debug("发送删除指令:", string(byteRequst))

	MapIPSendChan[structDevcieInfo.Addr] <- byteRequst

	structResult := <-removeReuslt

	var structReply CMDClientReplyJSON

	structReply.Cmd = "remove_employee_resp"
	structReply.Code = structResult.Ret
	structReply.Message = structResult.Msg

	byteReply, err := json.Marshal(structReply)
	if err != nil {
		return []byte(""), err
	}

	return byteReply, nil
}

var removeReuslt = make(chan CommonWssReply, 1)
var removeMutex sync.Mutex

//EmployeeRemoveResp 删除人员返回数据
func EmployeeRemoveResp(byteJSON []byte, strAddr string) ([]byte, error) {
	var structRequst CommonWssReply
	err := json.Unmarshal(byteJSON, &structRequst)
	if err != nil {
		return []byte(""), err
	}

	removeMutex.Lock()
	removeReuslt <- structRequst
	removeMutex.Unlock()

	return []byte(""), nil
}

type DoorStatutsInfoJSON struct {
	DoorStatus    int64 `json:"door_status"`
	DoorButton    bool  `json:"door_button"`
	DoorMangnetic bool  `json:"door_magnetic"`
	AlarmInput    bool  `json:"alarm_input"`
	PreventPry    bool  `json:"prevent_pry"`
}

type DoorStatusReplyJSON struct {
	Cmd  string              `json:"cmd"`
	Ret  int64               `json:"ret"`
	Msg  string              `json:"msg"`
	Data DoorStatutsInfoJSON `json:"data"`
}

var chanDoorStatus = make(chan DoorStatusReplyJSON, 1)

func GetDoorStautResp(byteJSON []byte, strAddr string) ([]byte, error) {

	var structRequst DoorStatusReplyJSON
	err := json.Unmarshal(byteJSON, &structRequst)
	if err != nil {
		return []byte(""), err
	}

	chanDoorStatus <- structRequst

	return []byte(""), nil

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

//SendOpenDoorCmdDelay _
func SendOpenDoorCmdDelay(structDevcieInfo mysqlitemanger.DeviceInfo, data OpenDoorInfoJSON) {
	timespan, err := time.ParseDuration(fmt.Sprintf("%ds", data.Duration))
	if err != nil {
		beego.Error(err)
		return
	}
	time.Sleep(timespan)

	var structRequst RemoteJSON
	structRequst.Cmd = "remote_ctrl"
	structRequst.Data.CtrlType = data.CtrlType
	structRequst.Data.Duration = data.Duration

	byteRequst, err := json.Marshal(structRequst)
	if err != nil {
		beego.Error(err)
		return
	}

	MapIPSendChan[structDevcieInfo.Addr] <- byteRequst

	//beego.Debug("开门指令:", string(byteRequst))
}

//OpenDoor __
func OpenDoor(byteJSON []byte, strAddr string) ([]byte, error) {
	var structRequest OpenDoorRequstJSON
	err := json.Unmarshal(byteJSON, &structRequest)
	if err != nil {
		return []byte(""), err
	}

	structDevcieInfo, err := mysqlitemanger.GetDeviceBySerial(structRequest.Data.DeviceID)
	if err != nil {
		return []byte(""), err
	}

	if structRequest.Data.Duration > 0 {
		structDelayData := structRequest.Data
		structDelayData.CtrlType = 0
		go SendOpenDoorCmdDelay(structDevcieInfo, structDelayData)
	}

	byteRequst, err := GetRemoteCmd(structRequest.Data)
	if err != nil {
		return []byte(""), nil
	}

	//beego.Debug("开门指令:", string(byteRequst), "地址", structDevcieInfo.Addr)

	if len(structDevcieInfo.Addr) > 0 && len(byteRequst) > 0 {
		MapIPSendChan[structDevcieInfo.Addr] <- byteRequst
	} else {
		return []byte(""), errors.New("指令或地址为空")
	}

	structResult := <-chanRemoteCtrl

	var structReply CMDClientReplyJSON

	structReply.Cmd = "open_door_resp"
	structReply.Code = structResult.Ret
	structReply.Message = structResult.Msg

	byteReply, err := json.Marshal(structReply)
	if err != nil {
		return []byte(""), err
	}

	return byteReply, nil
}

type GetDoorStatusCMD struct {
	Cmd string `json:"cmd"`
}

type DoorStatusReplyWSSJSON struct {
	Cmd     string              `json:"cmd"`
	Code    int64               `json:"code"`
	Message string              `json:"message"`
	Data    DoorStatutsInfoJSON `json:"data"`
}

func GetDoorStatus(byteJSON []byte, strAddr string) ([]byte, error) {
	var structRequest OpenDoorRequstJSON
	err := json.Unmarshal(byteJSON, &structRequest)
	if err != nil {
		return []byte(""), err
	}

	structDevcieInfo, err := mysqlitemanger.GetDeviceBySerial(structRequest.Data.DeviceID)
	if err != nil {
		return []byte(""), err
	}

	var structRequst GetDoorStatusCMD
	structRequst.Cmd = "get_door_status"

	byteRequst, err := json.Marshal(structRequst)
	if err != nil {
		return []byte(""), err
	}

	MapIPSendChan[structDevcieInfo.Addr] <- byteRequst

	structResult := <-chanDoorStatus

	var structReply DoorStatusReplyWSSJSON

	structReply.Cmd = "get_door_status_resp"
	structReply.Code = structResult.Ret
	structReply.Message = structResult.Msg
	structReply.Data = structResult.Data

	byteReply, err := json.Marshal(structReply)
	if err != nil {
		return []byte(""), err
	}

	return byteReply, nil

}

var chanIOCfg = make(chan []byte, 1)

//GetIOCfg __
func GetIOCfg(byteJSON []byte, strAddr string) ([]byte, error) {
	mapRecv := make(map[string]interface{}, 0)
	err := json.Unmarshal(byteJSON, &mapRecv)
	if err != nil {
		return []byte(""), err
	}

	strDeviceID := mapRecv["data"].(map[string]interface{})["deivceID"].(string)
	structDevcieInfo, err := mysqlitemanger.GetDeviceBySerial(strDeviceID)
	if err != nil {
		return []byte(""), err
	}

	mapRequst := make(map[string]interface{})
	mapRequst["cmd"] = "get_io_cfg"

	byteRequest, err := json.Marshal(mapRequst)
	if err != nil {
		return []byte(""), err
	}
	MapIPSendChan[structDevcieInfo.Addr] <- byteRequest

	byteReply := <-chanIOCfg

	return byteReply, nil
}

//GetIOCfgResp __
func GetIOCfgResp(byteJSON []byte, strAddr string) ([]byte, error) {
	chanIOCfg <- byteJSON

	return []byte(""), nil
}

var chanIOSet = make(chan []byte, 1)

//SetIOCfg __
func SetIOCfg(byteJSON []byte, strAddr string) ([]byte, error) {
	mapRecv := make(map[string]interface{}, 0)
	err := json.Unmarshal(byteJSON, &mapRecv)
	if err != nil {
		return []byte(""), err
	}

	strDeviceID := mapRecv["deviceID"].(string)
	structDeviceInfo, err := mysqlitemanger.GetDeviceBySerial(strDeviceID)
	if err != nil {
		return []byte(""), err
	}

	mapRequst := make(map[string]interface{}, 0)
	mapRequst["cmd"] = "set_io_cfg"
	mapRequst["data"] = mapRecv["data"]

	byteRequest, err := json.Marshal(mapRequst)
	if err != nil {
		return []byte(""), err
	}

	MapIPSendChan[structDeviceInfo.Addr] <- byteRequest

	byteReply := <-chanIOSet

	return byteReply, nil
}

//SetIOCfgResp __
func SetIOCfgResp(byteJSON []byte, strAddr string) ([]byte, error) {
	chanIOSet <- byteJSON

	return []byte(""), nil
}

// { "cmd" : "remote_ctrl_resp", "msg" : "remote ctrl OK", "ret" : 200 }

var chanRemoteCtrl = make(chan CommonWssReply, 1)

//RemoteCtrlResp __
func RemoteCtrlResp(byteJSON []byte, strAddr string) ([]byte, error) {
	var structReply CommonWssReply
	err := json.Unmarshal(byteJSON, &structReply)
	if err != nil {
		beego.Error(err)
		return []byte(""), err
	}

	chanRemoteCtrl <- structReply

	//beego.Debug("接收数据:", structReply)

	return []byte(""), nil
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

var chanChangePwd = make(chan CommonWssReply, 1)

//ChangePWDResp __
func ChangePWDResp(byteJSON []byte, strAddr string) ([]byte, error) {
	var structReply CommonWssReply
	err := json.Unmarshal(byteJSON, &structReply)
	if err != nil {
		beego.Error(err)
		return []byte(""), err
	}

	chanChangePwd <- structReply

	//beego.Debug("接收数据:", structReply)

	return []byte(""), nil
}

//ChangePWD _-
func ChangePWD(byteJSON []byte, strAddr string) ([]byte, error) {
	var structRcv ChangePWDJSON
	err := json.Unmarshal(byteJSON, &structRcv)
	if err != nil {
		return []byte(""), err
	}

	structRequest, err := GetChangePwdCMD(structRcv.Data)
	if err != nil {
		return []byte(""), err
	}

	structDevcieInfo, err := mysqlitemanger.GetDeviceBySerial(structRcv.Data.DeviceID)
	if err != nil {
		return []byte(""), err
	}

	byteRequst, err := json.Marshal(structRequest)
	if err != nil {
		return []byte(""), err
	}

	MapIPSendChan[structDevcieInfo.Addr] <- byteRequst

	structResult := <-chanChangePwd

	var structReply CMDClientReplyJSON

	structReply.Cmd = "change_password_resp"
	structReply.Code = structResult.Ret
	structReply.Message = structResult.Msg

	byteReply, err := json.Marshal(structReply)
	if err != nil {
		return []byte(""), err
	}

	MapAddrAndConn[structDevcieInfo.Addr] = false
	delete(MapAddrAndConn, structDevcieInfo.Addr)

	return byteReply, nil
}

var chanRestart = make(chan []byte, 1)

//RestartDeviceResp __
func RestartDeviceResp(byteJSON []byte, strAddr string) ([]byte, error) {
	chanRestart <- byteJSON

	return []byte(""), nil
}

//RestartDevice _
func RestartDevice(byteJSON []byte, strAddr string) ([]byte, error) {
	mapRecv := make(map[string]interface{}, 0)
	err := json.Unmarshal(byteJSON, &mapRecv)
	if err != nil {
		return []byte(""), err
	}

	strDeviceID := mapRecv["deviceID"].(string)

	structDeviceInfo, err := mysqlitemanger.GetDeviceBySerial(strDeviceID)
	if err != nil {
		return []byte(""), err
	}

	mapRequst := make(map[string]interface{}, 0)
	mapRequst["cmd"] = "restart"

	byteRequset, err := json.Marshal(mapRequst)
	if err != nil {
		return []byte(""), err
	}

	MapIPSendChan[structDeviceInfo.Addr] <- byteRequset

	byteReply := <-chanRestart

	if len(byteReply) <= 0 {
		return []byte(""), errors.New("接收数据为空")
	}

	return byteReply, nil
}

//CommonRequstJSON 公共json数据包
type CommonRequstJSON struct {
	Cmd  string      `json:"cmd"`
	Data interface{} `json:"data"`
}

//MessageParsing 数据解析分发
func MessageParsing(byteJSON []byte, strAddr string) {
	var structRequst CommonRequstJSON
	byteReply := make([]byte, 0)

	err := json.Unmarshal(byteJSON, &structRequst)
	if err != nil {
		beego.Error(err)
		return
	}

	// if structRequst.Cmd != "heartbeat_resp" && structRequst.Cmd != "add_employee" {
	// 	beego.Debug("接收数据:", string(byteJSON), "from：", strAddr)
	// }

	switch structRequst.Cmd {
	case "login":
		byteReply, err = Login(byteJSON, strAddr)
		break
	case "heartbeat_resp":
		err = HeartResp(byteJSON, strAddr)
		break
	case "http_client":
		byteReply, err = DealHTTPClient(byteJSON, strAddr)
		break
	case "db_path":
		byteReply, err = SetDBPath(byteJSON, strAddr)
		break
	case "remove_device":
		byteReply, err = RemoveDevice(byteJSON, strAddr)
		break
	case "set_upload_url_resp":
		break
	case "add_group_resp":
		break
	case "add_employee":
		go AddEmployee(byteJSON, strAddr)
		break
	case "add_user_resp":
		byteReply, err = AddUserResp(byteJSON, strAddr)
		break
	case "remove_employee":
		byteReply, err = RemoveEmployee(byteJSON, strAddr)
		break
	case "del_user_resp":
		byteReply, err = EmployeeRemoveResp(byteJSON, strAddr)
		break
	case "set_time_resp":
		break
	case "open_door":
		byteReply, err = OpenDoor(byteJSON, strAddr)
		break
	case "get_door_status":
		byteReply, err = GetDoorStatus(byteJSON, strAddr)
		break
	case "get_door_status_resp":
		byteReply, err = GetDoorStautResp(byteJSON, strAddr)
		break
	case "remote_ctrl_resp":
		byteReply, err = RemoteCtrlResp(byteJSON, strAddr)
		break
	case "change_password_resp":
		byteReply, err = ChangePWDResp(byteJSON, strAddr)
	case "change_password":
		byteReply, err = ChangePWD(byteJSON, strAddr)
	case "get_io_cfg_resp":
		byteReply, err = GetIOCfgResp(byteJSON, strAddr)
	case "get_io_cfg":
		byteReply, err = GetIOCfg(byteJSON, strAddr)
	case "set_io_cfg":
		byteReply, err = SetIOCfg(byteJSON, strAddr)
	case "set_io_cfg_resp":
		byteReply, err = SetIOCfgResp(byteJSON, strAddr)
	case "restart":
		byteReply, err = RestartDevice(byteJSON, strAddr)
	case "restart_resp":
		byteReply, err = RestartDeviceResp(byteJSON, strAddr)
	default:
		err = errors.New("接收到错误的数据")
		break
	}

	if err != nil {
		beego.Error(err)
	}

	if len(byteReply) > 0 {
		MapIPSendChan[strAddr] <- byteReply
	}
}
