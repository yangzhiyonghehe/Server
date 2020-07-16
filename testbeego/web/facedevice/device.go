package facedevice

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"../my_db"
)

var haiqinClient = http.Client{Transport: &http.Transport{
	Dial: func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, time.Second*2) //设置建立连接超时
		if err != nil {
			return nil, err
		}
		conn.SetDeadline(time.Now().Add(time.Second * 2)) //设置发送接受数据超时
		return conn, nil
	},
	ResponseHeaderTimeout: time.Second * 2,
}} //创建客户端

type FaceDeviceInfoReadBack struct {
	Name     string `json:"Name"`
	DeviceId int64  `json:"DeviceID"`
	Version  string `json:"Version"`
}

type ReadBackPackge struct {
	Operator string                 `json:"operator"`
	Info     FaceDeviceInfoReadBack `json:"info"`
}

func GetDeviceInfo(info my_db.DeviceInfo) (FaceDeviceInfoReadBack, error) {
	var structResult FaceDeviceInfoReadBack
	//尝试获取设备信息，如果无法获取，则说明设备不在线
	strUrl := fmt.Sprintf("http://%s/action/GetSysParam", info.Ip)
	buffer := bytes.NewBuffer([]byte(""))
	request, err := http.NewRequest("POST", strUrl, buffer)
	if err != nil {
		return structResult, err
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8") //添加请求头
	strBase64 := base64.StdEncoding.EncodeToString([]byte(info.Account + ":" + info.Psw))
	strAuthority := fmt.Sprintf("Basic %s", strBase64)
	request.Header.Set("Authorization", strAuthority)
	client := haiqinClient
	resp, err := client.Do(request.WithContext(context.TODO())) //发送请求
	if err != nil {
		return structResult, err
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if len(respBytes) == 0 {
		return structResult, err
	}

	//fmt.Println("设备信息:", string(respBytes))

	var FaceDeviceReadBackJson ReadBackPackge
	err = json.Unmarshal(respBytes, &FaceDeviceReadBackJson)
	if err != nil {
		return structResult, err
	}

	structResult = FaceDeviceReadBackJson.Info
	return structResult, nil
}

type SyncEmployee_Json struct {
	DeviceId          int64  `json:"DeviceID"`
	PersonType        int    `json:"PersonType"`
	Name              string `json:"Name"`
	TempValid         int    `json:"Tempvalid"`
	IdType            int    `json:"IdType"`
	CustomizeID       int64  `json:"CustomizeID"`
	IsCheckSimilarity int    `json:"isCheckSimilarity"`
}

type SendSyncPackge_Josn struct {
	Operator string            `json:"operator"`
	Info     SyncEmployee_Json `json:"info"`
	PicInfo  string            `json:"picinfo"`
}

type PostEmployeeToDeviceReply struct {
	Result string `json:"Result"`
	Detail string `json:"Detail"`
}

type PostEmployeeTodeviceReplayPackge struct {
	Operator string                    `json:"operator"`
	Code     int                       `json:"code"`
	Info     PostEmployeeToDeviceReply `json:"info"`
}

func AddDeviceEmployee(deviceInfo my_db.DeviceInfo, employeeInfo my_db.EmployeeInfo) error {
	var postSyncJson SendSyncPackge_Josn
	postSyncJson.Operator = "AddPerson"
	var infoJson SyncEmployee_Json
	infoJson.Name = employeeInfo.Name
	nIndex, err := strconv.ParseInt(deviceInfo.DevcieId, 0, 64)
	infoJson.DeviceId = nIndex
	if err != nil {
		return err
	}
	infoJson.CustomizeID = employeeInfo.Id
	infoJson.IdType = 0
	infoJson.IsCheckSimilarity = 0
	infoJson.PersonType = 0
	infoJson.TempValid = 0

	postSyncJson.Info = infoJson
	postSyncJson.PicInfo = employeeInfo.Pic

	bytePost, err := json.Marshal(postSyncJson)
	if err != nil {
		return err
	}

	strUrl := fmt.Sprintf("http://%s/action/AddPerson", deviceInfo.Ip)
	buffer := bytes.NewBuffer(bytePost)
	request, err := http.NewRequest("POST", strUrl, buffer)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8") //添加请求头
	strBase64 := base64.StdEncoding.EncodeToString([]byte(deviceInfo.Account + ":" + deviceInfo.Psw))
	strAuthority := fmt.Sprintf("Basic %s", strBase64)
	request.Header.Set("Authorization", strAuthority)
	client := haiqinClient
	resp, err := client.Do(request.WithContext(context.TODO())) //发送请求
	if err != nil {
		return err
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	var respJson PostEmployeeTodeviceReplayPackge
	err = json.Unmarshal(respBytes, &respJson)
	if err != nil {
		return err
	}

	if respJson.Code != 200 {
		strEror := fmt.Sprintf("添加用户到设备失败: %s", respJson.Info.Detail)

		return errors.New(strEror)
	}

	return nil
}

type DeleteEmployeeFromDeviceRequestInfo_Json struct {
	DeviceId    int64   `json:"DeviceID"`
	TotalNum    int     `json:"TotalNum"`
	IdType      int64   `json:"IdType"`
	CustomizeID []int64 `json:"CustomizeID"`
}

type DeleteEmployeFromDevicePackge_Json struct {
	Operator string                                   `json:"operator"`
	Info     DeleteEmployeeFromDeviceRequestInfo_Json `json:"info"`
}

func RemoveDeviceEmployee(deviceInfo my_db.DeviceInfo, employeeInfo my_db.EmployeeInfo) error {

	if deviceInfo.Type == 0 {
		var infoIndex DeleteEmployeeFromDeviceRequestInfo_Json
		nDeviceId, err := strconv.ParseInt(deviceInfo.DevcieId, 0, 64)
		if err != nil {
			return err
		}
		infoIndex.DeviceId = nDeviceId
		infoIndex.IdType = deviceInfo.Type
		infoIndex.TotalNum = 1
		infoIndex.CustomizeID = append(infoIndex.CustomizeID, employeeInfo.Id)

		var struct_PostPackge DeleteEmployeFromDevicePackge_Json
		struct_PostPackge.Operator = "DeletePerson"
		struct_PostPackge.Info = infoIndex

		bytePost, err := json.Marshal(struct_PostPackge)

		strUrl := fmt.Sprintf("http://%s/action/DeletePerson", deviceInfo.Ip)

		//fmt.Printf("删除人员请求包: %s", string(bytePost))

		buffer := bytes.NewBuffer(bytePost)
		request, err := http.NewRequest("POST", strUrl, buffer)
		if err != nil {
			return err
		}
		request.Header.Set("Content-Type", "application/json;charset=UTF-8") //添加请求头
		strBase64 := base64.StdEncoding.EncodeToString([]byte(deviceInfo.Account + ":" + deviceInfo.Psw))
		strAuthority := fmt.Sprintf("Basic %s", strBase64)
		request.Header.Set("Authorization", strAuthority)
		client := haiqinClient
		resp, err := client.Do(request.WithContext(context.TODO())) //发送请求
		if err != nil {
			return err
		}

		respBytes, err := ioutil.ReadAll(resp.Body)
		var unMarshJson PostEmployeeTodeviceReplayPackge
		err = json.Unmarshal(respBytes, &unMarshJson)
		if err != nil {
			return err
		}

		if unMarshJson.Code != 200 {
			strMsg := fmt.Sprintf("删除设备人员失败: %s", unMarshJson.Info.Detail)
			return errors.New(strMsg)
		}
	}

	return nil
}

type SubscribeUrlInfo_Json struct {
	Snap      string `json:"Snap"`
	Verify    string `json:"Verify"`
	HeartBeat string `json:"HeartBeat"`
}

type SubscribeInfo_Json struct {
	DeviceId             int64                 `json:"DeviceID"`
	Num                  int64                 `json:"Num"`
	Topics               []string              `json:"Topics"`
	SubscribeAddr        string                `json:"SubscribeAddr"`
	SubscribeUrl         SubscribeUrlInfo_Json `json:"SubscribeUrl"`
	BeatInterval         int                   `json:"BeatInterval"`
	ResumefromBreakpoint int                   `json:"ResumefromBreakpoint"` //断点续传
	Auth                 string                `json:"Auth"`
}

type SubscribeRequst_Json struct {
	Operator string             `json:"operator"`
	Info     SubscribeInfo_Json `json:"info"`
}

type subscribeReplyInfo_Json struct {
	Result string `json:"Result"`
	Detail string `json:"Detail"`
}

type DeviceSetRely_Json struct {
	Operator string                  `json:"operator"`
	Code     int64                   `json:"code"`
	Info     subscribeReplyInfo_Json `json:"info"`
}

func GetHostIp() (string, error) {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for i := 0; i < len(netInterfaces); i++ {
		if (netInterfaces[i].Flags & net.FlagUp) != 0 {
			addrs, _ := netInterfaces[i].Addrs()

			for _, address := range addrs {
				if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						return ipnet.IP.String(), nil
					}
				}
			}
		}
	}

	return "", errors.New("未找到Ip")
}

func SubscribeWithDevice(deviceInfo my_db.DeviceInfo, strHostIp string) error {
	var structSubscribeInfo SubscribeInfo_Json
	nDeviceInfo, err := strconv.ParseInt(deviceInfo.DevcieId, 0, 64)
	if err != nil {
		return err
	}
	structSubscribeInfo.DeviceId = nDeviceInfo
	structSubscribeInfo.Num = 2
	structSubscribeInfo.ResumefromBreakpoint = 0
	strHost := strHostIp

	structSubscribeInfo.SubscribeAddr = "http://" + strHost + ":80"
	structSubscribeInfo.Topics = []string{"Snap", "VerifyWithSnap"}
	structSubscribeInfo.Auth = "none"
	structSubscribeInfo.BeatInterval = 30

	var structSubscribeUrlInfo SubscribeUrlInfo_Json
	structSubscribeUrlInfo.HeartBeat = "/Subscribe/heartbeat"
	structSubscribeUrlInfo.Snap = "/Subscribe/Snap"
	structSubscribeUrlInfo.Verify = "/Subscribe/Verify"

	structSubscribeInfo.SubscribeUrl = structSubscribeUrlInfo

	var structRequst SubscribeRequst_Json
	structRequst.Operator = "Subscribe"
	structRequst.Info = structSubscribeInfo

	bytePost, err := json.Marshal(structRequst)
	if err != nil {
		return err
	}

	strUrl := fmt.Sprintf("http://%s/action/Subscribe", deviceInfo.Ip)

	//fmt.Printf("订阅请求包: %s", string(bytePost))

	buffer := bytes.NewBuffer(bytePost)
	request, err := http.NewRequest("POST", strUrl, buffer)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8") //添加请求头
	strBase64 := base64.StdEncoding.EncodeToString([]byte(deviceInfo.Account + ":" + deviceInfo.Psw))
	strAuthority := fmt.Sprintf("Basic %s", strBase64)
	request.Header.Set("Authorization", strAuthority)
	client := haiqinClient
	resp, err := client.Do(request.WithContext(context.TODO())) //发送请求
	if err != nil {
		return err
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	var unMarshJson DeviceSetRely_Json
	err = json.Unmarshal(respBytes, &unMarshJson)
	if err != nil {
		return err
	}

	if unMarshJson.Code != 200 {
		strMsg := fmt.Sprintf("抓拍认证信息订阅失败: %s", unMarshJson.Info.Detail)
		return errors.New(strMsg)
	}

	return nil
}

type UnSubscribeRequstInfo_Json struct {
	DeviceID int64    `json:"DeviceID"`
	Num      int64    `json:"Num"`
	Topics   []string `json:"Topics"`
}

type UnSubscribeRequst_Json struct {
	Operator string                     `json:"operator"`
	Info     UnSubscribeRequstInfo_Json `json:"info"`
}

func UnSubcribeDevice(deviceInfo my_db.DeviceInfo) error {
	var structRequst UnSubscribeRequst_Json
	structRequst.Operator = "Unsubscribe"
	structRequst.Info.Topics = []string{"Snap", "VerifyWithSnap"}
	structRequst.Info.Num = 2
	nDeviceId, err := strconv.ParseInt(deviceInfo.DevcieId, 0, 64)
	if err != nil {
		return err
	}
	structRequst.Info.DeviceID = nDeviceId

	bytePost, err := json.Marshal(structRequst)
	if err != nil {
		return err
	}

	strUrl := fmt.Sprintf("http://%s/action/Unsubscribe", deviceInfo.Ip)
	//fmt.Printf("订阅请求包: %s", string(bytePost))

	buffer := bytes.NewBuffer(bytePost)
	request, err := http.NewRequest("POST", strUrl, buffer)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8") //添加请求头
	strBase64 := base64.StdEncoding.EncodeToString([]byte(deviceInfo.Account + ":" + deviceInfo.Psw))
	strAuthority := fmt.Sprintf("Basic %s", strBase64)
	request.Header.Set("Authorization", strAuthority)
	client := haiqinClient
	resp, err := client.Do(request.WithContext(context.TODO())) //发送请求
	if err != nil {
		return err
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	var unMarshJson DeviceSetRely_Json
	err = json.Unmarshal(respBytes, &unMarshJson)
	if err != nil {
		return err
	}

	if unMarshJson.Code != 200 {
		strMsg := fmt.Sprintf("取消抓拍认证订阅失败: %s", unMarshJson.Info.Detail)
		return errors.New(strMsg)
	}
	return nil
}

type NetParamInfo_Json struct {
	IPAddr  string `json:"IPAddr"`
	Submask string `json:"Submask"`
	Gateway string `json:"Gateway"`
}

type NetParam_Json struct {
	Operator string            `json:"operator"`
	Info     NetParamInfo_Json `json:"info"`
}

func GetDeviceNetInfo(deviceInfo my_db.DeviceInfo) (NetParamInfo_Json, error) {
	var structReply NetParam_Json
	strUrl := fmt.Sprintf("http://%s/action/GetNetParam", deviceInfo.Ip)
	//fmt.Printf("订阅请求包: %s", string(bytePost))

	buffer := bytes.NewBuffer([]byte(""))
	request, err := http.NewRequest("POST", strUrl, buffer)
	if err != nil {
		return structReply.Info, err
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8") //添加请求头
	strBase64 := base64.StdEncoding.EncodeToString([]byte(deviceInfo.Account + ":" + deviceInfo.Psw))
	strAuthority := fmt.Sprintf("Basic %s", strBase64)
	request.Header.Set("Authorization", strAuthority)
	client := haiqinClient
	resp, err := client.Do(request.WithContext(context.TODO())) //发送请求
	if err != nil {
		return structReply.Info, err
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(respBytes, &structReply)
	if err != nil {
		return structReply.Info, err
	}

	return structReply.Info, nil
}

func SetDeviceNetInfo(data NetParamInfo_Json, deviceInfo my_db.DeviceInfo) error {
	var structRequstInfo NetParam_Json
	structRequstInfo.Info = data
	structRequstInfo.Operator = "SetNetParam"

	byteRequst, err := json.Marshal(structRequstInfo)
	if err != nil {
		return err
	}

	strUrl := fmt.Sprintf("http://%s/action/SetNetParam", deviceInfo.Ip)
	//fmt.Printf("订阅请求包: %s", string(bytePost))

	buffer := bytes.NewBuffer(byteRequst)
	request, err := http.NewRequest("POST", strUrl, buffer)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8") //添加请求头
	strBase64 := base64.StdEncoding.EncodeToString([]byte(deviceInfo.Account + ":" + deviceInfo.Psw))
	strAuthority := fmt.Sprintf("Basic %s", strBase64)
	request.Header.Set("Authorization", strAuthority)
	client := haiqinClient
	resp, err := client.Do(request.WithContext(context.TODO())) //发送请求
	if err != nil {
		return err
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	var unMarshJson DeviceSetRely_Json
	err = json.Unmarshal(respBytes, &unMarshJson)
	if err != nil {
		return err
	}

	if unMarshJson.Code != 200 {
		strMsg := fmt.Sprintf("取消抓拍认证订阅失败: %s", unMarshJson.Info.Detail)
		return errors.New(strMsg)
	}

	return nil
}

func GetDevcieSysParam(deviceInfo my_db.DeviceInfo) error {
	straction := "list"
	group := "CFGSYSBASICPARA"
	nRanId := rand.Int63n(99999999)

	strUrl := fmt.Sprintf("http://%s/action/getSysBasicPara", deviceInfo.Ip)
	//fmt.Printf("订阅请求包: %s", string(bytePost))

	//buffer := bytes.NewBuffer([]byte(""))
	var r http.Request
	r.ParseForm()
	r.Form.Add("action", straction)
	r.Form.Add("group", group)
	r.Form.Add("nRanId", fmt.Sprintf("%d", nRanId))
	bodystr := strings.TrimSpace(r.Form.Encode())
	request, err := http.NewRequest("GET", strUrl, strings.NewReader(bodystr))
	if err != nil {
		fmt.Println(err)
		return err
	}

	request.Header.Set("Content-Type", "application/json;charset=UTF-8") //添加请求头
	strBase64 := base64.StdEncoding.EncodeToString([]byte(deviceInfo.Account + ":" + deviceInfo.Psw))
	strAuthority := fmt.Sprintf("Basic %s", strBase64)
	request.Header.Set("Authorization", strAuthority)

	client := haiqinClient
	resp, err := client.Do(request.WithContext(context.TODO())) //发送请求
	if err != nil {
		return err
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Println("设备参数:", string(respBytes))

	return nil
}

type DoorConditionInfo_Json struct {
	FaceThreshold   int64 `json:"FaceThreshold"`
	IDCardThreshold int64 `json:"IDCardThreshold"`
	OpendoorWay     int64 `json:"OpendoorWay"`
	VerifyMode      int64 `json:"VerifyMode"`
	ControlType     int64 `json:"ControlType"`
	IOType          int64 `json:"IOType"`
	IOStayTime      int64 `json:"IOStayTime"`
	Endian          int64 `json:"Endian"`
	CardMode        int64 `json:"CardMode"`
	Wiegand         int64 `json:"Wiegand"`
	PublicMjCardNo  int64 `json:"PublicMjCardNo"`
	AutoMjCardBgnNo int64 `json:"AutoMjCardBgnNo"`
	AutoMjCardEndNo int64 `json:"AutoMjCardEndNo"`
}

type DoorConditionReplyInfo_Json struct {
	Operator string                 `json:"operator"`
	Info     DoorConditionInfo_Json `json:"info"`
}

func GetDoorConditionParam(deviceInfo my_db.DeviceInfo) (DoorConditionInfo_Json, error) {
	var structReply DoorConditionReplyInfo_Json

	strUrl := fmt.Sprintf("http://%s/action/GetDoorCondition", deviceInfo.Ip)
	//fmt.Printf("订阅请求包: %s", string(bytePost))

	buffer := bytes.NewBuffer([]byte(""))
	request, err := http.NewRequest("POST", strUrl, buffer)
	if err != nil {
		return structReply.Info, err
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8") //添加请求头
	strBase64 := base64.StdEncoding.EncodeToString([]byte(deviceInfo.Account + ":" + deviceInfo.Psw))
	strAuthority := fmt.Sprintf("Basic %s", strBase64)
	request.Header.Set("Authorization", strAuthority)
	client := haiqinClient
	resp, err := client.Do(request.WithContext(context.TODO())) //发送请求
	if err != nil {
		return structReply.Info, err
	}

	respBytes, err := ioutil.ReadAll(resp.Body)

	err = json.Unmarshal(respBytes, &structReply)
	if err != nil {
		return structReply.Info, err
	}

	return structReply.Info, nil
}

type SetDoorConditionRequstInfo_Json struct {
	Operator string                 `json:"operator"`
	Info     DoorConditionInfo_Json `json:"info"`
}

type SetDoorConditionInfo_Json struct {
	Result string `json:"result"`
}

type SetDoorConditionReplyInfo_Json struct {
	Code     int64                     `json:"code"`
	Operator string                    `json:"operator"`
	Info     SetDoorConditionInfo_Json `json:"info"`
}

func SetDoorConditionParam(deviceInfo my_db.DeviceInfo, data DoorConditionInfo_Json) error {
	var structRequstInfo SetDoorConditionRequstInfo_Json
	structRequstInfo.Operator = "SetDoorCondition"
	structRequstInfo.Info = data

	byteRequst, err := json.Marshal(structRequstInfo)
	if err != nil {
		return err
	}

	strUrl := fmt.Sprintf("http://%s/action/SetDoorCondition", deviceInfo.Ip)
	//fmt.Printf("订阅请求包: %s", string(bytePost))

	buffer := bytes.NewBuffer(byteRequst)
	request, err := http.NewRequest("POST", strUrl, buffer)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8") //添加请求头
	strBase64 := base64.StdEncoding.EncodeToString([]byte(deviceInfo.Account + ":" + deviceInfo.Psw))
	strAuthority := fmt.Sprintf("Basic %s", strBase64)
	request.Header.Set("Authorization", strAuthority)
	client := haiqinClient
	resp, err := client.Do(request.WithContext(context.TODO())) //发送请求
	if err != nil {
		return err
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var structReplyInfo SetDoorConditionReplyInfo_Json
	err = json.Unmarshal(respBytes, &structReplyInfo)
	if err != nil {
		return err
	}

	//fmt.Println("设置开门条件返回内容:", string(respBytes))

	return nil
}

//OpenDoorInfoJSON __
type OpenDoorInfoJSON struct {
	DeviceID string `json:"DeviceID"`
	Status   int64  `json:"status"`
	Msg      string `json:"msg"`
}

//OpenDoorJSON __
type OpenDoorJSON struct {
	Operator string           `json:"operator"`
	Info     OpenDoorInfoJSON `json:"info"`
}

//CommonReplyInfoJSON __
type CommonReplyInfoJSON struct {
	Result string `json:"Result"`
}

//CommonReplyJSON __
type CommonReplyJSON struct {
	Operator string              `json:"operator"`
	Code     int64               `json:"code"`
	Info     CommonReplyInfoJSON `json:"info"`
}

//OpenDoor __
func OpenDoor(deviceInfo my_db.DeviceInfo, data OpenDoorInfoJSON) error {
	var structRequstInfo OpenDoorJSON
	structRequstInfo.Operator = "OpenDoor"
	structRequstInfo.Info = data

	byteRequst, err := json.Marshal(structRequstInfo)
	if err != nil {
		return err
	}

	strUrl := fmt.Sprintf("http://%s/action/OpenDoor", deviceInfo.Ip)
	//fmt.Printf("订阅请求包: %s", string(bytePost))

	buffer := bytes.NewBuffer(byteRequst)
	request, err := http.NewRequest("POST", strUrl, buffer)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8") //添加请求头
	strBase64 := base64.StdEncoding.EncodeToString([]byte(deviceInfo.Account + ":" + deviceInfo.Psw))
	strAuthority := fmt.Sprintf("Basic %s", strBase64)
	request.Header.Set("Authorization", strAuthority)
	client := haiqinClient
	resp, err := client.Do(request.WithContext(context.TODO())) //发送请求
	if err != nil {
		return err
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var structReplyInfo CommonReplyJSON
	err = json.Unmarshal(respBytes, &structReplyInfo)
	if err != nil {
		return err
	}

	//fmt.Println("设置开门条件返回内容:", string(respBytes))

	return nil

}
