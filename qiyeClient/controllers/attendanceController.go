package controllers

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"time"

	"../check_device"
	"../facedevice"
	"../facedevicexiongmai"
	my_db "../mysqlitemanger"
	"../qiyewss"
	"github.com/astaxie/beego"
	"github.com/axgle/mahonia"
	"github.com/buger/jsonparser"
)

const TIME_LAYOUT = "2006-01-02 15:04:05"

/*--------------------------------------------------------*/

//验证路由
type VerifyController struct {
	beego.Controller
}

func (c *VerifyController) Post() {
	data := c.Ctx.Input.RequestBody

	//fmt.Printf("%s", data)

	//beego.Debug(string(data))

	// DeviceID, err := jsonparser.GetInt([]byte(data), "info", "DeviceID")
	// if err != nil {
	// 	beego.Error("添加验证内容错误", err)
	// 	return
	// }

	CreateTime, err := jsonparser.GetString([]byte(data), "info", "CreateTime")
	if err != nil {
		beego.Error("添加验证内容错误", err)
		return
	}

	// SanpPic, err := jsonparser.GetString([]byte(data), "SanpPic")
	// if err != nil {
	// 	beego.Error("添加验证内容错误", err)
	// 	return
	// }

	// Name, err := jsonparser.GetString([]byte(data), "info", "Name")
	// if err != nil {
	// 	beego.Error("添加验证内容错误", err)
	// 	return
	// }

	// VerifyStatus, err := jsonparser.GetInt([]byte(data), "info", "VerifyStatus")
	// if err != nil {
	// 	beego.Error("添加验证内容错误", err)
	// 	return
	// }

	//CustomizeID
	nCustomID, err := jsonparser.GetInt([]byte(data), "info", "CustomizeID")
	if err != nil {
		beego.Error("添加验证内容错误", err)
		return
	}

	// templateData, err := jsonparser.GetFloat([]byte(data), "info", "Temperature")
	// if err != nil {
	// 	beego.Warning("未获取到体温数据", err)
	// }

	timeSplitArray := strings.Split(CreateTime, "T")
	if len(timeSplitArray) != 2 {
		return
	}

	timeSplitArray_ := strings.Split(timeSplitArray[1], "Z")

	CreateTime = timeSplitArray[0] + " " + timeSplitArray_[0]
	timeAttendance, err := time.Parse(TIME_LAYOUT, CreateTime)
	if err != nil {
		return
	}

	timeSpan, err := time.ParseDuration("-8h")
	if err != nil {
		beego.Error(err)
		return
	}
	timeAttendance = timeAttendance.Add(timeSpan)
	beego.Debug("打卡时间:", timeAttendance.Format(TIME_LAYOUT))

	strUserID, err := my_db.GetUserIDByHQ(nCustomID)
	if err != nil {
		beego.Error(err)
		return
	}

	_, err = qiyewss.Checkin(strUserID, 0, timeAttendance.Unix())
	if err != nil {
		beego.Error(err)
		return
	}

	Reply := `{
			"operator": "VerifyPush",
			"code": 200,
			"info": {
				"Result": "Ok"
			}
		}`

	c.Data["json"] = json.RawMessage(string(Reply))
	c.ServeJSON()
}

/*------------------------------------------*/

//抓拍路由处理
type SnapController struct {
	beego.Controller
}

func (c *SnapController) Post() {
	data := c.Ctx.Input.RequestBody
	//beego.Debug(string(data))

	// DeviceID, err := jsonparser.GetInt([]byte(data), "info", "DeviceID")
	// if err != nil {
	// 	beego.Error("添加抓拍内容错误", err)
	// 	return
	// }

	// CreateTime, err := jsonparser.GetString([]byte(data), "info", "CreateTime")
	// if err != nil {
	// 	beego.Error("添加抓拍内容错误", err)
	// 	return
	// }

	// SanpPic, err := jsonparser.GetString([]byte(data), "SanpPic")
	// if err != nil {
	// 	beego.Error("添加抓拍内容错误", err)
	// 	return
	// }

	// templateData, err := jsonparser.GetFloat([]byte(data), "info", "Temperature")
	// if err != nil {
	// 	beego.Warning("未获取到体温数据", err)
	// }

	Reply := `{
			"operator": "SnapPush",
			"code": 200,
			"info": {
				"Result": "Ok"
			}
		}`

	data, err := json.Marshal(Reply)
	if err != nil {
		beego.Error("添加抓拍内容错误", err)
		return
	}

	responseStr := string(data)
	c.Data["json"] = responseStr
	c.ServeJSON()
}

//心跳包路由
type HeartController struct {
	beego.Controller
}

type infoStruct struct {
	DeviceID int    `json:"DeviceID"`
	Time     string `json:"Time"`
}

type ResponseHeart struct {
	Operator string     `json:"operator"`
	Info     infoStruct `json:"info"`
}

var timeLayoutStr = "2006-01-02 15:04:05"

func (c *HeartController) Post() {
	data := c.Ctx.Input.RequestBody
	enc := mahonia.NewDecoder("gb18030")
	goStr := enc.ConvertString(string(data))

	var unmarshData ResponseHeart
	json.Unmarshal([]byte(goStr), &unmarshData)

	currentTime := time.Now()
	currentTime.Unix()
	ts := currentTime.Format(timeLayoutStr)
	infoChild := infoStruct{unmarshData.Info.DeviceID, ts}
	responseHeartPack := ResponseHeart{Operator: "HeartBeat", Info: infoChild}
	responseByte, err := json.Marshal(responseHeartPack)
	if err != nil {
		return
	}

	responseStr := string(responseByte)
	c.Data["json"] = responseStr
	c.ServeJSON()
	//c.Ctx.WriteString
}

//XMUpdateContoller 雄迈设备数据
type XMUpdateContoller struct {
	beego.Controller
}

//XMUpdateRequst  雄迈设备更新数据
type XMUpdateRequst struct {
	RecordType int64  `json:"record_type"`
	DeviceID   string `json:"device_id"`
	UserID     string `json:"user_id"`
	CardID     string `json:"card_id"`
	GroupID    string `json:"group_id"`
	UserName   string `json:"user_name"`
	Time       string `json:"time"`
	PassResult int64  `json:"pass_result"`
	Message    string `json:"message"`
	Similarity int64  `json:"similarity"`
	JpgLen     int64  `json:"jpg_len"`
}

//Post post请求
func (c *XMUpdateContoller) Post() {
	var arrayData []XMUpdateRequst

	if c.Ctx.Input.Context.Request.MultipartForm == nil {
		beego.Error("c.Ctx.Input.Context.Request.MultipartForm == nil")
		return
	}

	if len(c.Ctx.Input.Context.Request.MultipartForm.Value) <= 0 {
		beego.Error("c.Ctx.Input.Context.Request.MultipartForm.Value) <= 0")
		return
	}

	if len(c.Ctx.Input.Context.Request.MultipartForm.File) <= 0 {
		beego.Error("c.Ctx.Input.Context.Request.MultipartForm.File) <= 0")
		return
	}

	data := c.Ctx.Input.Context.Request.MultipartForm.Value["data"]

	for _, dataIndex := range data {
		var structRqust XMUpdateRequst
		err := json.Unmarshal([]byte(dataIndex), &structRqust)
		if err != nil {
			beego.Error("更新雄迈设备数据错误", err)
			return
		}

		//beego.Debug("接收到数据:", structRqust)
		arrayData = append(arrayData, structRqust)
	}

	for _, value := range arrayData {
		if value.PassResult < 0 {

		} else {
			CreateTime := value.Time
			// timeSplitArray := strings.Split(CreateTime, "T")
			// if len(timeSplitArray) != 2 {
			// 	return
			// }

			// timeSplitArray_ := strings.Split(timeSplitArray[1], "Z")

			// CreateTime = timeSplitArray[0] + " " + timeSplitArray_[0]
			timeAttendance, err := time.Parse(TIME_LAYOUT, CreateTime)
			if err != nil {
				beego.Error(err)
				return
			}

			beego.Debug("打卡时间:", timeAttendance.Format(TIME_LAYOUT))
			timeSpan, err := time.ParseDuration("-8h")
			if err != nil {
				beego.Error(err)
				return
			}
			timeAttendance = timeAttendance.Add(timeSpan)

			_, err = qiyewss.Checkin(value.UserID, 0, timeAttendance.Unix())
			if err != nil {
				beego.Error(err)
				return
			}
		}

	}

	c.Data["json"] = ""
	c.ServeJSON()

	//c.GetFile("")
}

type GetServerConfController struct {
	beego.Controller
}

func (c *GetServerConfController) Post() {
	mapReply := make(map[string]interface{}, 0)
	mapReply["code"] = 20000
	mapReply["message"] = "请求成功"

	byteReply, err := json.Marshal(mapReply)
	if err != nil {
		beego.Error(err)
		return
	}

	c.Data["json"] = json.RawMessage(byteReply)
	c.ServeJSON()
}

type AddDeviceController struct {
	beego.Controller
}

func (c *AddDeviceController) Post() {
	data := c.Ctx.Input.RequestBody

	mapRequest := make(map[string]interface{}, 0)

	err := json.Unmarshal(data, &mapRequest)

	err = facedevicexiongmai.DeviceSearch()
	if err != nil {
		beego.Error("添加设备错误", err)
		return
	}

	spanindex, err := time.ParseDuration("0.1s")
	time.Sleep(spanindex)

	nDeviceType := -1
	strEntrilDeviceID := ""
	strXMPSW := ""

	mapDeviceInfo, err := facedevicexiongmai.GetDeviceInfoByIP(mapRequest["ip"].(string))

	if err == nil {
		nDeviceType = 2
		strEntrilDeviceID = mapDeviceInfo["serial"]
		strXMPSW = mapDeviceInfo["psw"]
	} else {
		for i := 0; i < 10; i++ {
			spanindex, err := time.ParseDuration("0.1s")
			time.Sleep(spanindex)
			mapDeviceInfo, err = facedevicexiongmai.GetDeviceInfoByIP(mapRequest["ip"].(string))
			if err == nil {
				nDeviceType = 2
				strEntrilDeviceID = mapDeviceInfo["serial"]
				strXMPSW = mapDeviceInfo["psw"]
				break
			}
		}
	}

	if nDeviceType == -1 {
		beego.Debug("判断雄迈设备超时")

		//判读是否是海清的设备
		var structDeviceInfoIndex my_db.DeviceInfo
		structDeviceInfoIndex.Account = "admin"
		structDeviceInfoIndex.Psw = "admin"
		structDeviceInfoIndex.Ip = mapRequest["ip"].(string)

		structHaiQinDeviceInfo, err := facedevice.GetDeviceInfo(structDeviceInfoIndex)
		if err == nil {
			nDeviceType = 1
			strEntrilDeviceID = fmt.Sprintf("%d", structHaiQinDeviceInfo.DeviceId)
		}

		structDeviceData, err := my_db.GetDeviceByDeviceId(strEntrilDeviceID)
		if structDeviceData.DevcieId == strEntrilDeviceID && len(strEntrilDeviceID) > 0 {
			mapReply := make(map[string]interface{}, 0)
			mapReply["code"] = 40000
			mapReply["message"] = "添加设备失败: 该设备已被绑定"

			byteReply, err := json.Marshal(mapReply)
			if err != nil {
				beego.Error(err)
				return
			}

			c.Data["json"] = json.RawMessage(byteReply)
			c.ServeJSON()
			return
		}
	}

	if nDeviceType == 1 {
		byteReply, err := AddDeviceHaiQIn(c)
		if len(byteReply) == 0 && err != nil {
			beego.Error(err)
			return
		}
		c.Data["json"] = json.RawMessage(byteReply)
		c.ServeJSON()
		return
	} else if nDeviceType == 2 {
		if strXMPSW != mapRequest["psw"].(string) {
			mapResult := make(map[string]interface{}, 0)
			mapResult["code"] = 50000
			mapResult["message"] = "添加设备失败: 设备ip或密码错误"
			byteReply, err := json.Marshal(mapResult)
			if err != nil {
				beego.Error(err)
				return
			}
			c.Data["json"] = json.RawMessage(byteReply)
			c.ServeJSON()
			return
		}

		byteReply, err := AddDeviceXiongmai(strXMPSW, strEntrilDeviceID, c)
		if len(byteReply) == 0 && err != nil {
			beego.Error(err)
			return
		}
		c.Data["json"] = json.RawMessage(byteReply)
		c.ServeJSON()
		return
	}

	mapReply := make(map[string]interface{}, 0)
	mapReply["code"] = 40000
	mapReply["message"] = "不存在该设备,请检查输入IP是否正确"

	byteReply, err := json.Marshal(mapReply)
	if err != nil {
		beego.Error("添加设备错误", err)
		return
	}

	c.Data["json"] = json.RawMessage(string(byteReply))
	c.ServeJSON()
}

func AddDeviceHaiQIn(c *AddDeviceController) ([]byte, error) {
	data := c.Ctx.Input.RequestBody

	mapRequest := make(map[string]interface{}, 0)
	mapReply := make(map[string]interface{}, 0)

	err := json.Unmarshal(data, &mapRequest)
	if err != nil {
		beego.Error(err)
		return []byte(""), err
	}

	var structDeviceInfoIndex my_db.DeviceInfo
	structDeviceInfoIndex.Account = "admin"
	structDeviceInfoIndex.Psw = mapRequest["psw"].(string)
	structDeviceInfoIndex.Ip = mapRequest["ip"].(string)

	faceDeviceInfo, err := facedevice.GetDeviceInfo(structDeviceInfoIndex)
	if err != nil {

		mapReply["code"] = 40000
		mapReply["message"] = "添加设备失败: 设备ip或密码错误"

		byteReply, errindex := json.Marshal(mapReply)
		if errindex != nil {
			return []byte(""), errindex
		}
		return byteReply, err
	}

	err = check_device.DeviceCheck(fmt.Sprintf("%d", faceDeviceInfo.DeviceId))
	if err != nil {
		beego.Error("添加设备错误", err)

		mapReply["code"] = 40000
		mapReply["message"] = "添加设备失败: " + err.Error()

		byteReply, errindex := json.Marshal(mapReply)
		if errindex != nil {
			return []byte(""), errindex
		}

		return byteReply, err
	}

	var structDeviceInfo my_db.DeviceInfo
	structDeviceInfo.Ip = mapRequest["ip"].(string)
	structDeviceInfo.Name = mapRequest["name"].(string)
	structDeviceInfo.Psw = "admin"
	structDeviceInfo.Account = mapRequest["psw"].(string)
	structDeviceInfo.DevcieId = fmt.Sprintf("%d", faceDeviceInfo.DeviceId)
	structDeviceInfo.CompanyID = 0
	structDeviceInfo.Type = 0

	Host := c.Ctx.Input.Host()
	arrayIndex := strings.Split(string(Host), ":")
	strHostIP := arrayIndex[0]

	err = facedevice.SubscribeWithDevice(structDeviceInfo, strHostIP)
	if err != nil {
		beego.Error("添加设备错误", err)

		mapReply["code"] = 40000
		mapReply["message"] = "添加设备失败: 设备ip或序列号错误"

		byteReply, errindex := json.Marshal(mapReply)
		if errindex != nil {
			return []byte(""), errindex
		}

		return byteReply, err
	}

	nID, err := my_db.AddDevice(structDeviceInfo)
	if err != nil {
		return []byte(""), err
	}

	mapData := make(map[string]interface{}, 0)
	mapData["nID"] = nID
	mapData["name"] = structDeviceInfo.Name
	mapData["id"] = nID
	mapData["serial"] = structDeviceInfo.DevcieId
	mapData["status"] = 1

	mapReply["code"] = 20000
	mapReply["data"] = mapData

	byteReply, err := json.Marshal(mapReply)
	if err != nil {
		beego.Error(err)
		return []byte(""), err
	}

	return byteReply, nil
}

func AddDeviceXiongmai(strPsw string, strSerial string, c *AddDeviceController) ([]byte, error) {
	data := c.Ctx.Input.RequestBody

	mapReply := make(map[string]interface{}, 0)
	mapRequest := make(map[string]interface{}, 0)
	err := json.Unmarshal(data, &mapRequest)
	if err != nil {
		beego.Error(err)
		return []byte(""), err
	}

	err = check_device.DeviceCheck(strSerial)
	if err != nil {
		beego.Error("添加设备错误", err)

		mapReply["code"] = 40000
		mapReply["message"] = "添加设备失败: " + err.Error()

		byteReply, errindex := json.Marshal(mapReply)
		if errindex != nil {
			return []byte(""), errindex
		}

		return byteReply, err
	}

	var structDeviceInfo my_db.DeviceInfo
	structDeviceInfo.Name = mapRequest["name"].(string)
	structDeviceInfo.DevcieId = strSerial
	structDeviceInfo.Ip = mapRequest["ip"].(string)
	structDeviceInfo.Type = 1
	structDeviceInfo.Psw = mapRequest["psw"].(string)

	structDeviceInfo.CompanyID = 0
	structDeviceInfo.Status = 1

	nCount, err := my_db.AddDevice(structDeviceInfo)
	if err != nil {
		return []byte(""), err
	}

	err = my_db.EnableDevice(strSerial)
	if err != nil {
		return []byte(""), err
	}

	//绑定设备服务器
	err = facedevicexiongmai.DeviceSetServer(strSerial, strPsw)
	if err != nil {
		structDeviceInfo.Status = 0
		errIndex := my_db.UpdateDevice(structDeviceInfo)
		if errIndex != nil {
			beego.Error("添加设备错误", errIndex)
			return []byte(""), errIndex
		}

		mapReply["code"] = 40000
		mapReply["message"] = "添加设备失败" + err.Error()
		byteReply, errindex := json.Marshal(mapReply)
		if err != nil {
			return []byte(""), errindex
		}
		return byteReply, err
	}

	// structDeviceInfo, err := my_db.GetDeviceByDeviceId(strSerial)
	// if err != nil {
	// 	return []byte(""), err
	// }

	mapReply["code"] = 20000
	mapReply["message"] = "请求成功"

	mapData := make(map[string]interface{}, 0)

	mapData["name"] = structDeviceInfo.Name
	mapData["id"] = nCount
	mapData["serial"] = structDeviceInfo.DevcieId

	mapReply["data"] = mapData

	byteReply, errindex := json.Marshal(mapReply)
	if errindex != nil {
		return []byte(""), errindex
	}
	return byteReply, nil
}

//人脸设备列表
type FaceDeviceListController struct {
	beego.Controller
}

func (c *FaceDeviceListController) Post() {
	//data := c.Ctx.Input.RequestBody

	//mapRequest := make(map[string]interface{}, 0)
	mapReply := make(map[string]interface{}, 0)
	mapData := make(map[string]interface{}, 0)

	//err := json.Unmarshal(data, &mapRequest)

	//海清的设备
	deviceInfoArray, err := my_db.GetDeviceListByType(0, 0)
	if err != nil {
		beego.Error("\r\n 获取设备列表失败", err)
		return
	}

	arrayDevice := make([]interface{}, 0)
	for _, structInfo := range deviceInfoArray {
		mapDeviceInfo := make(map[string]interface{}, 0)

		_, err := facedevice.GetDeviceInfo(structInfo)
		if err != nil {
			beego.Warning("\r\npost查找设备信息失败：", err)
			mapDeviceInfo["status"] = 0
		} else {
			mapDeviceInfo["status"] = 1
		}

		mapDeviceInfo["id"] = structInfo.Id
		mapDeviceInfo["serial"] = structInfo.DevcieId
		mapDeviceInfo["type"] = structInfo.Type

		arrayDevice = append(arrayDevice, mapDeviceInfo)
	}

	//雄迈的设备
	deviceInfoArray, err = my_db.GetDeviceListByType(0, 1)
	if err != nil {
		beego.Error("\r\n 获取设备列表失败", err)
		return
	}

	for _, structInfo := range deviceInfoArray {
		mapDeviceInfo := make(map[string]interface{}, 0)

		mapDeviceInfo["id"] = structInfo.Id
		mapDeviceInfo["serial"] = structInfo.DevcieId
		mapDeviceInfo["type"] = structInfo.Type
		if structInfo.ActiveTime > 0 {

			timeSpan, err := time.ParseDuration("10s")
			if err != nil {
				beego.Error(err)
				return
			}

			timeActv := time.Unix(structInfo.ActiveTime, 0)
			timeNow := time.Now()

			//timeIndex := time.Unix(timeNow.Unix(), 0)
			//			beego.Debug("timeIndex:", timeIndex.Format(rule_algorithm.TIME_LAYOUT))

			if timeNow.Sub(timeActv) > timeSpan {
				mapDeviceInfo["status"] = 0
			} else {
				mapDeviceInfo["status"] = 1
			}

		} else {
			mapDeviceInfo["status"] = 0
		}

		arrayDevice = append(arrayDevice, mapDeviceInfo)
		// err = my_db.UpdateDevice(structInfo)
		// if err != nil {
		// 	beego.Error(err)
		// 	return
		// }
	}

	mapData["deviceArray"] = arrayDevice

	mapReply["data"] = mapData
	mapReply["code"] = 20000
	mapReply["message"] = "请求成功"

	responseByte, err := json.Marshal(mapReply)
	if err != nil {
		return
	}

	c.Data["json"] = json.RawMessage(string(responseByte))
	c.ServeJSON()
}

// /api/management/device/unbind 删除设备
type RemoveDeviceController struct {
	beego.Controller
}

func (c *RemoveDeviceController) Post() {
	data := c.Ctx.Input.RequestBody

	mapRequest := make(map[string]interface{}, 0)

	err := json.Unmarshal(data, &mapRequest)
	if err != nil {
		beego.Error("删除设备出错:", err)
		return
	}

	structDeviceInfo, err := my_db.GetDeviceById(int64(mapRequest["id"].(float64)))
	if err != nil {
		beego.Error("删除设备出错", err)
		return
	}

	//雄迈设备
	if structDeviceInfo.Type == 1 {
		//解除绑定服务
		// err := facedevicexiongmai.DeviceUnbingServer(structDeviceInfo.DevcieId)
		// if err != nil {
		// 	beego.Error("删除设备错误:", err)
		// 	return
		// }

		if structDeviceInfo.Status == 1 {
			//让WSS服务停止发送心跳
			err = facedevicexiongmai.RemoveDeviceWSS(structDeviceInfo.DevcieId)
			if err != nil {
				beego.Error("删除设备错误", err)
				return
			}
		}

		//使设备失效
		err = my_db.UnableDevice(structDeviceInfo.DevcieId)
		if err != nil {
			beego.Error("删除设备错误", err)
			return
		}

	} else if structDeviceInfo.Type == 0 {
		//如果是海清的设备
		err = facedevice.UnSubcribeDevice(structDeviceInfo)
		if err != nil {
			beego.Warning("删除设备警告", err)
		}

		err = my_db.RemoveDevice(mapRequest["serial"].(string))
		if err != nil {
			beego.Error("删除设备出错:", err)
			return
		}
	} else {
		beego.Error("删除设备错误， 错误的设备类型")
		return
	}

	mapReply := make(map[string]interface{}, 0)
	mapReply["code"] = 20000
	mapReply["data"] = ""
	mapReply["message"] = "请求成功"

	byteReply, err := json.Marshal(mapReply)
	if err != nil {
		beego.Error("删除设备出错", err)
		return
	}

	c.Data["json"] = json.RawMessage(string(byteReply))
	c.ServeJSON()
}

///api/management/device/version
type DeviceGetVersionController struct {
	beego.Controller
}

type DeviceGetVersionRequst_Json struct {
	Id int64 `json:"id"`
}

type DeviceGetVersionReply_Json struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

func (c *DeviceGetVersionController) Post() {
	data := c.Ctx.Input.RequestBody
	var structRequstInfo DeviceGetVersionRequst_Json
	err := json.Unmarshal(data, &structRequstInfo)
	if err != nil {
		beego.Error("获取设备版本信息错误", err)
		return
	}

	structDeviceInfo, err := my_db.GetDeviceById(structRequstInfo.Id)
	if err != nil {
		beego.Error("获取设备版本信息错误", err)
		return
	}

	var structReply DeviceGetVersionReply_Json

	//海清设备
	if structDeviceInfo.Type == 0 {
		structDeviceData, err := facedevice.GetDeviceInfo(structDeviceInfo)
		if err != nil {
			beego.Error("获取设版本信息错误", err)
			return
		}
		structReply.Data = structDeviceData.Version
	}

	//雄迈设备
	if structDeviceInfo.Type == 1 {
		structReply.Data = structDeviceInfo.Version
	}

	structReply.Code = 20000
	structReply.Message = "请求成功"

	byteReply, err := json.Marshal(structReply)
	if err != nil {
		beego.Error("获取设备版本信息错误", err)
		return
	}
	c.Data["json"] = json.RawMessage(byteReply)
	c.ServeJSON()
}

///api/management/device/truncate
type DeviceTruncateController struct {
	beego.Controller
}

type DeviceTruncateReply_Json struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

func (c *DeviceTruncateController) Post() {

	data := c.Ctx.Input.RequestBody

	mapRequest := make(map[string]interface{}, 0)

	err := json.Unmarshal(data, &mapRequest)

	nID := int64(0)

	switch reflect.TypeOf(mapRequest["id"]).Kind() {
	case reflect.String:
		{
			nIDIndex, err := strconv.ParseInt(mapRequest["id"].(string), 0, 64)
			if err != nil {
				beego.Error(err)
				return
			}
			nID = nIDIndex
		}
		break
	case reflect.Float64:
		nID = int64(mapRequest["id"].(float64))
		break
	}

	if nID > 0 {
		structDeviceInfo, err := my_db.GetDeviceById(nID)
		if err != nil {
			beego.Error(err)
			return
		}

		if structDeviceInfo.Type == 0 {
			err = facedevice.RemoveAllDeviceEmployee(structDeviceInfo)
			if err != nil {
				beego.Error(err)
				return
			}

		}

		if structDeviceInfo.Type == 1 {
			var structXMRemoveInfo facedevicexiongmai.RemoveEmployeeDataJSON
			structXMRemoveInfo.EmployeeInfo.UserID = "ALL"
			structXMRemoveInfo.DeviceInfo.DeviceID = structDeviceInfo.DevcieId
			err = facedevicexiongmai.RemoveEmployeeWSS(structXMRemoveInfo)
			if err != nil {
				beego.Error(err)
				return
			}

		}

	} else {
		beego.Error("设备ID为空, ID:", nID)
		return
	}

	var structReply DeviceTruncateReply_Json
	structReply.Code = 20000
	structReply.Data = ""
	structReply.Message = "请求成功"

	byteReply, err := json.Marshal(structReply)
	if err != nil {
		beego.Error("清空设备人员数据错误", err)
		return
	}
	c.Data["json"] = json.RawMessage(string(byteReply))
	c.ServeJSON()
}

type RebootXMDeviceController struct {
	beego.Controller
}

func (c *RebootXMDeviceController) Post() {
	data := c.Ctx.Input.RequestBody

	mapRecv := make(map[string]interface{}, 0)

	err := json.Unmarshal(data, &mapRecv)

	nID := mapRecv["id"].(float64)

	structDeviceInfo, err := my_db.GetDeviceById(int64(nID))
	if err != nil {
		beego.Error(err)
		return
	}

	mapReply := make(map[string]interface{}, 0)

	err = facedevicexiongmai.RestartXMDevice(structDeviceInfo.DevcieId)
	if err != nil {
		mapReply["code"] = 50000
		mapReply["msg"] = err.Error()
	} else {
		mapReply["code"] = 20000
		mapReply["msg"] = "请求成功"
	}

	byteReply, err := json.Marshal(mapReply)
	if err != nil {
		beego.Error(err)
		return
	}

	c.Data["json"] = json.RawMessage(byteReply)
	c.ServeJSON()
}

type SyncEmployeeListController struct {
	beego.Controller
}

func (c *SyncEmployeeListController) Post() {
	data := c.Ctx.Input.RequestBody

	mapRequset := make(map[string]interface{}, 0)
	err := json.Unmarshal(data, &mapRequset)
	if err != nil {
		beego.Error(err)
		return
	}

	structDeviceInfo, err := my_db.GetDeviceById(int64(mapRequset["id"].(float64)))
	if err != nil {
		beego.Error(err)
		return
	}

	byteResult, err := qiyewss.GetUserInfoByPage()
	if err != nil {
		beego.Error(err)
		return
	}

	//beego.Debug("人员列表:", string(byteResult))

	mapResult := make(map[string]interface{}, 0)
	byteReply := make([]byte, 0)

	err = json.Unmarshal(byteResult, &mapResult)
	if err != nil {
		beego.Error(err)
		return
	}

	mapReply := make(map[string]interface{}, 0)
	if mapResult["errcode"].(float64) != 0 {
		mapReply["code"] = 50000
		mapReply["message"] = mapResult["errmsg"]

		byteReply, err = json.Marshal(mapReply)
		if err != nil {
			beego.Error(err)
			return
		}

		c.Data["json"] = json.RawMessage(byteReply)
		c.ServeJSON()
		return
	}

	mapData := make([]map[string]interface{}, 0)
	mapDataIndex := make(map[string]interface{}, 0)
	//mapBody := make(map[string]interface{}, 0)
	mapBody := mapReply["body"].(map[string]interface{})
	for _, value := range mapBody["userinfo"].([]map[string]interface{}) {
		if len(value["fa_list"].([]interface{})) > 0 {
			mapDataIndex["name"] = value["name"].(string)
			mapDataIndex["id"] = value["userid"].(string)

			ArraymapFace := value["fa_list"].([]map[string]interface{})
			mapFace0 := ArraymapFace[0]

			if structDeviceInfo.Type == 1 {
				//雄迈
				var structEmployeeXM facedevicexiongmai.AddEmployeeInfoJSON
				structEmployeeXM.EmployeInfo.UserID = value["userid"].(string)
				structEmployeeXM.EmployeInfo.CardID = ""
				structEmployeeXM.EmployeInfo.GroupID = "default"
				structEmployeeXM.EmployeInfo.Name = value["name"].(string)
				structEmployeeXM.EmployeInfo.StartTime = time.Now().Format(TIME_LAYOUT)
				spanindex, err := time.ParseDuration("8760h")
				if err != nil {
					beego.Error(err)
					return
				}
				structEmployeeXM.EmployeInfo.EndTime = time.Now().Add(spanindex).Format(TIME_LAYOUT)

				byteImage, err := base64.StdEncoding.DecodeString(mapFace0["data"].(string))
				if err != nil {
					beego.Error(err)
					return
				}
				structEmployeeXM.EmployeInfo.FaceFeature = strings.ToUpper(hex.EncodeToString(byteImage))
				structEmployeeXM.DeviceInfo.DeviceID = structDeviceInfo.DevcieId
				err = facedevicexiongmai.AddEmployeeWSS(structEmployeeXM)
				if err != nil {
					beego.Error(err)

					mapDataIndex["result"] = 0
					mapDataIndex["reason"] = err.Error()

					mapData = append(mapData, mapDataIndex)
					continue
				}

				mapDataIndex["result"] = 200
				mapDataIndex["reason"] = ""
				mapData = append(mapData, mapDataIndex)
			} else if structDeviceInfo.Type == 0 {
				//海清
				nUserID := rand.Int31()
				nUserIDIndex, err := my_db.GetUserIDByXM(value["userid"].(string))
				if err != nil {
					beego.Error(err)
					return
				}

				var structEmployeeHQ facedevice.EmployeeInfo
				structEmployeeHQ.Name = value["name"].(string)
				structEmployeeHQ.Pic = mapFace0["data"].(string)

				if nUserIDIndex <= 0 {
					err = my_db.InsertUserID(my_db.UserIDInfo{XMUserID: value["userid"].(string), HQUserID: int64(nUserID)})
					if err != nil {
						beego.Error(err)
						return
					}

					structEmployeeHQ.UserID = int64(nUserID)
				} else {
					structEmployeeHQ.UserID = nUserIDIndex
				}

				err = facedevice.AddDeviceEmployee(structDeviceInfo, structEmployeeHQ)
				if err != nil {
					beego.Error(err)

					mapDataIndex["result"] = 0
					mapDataIndex["reason"] = err.Error()
					mapData = append(mapData, mapDataIndex)
					continue
				}

				mapDataIndex["result"] = 200
				mapDataIndex["reason"] = ""
				mapData = append(mapData, mapDataIndex)
			}

		}
	}

	mapReply["code"] = 20000
	mapReply["message"] = "同步成功"
	mapReply["data"] = mapData

	byteReply, err = json.Marshal(mapReply)
	if err != nil {
		beego.Error(err)
		return
	}

	c.Data["json"] = json.RawMessage(byteReply)
	c.ServeJSON()
}

type GetEmployeeListController struct {
	beego.Controller
}

func (c *GetEmployeeListController) Post() {
	byteResult, err := qiyewss.GetUserInfoByPage()
	if err != nil {
		beego.Error(err)
		return
	}

	c.Data["json"] = json.RawMessage(byteResult)
	c.ServeJSON()
}

type SyncEmployeeController struct {
	beego.Controller
}

func (c *SyncEmployeeController) Post() {
	data := c.Ctx.Input.RequestBody

	mapRequest := make(map[string]interface{}, 0)
	err := json.Unmarshal(data, &mapRequest)
	if err != nil {
		beego.Error(err)
		return
	}

	structDeviceInfo, err := my_db.GetDeviceById(int64(mapRequest["deviceID"].(float64)))
	if err != nil {
		beego.Error(err)
		return
	}

	strName := mapRequest["name"].(string)
	strUserId := mapRequest["userid"].(string)
	strBase64Image := mapRequest["image"].(string)
	mapReply := make(map[string]interface{}, 0)
	byteReply := make([]byte, 0)

	if structDeviceInfo.Type == 1 {
		var structEmployeeXM facedevicexiongmai.AddEmployeeInfoJSON
		structEmployeeXM.EmployeInfo.UserID = strUserId
		structEmployeeXM.EmployeInfo.CardID = ""
		structEmployeeXM.EmployeInfo.GroupID = "default"
		structEmployeeXM.EmployeInfo.Name = strName
		structEmployeeXM.EmployeInfo.StartTime = time.Now().Format(TIME_LAYOUT)
		spanindex, err := time.ParseDuration("8760h")
		if err != nil {
			beego.Error(err)
			return
		}
		structEmployeeXM.EmployeInfo.EndTime = time.Now().Add(spanindex).Format(TIME_LAYOUT)

		byteImage, err := base64.StdEncoding.DecodeString(strBase64Image)
		if err != nil {
			beego.Error(err)
			return
		}
		structEmployeeXM.EmployeInfo.FaceFeature = strings.ToUpper(hex.EncodeToString(byteImage))
		structEmployeeXM.DeviceInfo.DeviceID = structDeviceInfo.DevcieId
		err = facedevicexiongmai.AddEmployeeWSS(structEmployeeXM)
		if err != nil {
			mapReply["code"] = 50000
			mapReply["message"] = err.Error()
			byteReply, err = json.Marshal(mapReply)
			if err != nil {
				beego.Error(err)
				return
			}

			c.Data["json"] = json.RawMessage(byteReply)
			c.ServeJSON()
			return
		}

	} else if structDeviceInfo.Type == 0 {
		//海清
		nUserID := rand.Int31()
		nUserIDIndex, err := my_db.GetUserIDByXM(strUserId)
		if err != nil {
			beego.Error(err)
			return
		}

		var structEmployeeHQ facedevice.EmployeeInfo
		structEmployeeHQ.Name = strName
		structEmployeeHQ.Pic = strBase64Image

		if nUserIDIndex <= 0 {
			err = my_db.InsertUserID(my_db.UserIDInfo{XMUserID: strUserId, HQUserID: int64(nUserID)})
			if err != nil {
				beego.Error(err)
				return
			}

			structEmployeeHQ.UserID = int64(nUserID)
		} else {
			structEmployeeHQ.UserID = nUserIDIndex
		}

		err = facedevice.AddDeviceEmployee(structDeviceInfo, structEmployeeHQ)
		if err != nil {
			mapReply["code"] = 50000
			mapReply["message"] = err.Error()

			byteReply, err = json.Marshal(mapReply)
			if err != nil {
				beego.Error(err)
				return
			}
			c.Data["json"] = json.RawMessage(byteReply)
			c.ServeJSON()
			return
		}
	}

	mapReply["code"] = 20000
	mapReply["message"] = "同步成功"

	byteReply, err = json.Marshal(mapReply)
	if err != nil {
		beego.Error(err)
		return
	}

	c.Data["json"] = json.RawMessage(byteReply)
	c.ServeJSON()
}
