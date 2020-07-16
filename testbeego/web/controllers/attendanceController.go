package controllers

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/nfnt/resize"

	"../check_device"
	"../confreader"
	"../facedevice"
	"../facedevicexiongmai"
	"../my_aes"
	"../my_db"
	"../rule_algorithm"
	"../xlsxtool"

	"github.com/astaxie/beego"
	"github.com/axgle/mahonia"
	"github.com/buger/jsonparser"
	"github.com/tealeg/xlsx"
)

/*--------------------------------------------------------*/

//验证路由
type VerifyController struct {
	beego.Controller
}

func (c *VerifyController) Post() {
	data := c.Ctx.Input.RequestBody

	//fmt.Printf("%s", data)

	//beego.Debug(string(data))

	DeviceID, err := jsonparser.GetInt([]byte(data), "info", "DeviceID")
	if err != nil {
		beego.Error("添加验证内容错误", err)
		return
	}

	CreateTime, err := jsonparser.GetString([]byte(data), "info", "CreateTime")
	if err != nil {
		beego.Error("添加验证内容错误", err)
		return
	}

	SanpPic, err := jsonparser.GetString([]byte(data), "SanpPic")
	if err != nil {
		beego.Error("添加验证内容错误", err)
		return
	}

	Name, err := jsonparser.GetString([]byte(data), "info", "Name")
	if err != nil {
		beego.Error("添加验证内容错误", err)
		return
	}

	VerifyStatus, err := jsonparser.GetInt([]byte(data), "info", "VerifyStatus")
	if err != nil {
		beego.Error("添加验证内容错误", err)
		return
	}

	//CustomizeID
	CustomId, err := jsonparser.GetInt([]byte(data), "info", "CustomizeID")
	if err != nil {
		beego.Error("添加验证内容错误", err)
		return
	}

	templateData, err := jsonparser.GetFloat([]byte(data), "info", "Temperature")
	if err != nil {
		beego.Warning("未获取到体温数据", err)
	}

	var structSnapInfo my_db.VerifyRecordInfo
	structSnapInfo.Id = -1
	structSnapInfo.XMDeviceID = ""
	structSnapInfo.CreateTime = CreateTime
	structSnapInfo.VerifyStatus = VerifyStatus
	structSnapInfo.VerifyPic = SanpPic
	structSnapInfo.Stanger = 0
	structSnapInfo.Name = Name
	structSnapInfo.DeviceId = DeviceID
	structSnapInfo.Template = fmt.Sprintf("%f", templateData)
	structSnapInfo.CutomId = CustomId

	arrayIndex := strings.Split(CreateTime, "T")

	structSnapInfo.CreateDate = arrayIndex[0]

	structDeviceInfo, err := my_db.GetDeviceByDeviceId(fmt.Sprintf("%d", structSnapInfo.DeviceId))
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

	if structDeviceInfo.IsAttendance <= 0 {
		c.Data["json"] = json.RawMessage(string(Reply))
		c.ServeJSON()
		return
	}

	structSnapInfo.DeviceName = structDeviceInfo.Name

	structEmployeeInfo, err := my_db.GetEmployeeById(CustomId)
	if err != nil {
		beego.Error(err)
		return
	}

	nMaxRuleID, err := my_db.GetMaxRuleID()
	if err != nil {
		beego.Error(err)
		return
	}

	//如果是固定时间
	if structEmployeeInfo.RuleID > nMaxRuleID {
		strAttendanceResult, strAttendanceTime, nRangeIndex, err := rule_algorithm.GetAttendnaceResultByDayRange(structEmployeeInfo.RuleID-nMaxRuleID, structSnapInfo.CreateTime)
		if err != nil {
			beego.Error(err)
			return
		}
		structSnapInfo.Result = strAttendanceResult
		structSnapInfo.RangeIndex = nRangeIndex
		structSnapInfo.AttendanceTime = strAttendanceTime
	} else if structEmployeeInfo.RuleID == 0 {
		structSnapInfo.Result = "无考勤规则"
	} else {
		//如果是轮班
		strAttendanceResult, strAttendanceTime, nRangeIndex, err := rule_algorithm.GetAttendanceShaduleResult(structSnapInfo.CreateTime, structEmployeeInfo.RuleID, structEmployeeInfo.Id)
		if err != nil {
			beego.Error("获取考勤记录列表错误", err)
			return
		}

		structSnapInfo.AttendanceTime = strAttendanceTime
		structSnapInfo.Result = strAttendanceResult
		structSnapInfo.RangeIndex = nRangeIndex
	}

	structSnapInfo.CompanyID = structDeviceInfo.CompanyID

	err = my_db.AddDeviceRecord(structSnapInfo)
	if err != nil {
		beego.Error("记录抓拍数据错误", err)
		return
	}

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

	DeviceID, err := jsonparser.GetInt([]byte(data), "info", "DeviceID")
	if err != nil {
		beego.Error("添加抓拍内容错误", err)
		return
	}

	CreateTime, err := jsonparser.GetString([]byte(data), "info", "CreateTime")
	if err != nil {
		beego.Error("添加抓拍内容错误", err)
		return
	}

	SanpPic, err := jsonparser.GetString([]byte(data), "SanpPic")
	if err != nil {
		beego.Error("添加抓拍内容错误", err)
		return
	}

	templateData, err := jsonparser.GetFloat([]byte(data), "info", "Temperature")
	if err != nil {
		beego.Warning("未获取到体温数据", err)
	}

	var structSnapInfo my_db.VerifyRecordInfo
	structSnapInfo.Id = -1
	structSnapInfo.XMDeviceID = ""
	structSnapInfo.CreateTime = CreateTime
	structSnapInfo.VerifyStatus = -1
	structSnapInfo.VerifyPic = SanpPic
	structSnapInfo.Stanger = 1
	structSnapInfo.Name = "陌生人"
	structSnapInfo.DeviceId = DeviceID
	structSnapInfo.Template = fmt.Sprintf("%f", templateData)

	structDeviceInfo, err := my_db.GetDeviceByDeviceId(fmt.Sprintf("%d", DeviceID))
	if err != nil {
		beego.Error(err)
		return
	}

	structSnapInfo.CompanyID = structDeviceInfo.CompanyID

	err = my_db.AddDeviceRecord(structSnapInfo)
	if err != nil {
		beego.Error("记录抓拍数据错误", err)
		return
	}

	// db, err := sql.Open("sqlite3", my_db.GetRootPath()+"/beegoServer.db")
	// if err != nil {
	// 	db.Close()
	// 	beego.Error("添加抓拍内容错误", err)
	// 	return
	// }

	// rows, err := db.Query("SELECT count(*) FROM VERIFY_PERSON")
	// if err != nil {
	// 	db.Close()
	// 	beego.Error("添加抓拍内容错误", err)
	// 	return
	// }

	// var count int
	// for rows.Next() {
	// 	err = rows.Scan(&count)
	// 	if err != nil {
	// 		db.Close()
	// 		beego.Error("添加抓拍内容错误", err)
	// 		return
	// 	}
	// }

	// index := count + 1

	// stmt, err := db.Prepare("insert into VERIFY_PERSON(ID, DEVICE_ID, CREATE_TIME, NAME, VERIFY_STATUS,VERIFY_PIC, CUSTOM_ID, STRANGER, TEMPLATE) values(?,?,?,?,?,?,?, ?, ?)")
	// if err != nil {
	// 	db.Close()
	// 	beego.Error("添加抓拍内容错误", err)
	// 	return
	// }
	// res, err := stmt.Exec(index, DeviceID, CreateTime, "陌生人", 0, SanpPic, -1, 1, templateData)
	// if err != nil {
	// 	db.Close()
	// 	beego.Error("添加抓拍内容错误", err)
	// 	return
	// }
	// affect, err := res.RowsAffected()
	// if err != nil {
	// 	db.Close()
	// 	beego.Error("添加抓拍内容错误", err)
	// 	return
	// }
	// beego.Debug(affect)
	// db.Close()

	Reply := `{
			"operator": "SnapPush",
			"code": 200,
			"info": {
				"Result": "Ok"
			}
		}`

	data, err = json.Marshal(Reply)
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
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	var arrayData []XMUpdateRequst
	var arrayImg []string

	if c.Ctx.Input.Context.Request.MultipartForm == nil {
		return
	}

	if len(c.Ctx.Input.Context.Request.MultipartForm.Value) <= 0 {
		return
	}

	if len(c.Ctx.Input.Context.Request.MultipartForm.File) <= 0 {
		return
	}

	data := c.Ctx.Input.Context.Request.MultipartForm.Value["data"]

	files := c.Ctx.Input.Context.Request.MultipartForm.File["file"]

	for _, dataIndex := range data {
		var structRqust XMUpdateRequst
		err := json.Unmarshal([]byte(dataIndex), &structRqust)
		if err != nil {
			beego.Error("更新雄迈设备数据错误", err)
			return
		}

		//beego.Debug("接收到数据:", string(dataIndex))
		arrayData = append(arrayData, structRqust)
	}

	for _, file := range files {
		//beego.Debug("接收到文件名:", file.Filename)
		fp, err := file.Open()
		byteRead := make([]byte, file.Size)
		if err != nil {
			beego.Error(err)
			return
		}
		_, err = fp.Read(byteRead)
		if err != nil {
			beego.Error(err)
			return
		}

		baseStore := base64.StdEncoding.EncodeToString(byteRead)
		strBaseFile := `data:image/jpeg;base64,` + baseStore
		arrayImg = append(arrayImg, strBaseFile)
		//beego.Debug("图片base64数据:", baseStore)
	}

	// Id           int64
	// DeviceId     int64
	// CreateTime   string
	// Name         string
	// VerifyStatus int
	// VerifyPic    string
	// CutomId      int64
	// Stanger      int64
	// Template     string

	for nIndex, value := range arrayData {
		var structSnapInfo my_db.VerifyRecordInfo
		structSnapInfo.Id = -1
		structSnapInfo.XMDeviceID = value.DeviceID
		structSnapInfo.VerifyStatus = value.PassResult
		structSnapInfo.VerifyPic = arrayImg[nIndex]

		arrayIndex := strings.Split(value.Time, " ")
		structSnapInfo.CreateDate = arrayIndex[0]
		structSnapInfo.CreateTime = arrayIndex[0] + "T" + arrayIndex[1] + "Z"
		if len(value.UserID) > 0 {
			nIndexCutomID, err := strconv.ParseInt(value.UserID, 0, 64)
			if err != nil {
				beego.Error(err)
				return
			}
			structSnapInfo.CutomId = nIndexCutomID
		} else {
			structSnapInfo.CutomId = -1
		}

		if value.PassResult == -8 {
			structSnapInfo.Stanger = 1
			structSnapInfo.Name = "陌生人"
		} else {
			structSnapInfo.Name = value.UserName
			structSnapInfo.Stanger = 0

			if value.PassResult == -5 {
				structDeviceInfoIndex, err := my_db.GetDeviceByDeviceId(value.DeviceID)
				if err != nil {
					beego.Error(err)
					return
				}
				if structDeviceInfoIndex.MultiCert > 0 {
					timeIndex, err := time.Parse(rule_algorithm.TIME_LAYOUT, value.Time)
					if err != nil {
						beego.Error(err)
						return
					}
					timeIndex = timeIndex.AddDate(0, -2, 0)

					arrayIndex := strings.Split(timeIndex.Format(rule_algorithm.TIME_LAYOUT), " ")
					strTimeIndex := arrayIndex[0] + "T" + arrayIndex[1] + "Z"
					nCount, err := my_db.GetRecordCountByContent(value.DeviceID, value.PassResult, strTimeIndex, value.UserName)
					if err != nil {
						beego.Error(err)
						return
					}
					if nCount > 0 {
						var openDoorParam facedevicexiongmai.OpenDoorInfoJSON
						openDoorParam.DeviceID = value.DeviceID
						openDoorParam.CtrlType = 1
						openDoorParam.Duration = 0
						err = facedevicexiongmai.OpenDoor(openDoorParam)
						if err != nil {
							beego.Error(err)
							return
						}

						openDoorParam.CtrlType = 0
						go facedevicexiongmai.OpenDoorDelayRouter(openDoorParam, 2)
					}

				}

			}
		}

		structSnapInfo.Template = ""
		structSnapInfo.DeviceType = 1

		structEmployeeInfo, err := my_db.GetEmployeeById(structSnapInfo.CutomId)
		if err != nil {
			beego.Error(err)
			return
		}

		nMaxRuleID, err := my_db.GetMaxRuleID()
		if err != nil {
			beego.Error(err)
			return
		}

		//如果是固定时间
		if structEmployeeInfo.RuleID > nMaxRuleID {
			strAttendanceResult, strAttendanceTime, nRangeIndex, err := rule_algorithm.GetAttendnaceResultByDayRange(structEmployeeInfo.RuleID-nMaxRuleID, structSnapInfo.CreateTime)
			if err != nil {
				beego.Error(err)
				return
			}
			structSnapInfo.AttendanceTime = strAttendanceTime
			structSnapInfo.Result = strAttendanceResult
			structSnapInfo.RangeIndex = nRangeIndex
		} else if structEmployeeInfo.RuleID <= 0 {
			structSnapInfo.Result = "无考勤规则"
		} else {
			//如果是轮班
			strAttendanceResult, strAttendanceTime, nRangeIndex, err := rule_algorithm.GetAttendanceShaduleResult(structSnapInfo.CreateTime, structEmployeeInfo.RuleID, structEmployeeInfo.Id)
			if err != nil {
				beego.Error("获取考勤记录列表错误", err)
				return
			}

			structSnapInfo.AttendanceTime = strAttendanceTime
			structSnapInfo.Result = strAttendanceResult
			structSnapInfo.RangeIndex = nRangeIndex
		}

		structDeviceInfo, err := my_db.GetDeviceByDeviceId(value.DeviceID)
		if err != nil {
			beego.Error(err)
			return
		}

		if structDeviceInfo.IsAttendance <= 0 {
			c.Data["json"] = ""
			c.ServeJSON()
			return
		}

		structSnapInfo.DeviceName = structDeviceInfo.Name
		structSnapInfo.CompanyID = structDeviceInfo.CompanyID
		err = my_db.AddDeviceRecord(structSnapInfo)
		if err != nil {
			beego.Error(err)
			return
		}
	}

	c.Data["json"] = ""
	c.ServeJSON()

	//c.GetFile("")
}

//登录
type LoginController struct {
	beego.Controller
}

//接收数据
type LoginRequest_Json struct {
	Password string `json:"password"`
	Username string `json:"username"`
}

//返回数据
type LoginResultData_Json struct {
	Token string `json:"token"`
}

type LoginResult_Json struct {
	Code    int64                `json:"code"`
	Message string               `json:"message"`
	Data    LoginResultData_Json `json:"data"`
}

func (c *LoginController) Post() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	data := c.Ctx.Input.RequestBody
	//beego.Error(string(data))

	//response := []byte(`{"code":20000,"message":"请求成功!","data":{"token":"L0Q0MllLOUZoRFg3WDQ1Uy90ZnNCNkNsc09sQVEvN2JvUEZENjlsa2R6ek1uck5TTUFIMDFNQXRsdDhxaHNmeU1KbTllWHpWZ3JiUUF3V0J6dnJ1c0E9PQ=="}}`)
	var structRequst LoginRequest_Json
	err := json.Unmarshal(data, &structRequst)
	if err != nil {
		beego.Error(err)
		return
	}

	structAdminInfoArray, err := my_db.GetAdminUserList()
	if err != nil {
		beego.Error(err)
		return
	}

	bLogSuc := false
	for _, structIndex := range structAdminInfoArray {
		if structIndex.Account == structRequst.Username && structIndex.Psw == structRequst.Password {
			bLogSuc = true
		}
	}

	var responseStruct LoginResult_Json

	if !bLogSuc {
		responseStruct.Code = 50000
		responseStruct.Message = confreader.GetValue("LoginFail")

		responseByte, err := json.Marshal(responseStruct)
		if err != nil {
			return
		}

		c.Data["json"] = json.RawMessage(string(responseByte))
		c.ServeJSON()
		return
	}

	strAes, _ := my_aes.Encrypt([]byte(structRequst.Username))
	// timeNow := time.Now()

	// //strDecry, _ := my_aes.Dncrypt(strAes)
	// //beego.Error("解密", strDecry)

	// structGetToken, err := my_db.GetToken(strAes)
	// if err != nil {
	// 	beego.Error(err)
	// 	return
	// }

	// var structToken my_db.TokenInfo
	// structToken.Token = strAes
	// structToken.CreateTime = timeNow.String()

	// if structGetToken.Token != "" {
	// 	err := my_db.UpdateToken(structToken)
	// 	if err != nil {
	// 		beego.Error(err)
	// 		return
	// 	}
	// } else {
	// 	err := my_db.AddToken(structToken)
	// 	if err != nil {
	// 		beego.Error(err)
	// 		return
	// 	}
	// }

	//beego.Error(strAes)
	responseStruct.Code = 20000
	responseStruct.Message = confreader.GetValue("SucRequest")
	responseStruct.Data.Token = strAes
	responseByte, err := json.Marshal(responseStruct)
	if err != nil {
		return
	}

	c.Data["json"] = json.RawMessage(string(responseByte))
	c.ServeJSON()
}

//登出
type LoginOutController struct {
	beego.Controller
}

type LoginOutReply_Json struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

func (c *LoginOutController) Post() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	data := c.Ctx.Input.Header("X-Token")

	//处理
	err := my_db.RemoveToken(string(data))
	if err != nil {
		beego.Error(err)
		return
	}

	//返回
	var structReply LoginOutReply_Json
	structReply.Code = 20000
	structReply.Data = ""
	structReply.Message = "success"

	byteReply, err := json.Marshal(structReply)
	if err != nil {
		beego.Error(err)
		return
	}

	c.Data["json"] = json.RawMessage(string(byteReply))
	c.ServeJSON()
}

//注册
type RegisterController struct {
	beego.Controller
}

//接收
type RegisterRequest_Json struct {
	Name       string `json:"name"`
	Mobile     string `json:"mobile"`
	Password   string `json:"password"`
	Repassword string `json:"repassword"`
	Username   string `json:"username"`
}

//发送
type RegisterReply_Json struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
	Data    string `json:""`
}

func (c *RegisterController) Post() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	data := c.Ctx.Input.RequestBody

	var structRequst RegisterRequest_Json
	err := json.Unmarshal(data, &structRequst)
	if err != nil {
		beego.Error(err)
		return
	}

	adminArray, err := my_db.GetAdminUserList()
	if err != nil {
		beego.Error(err)
		return
	}

	bRepeat := false
	for _, structIndex := range adminArray {
		if structIndex.Account == structRequst.Username {
			bRepeat = true
		}
	}

	var structReply RegisterReply_Json
	if bRepeat {
		structReply.Code = 50000
		structReply.Data = ""
		structReply.Message = confreader.GetValue("AccountContext")
	} else {
		var structInfo my_db.AdminInfo
		structInfo.Name = structRequst.Name
		structInfo.Account = structRequst.Username
		structInfo.PhoneName = structRequst.Mobile
		structInfo.Psw = structRequst.Password
		structInfo.CompanyId = -1
		structInfo.ISRoot = 1
		err := my_db.AddAdminUser(structInfo)
		if err != nil {
			beego.Error("\r\n添加管理员错误:", err)
			return
		}

		structReply.Code = 20000
		structReply.Data = ""
		structReply.Message = confreader.GetValue("RegSuc")
	}

	byteReply, err := json.Marshal(structReply)
	if err != nil {
		beego.Error(err)
		return
	}

	c.Data["json"] = json.RawMessage(string(byteReply))
	c.ServeJSON()
}

//获取信息
type InfoController struct {
	beego.Controller
}

type InfoCompny_Json struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

type InfoResultData_Json struct {
	Roles        []string          `json:"roles"`
	Name         string            `json:"name"`
	Avatar       string            `json:"avatar"`
	Company      int64             `json:"company"`
	Companys     []InfoCompny_Json `json:"companys"`
	Introduction string            `json:"introduction"`
}

type InfoResult_Json struct {
	Code    int64               `json:"code"`
	Message string              `json:"message"`
	Data    InfoResultData_Json `json:"data"`
}

func (c *InfoController) Get() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	Xtoken := c.Ctx.Input.Header("X-Token")
	strAccount, err := my_aes.Dncrypt(string(Xtoken))
	if err != nil {
		beego.Error("获取用户信息错误", err)
		return
	}

	var responseStruct InfoResult_Json
	responseStruct.Data.Companys = make([]InfoCompny_Json, 0)

	structAdminInfo, err := my_db.GetAdminByAccount(strAccount)
	if err != nil {
		beego.Error("获取用户信息错误", err)
		return
	}

	responseStruct.Code = 20000
	responseStruct.Message = ""
	responseStruct.Data.Roles = append(responseStruct.Data.Roles, "admin")
	responseStruct.Data.Name = structAdminInfo.Name
	responseStruct.Data.Avatar = `https:\/\/qiniu.organize.wanglitiaoyi.cn\/2019\/10\/noavatar.png`
	responseStruct.Data.Introduction = ""
	responseStruct.Data.Company = -1

	if structAdminInfo.CompanyId != -1 {
		structCompanyInfo, err := my_db.GetCompnyById(structAdminInfo.CompanyId)
		if err != nil {
			beego.Error("获取用户信息错误", err)
			return
		}

		responseStruct.Data.Company = structCompanyInfo.Id
		//responseStruct.Data.Companys = append(responseStruct.Data.Companys, InfoCompny_Json{Id: structCompanyInfo.Id, Name: structCompanyInfo.Name})

		if len(structAdminInfo.Companys) > 0 {
			arraySplite := strings.Split(structAdminInfo.Companys, ",")
			for _, varIndex := range arraySplite {
				nCompanyIDIndex, err := strconv.ParseInt(varIndex, 0, 64)
				if err != nil {
					beego.Error(err)
					return
				}
				structCompanyInfo, err = my_db.GetCompnyById(nCompanyIDIndex)
				if err != nil {
					beego.Error(err)
					return
				}
				responseStruct.Data.Companys = append(responseStruct.Data.Companys, InfoCompny_Json{Id: structCompanyInfo.Id, Name: structCompanyInfo.Name})
			}
		}

	}

	responseByte, err := json.Marshal(responseStruct)
	if err != nil {
		return
	}

	c.Data["json"] = json.RawMessage(string(responseByte))
	c.ServeJSON()
}

//检查
type CheckController struct {
	beego.Controller
}

type CheckResultData_Json struct {
	Error int64  `json:"error"`
	Next  string `json:"next"`
	Name  string `json:"name"`
}

type CheckResult_Json struct {
	Code    int64                `json:"code"`
	Message string               `json:"message"`
	Data    CheckResultData_Json `json:"data"`
}

func (c *CheckController) Get() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	Xtoken := c.Ctx.Input.Header("X-Token")
	strAccount, err := my_aes.Dncrypt(string(Xtoken))
	if err != nil {
		beego.Error("检查用户公司信息错误", err)
		return
	}

	var responseStruct CheckResult_Json

	structAdminInfo, err := my_db.GetAdminByAccount(strAccount)
	if err != nil {
		beego.Error("检查用户公司信息错误", err)
		return
	}

	if structAdminInfo.CompanyId == -1 {
		responseStruct.Data.Error = 1
		responseStruct.Data.Next = "/company/create"
	} else {
		structCompanyInfo, err := my_db.GetCompnyById(structAdminInfo.CompanyId)
		if err != nil {
			beego.Error("检查用户公司信息错误", err)
			return
		}
		responseStruct.Data.Name = structCompanyInfo.Name
	}

	responseStruct.Code = 20000
	responseStruct.Message = ""

	responseByte, err := json.Marshal(responseStruct)
	if err != nil {
		return
	}

	c.Data["json"] = json.RawMessage(string(responseByte))
	c.ServeJSON()
}

//人脸设备列表
type FaceDeviceListController struct {
	beego.Controller
}

type ExtendInfo_Json struct {
	LiveThreshold        int64  `json:"LiveThreshold"`
	LiveFrameNum         int64  `json:"LiveFrameNum"`
	LedLightType         int64  `json:"LedLightType"`
	LedTimeBeg           string `json:"LedTimeBeg"`
	LedTimeEnd           string `json:"LedTimeEnd"`
	LedBrightness        int64  `json:"LedBrightness"`
	LedDisableAfterSec   int64  `json:"LedDisableAfterSec"`
	LcdBLDisable         int64  `json:"LcdBLDisable"`
	LcdBLDisableAfterSec int64  `json:"LcdBLDisableAfterSec"`
	Endian               int64  `json:"Endian"`
	ControlType          int64  `json:"ControlType"`
	Wiegand              int64  `json:"Wiegand"`
	PublicMjCardNo       int64  `json:"PublicMjCardNo"`
	AutoMjCardBgnNo      int64  `json:"AutoMjCardBgnNo"`
	AutoMjCardEndNo      int64  `json:"AutoMjCardEndNo"`
	IOStayTime           int64  `json:"IOStayTime"`
	IPAddr               string `json:"IPAddr"`
	Submask              string `json:"Submask"`
	Gateway              string `json:"Gateway"`
	IOType               int64  `json:"iotype"`
	AutoRebootDay        int64  `json:"AutoRebootDay"`
	CardMode             int64  `json:"CardMode"`
	PSW                  string `json:"password"`
	Delay                int64  `json:"delay"`
}

type FaceDaviceInfo_Json struct {
	Id                int64       `json:"id"`
	Name              string      `json:"name"`
	Serial            string      `json:"serial"`
	Model             string      `json:"model"`
	Type              int64       `json:"type"`
	Extend            interface{} `json:"extend"`
	Status            int64       `json:"status"`
	Account           string      `json:"account"`
	Psw               string      `json:"psw"`
	Ip                string      `json:"ip"`
	Checkatt          int64       `json:"checkatt"`
	Isshow_doorstatus bool        `json:"isshow_doorstatus"`
	Isopen            int64       `json:"isopen"`
	Multi_cert        int64       `json:"multi_cert"`
}

type FaceDeviceListResultData_Json struct {
	Item  []FaceDaviceInfo_Json `json:"items"`
	Total int64                 `json:"total"`
}

type FaceDeviceListResult_Json struct {
	Code    int64                         `json:"code"`
	Message string                        `json:"message"`
	Data    FaceDeviceListResultData_Json `json:"data"`
}

func (c *FaceDeviceListController) Get() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	Xtoken := c.Ctx.Input.Header("X-Token")
	Account, err := my_aes.Dncrypt(Xtoken)

	structAdminInfo, err := my_db.GetAdminByAccount(Account)
	if err != nil {
		beego.Error(err)
		return
	}

	var responseStruct FaceDeviceListResult_Json
	responseStruct.Data.Item = make([]FaceDaviceInfo_Json, 0)

	//海清的设备
	deviceInfoArray, err := my_db.GetDeviceListByType(structAdminInfo.CompanyId, 0)
	if err != nil {
		beego.Error("\r\n 获取设备列表失败", err)
		return
	}
	for _, structInfo := range deviceInfoArray {
		var FaceDeviceInfo FaceDaviceInfo_Json
		structDeviceInfo, err := facedevice.GetDeviceInfo(structInfo)
		if err != nil {
			beego.Warning("\r\npost查找设备信息失败：", err)
		}
		if fmt.Sprintf("%d", structDeviceInfo.DeviceId) == structInfo.DevcieId {
			FaceDeviceInfo.Status = 1
		} else {
			FaceDeviceInfo.Status = 0
		}

		FaceDeviceInfo.Id = structInfo.Id
		FaceDeviceInfo.Model = ""
		FaceDeviceInfo.Name = structInfo.Name
		FaceDeviceInfo.Serial = structInfo.DevcieId
		FaceDeviceInfo.Type = structInfo.Type
		FaceDeviceInfo.Account = structInfo.Account
		FaceDeviceInfo.Psw = structInfo.Psw
		FaceDeviceInfo.Ip = structInfo.Ip
		FaceDeviceInfo.Checkatt = structInfo.IsAttendance
		if FaceDeviceInfo.Status == 1 {
			//获取扩展信息
			var structDeviceExpendInfo ExtendInfo_Json

			//获取网络信息
			structNetInfo, err := facedevice.GetDeviceNetInfo(structInfo)
			if err != nil {
				beego.Warning("获取设备列表警告", err)
				//return
			}

			structDeviceExpendInfo.IPAddr = structNetInfo.IPAddr
			structDeviceExpendInfo.Submask = structNetInfo.Submask
			structDeviceExpendInfo.Gateway = structNetInfo.Gateway

			//获取开门条件参数
			structDoorConditionInfo, err := facedevice.GetDoorConditionParam(structInfo)
			if err != nil {
				beego.Warning("获取设备列表警告", err)

			}
			structDeviceExpendInfo.ControlType = structDoorConditionInfo.ControlType
			structDeviceExpendInfo.Endian = structDoorConditionInfo.Endian
			structDeviceExpendInfo.IOStayTime = structDoorConditionInfo.IOStayTime
			structDeviceExpendInfo.IOType = structDoorConditionInfo.IOType
			structDeviceExpendInfo.CardMode = structDoorConditionInfo.CardMode
			structDeviceExpendInfo.PublicMjCardNo = structDoorConditionInfo.PublicMjCardNo
			structDeviceExpendInfo.AutoMjCardBgnNo = structDoorConditionInfo.AutoMjCardBgnNo
			structDeviceExpendInfo.AutoMjCardEndNo = structDoorConditionInfo.AutoMjCardEndNo
			structDeviceExpendInfo.Wiegand = structDoorConditionInfo.Wiegand
			structDeviceExpendInfo.IOType = structDoorConditionInfo.IOType
			structDeviceExpendInfo.IOStayTime = structDoorConditionInfo.IOStayTime

			FaceDeviceInfo.Extend = structDeviceExpendInfo
		}
		responseStruct.Data.Item = append(responseStruct.Data.Item, FaceDeviceInfo)
	}

	//雄迈的设备
	deviceInfoArray, err = my_db.GetDeviceListByType(structAdminInfo.CompanyId, 1)
	if err != nil {
		beego.Error("\r\n 获取设备列表失败", err)
		return
	}

	// err = facedevicexiongmai.DeviceSearch()
	// if err != nil {
	// 	beego.Error("添加设备错误", err)
	// 	return
	// }

	// spanindex, err := time.ParseDuration("0.1s")
	// time.Sleep(spanindex)

	for _, structInfo := range deviceInfoArray {
		var FaceDeviceInfo FaceDaviceInfo_Json

		// _, err := facedevicexiongmai.GetDeviceInfoByIP(structInfo.Ip)
		// if err != nil {
		// 	FaceDeviceInfo.Status = 0
		// 	structInfo.Status = 0
		// } else {
		// 	FaceDeviceInfo.Status = 1
		// 	structInfo.Status = 1
		// }

		mapExtend := make(map[string]interface{}, 0)

		FaceDeviceInfo.Ip = structInfo.Ip
		FaceDeviceInfo.Id = structInfo.Id
		FaceDeviceInfo.Model = ""
		FaceDeviceInfo.Name = structInfo.Name
		FaceDeviceInfo.Serial = structInfo.DevcieId
		FaceDeviceInfo.Type = structInfo.Type
		FaceDeviceInfo.Psw = structInfo.Psw
		FaceDeviceInfo.Checkatt = structInfo.IsAttendance
		FaceDeviceInfo.Multi_cert = structInfo.MultiCert

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
				FaceDeviceInfo.Status = 0
			} else {
				FaceDeviceInfo.Status = 1
			}

		} else {
			FaceDeviceInfo.Status = 0
		}

		if FaceDeviceInfo.Status == 1 {
			mapIOResult, err := facedevicexiongmai.GetXMIOCfg(structInfo.DevcieId)
			if err != nil {
				beego.Error(err)
				FaceDeviceInfo.Status = 0
				responseStruct.Data.Item = append(responseStruct.Data.Item, FaceDeviceInfo)
				continue
			}

			parmDoorReplyIndex := (mapIOResult["data"].(map[string]interface{})["door_relay"]).(map[string]interface{})

			//			beego.Debug(reflect.TypeOf(parmDoorReplyIndex["duration"]))

			fDuration := parmDoorReplyIndex["duration"].(float64)
			bEnable := parmDoorReplyIndex["enable"].(bool)
			if bEnable {
				mapExtend["delay"] = int64(fDuration)
			}

			nStatusIndex, err := facedevicexiongmai.GetDoorStatus(structInfo.DevcieId)
			if err != nil {
				beego.Error(err)
			}

			mapExtend["iotype"] = nStatusIndex

			if structInfo.IsShowDoorStatus > 0 {
				FaceDeviceInfo.Isshow_doorstatus = true

				FaceDeviceInfo.Isopen = nStatusIndex

			} else {
				FaceDeviceInfo.Isshow_doorstatus = false
			}
		}
		mapExtend["password"] = structInfo.Psw

		FaceDeviceInfo.Extend = mapExtend
		responseStruct.Data.Item = append(responseStruct.Data.Item, FaceDeviceInfo)

		// err = my_db.UpdateDevice(structInfo)
		// if err != nil {
		// 	beego.Error(err)
		// 	return
		// }
	}

	responseStruct.Code = 20000
	responseStruct.Message = confreader.GetValue("SucRequest")

	responseByte, err := json.Marshal(responseStruct)
	if err != nil {
		return
	}

	c.Data["json"] = json.RawMessage(string(responseByte))
	c.ServeJSON()
}

//deviceOption
type DeviceOptionController struct {
	beego.Controller
}

type DeviceOptionInfo_Json struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

type DeviceOptionResult_Json struct {
	Code    int64                   `json:"code"`
	Message string                  `json:"message"`
	Data    []DeviceOptionInfo_Json `json:"data"`
}

func (c *DeviceOptionController) Get() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	Xtoken := c.Ctx.Input.Header("X-Token")
	strAdmin, err := my_aes.Dncrypt(string(Xtoken))
	if err != nil {
		beego.Error(err)
		return
	}

	structAdminInfo, err := my_db.GetAdminByAccount(strAdmin)
	if err != nil {
		beego.Error(err)
		return
	}

	DeviceArray, err := my_db.GetDeviceList(structAdminInfo.CompanyId)
	if err != nil {
		beego.Error(err)
		return
	}

	var responseStruct DeviceOptionResult_Json
	responseStruct.Data = make([]DeviceOptionInfo_Json, 0)
	for _, deviceInfoIndex := range DeviceArray {
		var structDeviceInfo DeviceOptionInfo_Json
		structDeviceInfo.Id = deviceInfoIndex.Id
		structDeviceInfo.Name = deviceInfoIndex.Name

		responseStruct.Data = append(responseStruct.Data, structDeviceInfo)
	}

	responseStruct.Code = 20000
	responseStruct.Message = confreader.GetValue("SucRequest")

	responseByte, err := json.Marshal(responseStruct)
	if err != nil {
		return
	}

	c.Data["json"] = json.RawMessage(string(responseByte))
	c.ServeJSON()
}

///api/management/device/bind
type AddDeviceController struct {
	beego.Controller
}

//接收的数据
type AddDeviceRequstInfo_Json struct {
	Account           string        `json:"account"`
	Ip                string        `json:"ip"`
	Name              string        `json:"name"`
	Persons           []interface{} `json:"persons"`
	Psw               string        `json:"psw"`
	Serial            string        `json:"serial"`
	Checkatt          int64         `json:"checkatt"`
	Isshow_doorstatus bool          `json:"isshow_doorstatus"`
	Multi_cert        int64         `json:"multi_cert"`
}

//发送的数据
type AddDeviceChildInfo_Json struct {
	Id     int64  `json:"id"`
	Name   string `json:"name"`
	Serial string `json:"serial"`
	Status int64  `json:"status"`
}

type AddDeviceReplyInfo_Json struct {
	Code    int64                   `json:"code"`
	Message string                  `json:"message"`
	Data    AddDeviceChildInfo_Json `json:"data"`
}

//AddDeviceHaiQIn  添加海清的设备
func AddDeviceHaiQIn(c *AddDeviceController) ([]byte, error) {
	data := c.Ctx.Input.RequestBody
	var structRequst AddDeviceRequstInfo_Json
	err := json.Unmarshal(data, &structRequst)
	if err != nil {
		beego.Error("添加设备失败:", err)
		return []byte(""), err
	}

	var structDeviceInfoIndex my_db.DeviceInfo
	structDeviceInfoIndex.Account = "admin"
	structDeviceInfoIndex.Psw = structRequst.Psw
	structDeviceInfoIndex.Ip = structRequst.Ip

	faceDeviceInfo, err := facedevice.GetDeviceInfo(structDeviceInfoIndex)
	if err != nil {
		var structReplyInfo AddDeviceReplyInfo_Json
		structReplyInfo.Code = 40000
		structReplyInfo.Message = "添加设备失败: 设备ip或密码错误"

		byteReply, errindex := json.Marshal(structReplyInfo)
		if errindex != nil {
			return []byte(""), errindex
		}
		return byteReply, err
	}

	err = check_device.DeviceCheck(fmt.Sprintf("%d", faceDeviceInfo.DeviceId))
	if err != nil {
		beego.Error("添加设备错误", err)

		var structReplyInfo AddDeviceReplyInfo_Json
		structReplyInfo.Code = 40000
		structReplyInfo.Message = "添加设备失败: " + err.Error()

		byteReply, errindex := json.Marshal(structReplyInfo)
		if errindex != nil {
			return []byte(""), errindex
		}

		return byteReply, err
	}

	var structDeviceInfo my_db.DeviceInfo
	structDeviceInfo.Ip = structRequst.Ip
	structDeviceInfo.Name = structRequst.Name
	structDeviceInfo.Psw = "admin"
	structDeviceInfo.Account = structRequst.Psw
	structDeviceInfo.DevcieId = fmt.Sprintf("%d", faceDeviceInfo.DeviceId)

	XToken := c.Ctx.Input.Header("X-Token")
	strAdmin, err := my_aes.Dncrypt(string(XToken))
	if err != nil {
		beego.Error("获取规则列表错误", err)
	}

	structAdminInfo, err := my_db.GetAdminByAccount(strAdmin)
	if err != nil {
		beego.Error(err)
		return []byte(""), err
	}

	structDeviceInfo.CompanyID = structAdminInfo.CompanyId
	if err != nil {
		//		beego.Error("解密错误:", err)
		return []byte(""), err
	}
	structDeviceInfo.Type = 0

	//订阅抓拍认证信息
	Host := c.Ctx.Input.Host()
	arrayIndex := strings.Split(string(Host), ":")
	strHostIP := arrayIndex[0]

	err = facedevice.SubscribeWithDevice(structDeviceInfo, strHostIP)
	if err != nil {
		beego.Error("添加设备错误", err)

		var structReplyInfo AddDeviceReplyInfo_Json
		structReplyInfo.Code = 40000
		structReplyInfo.Message = "添加设备失败: 设备ip或序列号错误"

		byteReply, errindex := json.Marshal(structReplyInfo)
		if errindex != nil {
			return []byte(""), errindex
		}

		return byteReply, err
	}

	structDeviceInfo.IsAttendance = structRequst.Checkatt
	if structRequst.Isshow_doorstatus {
		structDeviceInfo.IsShowDoorStatus = 1
	} else {
		structDeviceInfo.IsShowDoorStatus = 0
	}

	nID, err := my_db.AddDevice(structDeviceInfo)
	if err != nil {
		return []byte(""), err
	}

	var structDeviceAndEployeeInfo my_db.DeviceAndPersonInfo
	structDeviceAndEployeeInfo.DeviceId = nID
	for _, personID := range structRequst.Persons {
		strtype := reflect.TypeOf(personID)
		strName := strtype.Name()
		if strName == "float64" {
			structDeviceAndEployeeInfo.EmployeeId = append(structDeviceAndEployeeInfo.EmployeeId, int64(personID.(float64)))
		}
	}
	err = my_db.AddDviceAndEmployee(structDeviceAndEployeeInfo)
	if err != nil {
		return []byte(""), err
	}

	var structGroupDeviceInfo my_db.GroupDeviceInfo
	structGroupDeviceInfo.GroupName = "default"
	structGroupDeviceInfo.GroupID = 1
	structGroupDeviceInfo.ID = nID
	structGroupDeviceInfo.Name = structRequst.Name
	_, err = my_db.AddGroupDevice(structGroupDeviceInfo)
	if err != nil {
		beego.Error(err)
		return []byte(""), err
	}

	var structReplyInfo AddDeviceReplyInfo_Json
	structReplyInfo.Code = 20000
	structReplyInfo.Message = confreader.GetValue("SucRequest")

	var structChildDeviceInfo AddDeviceChildInfo_Json
	structChildDeviceInfo.Name = structDeviceInfo.Name
	structChildDeviceInfo.Id = nID
	structChildDeviceInfo.Serial = structDeviceInfo.DevcieId
	structChildDeviceInfo.Status = 1

	structReplyInfo.Data = structChildDeviceInfo

	byteReply, errindex := json.Marshal(structReplyInfo)
	if errindex != nil {
		return []byte(""), errindex
	}
	return byteReply, nil
}

//AddDeviceXiongmai 添加雄迈的设备
func AddDeviceXiongmai(strPsw string, strSerial string, c *AddDeviceController) ([]byte, error) {
	//Xtoken := c.Ctx.Input.Header("X-Token")

	data := c.Ctx.Input.RequestBody
	var structRequst AddDeviceRequstInfo_Json
	err := json.Unmarshal(data, &structRequst)
	if err != nil {
		beego.Error("添加设备失败:", err)
		return []byte(""), err
	}

	err = check_device.DeviceCheck(strSerial)
	if err != nil {
		beego.Error("添加设备错误", err)

		var structReplyInfo AddDeviceReplyInfo_Json
		structReplyInfo.Code = 40000
		structReplyInfo.Message = "添加设备失败: " + err.Error()

		byteReply, errindex := json.Marshal(structReplyInfo)
		if errindex != nil {
			return []byte(""), errindex
		}

		return byteReply, err
	}

	var structDeviceInfo my_db.DeviceInfo
	structDeviceInfo.Name = structRequst.Name
	structDeviceInfo.DevcieId = strSerial
	structDeviceInfo.Ip = structRequst.Ip
	structDeviceInfo.Type = 1
	structDeviceInfo.Psw = structRequst.Psw
	XToken := c.Ctx.Input.Header("X-Token")
	strAdmin, err := my_aes.Dncrypt(string(XToken))
	if err != nil {
		beego.Error("获取规则列表错误", err)
	}

	structAdminInfo, err := my_db.GetAdminByAccount(strAdmin)
	if err != nil {
		beego.Error(err)
		return []byte(""), err
	}

	structDeviceInfo.CompanyID = structAdminInfo.CompanyId
	structDeviceInfo.Status = 1
	structDeviceInfo.IsAttendance = structRequst.Checkatt
	if structRequst.Isshow_doorstatus {
		structDeviceInfo.IsShowDoorStatus = 1
	} else {
		structDeviceInfo.IsShowDoorStatus = 0
	}

	structDeviceInfo.MultiCert = structRequst.Multi_cert
	nCount, err := my_db.AddDevice(structDeviceInfo)
	if err != nil {
		return []byte(""), err
	}

	err = my_db.EnableDevice(strSerial)
	if err != nil {
		return []byte(""), err
	}

	var structGroupDeviceInfo my_db.GroupDeviceInfo
	structGroupDeviceInfo.GroupName = "default"
	structGroupDeviceInfo.GroupID = 1
	structGroupDeviceInfo.ID = nCount
	structGroupDeviceInfo.Name = structRequst.Name
	_, err = my_db.AddGroupDevice(structGroupDeviceInfo)
	if err != nil {
		beego.Error(err)
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

		var structReplyInfo AddDeviceReplyInfo_Json
		structReplyInfo.Code = 40000
		structReplyInfo.Message = "添加设备失败" + err.Error()
		byteReply, errindex := json.Marshal(structReplyInfo)
		if err != nil {
			return []byte(""), errindex
		}
		return byteReply, err
	}

	// structDeviceInfo, err := my_db.GetDeviceByDeviceId(strSerial)
	// if err != nil {
	// 	return []byte(""), err
	// }

	var structReplyInfo AddDeviceReplyInfo_Json
	structReplyInfo.Code = 20000
	structReplyInfo.Message = confreader.GetValue("SucRequest")

	var structChildDeviceInfo AddDeviceChildInfo_Json
	structChildDeviceInfo.Name = structDeviceInfo.Name
	structChildDeviceInfo.Id = nCount
	structChildDeviceInfo.Serial = structDeviceInfo.DevcieId

	structReplyInfo.Data = structChildDeviceInfo

	byteReply, errindex := json.Marshal(structReplyInfo)
	if errindex != nil {
		return []byte(""), errindex
	}
	return byteReply, nil
}

//Post AddDeviceController Post请求
func (c *AddDeviceController) Post() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	data := c.Ctx.Input.RequestBody
	var structRequst AddDeviceRequstInfo_Json
	err := json.Unmarshal(data, &structRequst)
	if err != nil {
		beego.Error("添加设备失败:", err)
		return
	}

	//判断是否是雄迈的设备
	var structReply AddDeviceReplyInfo_Json

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

	mapDeviceInfo, err := facedevicexiongmai.GetDeviceInfoByIP(structRequst.Ip)
	if err == nil {
		nDeviceType = 2
		strEntrilDeviceID = mapDeviceInfo["serial"]
		strXMPSW = mapDeviceInfo["psw"]
	} else {
		for i := 0; i < 10; i++ {
			spanindex, err := time.ParseDuration("0.1s")
			time.Sleep(spanindex)
			mapDeviceInfo, err = facedevicexiongmai.GetDeviceInfoByIP(structRequst.Ip)
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
	}

	//判读是否是海清的设备
	var structDeviceInfoIndex my_db.DeviceInfo
	structDeviceInfoIndex.Account = "admin"
	structDeviceInfoIndex.Psw = "admin"
	structDeviceInfoIndex.Ip = structRequst.Ip

	structHaiQinDeviceInfo, err := facedevice.GetDeviceInfo(structDeviceInfoIndex)
	if err == nil {
		nDeviceType = 1
		strEntrilDeviceID = fmt.Sprintf("%d", structHaiQinDeviceInfo.DeviceId)
	}

	structDeviceData, err := my_db.GetDeviceByDeviceId(strEntrilDeviceID)
	if structDeviceData.DevcieId == strEntrilDeviceID && len(strEntrilDeviceID) > 0 {
		structCompnyInfoIndex, err := my_db.GetCompnyById(structDeviceData.CompanyID)
		if err != nil {
			beego.Error(err)
			return
		}

		var structReplyInfo AddDeviceReplyInfo_Json
		structReplyInfo.Code = 40000
		structReplyInfo.Message = confreader.GetValue("AlreadyBind") + structCompnyInfoIndex.Name

		byteReply, err := json.Marshal(structReplyInfo)
		if err != nil {
			beego.Error(err)
			return
		}

		beego.Error(structReplyInfo.Message)

		c.Data["json"] = json.RawMessage(byteReply)
		c.ServeJSON()
		return
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
		if strXMPSW != structRequst.Psw {
			mapResult := make(map[string]interface{}, 0)
			mapResult["code"] = 50000
			mapResult["message"] = confreader.GetValue("ErrIpOrPWD")
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

	structReply.Code = 40000
	structReply.Message = confreader.GetValue("NoDeviceFind")

	byteReply, err := json.Marshal(structReply)
	if err != nil {
		beego.Error("添加设备错误", err)
		return
	}

	c.Data["json"] = json.RawMessage(string(byteReply))
	c.ServeJSON()
}

// /api/management/device/unbind 删除设备
type RemoveDeviceController struct {
	beego.Controller
}

type RemoveDeviceExtendInfo_Json struct {
	LcdBLDisable         int64  `json:"LcdBLDisable"`
	LcdBLDisableAfterSec int64  `json:"LcdBLDisableAfterSec"`
	LedBrightness        int64  `json:"LedBrightness"`
	LedDisableAfterSec   int64  `json:"LedDisableAfterSec"`
	LedLightType         int64  `json:"LedLightType"`
	LedTimeBeg           string `json:"LedTimeBeg"`
	LedTimeEnd           string `json:"LedTimeEnd"`
	LiveFrameNum         int64  `json:"LiveFrameNum"`
	LiveThreshold        int64  `json:"LiveThreshold"`
}

//接收到的数据
type RemoveDeviceRequstInfo_Json struct {
	Id     int64                       `json:"id"`
	Model  string                      `json:"model"`
	Name   string                      `json:"name"`
	Serial string                      `json:"serial"`
	Status int64                       `json:"status"`
	Type   int                         `json:"type"`
	Extend RemoveDeviceExtendInfo_Json `json:"extend"`
}

//返回数据
type RemoveReplyInfo_Json struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

func (c *RemoveDeviceController) Post() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	data := c.Ctx.Input.RequestBody
	var structRequstInfo RemoveDeviceRequstInfo_Json
	err := json.Unmarshal(data, &structRequstInfo)
	if err != nil {
		beego.Error("删除设备出错:", err)
		return
	}

	structDeviceInfo, err := my_db.GetDeviceById(structRequstInfo.Id)
	if err != nil {
		beego.Error("删除设备出错", err)
		return
	}

	if structRequstInfo.Id <= 0 {
		beego.Error("错误的设备ID")
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

		err = my_db.RemoveDevice(structRequstInfo.Serial)
		if err != nil {
			beego.Error("删除设备出错:", err)
			return
		}
	} else {
		beego.Error("删除设备错误， 错误的设备类型")
		return
	}

	err = my_db.RemoveDeviceAndEmployee(structRequstInfo.Id)
	if err != nil {
		beego.Error("删除设备出错", err)
		return
	}

	err = my_db.RemoveGroupDevice(structRequstInfo.Id)
	if err != nil {
		beego.Error(err)
		return
	}

	var structReply RemoveReplyInfo_Json
	structReply.Code = 20000
	structReply.Data = ""
	structReply.Message = confreader.GetValue("SucRequest")

	byteReply, err := json.Marshal(structReply)
	if err != nil {
		beego.Error("删除设备出错", err)
		return
	}

	c.Data["json"] = json.RawMessage(string(byteReply))
	c.ServeJSON()
}

///api/management/device/selectedPersons
type DeviceSelectEmployeeController struct {
	beego.Controller
}

type DeviceSelectEmployeeReply_Json struct {
	Code    int64   `json:"code"`
	Message string  `json:"message"`
	Data    []int64 `json:"data"`
}

func (c *DeviceSelectEmployeeController) Get() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	DeviceId := c.Input().Get("device_id")

	nDeviceId, err := strconv.ParseInt(DeviceId, 0, 64)
	if err != nil {
		beego.Error("获取设备关联人员错误", err)
		return
	}

	deviceEmployeeInfo, err := my_db.GetDeviceAndEmployee(nDeviceId)
	if err != nil {
		beego.Error("获取设备关联人员错误", err)
		return
	}

	var structReplyInfo DeviceSelectEmployeeReply_Json
	structReplyInfo.Data = make([]int64, 0)
	for _, employeeIndex := range deviceEmployeeInfo.EmployeeId {
		structReplyInfo.Data = append(structReplyInfo.Data, employeeIndex)
	}

	structReplyInfo.Code = 20000
	structReplyInfo.Message = confreader.GetValue("SucRequest")
	byteReply, err := json.Marshal(structReplyInfo)
	if err != nil {
		beego.Error("获取设备关联人员错误:", err)
		return
	}
	c.Data["json"] = json.RawMessage(string(byteReply))
	c.ServeJSON()
}

///api/management/device/update
type DeviceUpdateController struct {
	beego.Controller
}

//接收数据
type DeviceUpdateExtendInfo_Json struct {
	LedBrightness        int64       `json:"LedBrightness"`
	LedDisableAfterSec   int64       `json:"LedDisableAfterSec"`
	LcdBLDisable         int64       `json:"LcdBLDisable"`
	LcdBLDisableAfterSec int64       `json:"LcdBLDisableAfterSec"`
	Endian               int64       `json:"Endian"`
	ControlType          int64       `json:"ControlType"`
	Wiegand              int64       `json:"Wiegand"`
	PublicMjCardNo       int64       `json:"PublicMjCardNo"`
	AutoMjCardBgnNo      int64       `json:"AutoMjCardBgnNo"`
	AutoMjCardEndNo      int64       `json:"AutoMjCardEndNo"`
	IOStayTime           int64       `json:"IOStayTime"`
	IOType               int64       `json:"IOType"`
	AutoRebootDay        int64       `json:"AutoRebootDay"`
	LiveThreshold        int64       `json:"LiveThreshold"`
	LiveFrameNum         int64       `json:"LiveFrameNum"`
	LedLightType         int64       `json:"LedLightType"`
	IPAddr               string      `json:"IPAddr"`
	Submask              string      `json:"Submask"`
	Gateway              string      `json:"Gateway"`
	LedTimeBeg           string      `json:"LedTimeBeg"`
	LedTimeEnd           string      `json:"LedTimeEnd"`
	CardMode             int64       `json:"cardMode"`
	FaceThreshold        int64       `json:"FaceThreshold"`
	IDCardThreshold      int64       `json:"IDCardThreshold"`
	OpendoorWay          int64       `json:"OpendoorWay"`
	VerifyMode           int64       `json:"VerifyMode"`
	Delay                interface{} `json:"delay"`
	Password             string      `json:"password"`
}

type DeviceUpdateRequst_Json struct {
	Account           string                      `json:"account"`
	Id                int64                       `json:"id"`
	Ip                string                      `json:"ip"`
	Model             string                      `json:"model"`
	Name              string                      `json:"name"`
	Psw               string                      `json:"psw"`
	Serial            string                      `json:"serial"`
	Status            int64                       `json:"status"`
	Type              int                         `json:"type"`
	Persons           []interface{}               `json:"persons"`
	Extend            DeviceUpdateExtendInfo_Json `json:"extend"`
	Checkatt          int64                       `json:"checkatt"`
	Isshow_doorstatus bool                        `json:"isshow_doorstatus"`
	Multi_cert        int64                       `json:"multi_cert"`
}

//发送数据
type DeviceUpdateReply_Json struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

func (c *DeviceUpdateController) Post() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	data := c.Ctx.Input.RequestBody
	var structRequst DeviceUpdateRequst_Json
	err := json.Unmarshal(data, &structRequst)
	if err != nil {
		beego.Error("更新设备信息错误:", err)
		return
	}

	//从设备获取信息
	var structDeviceInfo my_db.DeviceInfo
	structDeviceInfo.Ip = structRequst.Ip
	structDeviceInfo.Name = structRequst.Name
	structDeviceInfo.Id = structRequst.Id
	structDeviceInfo.Psw = structRequst.Psw
	structDeviceInfo.Account = "admin"
	structDeviceInfo.DevcieId = structRequst.Serial
	structDeviceInfo.IsAttendance = structRequst.Checkatt
	if structRequst.Isshow_doorstatus {
		structDeviceInfo.IsShowDoorStatus = 1
	} else {
		structDeviceInfo.IsShowDoorStatus = 0
	}
	structDeviceInfo.MultiCert = structRequst.Multi_cert

	if structRequst.Status > 0 {
		//海清的设备
		if structRequst.Type == 0 {
			//更新开门控制
			structDeviceDoorConditionInfo, err := facedevice.GetDoorConditionParam(structDeviceInfo)
			if err != nil {
				beego.Error("更新设备信息错误", err)
				return
			}

			structDeviceDoorConditionInfo.AutoMjCardEndNo = structRequst.Extend.AutoMjCardEndNo
			structDeviceDoorConditionInfo.CardMode = structRequst.Extend.CardMode
			structDeviceDoorConditionInfo.ControlType = structRequst.Extend.ControlType
			structDeviceDoorConditionInfo.Endian = structRequst.Extend.Endian
			//structDeviceDoorConditionInfo.FaceThreshold = structRequst.Extend.FaceThreshold
			//structDeviceDoorConditionInfo.IDCardThreshold = structRequst.Extend.IDCardThreshold
			structDeviceDoorConditionInfo.IOStayTime = structRequst.Extend.IOStayTime
			structDeviceDoorConditionInfo.IOType = structRequst.Extend.IOType
			structDeviceDoorConditionInfo.OpendoorWay = structRequst.Extend.OpendoorWay
			structDeviceDoorConditionInfo.PublicMjCardNo = structRequst.Extend.PublicMjCardNo
			//structDeviceDoorConditionInfo.VerifyMode = structRequst.Extend.VerifyMode
			structDeviceDoorConditionInfo.Wiegand = structRequst.Extend.Wiegand

			err = facedevice.SetDoorConditionParam(structDeviceInfo, structDeviceDoorConditionInfo)
			if err != nil {
				beego.Error("更新设备信息错误", err)
				return
			}

			//更新网络信息
			var structDeviceNetInfo facedevice.NetParamInfo_Json
			structDeviceNetInfo.Gateway = structRequst.Extend.Gateway
			structDeviceNetInfo.IPAddr = structRequst.Extend.IPAddr
			structDeviceNetInfo.Submask = structRequst.Extend.Submask
			err = facedevice.SetDeviceNetInfo(structDeviceNetInfo, structDeviceInfo)
			if err != nil {
				beego.Error("更新设备信息错误", err)
				return
			}
		}

		//雄迈的设备
		if structRequst.Type == 1 {
			//strPsw := structRequst.Extend.Password
			//strDely := structRequst.Extend.Delay

			mapIOCfgIndex, err := facedevicexiongmai.GetXMIOCfg(structRequst.Serial)
			if err != nil {
				beego.Error(err)
				return
			}

			parmDoorReplyIndex := (mapIOCfgIndex["data"].(map[string]interface{})["door_relay"]).(map[string]interface{})

			//strDelay := structRequst.Extend.Delay

			mapIOParam := mapIOCfgIndex["data"].(map[string]interface{})

			// fDuration := parmDoorReplyIndex["duration"].(float64)
			// bEnable := parmDoorReplyIndex["enable"].(bool)

			// if len(strDelay) > 5 {
			// 	beego.Error("错误的延时数据")
			// 	return
			// }

			// nDelay, err := strconv.ParseInt(strDelay, 0, 64)
			// if err != nil {
			// 	beego.Error(err)
			// 	return
			// }

			typeIndex := reflect.TypeOf(structRequst.Extend.Delay).Kind()
			switch typeIndex {
			case reflect.Float64:
				parmDoorReplyIndex["duration"] = int64(structRequst.Extend.Delay.(float64))
				break
			case reflect.String:
				{
					nDelayIndex, err := strconv.ParseInt(structRequst.Extend.Delay.(string), 0, 64)
					if err != nil {
						beego.Error(err)
						return
					}
					parmDoorReplyIndex["duration"] = nDelayIndex
				}
				break
			}

			parmDoorReplyIndex["enable"] = true
			mapIOParam["door_relay"] = parmDoorReplyIndex

			err = facedevicexiongmai.SetXMIOCfg(structRequst.Serial, mapIOParam)
			if err != nil {
				beego.Error(err)
				return
			}

		}

	}

	//更新数据库
	err = my_db.UpdateDevice(structDeviceInfo)
	if err != nil {
		beego.Error("更新设备信息错误:", err)
		return
	}

	//更新设备人员表
	var structDeviceAndEmployeeInfo my_db.DeviceAndPersonInfo
	structDeviceAndEmployeeInfo.DeviceId = structRequst.Id
	for _, personIndex := range structRequst.Persons {
		strtype := reflect.TypeOf(personIndex)
		strName := strtype.Name()
		if strName == "float64" {
			structDeviceAndEmployeeInfo.EmployeeId = append(structDeviceAndEmployeeInfo.EmployeeId, int64(personIndex.(float64)))
		}
	}
	my_db.UpdateDeviceAndEmployee(structDeviceAndEmployeeInfo)

	var structReply DeviceUpdateReply_Json
	structReply.Code = 20000
	structReply.Message = confreader.GetValue("SucRequest")
	structReply.Data = ""

	byteReply, err := json.Marshal(structReply)
	if err != nil {
		beego.Error("更新设备信息失败", err)
		return
	}

	c.Data["json"] = json.RawMessage(string(byteReply))
	c.ServeJSON()
}

///api/management/device/getExtend
type DeviceGetExtendController struct {
	beego.Controller
}

//返回数据
type DeviceExtendDataInfo_Jons struct {
	LcdBLDisable         int64  `json:"LcdBLDisable"`
	LcdBLDisableAfterSec int64  `json:"LcdBLDisableAfterSec"`
	LedBrightness        int64  `json:"LedBrightness"`
	LedDisableAfterSec   int64  `json:"LedDisableAfterSec"`
	LedLightType         int64  `json:"LedLightType"`
	LedTimeBeg           string `json:"LedTimeBeg"`
	LedTimeEnd           string `json:"LedTimeEnd"`
	LiveFrameNum         int64  `json:"LiveFrameNum"`
	LiveThreshold        int64  `json:"LiveThreshold"`
}

type DeviceExtendReplyInfo_Json struct {
	Code    int64                     `json:"code"`
	Message string                    `json:"message"`
	Data    DeviceExtendDataInfo_Jons `json:"data"`
}

func (c *DeviceGetExtendController) Get() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))
	// Id := c.Input().Get("id")

	// nId, err := strconv.ParseInt(Id, 0, 64)
	// if err != nil {
	// 	beego.Error("获取设备配置信息错误", err)
	// 	return
	// }
	// structDeviceInfo, err := my_db.GetDeviceById(nId)
	// if err != nil {
	// 	beego.Error("获取设备配置信息错误", err)
	// 	return
	// }

	var structReplyInfo DeviceExtendReplyInfo_Json
	structReplyInfo.Code = 20000
	structReplyInfo.Message = confreader.GetValue("NoFunction")

	byteReply, err := json.Marshal(structReplyInfo)
	if err != nil {
		beego.Error("获取设备配置信息错误", err)
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
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

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
	structReply.Message = confreader.GetValue("SucRequest")

	byteReply, err := json.Marshal(structReply)
	if err != nil {
		beego.Error("获取设备版本信息错误", err)
		return
	}
	c.Data["json"] = json.RawMessage(byteReply)
	c.ServeJSON()
}

///api/management/person/personOrgOptions
type PerSonOrgOptionsController struct {
	beego.Controller
}

type OrgPersonInfo_Json struct {
	Id                 int64  `json:"id"`
	Name               string `json:"name"`
	Phone              string `json:"phone"`
	Pic                string `json:"pic"`
	Status             int64  `json:"status"`
	Msg                string `json:"msg"`
	UserId             int64  `json:"user_id"`
	AdminId            int64  `json:"admin_id"`
	Organizations      string `json:"organizations"`
	Attendance_rule_id int64  `json:"attendance_rule_id"`
	CompanyId          int64  `json:"company_id"`
	Sort               int64  `json:"sort"`
	Is_gateman         int64  `json:"is_gateman"`
	Is_default         int64  `json:"is_default"`
}

type ChildOrgOptionInfo_Json struct {
	Id       string               `json:"id"`
	Name     string               `json:"name"`
	Children []OrgPersonInfo_Json `json:"children"`
}

type OrgOptionsInfo_Json struct {
	Id       string        `json:"id"`
	Name     string        `json:"name"`
	Children []interface{} `json:"children"`
}

type PerSonOrgOptionsResult_Json struct {
	Code    int64                 `json:"code"`
	Message string                `json:"message"`
	Data    []OrgOptionsInfo_Json `json:"data"`
}

func (c *PerSonOrgOptionsController) Post() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	//xtoken := c.Ctx.Input.Header("X-Token")
	//beego.Debug(string(xtoken))

	groupArray, err := my_db.GetGroupListByPurview(0)
	if err != nil {
		beego.Error("获取组织列表错误:", err)
		return
	}

	var structRplyInfo PerSonOrgOptionsResult_Json
	structRplyInfo.Data = make([]OrgOptionsInfo_Json, 0)
	for _, groupInfoIndex := range groupArray {
		var structGroupInfo OrgOptionsInfo_Json
		structGroupInfo.Name = groupInfoIndex.Name
		structGroupInfo.Id = groupInfoIndex.Id
		structGroupInfo.Children = make([]interface{}, 0)

		EmployeeIdArray, err := my_db.GetEmployeeByGroup(groupInfoIndex.Id)
		if err != nil {
			beego.Error("获取组织列表错误， 获取组织下的人员列表错误", err)
			return
		}

		for _, employeeIndex := range EmployeeIdArray {
			structEmployeeInfo, err := my_db.GetEmployeeById(employeeIndex)
			if err != nil {
				beego.Error("获取组织列表错误， 获取人员数据错误", err)
				return
			}

			var structPerSonInfo OrgPersonInfo_Json
			structPerSonInfo.Id = structEmployeeInfo.Id
			structPerSonInfo.Name = structEmployeeInfo.Name
			structPerSonInfo.Phone = structEmployeeInfo.Phone
			structPerSonInfo.Pic = structEmployeeInfo.Pic
			structPerSonInfo.Sort = structEmployeeInfo.Sort
			structPerSonInfo.Status = structEmployeeInfo.Status
			structPerSonInfo.Organizations = ""
			structPerSonInfo.UserId = 0
			structPerSonInfo.AdminId = 0
			structPerSonInfo.Attendance_rule_id = 0
			structPerSonInfo.CompanyId = 3
			structPerSonInfo.Is_default = 1
			structPerSonInfo.Is_gateman = 0
			structPerSonInfo.Msg = ""

			structGroupInfo.Children = append(structGroupInfo.Children, structPerSonInfo)
		}

		arrayChildGroupInfo, err := my_db.GetGroupListByParent(groupInfoIndex.Id)
		if err != nil {
			beego.Error("获取组织人员数据错误", err)
			return
		}

		for _, structChildIndex := range arrayChildGroupInfo {
			var structChildGroup ChildOrgOptionInfo_Json
			structChildGroup.Id = structChildIndex.Id
			structChildGroup.Name = structChildIndex.Name
			structChildGroup.Children = make([]OrgPersonInfo_Json, 0)

			arrayEemployeeId, err := my_db.GetEmployeeByGroup(structChildIndex.Id)
			if err != nil {
				beego.Error("获取组织人员错误", err)
				return
			}
			for _, nEmployeeIdIndex := range arrayEemployeeId {
				structEmployeeInfo, err := my_db.GetEmployeeById(nEmployeeIdIndex)
				if err != nil {
					beego.Error("获取组织人员错误", err)
					return
				}

				var structPerSonInfo OrgPersonInfo_Json
				structPerSonInfo.Id = structEmployeeInfo.Id
				structPerSonInfo.Name = structEmployeeInfo.Name
				structPerSonInfo.Phone = structEmployeeInfo.Phone
				structPerSonInfo.Pic = structEmployeeInfo.Pic
				structPerSonInfo.Sort = structEmployeeInfo.Sort
				structPerSonInfo.Status = structEmployeeInfo.Status
				structPerSonInfo.Organizations = ""
				structPerSonInfo.UserId = 0
				structPerSonInfo.AdminId = 0
				structPerSonInfo.Attendance_rule_id = 0
				structPerSonInfo.CompanyId = 3
				structPerSonInfo.Is_default = 1
				structPerSonInfo.Is_gateman = 0
				structPerSonInfo.Msg = ""

				structChildGroup.Children = append(structChildGroup.Children, structPerSonInfo)
			}

			structGroupInfo.Children = append(structGroupInfo.Children, structChildGroup)
		}

		structRplyInfo.Data = append(structRplyInfo.Data, structGroupInfo)
	}

	structRplyInfo.Code = 20000
	structRplyInfo.Message = confreader.GetValue("SucRequest")

	responseByte, err := json.Marshal(structRplyInfo)
	if err != nil {
		return
	}

	c.Data["json"] = json.RawMessage(string(responseByte))
	c.ServeJSON()
}

///api/management/person/create
type AddEmployeeContoller struct {
	beego.Controller
}

//接收数据
type AddEmployeeRequstInfo struct {
	DeviceId       []int64     `json:"devices"`
	Name           string      `json:"name"`
	Organizations  string      `json:"organizations"`
	Phone          string      `json:"phone"`
	Pic            string      `json:"pic"`
	Sort           interface{} `json:"sort"`
	Status         string      `json:"status"`
	Sync           bool        `json:"sync"`
	InCharges      []int64     `json:"in_charges"`
	Rights         []string    `json:"rights"`
	Remote_devices []int64     `json:"remote_devices"`
	Is_admin       bool        `json:"is_admin"`
	Working        bool        `json:"working"`
	Psw            string      `json:"password"`
}

//发送数据
type EmployeeReplyInfo struct {
	Company_id     int64    `json:"company_id"`
	Id             int64    `json:"id"`
	Name           string   `json:"name"`
	Organizations  string   `json:"organizations"`
	Phone          string   `json:"phone"`
	Pic            string   `json:"pic"`
	Sort           int64    `json:"sort"`
	Status         int64    `json:"status"`
	InCharges      []int64  `json:"in_charges"`
	Rights         []string `json:"rights"`
	Remote_devices []int64  `json:"remote_devices"`
	Is_admin       bool     `json:"is_admin"`
	Device         []int64  `json:"devices"`
	Working        bool     `json:"working"`
	Psw            string   `json:"password"`
}

type AddEmployeeReplyInfo struct {
	Code    int64             `json:"code"`
	Message string            `json:"message"`
	Data    EmployeeReplyInfo `json:"data"`
}

func (c *AddEmployeeContoller) Post() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	data := c.Ctx.Input.RequestBody
	var structRequstInfo AddEmployeeRequstInfo
	err := json.Unmarshal(data, &structRequstInfo)
	if err != nil {
		beego.Error("\r\n添加人员， 解析json失败:", err)
		return
	}
	var structEmployee my_db.EmployeeInfo
	structEmployee.Id = -1
	structEmployee.Name = structRequstInfo.Name
	structEmployee.Phone = structRequstInfo.Phone
	structEmployee.Pic = structRequstInfo.Pic

	switch reflect.TypeOf(structRequstInfo.Sort).Kind() {
	case reflect.String:
		{
			nSortIndex, err := strconv.ParseInt(structRequstInfo.Sort.(string), 0, 64)
			if err != nil {
				beego.Error(err)
				return
			}
			structEmployee.Sort = nSortIndex
		}
		break
	case reflect.Float64:
		structEmployee.Sort = int64(structRequstInfo.Sort.(float64))
		break
	}

	structEmployee.Status = 0

	XToken := c.Ctx.Input.Header("X-Token")
	strAdmin, err := my_aes.Dncrypt(string(XToken))
	if err != nil {
		beego.Error("获取规则列表错误", err)
	}

	structRootAdminInfo, err := my_db.GetAdminByAccount(strAdmin)
	if err != nil {
		beego.Error(err)
		return
	}

	if structRequstInfo.Is_admin {
		structEmployee.IsAdmin = 1

		var structAdminInfo my_db.AdminInfo
		structAdminInfo.Account = structRequstInfo.Phone
		structAdminInfo.Psw = structRequstInfo.Psw
		structAdminInfo.Account = structRequstInfo.Phone
		structAdminInfo.CompanyId = structRootAdminInfo.CompanyId
		structAdminInfo.ISRoot = 0
		structAdminInfo.Companys = fmt.Sprintf("%d", structRootAdminInfo.CompanyId)

		err = my_db.AddAdminUser(structAdminInfo)
		if err != nil {
			beego.Error(err)

			mapResult := make(map[string]interface{}, 0)
			mapResult["code"] = 50000
			mapResult["message"] = err.Error()
			byteResult, err := json.Marshal(mapResult)
			if err != nil {
				beego.Error(err)
				return
			}
			c.Data["json"] = json.RawMessage(byteResult)
			c.ServeJSON()

			return
		}
		structEmployee.AdminAccount = structRequstInfo.Phone

	} else {
		structEmployee.IsAdmin = 0
	}

	if structRequstInfo.Working {
		structEmployee.Working = 1
	} else {
		structEmployee.Working = 0
	}

	structEmployee.CompanyID = structRootAdminInfo.CompanyId

	for nIndex, nRgiht := range structRequstInfo.Rights {
		if nIndex == 0 {
			structEmployee.Rights += nRgiht
		} else {
			structEmployee.Rights += ","
			structEmployee.Rights += nRgiht
		}
	}

	for nIndex, nRemoteDevice := range structRequstInfo.Remote_devices {
		if nIndex == 0 {
			structEmployee.RemoteDevice = fmt.Sprintf("%d", nRemoteDevice)
		} else {
			structEmployee.RemoteDevice += ","
			structEmployee.RemoteDevice += fmt.Sprintf("%d", nRemoteDevice)
		}
	}

	for nIndex, nIncharge := range structRequstInfo.InCharges {
		if nIndex == 0 {
			structEmployee.Incharges += fmt.Sprintf("%d", nIncharge)
		} else {
			structEmployee.Incharges += ","
			structEmployee.Incharges += fmt.Sprintf("%d", nIncharge)
		}
	}

	structEmployee.RuleID = 0
	nId, err := my_db.AddEmployee(structEmployee)
	if err != nil {
		beego.Error("\r\n添加员工失败:", err)
		return
	}

	structEmployee.Id = nId
	//关联设备人员表
	var structEmployeeAndDevice my_db.EmployeeAndDeviceInfo
	structEmployeeAndDevice.DevcieId = structRequstInfo.DeviceId
	structEmployeeAndDevice.EmployeeId = nId
	err = my_db.AddEmployeeAndDevice(structEmployeeAndDevice)
	if err != nil {
		beego.Error("添加人员错误，关联设备人员表错误:", err)
		return
	}

	//关联组织人员表
	strOrgArray := strings.Split(structRequstInfo.Organizations, ",")
	for _, strOrgIndex := range strOrgArray {
		arrayChild, err := my_db.GetGroupListByParent("org-" + strOrgIndex)
		if err != nil {
			beego.Error("添加人员错误", err)
			return
		}

		//如果该组有子组 则无法关联到用户组织表
		if len(arrayChild) > 0 {
			continue
		}

		var structEmployeeAndGroup my_db.GroupAndEmployee
		structEmployeeAndGroup.EmployeeId = nId
		structEmployeeAndGroup.GroupId = "org-" + strOrgIndex
		err = my_db.AddGroupAndEmployee(structEmployeeAndGroup)
		if err != nil {
			beego.Error("添加人员错误:", err)
			return
		}
	}

	var structChildInfo EmployeeReplyInfo
	structChildInfo.Id = nId
	structChildInfo.Name = structEmployee.Name
	structChildInfo.Organizations = structRequstInfo.Organizations
	structChildInfo.Phone = structEmployee.Phone
	structChildInfo.Pic = structEmployee.Pic
	structChildInfo.Sort = structEmployee.Sort
	structChildInfo.Status = structEmployee.Status

	structChildInfo.Remote_devices = make([]int64, 0)
	structChildInfo.Rights = make([]string, 0)
	structChildInfo.InCharges = make([]int64, 0)
	structChildInfo.Device = make([]int64, 0)

	if len(structRequstInfo.Remote_devices) > 0 {
		structChildInfo.Remote_devices = structRequstInfo.Remote_devices
	}

	if len(structRequstInfo.Rights) > 0 {
		structChildInfo.Rights = structRequstInfo.Rights
	}

	if len(structRequstInfo.InCharges) > 0 {
		structChildInfo.InCharges = structRequstInfo.InCharges
	}

	if len(structRequstInfo.DeviceId) > 0 {
		structChildInfo.Device = structRequstInfo.DeviceId
	}

	structChildInfo.Is_admin = structRequstInfo.Is_admin
	structChildInfo.Working = structRequstInfo.Working
	structChildInfo.Psw = structRequstInfo.Psw

	var structReply AddEmployeeReplyInfo
	structReply.Code = 20000
	structReply.Message = confreader.GetValue("SucRequest")
	structReply.Data = structChildInfo

	byteReply, err := json.Marshal(structReply)
	if err != nil {
		beego.Error("\r\n添加员工失败， 打包json失败:", err)
		return
	}

	c.Data["json"] = json.RawMessage(string(byteReply))
	c.ServeJSON()
}

///api/management/person/sync
type SyncEmployeeController struct {
	beego.Controller
}

//接收数据
type SyncEmployeeRequstInfo_Json struct {
	Id      int64   `json:"id"`
	Name    string  `json:"name"`
	Next    int64   `json:"next"`
	Persons []int64 `json:"persons"`
}

//发送数据
type SyncEmployeeDataReply_Json struct {
	Next  int64 `json:"next"`
	Res   bool  `json:"res"`
	Total int64 `json:"total"`
}

type SyncEmployeeReplyInfo_Json struct {
	Code    int64                      `json:"code"`
	Message string                     `json:"message"`
	Data    SyncEmployeeDataReply_Json `json:"data"`
}

func (c *SyncEmployeeController) Post() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	data := c.Ctx.Input.RequestBody
	var structRequst SyncEmployeeRequstInfo_Json
	err := json.Unmarshal(data, &structRequst)
	if err != nil {
		beego.Error("同步数据到设备错误", err)
		return
	}

	XToken := c.Ctx.Input.Header("X-Token")
	strAdmin, err := my_aes.Dncrypt(string(XToken))
	if err != nil {
		beego.Error("获取规则列表错误", err)
	}

	structAdminInfo, err := my_db.GetAdminByAccount(strAdmin)
	if err != nil {
		beego.Error(err)
		return
	}

	structDeviceInfo, err := my_db.GetDeviceById(structRequst.Id)
	if err != nil {
		beego.Error(err)
		return
	}

	bSyncSuc := true
	strMessage := ""

	var structReplyInfo SyncEmployeeReplyInfo_Json

	if len(structRequst.Persons) > 0 {
		for _, nPersonIDIndex := range structRequst.Persons {
			structEmployeeInfoIndex, err := my_db.GetEmployeeById(nPersonIDIndex)
			if err != nil {
				beego.Error("同步数据到设备错误", err)
				return
			}

			// bExist, err := my_db.ISEmployeeAndDevice(nPersonIDIndex, structRequst.Id)
			// if err != nil {
			// 	beego.Error(err)
			// 	return
			// }

			// if !bExist {
			// 	mapReply := make(map[string]interface{}, 0)
			// 	mapReply["code"] = 20000
			// 	mapReply["message"] = "不是该用户的权限设备"
			// 	byteReply, err := json.Marshal(mapReply)
			// 	if err != nil {
			// 		beego.Error(err)
			// 		return
			// 	}

			// 	c.Data["json"] = json.RawMessage(byteReply)
			// 	c.ServeJSON()

			// 	return
			// }

			//arrayEmployeeAndDevice, err := my_db.GetEmployeeAndDvice(nPersonIDIndex)
			nSyncOpt, err := my_db.GetSycOptByContent(structRequst.Id, nPersonIDIndex)
			if err != nil {
				beego.Error(err)
				return
			}

			err = my_db.RemoveOptByContent(structRequst.Id, nPersonIDIndex)
			if err != nil {
				beego.Error(err)
				return
			}

			if nSyncOpt == 1 || nSyncOpt == 0 {
				//添加
				//海清设备
				if structDeviceInfo.Type == 0 {
					err := facedevice.AddDeviceEmployee(structDeviceInfo, structEmployeeInfoIndex)
					if err != nil {
						strMessage = fmt.Sprintf("%s", err)
						beego.Warning(err)
					}
				}

				if structDeviceInfo.Type == 1 {
					var structEmployeeXM facedevicexiongmai.AddEmployeeInfoJSON
					structEmployeeXM.EmployeInfo.UserID = fmt.Sprintf("%d", structEmployeeInfoIndex.Id)
					structEmployeeXM.EmployeInfo.CardID = ""
					if structDeviceInfo.MultiCert > 0 {
						structEmployeeXM.EmployeInfo.GroupID = "Forbid"
					} else {
						structEmployeeXM.EmployeInfo.GroupID = "default"
					}

					structEmployeeXM.EmployeInfo.Name = structEmployeeInfoIndex.Name
					structEmployeeXM.EmployeInfo.StartTime = time.Now().Format(rule_algorithm.TIME_LAYOUT)
					spanindex, err := time.ParseDuration("8760h")
					if err != nil {
						beego.Error(err)
						return
					}
					structEmployeeXM.EmployeInfo.EndTime = time.Now().Add(spanindex).Format(rule_algorithm.TIME_LAYOUT)
					arrayImg := strings.Split(structEmployeeInfoIndex.Pic, ",")
					byteDecode, err := base64.StdEncoding.DecodeString(arrayImg[1])
					if err != nil {
						beego.Error(err)
						return
					}

					var imageIndex image.Image
					imageIndex, _, err = image.Decode(bytes.NewBuffer(byteDecode))
					if err != nil {
						beego.Error(err)
						imageIndex, err = png.Decode(bytes.NewBuffer(byteDecode))
						if err != nil {
							beego.Error(err)
							bSyncSuc = false
							strMessage = err.Error() + "UserID:" + structEmployeeXM.EmployeInfo.UserID + ", name:" + structEmployeeXM.EmployeInfo.Name
							break
						}

					}

					imageIndex = resize.Resize(300, 400, imageIndex, resize.Lanczos3)

					var buf bytes.Buffer
					err = jpeg.Encode(&buf, imageIndex, nil)
					if err != nil {
						beego.Error(err)
						continue
					}
					//jpeg.Encode()

					structEmployeeXM.EmployeInfo.FaceFeature = strings.ToUpper(hex.EncodeToString(buf.Bytes()))

					structEmployeeXM.DeviceInfo.DeviceID = structDeviceInfo.DevcieId

					err = facedevicexiongmai.AddEmployeeWSS(structEmployeeXM)
					if err != nil {
						strMessage = fmt.Sprintf("%s", err)
						beego.Warning("删除设备人员失败", err)
					}
				}

			} else if nSyncOpt == -1 {
				//删除
				if structDeviceInfo.Type == 0 {
					err = facedevice.RemoveDeviceEmployee(structDeviceInfo, structEmployeeInfoIndex)
					if err != nil {
						beego.Warning("删除设备人员失败", err)
						structEmployeeInfoIndex.Status = 0
						strMessage = fmt.Sprintf("%s", err)
						bSyncSuc = false
					}
				}

				if structDeviceInfo.Type == 1 {
					var structXMRemoveInfo facedevicexiongmai.RemoveEmployeeDataJSON
					structXMRemoveInfo.EmployeeInfo.UserID = fmt.Sprintf("%d", structEmployeeInfoIndex.Id)
					structXMRemoveInfo.DeviceInfo.DeviceID = structDeviceInfo.DevcieId
					err = facedevicexiongmai.RemoveEmployeeWSS(structXMRemoveInfo)
					if err != nil {
						beego.Warning("删除人员错误", err)
						structEmployeeInfoIndex.Status = 0
						strMessage = fmt.Sprintf("%s", err)
						bSyncSuc = false
					}
				}

			} else if nSyncOpt == 2 {
				//更新
				//删除
				if structDeviceInfo.Type == 0 {
					facedevice.RemoveDeviceEmployee(structDeviceInfo, structEmployeeInfoIndex)

					err := facedevice.AddDeviceEmployee(structDeviceInfo, structEmployeeInfoIndex)
					if err != nil {
						strMessage = fmt.Sprintf("%s", err)
						beego.Warning(err)
					}
				}

				if structDeviceInfo.Type == 1 {
					var structXMRemoveInfo facedevicexiongmai.RemoveEmployeeDataJSON
					structXMRemoveInfo.EmployeeInfo.UserID = fmt.Sprintf("%d", structEmployeeInfoIndex.Id)
					structXMRemoveInfo.DeviceInfo.DeviceID = structDeviceInfo.DevcieId
					facedevicexiongmai.RemoveEmployeeWSS(structXMRemoveInfo)

					var structEmployeeXM facedevicexiongmai.AddEmployeeInfoJSON
					structEmployeeXM.EmployeInfo.UserID = fmt.Sprintf("%d", structEmployeeInfoIndex.Id)
					structEmployeeXM.EmployeInfo.CardID = ""
					if structDeviceInfo.MultiCert > 0 {
						structEmployeeXM.EmployeInfo.GroupID = "Forbid"
					} else {
						structEmployeeXM.EmployeInfo.GroupID = "default"
					}

					structEmployeeXM.EmployeInfo.Name = structEmployeeInfoIndex.Name
					structEmployeeXM.EmployeInfo.StartTime = time.Now().Format(rule_algorithm.TIME_LAYOUT)
					spanindex, err := time.ParseDuration("8760h")
					if err != nil {
						beego.Error(err)
						return
					}
					structEmployeeXM.EmployeInfo.EndTime = time.Now().Add(spanindex).Format(rule_algorithm.TIME_LAYOUT)
					arrayImg := strings.Split(structEmployeeInfoIndex.Pic, ",")
					byteDecode, err := base64.StdEncoding.DecodeString(arrayImg[1])
					if err != nil {
						beego.Error(err)
						return
					}

					var imageIndex image.Image
					imageIndex, _, err = image.Decode(bytes.NewBuffer(byteDecode))
					if err != nil {
						beego.Error(err)
						imageIndex, err = png.Decode(bytes.NewBuffer(byteDecode))
						if err != nil {
							beego.Error(err)
							bSyncSuc = false
							strMessage = err.Error() + "UserID:" + structEmployeeXM.EmployeInfo.UserID + ", name:" + structEmployeeXM.EmployeInfo.Name
							break
						}

					}

					imageIndex = resize.Resize(300, 400, imageIndex, resize.Lanczos3)

					var buf bytes.Buffer
					err = jpeg.Encode(&buf, imageIndex, nil)
					if err != nil {
						beego.Error(err)
						continue
					}
					//jpeg.Encode()

					structEmployeeXM.EmployeInfo.FaceFeature = strings.ToUpper(hex.EncodeToString(buf.Bytes()))

					structEmployeeXM.DeviceInfo.DeviceID = structDeviceInfo.DevcieId

					err = facedevicexiongmai.AddEmployeeWSS(structEmployeeXM)
					if err != nil {
						strMessage = fmt.Sprintf("%s", err)
						beego.Warning("删除设备人员失败", err)
					}
				}

			}

			err = my_db.UpdateEmployee(structEmployeeInfoIndex)
			if err != nil {
				beego.Error("同步数据错误", err)
				return

			}
		}
	} else {
		arrayPersons, err := my_db.GetEmployeeList(structAdminInfo.CompanyId, "")
		if err != nil {
			beego.Error("同步数据错误", err)
			return
		}

		for _, structEmployeeInfoIndex := range arrayPersons {
			if structEmployeeInfoIndex.Status == 0 {
				structEmployeeInfoIndex.Status = 1

				arrayEmployeeAndDevice, err := my_db.GetEmployeeAndDvice(structEmployeeInfoIndex.Id)
				if err != nil {
					beego.Error(err)
					return
				}

				//	mapDevice := make(map[string]my_db.DeviceInfo, 0)
				for _, nDeviceIDIndex := range arrayEmployeeAndDevice.DevcieId {
					nOptIndex, err := my_db.GetSycOptByContent(nDeviceIDIndex, structEmployeeInfoIndex.Id)
					if err != nil {
						beego.Error(err)
						return
					}

					structDeviceInfo, err = my_db.GetDeviceById(nDeviceIDIndex)
					if err != nil {
						beego.Error(err)
						return
					}

					err = my_db.RemoveOptByContent(nDeviceIDIndex, structEmployeeInfoIndex.Id)
					if err != nil {
						beego.Error(err)
						return
					}

					if nOptIndex == 1 || nOptIndex == 0 {
						//添加
						//海清设备
						if structDeviceInfo.Type == 0 {
							err := facedevice.AddDeviceEmployee(structDeviceInfo, structEmployeeInfoIndex)
							if err != nil {
								strMessage = fmt.Sprintf("%s", err)
								beego.Warning("添加设备人员失败:", err)
							}
						}

						if structDeviceInfo.Type == 1 {
							var structEmployeeXM facedevicexiongmai.AddEmployeeInfoJSON
							structEmployeeXM.EmployeInfo.UserID = fmt.Sprintf("%d", structEmployeeInfoIndex.Id)
							structEmployeeXM.EmployeInfo.CardID = ""
							if structDeviceInfo.MultiCert > 0 {
								structEmployeeXM.EmployeInfo.GroupID = "Forbid"
							} else {
								structEmployeeXM.EmployeInfo.GroupID = "default"
							}

							structEmployeeXM.EmployeInfo.Name = structEmployeeInfoIndex.Name
							structEmployeeXM.EmployeInfo.StartTime = time.Now().Format(rule_algorithm.TIME_LAYOUT)
							spanindex, err := time.ParseDuration("8760h")
							if err != nil {
								beego.Error(err)
								return
							}
							structEmployeeXM.EmployeInfo.EndTime = time.Now().Add(spanindex).Format(rule_algorithm.TIME_LAYOUT)
							arrayImg := strings.Split(structEmployeeInfoIndex.Pic, ",")
							byteDecode, err := base64.StdEncoding.DecodeString(arrayImg[1])
							if err != nil {
								beego.Error(err)
								return
							}

							var imageIndex image.Image
							imageIndex, _, err = image.Decode(bytes.NewBuffer(byteDecode))
							if err != nil {
								beego.Error(err)
								imageIndex, err = png.Decode(bytes.NewBuffer(byteDecode))
								if err != nil {
									beego.Error(err)
									bSyncSuc = false
									strMessage = err.Error() + "UserID:" + structEmployeeXM.EmployeInfo.UserID + ", name:" + structEmployeeXM.EmployeInfo.Name
									break
								}

							}

							imageIndex = resize.Resize(300, 400, imageIndex, resize.Lanczos3)

							var buf bytes.Buffer
							err = jpeg.Encode(&buf, imageIndex, nil)
							if err != nil {
								beego.Error(err)
								continue
							}
							//jpeg.Encode()

							structEmployeeXM.EmployeInfo.FaceFeature = strings.ToUpper(hex.EncodeToString(buf.Bytes()))

							structEmployeeXM.DeviceInfo.DeviceID = structDeviceInfo.DevcieId

							err = facedevicexiongmai.AddEmployeeWSS(structEmployeeXM)
							if err != nil {
								strMessage = fmt.Sprintf("%s", err)
								beego.Warning("添加设备人员失败:", err, "人员姓名:", structEmployeeXM.EmployeInfo.Name, "人员ID：", structEmployeeXM.EmployeInfo.UserID)
							} else {
								beego.Debug("录入成功,人员姓名：", structEmployeeXM.EmployeInfo.Name, "人员ID：", structEmployeeXM.EmployeInfo.UserID)
							}

						}

					} else if nOptIndex == -1 {
						//删除
						if structDeviceInfo.Type == 0 {
							err = facedevice.RemoveDeviceEmployee(structDeviceInfo, structEmployeeInfoIndex)
							if err != nil {
								strMessage = fmt.Sprintf("%s", err)
								beego.Warning("删除设备人员失败", err)
								structEmployeeInfoIndex.Status = 0
								bSyncSuc = false
							}
						}

						if structDeviceInfo.Type == 1 {
							var structXMRemoveInfo facedevicexiongmai.RemoveEmployeeDataJSON
							structXMRemoveInfo.EmployeeInfo.UserID = fmt.Sprintf("%d", structEmployeeInfoIndex.Id)
							structXMRemoveInfo.DeviceInfo.DeviceID = structDeviceInfo.DevcieId
							err = facedevicexiongmai.RemoveEmployeeWSS(structXMRemoveInfo)
							if err != nil {
								strMessage = fmt.Sprintf("%s", err)
								beego.Warning("删除人员错误", err)
								structEmployeeInfoIndex.Status = 0
								bSyncSuc = false
							}
						}

					} else if nOptIndex == 2 {
						//更新
						//删除
						if structDeviceInfo.Type == 0 {
							facedevice.RemoveDeviceEmployee(structDeviceInfo, structEmployeeInfoIndex)

							err := facedevice.AddDeviceEmployee(structDeviceInfo, structEmployeeInfoIndex)
							if err != nil {
								strMessage = fmt.Sprintf("%s", err)
								beego.Warning(err)
							}
						}

						if structDeviceInfo.Type == 1 {
							var structXMRemoveInfo facedevicexiongmai.RemoveEmployeeDataJSON
							structXMRemoveInfo.EmployeeInfo.UserID = fmt.Sprintf("%d", structEmployeeInfoIndex.Id)
							structXMRemoveInfo.DeviceInfo.DeviceID = structDeviceInfo.DevcieId
							facedevicexiongmai.RemoveEmployeeWSS(structXMRemoveInfo)

							var structEmployeeXM facedevicexiongmai.AddEmployeeInfoJSON
							structEmployeeXM.EmployeInfo.UserID = fmt.Sprintf("%d", structEmployeeInfoIndex.Id)
							structEmployeeXM.EmployeInfo.CardID = ""
							if structDeviceInfo.MultiCert > 0 {
								structEmployeeXM.EmployeInfo.GroupID = "Forbid"
							} else {
								structEmployeeXM.EmployeInfo.GroupID = "default"
							}

							structEmployeeXM.EmployeInfo.Name = structEmployeeInfoIndex.Name
							structEmployeeXM.EmployeInfo.StartTime = time.Now().Format(rule_algorithm.TIME_LAYOUT)
							spanindex, err := time.ParseDuration("8760h")
							if err != nil {
								beego.Error(err)
								return
							}
							structEmployeeXM.EmployeInfo.EndTime = time.Now().Add(spanindex).Format(rule_algorithm.TIME_LAYOUT)
							arrayImg := strings.Split(structEmployeeInfoIndex.Pic, ",")
							byteDecode, err := base64.StdEncoding.DecodeString(arrayImg[1])
							if err != nil {
								beego.Error(err)
								return
							}

							var imageIndex image.Image
							imageIndex, _, err = image.Decode(bytes.NewBuffer(byteDecode))
							if err != nil {
								beego.Error(err)
								imageIndex, err = png.Decode(bytes.NewBuffer(byteDecode))
								if err != nil {
									beego.Error(err)
									bSyncSuc = false
									strMessage = err.Error() + "UserID:" + structEmployeeXM.EmployeInfo.UserID + ", name:" + structEmployeeXM.EmployeInfo.Name
									break
								}

							}

							imageIndex = resize.Resize(300, 400, imageIndex, resize.Lanczos3)

							var buf bytes.Buffer
							err = jpeg.Encode(&buf, imageIndex, nil)
							if err != nil {
								beego.Error(err)
								continue
							}
							//jpeg.Encode()

							structEmployeeXM.EmployeInfo.FaceFeature = strings.ToUpper(hex.EncodeToString(buf.Bytes()))

							structEmployeeXM.DeviceInfo.DeviceID = structDeviceInfo.DevcieId

							err = facedevicexiongmai.AddEmployeeWSS(structEmployeeXM)
							if err != nil {
								strMessage = fmt.Sprintf("%s", err)
								beego.Warning("删除设备人员失败", err)
							}
						}
					}
				}
			}

			//err = my_db.UpdateEmployee(structEmployeeInfoIndex)
			if err != nil {
				beego.Error("同步数据错误", err)
				return
			}
		}
	}

	if bSyncSuc {
		structReplyInfo.Code = 20000
		structReplyInfo.Message = ""
	} else {
		structReplyInfo.Code = 50000
		structReplyInfo.Message = strMessage
	}

	structReplyInfo.Data.Next = 1
	structReplyInfo.Data.Res = true
	structReplyInfo.Data.Total = 1

	byteReply, err := json.Marshal(structReplyInfo)
	if err != nil {
		beego.Error("同步数据错误", err)
		return
	}

	c.Data["json"] = json.RawMessage(string(byteReply))
	c.ServeJSON()
}

//getRootOrgOptions
type RootOrgOptiponsController struct {
	beego.Controller
}

type RootOrgOptiponInfoJSON struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

type RootOrgOptionDataJSON struct {
	All  bool                     `json:"all"`
	Root []RootOrgOptiponInfoJSON `json:"root"`
}

type RootOrgOptiponsResultJSON struct {
	Code    int64                 `json:"code"`
	Message string                `json:"message"`
	Data    RootOrgOptionDataJSON `json:"data"`
}

func (c *RootOrgOptiponsController) Get() {
	structGroupArray, err := my_db.GetGroupListByPurview(0)
	if err != nil {
		beego.Error("\r\n获取root组失败:", err)
		return
	}

	var responseStruct RootOrgOptiponsResultJSON
	responseStruct.Data.Root = make([]RootOrgOptiponInfoJSON, 0)

	for _, structGroupInfo := range structGroupArray {
		var structGroup RootOrgOptiponInfoJSON
		strArrayIndex := strings.Split(structGroupInfo.Id, "-")
		structGroup.Id, err = strconv.ParseInt(strArrayIndex[1], 0, 64)
		if err != nil {
			beego.Error("获取root组错误：", err)
			return
		}
		structGroup.Name = structGroupInfo.Name

		responseStruct.Data.Root = append(responseStruct.Data.Root, structGroup)
	}

	responseStruct.Code = 20000
	responseStruct.Message = ""

	XToken := c.Ctx.Input.Header("X-Token")
	strAdmin, err := my_aes.Dncrypt(string(XToken))
	if err != nil {
		beego.Error("获取规则列表错误", err)
	}

	structAdminInfo, err := my_db.GetAdminByAccount(strAdmin)
	if err != nil {
		beego.Error(err)
		return
	}

	if structAdminInfo.ISRoot > 0 {
		responseStruct.Data.All = true
	} else {
		responseStruct.Data.All = false
	}

	responseByte, err := json.Marshal(responseStruct)
	if err != nil {
		return
	}

	c.Data["json"] = json.RawMessage(string(responseByte))
	c.ServeJSON()
}

//organization/list
type OrganizationlistController struct {
	beego.Controller
}

type OrganizationChildInfo_Json struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
	Sort int64  `json:"sort"`
}

type OrganizationInfo_Json struct {
	Id       int64                        `json:"id"`
	Name     string                       `json:"name"`
	Sort     int64                        `json:"sort"`
	Children []OrganizationChildInfo_Json `json:"children"`
}

type OrganizationsInfo_Json struct {
	Items []OrganizationInfo_Json `json:"items"`
	Total int64                   `json:"total"`
}

type OrganizationlistResult_Json struct {
	Code    int64                  `json:"code"`
	Message string                 `json:"message"`
	Data    OrganizationsInfo_Json `json:"data"`
}

func (c *OrganizationlistController) Get() {
	var responseStruct OrganizationlistResult_Json
	responseStruct.Data.Items = make([]OrganizationInfo_Json, 0)

	nToltalIndex := int64(0)
	structGroupArray, err := my_db.GetGroupListByPurview(0)
	if err != nil {
		beego.Error("\r\n获取组数据失败:", err)
		return
	}

	for _, structGroupInfo := range structGroupArray {
		structChildArray, err := my_db.GetGroupListByParent(structGroupInfo.Id)
		if err != nil {
			beego.Error("\r\n 获取子组失败:", err)
			return
		}
		var structGroup OrganizationInfo_Json
		structGroup.Children = make([]OrganizationChildInfo_Json, 0)
		strArrayIndex := strings.Split(structGroupInfo.Id, "-")
		structGroup.Id, err = strconv.ParseInt(strArrayIndex[1], 0, 64)
		if err != nil {
			beego.Error("获取组数据错误", err)
			return
		}
		structGroup.Name = structGroupInfo.Name
		structGroup.Sort = structGroupInfo.Sort
		for _, structChildInfo := range structChildArray {
			var strutChildGroup OrganizationChildInfo_Json
			strArrayIndex := strings.Split(structGroupInfo.Id, "-")
			nIdIndex, err := strconv.ParseInt(strArrayIndex[1], 0, 64)
			if err != nil {
				beego.Error("获取组数据错误", err)
				return
			}
			structGroup.Id = nIdIndex

			strutChildGroup.Name = structChildInfo.Name
			arrayIndex := strings.Split(structChildInfo.Id, "-")
			nIdIndex, err = strconv.ParseInt(arrayIndex[1], 0, 64)
			if err != nil {
				beego.Error("获取组数据错误", err)
				return
			}
			strutChildGroup.Id = nIdIndex
			strutChildGroup.Sort = structChildInfo.Sort

			structGroup.Children = append(structGroup.Children, strutChildGroup)
		}
		responseStruct.Data.Items = append(responseStruct.Data.Items, structGroup)
		nToltalIndex++
	}

	responseStruct.Code = 20000
	responseStruct.Message = "list 请求成功"
	responseStruct.Data.Total = nToltalIndex

	responseByte, err := json.Marshal(responseStruct)
	if err != nil {
		return
	}

	c.Data["json"] = json.RawMessage(responseByte)
	c.ServeJSON()
}

//personlist
type PersonlistController struct {
	beego.Controller
}

type PersonInfo_Json struct {
	Id                 int64    `json:"id"`
	Name               string   `json:"name"`
	Phone              string   `json:"phone"`
	Pic                string   `json:"pic"`
	Status             int64    `json:"status"`
	Msg                string   `json:"msg"`
	User_id            int64    `json:"user_id"`
	Admin_id           int64    `json:"admin_id"`
	Organizations      string   `json:"organizations"`
	Attendance_rule_id int64    `json:"attendance_rule_id"`
	Company_id         int64    `json:"company_id"`
	Sort               int64    `json:"sort"`
	Is_gateman         int      `json:"is_gateman"`
	Is_default         int      `json:"is_default"`
	Invcode            string   `json:"invcode"`
	Organization       string   `json:"organization"`
	Is_admin           bool     `json:"is_admin"`
	Devices            []int64  `json:"devices"`
	Is_init            int64    `json:"is_init"`
	Remote_devices     []int64  `json:"remote_devices"`
	Rights             []string `json:"rights"`
	In_charges         []int64  `json:"in_charges"`
	Working            bool     `json:"working"`
	Psw                string   `json:"password"`
	//Take_index
}

type PersonsInfo_Json struct {
	Items []PersonInfo_Json `json:"items"`
	Total int64             `json:"total"`
}

type PersonlistResult_Json struct {
	Code    int64            `json:"code"`
	Message string           `json:"message"`
	Data    PersonsInfo_Json `json:"data"`
}

func (c *PersonlistController) Get() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	//Page := c.Input().Get("page")
	//Limit := c.Input().Get("limit")
	Keyword := c.Input().Get("keyword")
	//beego.Debug(c.Input().Get("organizations[]"))
	mapOrgnization := make(map[string]int64, 0)
	arrayOrgnizations, key := c.Ctx.Request.Form["organizations[]"]
	if key {
		for _, value := range arrayOrgnizations {
			mapOrgnization["org-"+value] = 1
		}
	}

	XToken := c.Ctx.Input.Header("X-Token")
	strAdmin, err := my_aes.Dncrypt(string(XToken))
	if err != nil {
		beego.Error("获取规则列表错误", err)
	}

	structAdminInfo, err := my_db.GetAdminByAccount(strAdmin)
	if err != nil {
		beego.Error(err)
		return
	}

	var EmployeeArray []my_db.EmployeeInfo

	if structAdminInfo.ISRoot > 0 {
		EmployeeArray, err = my_db.GetEmployeeList(structAdminInfo.CompanyId, Keyword)
		if err != nil {
			beego.Error("获取人员列表失败:", err)
			return
		}
	} else {
		structEmployeeInfo, err := my_db.GetEmployeeByAdminAccount(structAdminInfo.Account, structAdminInfo.CompanyId)
		if err != nil {
			beego.Error(err)
			return
		}

		bAllCharge := false
		arrayIncharge := strings.Split(structEmployeeInfo.Incharges, ",")
		for _, strIncharge := range arrayIncharge {
			if strIncharge == "0" {
				bAllCharge = true
			}
		}

		if bAllCharge {
			EmployeeArray, err = my_db.GetEmployeeList(structAdminInfo.CompanyId, Keyword)
			if err != nil {
				beego.Error("获取人员列表失败:", err)
				return
			}
		} else {
			arrayGroupID, err := my_db.GetGroupByEmployee(structEmployeeInfo.Id)
			if err != nil {
				beego.Error(err)
				return
			}

			for _, strGroupID := range arrayGroupID {
				EmployeeArrayIndex, err := my_db.GetGroupEmployeeList(structAdminInfo.CompanyId, strGroupID, Keyword)
				if err != nil {
					beego.Error(err)
					return
				}
				EmployeeArray = append(EmployeeArray, EmployeeArrayIndex...)
			}
		}
	}

	var responseStruct PersonlistResult_Json
	responseStruct.Data.Items = make([]PersonInfo_Json, 0)

	for _, EmployeeIndex := range EmployeeArray {
		if len(mapOrgnization) > 0 {
			arrayGroupID, err := my_db.GetGroupByEmployee(EmployeeIndex.Id)
			if err != nil {
				beego.Error(err)
				return
			}
			bFind := false
			for _, valueOrg := range arrayGroupID {
				_, key := mapOrgnization[valueOrg]
				if key {
					bFind = true
				}
			}

			if !bFind {
				continue
			}
		}

		var structChild PersonInfo_Json

		structChild.Id = EmployeeIndex.Id
		structChild.Name = EmployeeIndex.Name
		structChild.Phone = EmployeeIndex.Phone
		structChild.Pic = EmployeeIndex.Pic
		structChild.Status = EmployeeIndex.Status
		structChild.Msg = ""
		structChild.User_id = 0
		structChild.Company_id = 3
		structChild.Sort = EmployeeIndex.Sort
		structChild.Is_gateman = 0
		structChild.Is_default = 1
		structChild.Invcode = `OXBPUy9nekJWenBIcE9EQnpjaWF0UEEwUHdMMEFmZWdrNmxwV1B4M243bGt5MnFCNGJudExnK3g3ekhtQmpITlVvQ0FYWW1Ec1pyZkFwVW1WZE9vSnc9PQ==`
		if EmployeeIndex.Working > 0 {
			structChild.Working = true
		} else {
			structChild.Working = false
		}

		if EmployeeIndex.IsAdmin > 0 {
			structChild.Is_admin = true

			strcutAdminInfo, err := my_db.GetAdminByAccount(EmployeeIndex.AdminAccount)
			if err != nil {
				beego.Error(err)
				return
			}
			structChild.Psw = strcutAdminInfo.Psw
		} else {
			structChild.Is_admin = false
		}

		structChild.Remote_devices = make([]int64, 0)
		structChild.Rights = make([]string, 0)
		structChild.In_charges = make([]int64, 0)

		if len(EmployeeIndex.Rights) > 0 {
			arrayRight := strings.Split(EmployeeIndex.Rights, ",")
			for _, strRight := range arrayRight {
				structChild.Rights = append(structChild.Rights, strRight)
			}
		}

		if len(EmployeeIndex.RemoteDevice) > 0 {
			arrayRemote := strings.Split(EmployeeIndex.RemoteDevice, ",")
			for _, strRemoteDevice := range arrayRemote {
				nDeviceIDIndex, err := strconv.ParseInt(strRemoteDevice, 0, 64)
				if err != nil {
					beego.Error(err)
					return
				}
				structChild.Remote_devices = append(structChild.Remote_devices, nDeviceIDIndex)
			}
		}

		if len(EmployeeIndex.Incharges) > 0 {
			arrayIncharge := strings.Split(EmployeeIndex.Incharges, ",")
			for _, strCharge := range arrayIncharge {
				nID, err := strconv.ParseInt(strCharge, 0, 64)
				if err != nil {
					beego.Error(err)
					return
				}
				structChild.In_charges = append(structChild.In_charges, nID)
			}
		}

		GroupArray, err := my_db.GetGroupByEmployee(EmployeeIndex.Id)
		if err != nil {
			beego.Error("获取人员列表错误:", err)
			return
		}
		for nIndex, strGroupIndex := range GroupArray {
			structGroupInfo, err := my_db.GetGroupById(strGroupIndex)
			if err != nil {
				beego.Error("获取人员列表错误", err)
				return
			}
			arrayIndex := strings.Split(strGroupIndex, "-")

			if nIndex == 0 {
				structChild.Organizations += arrayIndex[1]
				structChild.Organization += structGroupInfo.Name
			} else {
				structChild.Organizations += ("," + arrayIndex[1])
				structChild.Organization += ("," + structGroupInfo.Name)
			}
		}

		structChild.Attendance_rule_id = EmployeeIndex.RuleID
		structEmployeeAndDevice, err := my_db.GetEmployeeAndDvice(EmployeeIndex.Id)
		if err != nil {
			beego.Error(err)
			return
		}

		structChild.Devices = make([]int64, 0)
		for _, nDeviceID := range structEmployeeAndDevice.DevcieId {
			structChild.Devices = append(structChild.Devices, nDeviceID)
		}

		responseStruct.Data.Items = append(responseStruct.Data.Items, structChild)
	}

	responseStruct.Code = 20000
	responseStruct.Message = confreader.GetValue("SucRequest")
	responseStruct.Data.Total = int64(len(responseStruct.Data.Items))

	responseByte, err := json.Marshal(responseStruct)
	if err != nil {
		return
	}

	c.Data["json"] = json.RawMessage(string(responseByte))
	c.ServeJSON()
}

//GetAllOrgOptionsController  ___
type GetAllOrgOptionsController struct {
	beego.Controller
}

//OrgOptionChildInfoJSON  ___
type OrgOptionChildInfoJSON struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

//OrgOptionInfoJSON  ___
type OrgOptionInfoJSON struct {
	ID       int64                    `json:"id"`
	Name     string                   `json:"name"`
	Sort     int64                    `json:"sort"`
	Valid    bool                     `json:"valid"`
	Children []OrgOptionChildInfoJSON `json:"children"`
}

//AllOrgOptionInfoJSON  ___
type AllOrgOptionInfoJSON struct {
	All  bool                `json:"all"`
	Tree []OrgOptionInfoJSON `json:"tree"`
}

//AllOrgOptionListJSON  ___
type AllOrgOptionListJSON struct {
	Code    int64                `json:"code"`
	Message string               `json:"message"`
	Data    AllOrgOptionInfoJSON `json:"data"`
}

//Get ___
func (c *GetAllOrgOptionsController) Get() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	RootGroupArray, err := my_db.GetGroupListByPurview(0)
	if err != nil {
		beego.Error("获取组织架构失败", err)
		return
	}

	var responseStruct AllOrgOptionListJSON
	responseStruct.Data.Tree = make([]OrgOptionInfoJSON, 0)
	for _, GroupInfo := range RootGroupArray {
		ChildGroupArray, err := my_db.GetGroupListByParent(GroupInfo.Id)
		if err != nil {
			beego.Error("获取组织架构失败", err)
			return
		}
		var structGroupInfo OrgOptionInfoJSON
		structGroupInfo.Children = make([]OrgOptionChildInfoJSON, 0)
		nIndexGroupID, err := strconv.ParseInt(strings.Split(GroupInfo.Id, "-")[1], 0, 64)
		if err != nil {
			beego.Error("获取组织架构失败", err)
			return
		}

		structGroupInfo.ID = nIndexGroupID
		structGroupInfo.Name = GroupInfo.Name
		structGroupInfo.Valid = true

		for _, childGroupInfo := range ChildGroupArray {
			var structChildInfo OrgOptionChildInfoJSON
			structChildInfo.Name = childGroupInfo.Name
			nIndexId, err := strconv.ParseInt(strings.Split(childGroupInfo.Id, "-")[1], 0, 64)
			if err != nil {
				beego.Error("获取组织架构失败", err)
				return
			}
			structChildInfo.ID = nIndexId

			structGroupInfo.Children = append(structGroupInfo.Children, structChildInfo)
		}
		responseStruct.Data.Tree = append(responseStruct.Data.Tree, structGroupInfo)
	}

	responseStruct.Code = 20000
	responseStruct.Message = confreader.GetValue("SucRequest")

	responseStruct.Data.All = true

	responseByte, err := json.Marshal(responseStruct)
	if err != nil {
		return
	}

	c.Data["json"] = json.RawMessage(string(responseByte))
	c.ServeJSON()
}

//createGroup
type CreateGroupController struct {
	beego.Controller
}

//接收到的
type CreateGroupRequset_Json struct {
	Name        string `json:"name"`
	Parent_id   int64  `json:"parent_id"`
	Sort        string `json:"sort"`
	Titserialle string `json:"titserialle"`
}

//发送出去的
type CreateGroupReplyInfo_Json struct {
	Name       string `json:"name"`
	Parent_id  int64  `json:"parent_id"`
	Company_id int64  `json:"company_id"`
	Sort       int64  `json:"sort"`
	Id         string `json:"id"`
}

type CreateGroupReply_Json struct {
	Code    int64                     `json:"code"`
	Message string                    `json:"message"`
	Data    CreateGroupReplyInfo_Json `json:"data"`
}

func (c *CreateGroupController) Post() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	data := c.Ctx.Input.RequestBody

	beego.Debug(string(data))

	var createInfo CreateGroupRequset_Json
	json.Unmarshal(data, &createInfo)

	//处理
	var groupInfo my_db.GroupInfo
	groupInfo.Id = "-1"
	groupInfo.Name = createInfo.Name
	groupInfo.Parent = createInfo.Parent_id
	nIndex, err := strconv.ParseInt(createInfo.Sort, 0, 64)
	if err != nil {
		beego.Error("添加组错误", err)
		return
	}
	groupInfo.Sort = nIndex
	if createInfo.Parent_id == 0 {
		groupInfo.Purview = 0
	} else {
		groupInfo.Purview = 1
	}

	nId, err := my_db.AddGroup(groupInfo)
	if err != nil {
		beego.Error("添加组错误", err)
		return
	}

	//返回
	var responseStruct CreateGroupReply_Json
	responseStruct.Code = 20000
	responseStruct.Message = confreader.GetValue("SucRequest")

	var structInfo CreateGroupReplyInfo_Json
	structInfo.Company_id = 3
	structInfo.Id = fmt.Sprintf("%d", nId)
	structInfo.Name = groupInfo.Name
	structInfo.Parent_id = groupInfo.Parent
	structInfo.Sort = groupInfo.Sort

	responseStruct.Data = structInfo

	responseByte, err := json.Marshal(responseStruct)
	if err != nil {
		return
	}

	c.Data["json"] = json.RawMessage(string(responseByte))
	c.ServeJSON()

}

//更新组数据
type UpdateGroupController struct {
	beego.Controller
}

//接收到的数据
type UpdateGroupChildInfo_Json struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

type UpdateRequsetInfo_Json struct {
	Id       int64                       `json:"id"`
	Name     string                      `json:"name"`
	Sort     string                      `json:"sort"`
	Children []UpdateGroupChildInfo_Json `json:"children"`
}

//发送的数据
type UpdateReplyInfo_Json struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

func (c *UpdateGroupController) Post() {

	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	data := c.Ctx.Input.RequestBody

	beego.Debug(string(data))

	//解析数据
	var structRequset UpdateRequsetInfo_Json
	err := json.Unmarshal(data, &structRequset)
	if err != nil {
		beego.Error(err)
		return
	}

	///更新操作
	var structGroupInfo my_db.GroupInfo
	structGroupInfo.Id = fmt.Sprintf("org-%d", structRequset.Id)
	structGroupInfo.Name = structRequset.Name
	nIndex, err := strconv.ParseInt(structRequset.Sort, 0, 64)
	if err != nil {
		beego.Error("更新组数据错误", err)
		return
	}
	structGroupInfo.Sort = nIndex
	err = my_db.UpdateGroup(structGroupInfo)
	if err != nil {
		beego.Error("更新组数据错误", err)
		return
	}

	//返回结果
	var structReply UpdateReplyInfo_Json
	structReply.Code = 20000
	structReply.Message = confreader.GetValue("SucRequest")
	structReply.Data = ""

	responseByte, err := json.Marshal(structReply)
	if err != nil {
		return
	}

	c.Data["json"] = json.RawMessage(string(responseByte))
	c.ServeJSON()

}

///api/management/organization/delete
type DeleteGroupController struct {
	beego.Controller
}

//接收到的信息
type DeleteGroupChildInfo struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

type DeleteGroupRequestInfo_Json struct {
	Id       int64                  `json:"id"`
	Name     string                 `json:"name"`
	Sort     interface{}            `json:"sort"`
	Children []DeleteGroupChildInfo `json:"children"`
}

//发送信息
type DeleteGroupReplyInfo_Json struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

func (c *DeleteGroupController) Post() {

	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	data := c.Ctx.Input.RequestBody
	var structRequst DeleteGroupRequestInfo_Json
	err := json.Unmarshal(data, &structRequst)
	if err != nil {
		beego.Error("删除组织错误，解析json数据失败", err)
		return
	}

	err = my_db.RemoveGroup(structRequst.Id)
	if err != nil {
		beego.Error("删除组织错误:", err)
		return
	}

	var structRplyInfo DeleteGroupReplyInfo_Json
	structRplyInfo.Code = 20000
	structRplyInfo.Data = ""
	structRplyInfo.Message = confreader.GetValue("SucRequest")

	byteReply, err := json.Marshal(structRplyInfo)
	if err != nil {
		beego.Error("删除组织错误:", err)
		return
	}

	c.Data["json"] = json.RawMessage(string(byteReply))
	c.ServeJSON()
}

//person/personSelectedDevice
type GetPersonSelectDevcieController struct {
	beego.Controller
}

//接收到的数据
type PersonSelectDeviceRequstInfo_Json struct {
	Person_id int64 `json:"person_id"`
}

//发送的数据
type PersonSelectDeviceReplyInfo_Json struct {
	Code    int64   `json:"code"`
	Message string  `json:"message"`
	Data    []int64 `json:"data"`
}

func (c *GetPersonSelectDevcieController) Post() {

	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	strPerson_id := c.Input().Get("person_id")
	//beego.Debug(strPerson_id)

	nPersonId, err := strconv.ParseInt(string(strPerson_id), 0, 64)
	// //处理请求
	structEmployeeAndDevice, err := my_db.GetEmployeeAndDvice(nPersonId)
	if err != nil {
		beego.Error("获取设备选中内容错误", err)
		return
	}

	//返回结果
	var structReply PersonSelectDeviceReplyInfo_Json
	structReply.Code = 20000
	structReply.Message = confreader.GetValue("SucRequest")
	structReply.Data = make([]int64, 0)

	for _, nDeviceIdIndex := range structEmployeeAndDevice.DevcieId {
		structReply.Data = append(structReply.Data, nDeviceIdIndex)
	}

	responseByte, err := json.Marshal(structReply)
	if err != nil {
		return
	}

	c.Data["json"] = json.RawMessage(string(responseByte))
	c.ServeJSON()

}

//删除 /api/management/person/delete
type PersonDeleteRequestController struct {
	beego.Controller
}

//接收到的数据
type PersonDeleteInfoRequset struct {
	Admin_id           int64  `json:"admin_id"`
	Attendance_rule_id int64  `json:"attendance_rule_id"`
	Company_id         int64  `json:"company_id"`
	Id                 int64  `json:"id"`
	Invcode            string `json:"invcode"`
	Is_admin           bool   `json:"is_admin"`
	Is_default         int    `json:"is_default"`
	Is_gateman         int    `json:"is_gateman"`
	Msg                string `json:"msg"`
	Name               string `json:"name"`
	Organization       string `json:"organization"`
	Organizations      string `json:"organizations"`
	Phone              string `json:"phone"`
	Pic                string `json:"pic"`
	Sort               int64  `json:"sort"`
	Status             int64  `json:"status"`
	User_id            int64  `json:"user_id"`
}

//返回的数据
type PersonDeleteInfoReply struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

func (c *PersonDeleteRequestController) Post() {

	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	data := c.Ctx.Input.RequestBody

	//解析数据
	var structRequestInfo PersonDeleteInfoRequset
	err := json.Unmarshal(data, &structRequestInfo)
	if err != nil {
		beego.Error(err)
		return
	}

	// nId, err := strconv.ParseInt(structRequestInfo.Id, 0, 64)
	// if err != nil {
	// 	beego.Error("删除人员错误", err)
	// 	return
	// }
	structEmployeeInfo, err := my_db.GetEmployeeById(structRequestInfo.Id)
	if err != nil {
		beego.Error("删除人员错误", err)
		return
	}

	if structRequestInfo.Status == 1 {
		structEmployeeAndDeviceInfo, err := my_db.GetEmployeeAndDvice(structRequestInfo.Id)
		if err != nil {
			beego.Error("删除人员错误", err)
			return
		}

		for _, nDeviceIdIndex := range structEmployeeAndDeviceInfo.DevcieId {
			structDeviceInfo_, err := my_db.GetDeviceById(nDeviceIdIndex)
			if err != nil {
				beego.Error("删除人员错误", err)
				return
			}

			if structDeviceInfo_.Type == 0 {
				err = facedevice.RemoveDeviceEmployee(structDeviceInfo_, structEmployeeInfo)
				if err != nil {
					beego.Warning("删除人员警告:", err)
				}
			}

			if structDeviceInfo_.Type == 1 {
				var structXMRemoveInfo facedevicexiongmai.RemoveEmployeeDataJSON
				structXMRemoveInfo.EmployeeInfo.UserID = fmt.Sprintf("%d", structEmployeeInfo.Id)
				structXMRemoveInfo.DeviceInfo.DeviceID = structDeviceInfo_.DevcieId
				err = facedevicexiongmai.RemoveEmployeeWSS(structXMRemoveInfo)
				if err != nil {
					beego.Warning("删除人员警告", err)
				}
			}
		}
	}

	if structRequestInfo.Is_admin {
		structEmployeeInfo, err := my_db.GetEmployeeById(structRequestInfo.Id)
		if err != nil {
			beego.Error(err)
			return
		}
		err = my_db.DeleteAdminUser(structEmployeeInfo.AdminAccount)
		if err != nil {
			beego.Error(err)
			return
		}
	}

	//处理事件
	err = my_db.RemoveEmployeeById(structRequestInfo.Id)
	if err != nil {
		beego.Error("删除人员错误", err)
		return
	}

	err = my_db.RemoveGroupAndEmployeeByEmployeeId(structRequestInfo.Id)
	if err != nil {
		beego.Error("删除人员错误", err)
		return
	}

	err = my_db.RemoveDeviceAndEmployeeByEmployeeId(structRequestInfo.Id)
	if err != nil {
		beego.Error("删除人员错误", err)
		return
	}

	err = my_db.RemoveAttendRecordByCustomId(structRequestInfo.Id)
	if err != nil {
		beego.Error("删除人员错误", err)
		return
	}

	err = my_db.DeleteRangeAndEmployeeByEmployeeID(structRequestInfo.Id)
	if err != nil {
		beego.Error(err)
		return
	}

	//返回结果
	var structReply PersonDeleteInfoReply
	structReply.Code = 20000
	structReply.Message = confreader.GetValue("SucRequest")
	structReply.Data = ""

	responseByte, err := json.Marshal(structReply)
	if err != nil {
		return
	}

	c.Data["json"] = json.RawMessage(string(responseByte))
	c.ServeJSON()

}

//更新人员 /api/management/person/update
type PersonUpdateController struct {
	beego.Controller
}

//接收数据
type PersonInfoRequst_Json struct {
	Admin_id           int64       `json:"admin_id"`
	Attendance_rule_id int64       `json:"attendance_rule_id"`
	Company_id         int64       `json:"company_id"`
	Devices            []int64     `json:"devices"`
	Id                 int64       `json:"id"`
	Invcode            string      `json:"invcode"`
	Is_admin           bool        `json:"is_admin"`
	Is_default         int         `json:"is_default"`
	Is_gateman         int         `json:"is_gateman"`
	Msg                string      `json:"msg"`
	Name               string      `json:"name"`
	Organization       string      `json:"organization"`
	Organizations      string      `json:"organizations"`
	Phone              string      `json:"phone"`
	Pic                string      `json:"pic"`
	Sort               interface{} `json:"sort"`
	Status             int64       `json:"status"`
	User_id            int64       `json:"user_id"`
	InCharges          []int64     `json:"in_charges"`
	Rights             []string    `json:"rights"`
	Remote_devices     []int64     `json:"remote_devices"`
	Psw                string      `json:"password"`
	Working            bool        `json:"working"`
	Sync               bool        `json:"sync"`
}

//发送数据
type PersonDelteReply_Json struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

func (c *PersonUpdateController) Post() {

	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	data := c.Ctx.Input.RequestBody

	//解析数据
	var structRequst PersonInfoRequst_Json
	err := json.Unmarshal(data, &structRequst)
	if err != nil {
		beego.Error(err)
		return
	}

	//获取数据库存放的人员设备
	mapDeviceInfo := make(map[int64]my_db.DeviceInfo, 0)
	structEmployeeAndDeviceInfo, err := my_db.GetEmployeeAndDvice(structRequst.Id)
	if err != nil {
		beego.Error("更新人员数据错误", err)
		return
	}

	for _, nDeviceIndex := range structEmployeeAndDeviceInfo.DevcieId {
		structDeviceInfo, err := my_db.GetDeviceById(nDeviceIndex)
		if err != nil {
			beego.Error("更新人员数据错误", err)
			return
		}
		mapDeviceInfo[nDeviceIndex] = structDeviceInfo
	}

	//删除设备 添加设备
	arrayAddDeviceInfo := make([]my_db.DeviceInfo, 0)
	arrayRemoveDeviceInfo := make([]my_db.DeviceInfo, 0)
	arrayUpdateDeviceInfo := make([]my_db.DeviceInfo, 0)

	mapRequstDeviceInfo := make(map[int64]my_db.DeviceInfo, 0)

	if structRequst.Working {
		//比较原来保存的设备和现在要设置的设备的差异
		for _, nDeviceIDIndex := range structRequst.Devices {
			structDeviceInfo, err := my_db.GetDeviceById(nDeviceIDIndex)
			if err != nil {
				beego.Error("更新设备错误", err)
				return
			}

			_, key := mapDeviceInfo[nDeviceIDIndex]
			if key {
				arrayUpdateDeviceInfo = append(arrayUpdateDeviceInfo, structDeviceInfo)
			} else {
				arrayAddDeviceInfo = append(arrayAddDeviceInfo, structDeviceInfo)
			}

			mapRequstDeviceInfo[nDeviceIDIndex] = structDeviceInfo
		}

		for _, mapIterDeviceInfo := range mapDeviceInfo {
			_, key := mapRequstDeviceInfo[mapIterDeviceInfo.Id]
			if !key {
				arrayRemoveDeviceInfo = append(arrayRemoveDeviceInfo, mapIterDeviceInfo)
			}
		}
	} else {
		for _, valueDeviceID := range structEmployeeAndDeviceInfo.DevcieId {
			err = my_db.InserSyncOpt(valueDeviceID, structRequst.Id, -1)
			if err != nil {
				beego.Error(err)
				return
			}
		}
	}

	//获取人员信息
	structEmployeeInfo, err := my_db.GetEmployeeById(structRequst.Id)
	if err != nil {
		beego.Error("更新人员信息错误", err)
		return
	}

	err = my_db.DeleteAdminUser(structEmployeeInfo.AdminAccount)
	if err != nil {
		beego.Error(err)
		return
	}

	err = my_db.DeleteAdminUser(structRequst.Phone)
	if err != nil {
		beego.Error(err)
		return
	}

	XToken := c.Ctx.Input.Header("X-Token")
	strAdmin, err := my_aes.Dncrypt(string(XToken))
	if err != nil {
		beego.Error("获取规则列表错误", err)
	}

	structRootAdminInfo, err := my_db.GetAdminByAccount(strAdmin)
	if err != nil {
		beego.Error(err)
		return
	}

	if structRequst.Is_admin {
		var structAdminInfo my_db.AdminInfo
		structAdminInfo.Name = structRequst.Name
		structAdminInfo.Account = structRequst.Phone
		structAdminInfo.PhoneName = structRequst.Phone
		structAdminInfo.Psw = structRequst.Psw
		structAdminInfo.CompanyId = structRootAdminInfo.CompanyId
		structAdminInfo.ISRoot = 0
		structAdminInfo.Companys = fmt.Sprintf("%d", structRootAdminInfo.CompanyId)
		err = my_db.AddAdminUser(structAdminInfo)
		if err != nil {
			beego.Error(err)

			mapResult := make(map[string]interface{}, 0)
			mapResult["code"] = 50000
			mapResult["message"] = err.Error()
			byteResult, err := json.Marshal(mapResult)
			if err != nil {
				beego.Error(err)
				return
			}
			c.Data["json"] = json.RawMessage(byteResult)
			c.ServeJSON()

			return
		}
	}

	structEmployeeInfo.AdminAccount = structRequst.Phone
	structEmployeeInfo.Pic = structRequst.Pic
	switch reflect.TypeOf(structRequst.Sort).Kind() {
	case reflect.Float64:
		structEmployeeInfo.Sort = int64(structRequst.Sort.(float64))
		break
	case reflect.String:
		{
			nSortIndex, err := strconv.ParseInt(structRequst.Sort.(string), 0, 64)
			if err != nil {
				beego.Error(err)
				return
			}
			structEmployeeInfo.Sort = nSortIndex
		}
		break
	}

	//structEmployeeInfo.Sort = structRequst.Sort
	structEmployeeInfo.Phone = structRequst.Phone
	structEmployeeInfo.Name = structRequst.Name
	if structRequst.Is_admin {
		structEmployeeInfo.IsAdmin = 1
	} else {
		structEmployeeInfo.IsAdmin = 0
	}

	if structRequst.Working {
		structEmployeeInfo.Working = 1
	} else {
		structEmployeeInfo.Working = 0
	}

	structEmployeeInfo.Rights = ""
	for nIndex, nRgiht := range structRequst.Rights {
		if nIndex == 0 {
			structEmployeeInfo.Rights = nRgiht
		} else {
			structEmployeeInfo.Rights += ","
			structEmployeeInfo.Rights += nRgiht
		}
	}

	structEmployeeInfo.RemoteDevice = ""
	for nIndex, nRemoteDevice := range structRequst.Remote_devices {
		if nIndex == 0 {
			structEmployeeInfo.RemoteDevice = fmt.Sprintf("%d", nRemoteDevice)
		} else {
			structEmployeeInfo.RemoteDevice += ","
			structEmployeeInfo.RemoteDevice += fmt.Sprintf("%d", nRemoteDevice)
		}
	}

	structEmployeeInfo.Incharges = ""
	for nIndex, nIncharge := range structRequst.InCharges {
		if nIndex == 0 {
			structEmployeeInfo.Incharges = fmt.Sprintf("%d", nIncharge)
		} else {
			structEmployeeInfo.Incharges += ","
			structEmployeeInfo.Incharges += fmt.Sprintf("%d", nIncharge)
		}
	}

	//根据添加列表添加设备
	for _, structAddInfoIndex := range arrayAddDeviceInfo {
		var structDeviceAndEmployeeAddInfo my_db.DeviceAndPersonInfo
		if structAddInfoIndex.Type == 0 {
			structDeviceAndEmployeeAddInfo.DeviceId = structAddInfoIndex.Id
			structDeviceAndEmployeeAddInfo.EmployeeId = append(structDeviceAndEmployeeAddInfo.EmployeeId, structEmployeeInfo.Id)

		}

		if structAddInfoIndex.Type == 1 {
			structDeviceAndEmployeeAddInfo.DeviceId = structAddInfoIndex.Id
			structDeviceAndEmployeeAddInfo.EmployeeId = append(structDeviceAndEmployeeAddInfo.EmployeeId, structEmployeeInfo.Id)
		}

		err = my_db.AddDviceAndEmployee(structDeviceAndEmployeeAddInfo)
		if err != nil {
			beego.Error("更新人员信息错误", err)
			return
		}

		err = my_db.InserSyncOpt(structAddInfoIndex.Id, structRequst.Id, 1)
		if err != nil {
			beego.Error(err)
			return
		}
	}

	//根据更新列表更新设备
	for _, structUpdateInfoIndex := range arrayUpdateDeviceInfo {
		var structDeviceAndEmployeeAddInfo my_db.DeviceAndPersonInfo
		if structUpdateInfoIndex.Type == 0 {
			structDeviceAndEmployeeAddInfo.DeviceId = structUpdateInfoIndex.Id
			structDeviceAndEmployeeAddInfo.EmployeeId = append(structDeviceAndEmployeeAddInfo.EmployeeId, structEmployeeInfo.Id)
		}

		if structUpdateInfoIndex.Type == 1 {
			structDeviceAndEmployeeAddInfo.DeviceId = structUpdateInfoIndex.Id
			structDeviceAndEmployeeAddInfo.EmployeeId = append(structDeviceAndEmployeeAddInfo.EmployeeId, structEmployeeInfo.Id)
		}

		err = my_db.RemoveDeviceAndEmoloyeeByAll(structEmployeeInfo.Id, structUpdateInfoIndex.Id)
		if err != nil {
			beego.Error("更新人员信息错误", err)
			return
		}

		err = my_db.AddDviceAndEmployee(structDeviceAndEmployeeAddInfo)
		if err != nil {
			beego.Error("更新人员信息错误", err)
			return
		}
		err = my_db.InserSyncOpt(structUpdateInfoIndex.Id, structRequst.Id, 2)
		if err != nil {
			beego.Error(err)
			return
		}
	}

	//根据删除列表删除设备
	for _, structRemoveInfoIndex := range arrayRemoveDeviceInfo {
		err = my_db.RemoveDeviceAndEmoloyeeByAll(structEmployeeInfo.Id, structRemoveInfoIndex.Id)
		if err != nil {
			beego.Error("更新人员信息错误", err)
			return
		}

		err = my_db.InserSyncOpt(structRemoveInfoIndex.Id, structRequst.Id, -1)
		if err != nil {
			beego.Error(err)
			return
		}
	}

	if !structRequst.Working && structRequst.Sync {
		err = my_db.RemoveDeviceAndEmployeeByEmployeeId(structRequst.Id)
		if err != nil {
			beego.Error(err)
			return
		}
	}

	structEmployeeInfo.Status = 0
	err = my_db.UpdateEmployee(structEmployeeInfo)
	if err != nil {
		beego.Error("更新人员信息错误", err)
		return
	}

	err = my_db.RemoveGroupAndEmployeeByEmployeeId(structEmployeeInfo.Id)
	if err != nil {
		beego.Error("更新人员信息错误", err)
		return
	}
	var structGroupAndEmployeeInfo my_db.GroupAndEmployee
	structGroupAndEmployeeInfo.EmployeeId = structEmployeeInfo.Id
	arrayGroupId := strings.Split(structRequst.Organizations, ",")
	for _, strGroupIdIndex := range arrayGroupId {
		arrayChild, err := my_db.GetGroupListByParent("org-" + strGroupIdIndex)
		if err != nil {
			beego.Error("更新人员信息错误", err)
			return
		}

		if len(arrayChild) > 0 {
			continue
		}

		structGroupAndEmployeeInfo.GroupId = "org-" + strGroupIdIndex
		err = my_db.AddGroupAndEmployee(structGroupAndEmployeeInfo)
		if err != nil {
			beego.Error("更新人员信息错误", err)
			return
		}
	}

	//返回结果
	var structReply PersonDelteReply_Json
	structReply.Code = 20000
	structReply.Message = confreader.GetValue("SucRequest")
	structReply.Data = ""

	responseByte, err := json.Marshal(structReply)
	if err != nil {
		return
	}

	c.Data["json"] = json.RawMessage(string(responseByte))
	c.ServeJSON()
}

//考勤规则列表 /api/management/attendance_rule/list
type RuleGetListContrller struct {
	beego.Controller
}

type FromInfoJSON struct {
	Color  string                 `json:"color"`
	Name   string                 `json:"name"`
	Type   int64                  `json:"type"`
	Range1 map[string]interface{} `json:"range1"`
	Range2 map[string]interface{} `json:"range2"`
	Range3 map[string]interface{} `json:"range3"`
}

type TagInfoJSON struct {
	Active   bool         `json:"active"`
	Id       int64        `json:"id"`
	Index    int64        `json:"index"`
	Selected bool         `json:"selected"`
	From     FromInfoJSON `json:"form"`
	Name     string       `json:"name"`
}

//轮班
type RuleItemJSON struct {
	Late_time     int64         `json:"late_time"`
	Left_early    int64         `json:"left_early"`
	Work_begin    string        `json:"work_begin"`
	Work_end      string        `json:"work_end"`
	Offwork_begin string        `json:"offwork_begin"`
	Offwork_end   string        `json:"offwork_end"`
	ID            int64         `json:"id"`
	Name          string        `json:"name"`
	Time1         string        `json:"time1"`
	Time2         string        `json:"time2"`
	Days          string        `json:"days"`
	Device_id     int64         `json:"device_id"`
	Company_id    int64         `json:"company_id"`
	Type          int64         `json:"type"`
	Tags          []TagInfoJSON `json:"tags"`
}

//一日N轮
//RuleAddInfoRequstJSON 接收数据
type DayRangeInfoJSON struct {
	Checktype int64                  `json:"checktype"`
	Days      string                 `json:"days"`
	Name      string                 `json:"name"`
	Person    []interface{}          `json:"person"`
	Range1    map[string]interface{} `json:"range1"`
	Range2    map[string]interface{} `json:"range2"`
	Range3    map[string]interface{} `json:"range3"`
	Sync      bool                   `json:"sync"`
	Type      int64                  `json:"type"`
	ID        int64                  `json:"id"`
}

type RulesInfoJSON struct {
	Items []interface{} `json:"items"`
	Total int64         `json:"total"`
}

type RuleGetListReplyJSON struct {
	Code    int64         `json:"code"`
	Message string        `json:"message"`
	Data    RulesInfoJSON `json:"data"`
}

func (c *RuleGetListContrller) Get() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	// Page := c.Input().Get("page")
	// Limit := c.Input().Get("limit")
	// Sort := c.Input().Get("+id")

	XToken := c.Ctx.Input.Header("X-Token")
	strAdmin, err := my_aes.Dncrypt(string(XToken))
	if err != nil {
		beego.Error("获取规则列表错误", err)
	}

	//beego.Debug(Page, Limit, Sort)
	//处理
	structRuleInfoArray, err := my_db.GetRuleList(strAdmin)
	if err != nil {
		beego.Error("获取规则列表错误", err)
		return
	}

	var structReply RuleGetListReplyJSON
	structReply.Data.Items = make([]interface{}, 0)
	structReply.Data.Total = 0
	for _, structRuleInfoIndex := range structRuleInfoArray {
		var structChildInfo RuleItemJSON
		structChildInfo.Name = structRuleInfoIndex.Name
		structChildInfo.Company_id = 3
		structChildInfo.Days = structRuleInfoIndex.Days
		structChildInfo.Device_id = 0
		structChildInfo.ID = structRuleInfoIndex.Id
		structChildInfo.Late_time = structRuleInfoIndex.Late_time
		structChildInfo.Left_early = structRuleInfoIndex.Left_early
		structChildInfo.Offwork_begin = structRuleInfoIndex.Offwork_begin
		structChildInfo.Offwork_end = structRuleInfoIndex.Offwork_end
		structChildInfo.Time1 = structRuleInfoIndex.Time1
		structChildInfo.Time2 = structRuleInfoIndex.Time2
		structChildInfo.Work_begin = structRuleInfoIndex.Work_begin
		structChildInfo.Work_end = structRuleInfoIndex.Work_end
		structChildInfo.Type = structRuleInfoIndex.Type
		structChildInfo.Tags = make([]TagInfoJSON, 0)

		if structRuleInfoIndex.Type == 1 && len(structRuleInfoIndex.Tags) > 0 {
			arrayTags := strings.Split(structRuleInfoIndex.Tags, ",")
			for _, strTagID := range arrayTags {
				nTagID, err := strconv.ParseInt(strTagID, 0, 64)
				if err != nil {
					beego.Error(err)
					return
				}
				structTagInfo, err := my_db.GetTagByID(nTagID)
				if err != nil {
					beego.Error(err)
					return
				}

				var structTagInfojson TagInfoJSON
				if structTagInfo.Active > 0 {
					structTagInfojson.Active = true
				} else {
					structTagInfojson.Active = false
				}

				structTagInfojson.Id = structTagInfo.ID
				structTagInfojson.Index = structTagInfo.Index
				if structTagInfo.Selected > 0 {
					structTagInfojson.Selected = true
				} else {
					structTagInfojson.Selected = false
				}

				structTagInfojson.From.Color = structTagInfo.Color
				structTagInfojson.From.Name = structTagInfo.Name

				structTagInfojson.From.Range1 = make(map[string]interface{}, 0)
				structTagInfojson.From.Range2 = make(map[string]interface{}, 0)
				structTagInfojson.From.Range3 = make(map[string]interface{}, 0)

				if structTagInfo.Range1 > 0 {
					structRange1, err := my_db.GetRangeByID(structTagInfo.Range1)
					if err != nil {
						beego.Error(err)
						return
					}
					structTagInfojson.From.Range1["late_time"] = structRange1.LateTime
					structTagInfojson.From.Range1["left_early"] = structRange1.LeftEarly
					structTagInfojson.From.Range1["offwork_begin"] = structRange1.OffWorkBegin
					structTagInfojson.From.Range1["offwork_end"] = structRange1.OffWorkEnd
					if structRange1.OffWorkCheck > 0 {
						structTagInfojson.From.Range1["offworkcheck"] = true
					} else {
						structTagInfojson.From.Range1["offworkcheck"] = false
					}

					structTagInfojson.From.Range1["time1"] = structRange1.Time1
					structTagInfojson.From.Range1["time2"] = structRange1.Time2
					structTagInfojson.From.Range1["work_begin"] = structRange1.WorkBegin
					structTagInfojson.From.Range1["work_end"] = structRange1.WorkEnd
					if structRange1.WorkCheck > 0 {
						structTagInfojson.From.Range1["workcheck"] = true
					} else {
						structTagInfojson.From.Range1["workcheck"] = false
					}
				}

				if structTagInfo.Range2 > 0 {
					structRange2, err := my_db.GetRangeByID(structTagInfo.Range2)
					if err != nil {
						beego.Error(err)
						return
					}
					structTagInfojson.From.Range2["late_time"] = structRange2.LateTime
					structTagInfojson.From.Range2["left_early"] = structRange2.LeftEarly
					structTagInfojson.From.Range2["offwork_begin"] = structRange2.OffWorkBegin
					structTagInfojson.From.Range2["offwork_end"] = structRange2.OffWorkEnd
					if structRange2.OffWorkCheck > 0 {
						structTagInfojson.From.Range2["offworkcheck"] = true
					} else {
						structTagInfojson.From.Range2["offworkcheck"] = false
					}

					structTagInfojson.From.Range2["time1"] = structRange2.Time1
					structTagInfojson.From.Range2["time2"] = structRange2.Time2
					structTagInfojson.From.Range2["work_begin"] = structRange2.WorkBegin
					structTagInfojson.From.Range2["work_end"] = structRange2.WorkEnd
					if structRange2.WorkCheck > 0 {
						structTagInfojson.From.Range2["workcheck"] = true
					} else {
						structTagInfojson.From.Range2["workcheck"] = false
					}
				}

				if structTagInfo.Range3 > 0 {
					structRange3, err := my_db.GetRangeByID(structTagInfo.Range3)
					if err != nil {
						beego.Error(err)
						return
					}
					structTagInfojson.From.Range3["late_time"] = structRange3.LeftEarly
					structTagInfojson.From.Range3["offwork_begin"] = structRange3.OffWorkBegin
					structTagInfojson.From.Range3["offwork_end"] = structRange3.OffWorkEnd
					if structRange3.OffWorkCheck > 0 {
						structTagInfojson.From.Range3["offworkcheck"] = true
					} else {
						structTagInfojson.From.Range3["offworkcheck"] = false
					}

					structTagInfojson.From.Range3["time1"] = structRange3.Time1
					structTagInfojson.From.Range3["time2"] = structRange3.Time2
					structTagInfojson.From.Range3["work_begin"] = structRange3.WorkBegin
					structTagInfojson.From.Range3["work_end"] = structRange3.WorkEnd
					if structRange3.WorkCheck > 0 {
						structTagInfojson.From.Range3["workcheck"] = true
					} else {
						structTagInfojson.From.Range3["workcheck"] = false
					}
				}

				structChildInfo.Tags = append(structChildInfo.Tags, structTagInfojson)
			}
		}
		structReply.Data.Items = append(structReply.Data.Items, structChildInfo)
		structReply.Data.Total++
	}

	mapDayRange, err := my_db.GetDayRangeList()
	if err != nil {
		beego.Error(err)
		return
	}

	for nFarther, ruleValue := range mapDayRange {
		var structRuleInfo DayRangeInfoJSON
		structRuleInfo.Checktype = int64(len(ruleValue))
		structRuleInfo.Days = ruleValue[0].Days
		structRuleInfo.Name = ruleValue[0].FatherName
		structRuleInfo.Person = make([]interface{}, 0)
		arrayEmployee, err := my_db.GetEmployeeListByRuleID(nFarther)
		if err != nil {
			beego.Error(err)
			return
		}

		for _, valueIndex := range arrayEmployee {
			structRuleInfo.Person = append(structRuleInfo.Person, valueIndex)
		}

		//copy(structRuleInfo.Person, arrayEmployee)

		if len(ruleValue) > 0 {
			structRuleInfo.Range1 = make(map[string]interface{})
			structRuleInfo.Range1["late_time"] = ruleValue[0].LateTime
			structRuleInfo.Range1["left_early"] = ruleValue[0].LeftEarly
			structRuleInfo.Range1["offwork_begin"] = ruleValue[0].OffworkBegin
			structRuleInfo.Range1["offwork_end"] = ruleValue[0].OffworkEnd
			if ruleValue[0].OffworkCheck > 0 {
				structRuleInfo.Range1["offworkcheck"] = true
			} else {
				structRuleInfo.Range1["offworkcheck"] = false
			}
			structRuleInfo.Range1["time1"] = ruleValue[0].Time1
			structRuleInfo.Range1["time2"] = ruleValue[0].Time2
			structRuleInfo.Range1["work_begin"] = ruleValue[0].WorkBegin
			structRuleInfo.Range1["work_end"] = ruleValue[0].WorkEnd
			if ruleValue[0].WorkCheck > 0 {
				structRuleInfo.Range1["workcheck"] = true
			} else {
				structRuleInfo.Range1["workcheck"] = false
			}

		}

		if len(ruleValue) > 1 {
			structRuleInfo.Range2 = make(map[string]interface{})
			structRuleInfo.Range2["late_time"] = ruleValue[1].LateTime
			structRuleInfo.Range2["left_early"] = ruleValue[1].LeftEarly
			structRuleInfo.Range2["offwork_begin"] = ruleValue[1].OffworkBegin
			structRuleInfo.Range2["offwork_end"] = ruleValue[1].OffworkEnd
			if ruleValue[1].OffworkCheck > 0 {
				structRuleInfo.Range2["offworkcheck"] = true
			} else {
				structRuleInfo.Range2["offworkcheck"] = false
			}
			structRuleInfo.Range2["time1"] = ruleValue[1].Time1
			structRuleInfo.Range2["time2"] = ruleValue[1].Time2
			structRuleInfo.Range2["work_begin"] = ruleValue[1].WorkBegin
			structRuleInfo.Range2["work_end"] = ruleValue[1].WorkEnd
			if ruleValue[1].WorkCheck > 0 {
				structRuleInfo.Range2["workcheck"] = true
			} else {
				structRuleInfo.Range2["workcheck"] = false
			}
		}

		if len(ruleValue) > 2 {
			structRuleInfo.Range3 = make(map[string]interface{})
			structRuleInfo.Range3["late_time"] = ruleValue[2].LateTime
			structRuleInfo.Range3["left_early"] = ruleValue[2].LeftEarly
			structRuleInfo.Range3["offwork_begin"] = ruleValue[2].OffworkBegin
			structRuleInfo.Range3["offwork_end"] = ruleValue[2].OffworkEnd
			if ruleValue[2].OffworkCheck > 0 {
				structRuleInfo.Range3["offworkcheck"] = true
			} else {
				structRuleInfo.Range3["offworkcheck"] = false
			}
			structRuleInfo.Range3["time1"] = ruleValue[2].Time1
			structRuleInfo.Range3["time2"] = ruleValue[2].Time2
			structRuleInfo.Range3["work_begin"] = ruleValue[2].WorkBegin
			structRuleInfo.Range3["work_end"] = ruleValue[2].WorkEnd
			if ruleValue[2].WorkCheck > 0 {
				structRuleInfo.Range3["workcheck"] = true
			} else {
				structRuleInfo.Range3["workcheck"] = false
			}

		}

		structRuleInfo.Type = 0
		nIDBegin, err := my_db.GetMaxRuleID()
		if err != nil {
			beego.Error(err)
			return
		}
		structRuleInfo.ID = nIDBegin + nFarther
		structReply.Data.Items = append(structReply.Data.Items, structRuleInfo)
	}

	//返回数据

	structReply.Code = 20000
	structReply.Message = confreader.GetValue("SucRequest")

	responseByte, err := json.Marshal(structReply)
	if err != nil {
		return
	}

	c.Data["json"] = json.RawMessage(string(responseByte))
	c.ServeJSON()
}

//添加规则  /api/management/attendance_rule/create
type RuleAddController struct {
	beego.Controller
}

//RuleAddInfoRequstJSON 接收数据
type RuleAddInfoRequstJSON struct {
	Checktype int64                  `json:"checktype"`
	Days      string                 `json:"days"`
	Name      string                 `json:"name"`
	Person    []interface{}          `json:"person"`
	Range1    map[string]interface{} `json:"range1"`
	Range2    map[string]interface{} `json:"range2"`
	Range3    map[string]interface{} `json:"range3"`
	Sync      bool                   `json:"sync"`
	ID        int64                  `json:"id"`
	Type      int64                  `json:"type"`
}

type RuleAddInfoReplyJSON struct {
	Code    int64                 `json:"code"`
	Message string                `json:"message"`
	Data    RuleAddInfoRequstJSON `json:"data"`
}

func (c *RuleAddController) Post() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	data := c.Ctx.Input.RequestBody
	var structRequst RuleAddInfoRequstJSON
	err := json.Unmarshal(data, &structRequst)
	if err != nil {
		beego.Error(err)
		return
	}

	var structDayRangeInfo my_db.DayRangeInfo
	if structRequst.Checktype > 0 {
		structDayRangeInfo.ID = -1
		structDayRangeInfo.Father = -1
		structDayRangeInfo.LateTime = int64(structRequst.Range1["late_time"].(float64))
		structDayRangeInfo.LeftEarly = int64(structRequst.Range1["left_early"].(float64))
		structDayRangeInfo.OffworkBegin = structRequst.Range1["offwork_begin"].(string)
		structDayRangeInfo.OffworkEnd = structRequst.Range1["offwork_end"].(string)
		if structRequst.Range1["offworkcheck"].(bool) {
			structDayRangeInfo.OffworkCheck = 1
		} else {
			structDayRangeInfo.OffworkCheck = 0
		}

		structDayRangeInfo.Time1 = structRequst.Range1["time1"].(string)
		structDayRangeInfo.Time2 = structRequst.Range1["time2"].(string)
		structDayRangeInfo.WorkBegin = structRequst.Range1["work_begin"].(string)
		structDayRangeInfo.WorkEnd = structRequst.Range1["work_end"].(string)
		if structRequst.Range1["workcheck"].(bool) {
			structDayRangeInfo.WorkCheck = 1
		} else {
			structDayRangeInfo.WorkCheck = 0
		}

		structDayRangeInfo.Days = structRequst.Days
		structDayRangeInfo.FatherName = structRequst.Name

		structDayRangeInfo.ID, structDayRangeInfo.Father, err = my_db.AddDayRange(structDayRangeInfo)
		structRequst.ID = structDayRangeInfo.Father
		if err != nil {
			beego.Error(err)
			return
		}
	}

	if structRequst.Checktype > 1 {
		structDayRangeInfo.LateTime = int64(structRequst.Range2["late_time"].(float64))
		structDayRangeInfo.LeftEarly = int64(structRequst.Range2["left_early"].(float64))
		structDayRangeInfo.OffworkBegin = structRequst.Range2["offwork_begin"].(string)
		structDayRangeInfo.OffworkEnd = structRequst.Range2["offwork_end"].(string)
		if structRequst.Range2["offworkcheck"].(bool) {
			structDayRangeInfo.OffworkCheck = 1
		} else {
			structDayRangeInfo.OffworkCheck = 0
		}

		structDayRangeInfo.Time1 = structRequst.Range2["time1"].(string)
		structDayRangeInfo.Time2 = structRequst.Range2["time2"].(string)
		structDayRangeInfo.WorkBegin = structRequst.Range2["work_begin"].(string)
		structDayRangeInfo.WorkEnd = structRequst.Range2["work_end"].(string)
		if structRequst.Range2["workcheck"].(bool) {
			structDayRangeInfo.WorkCheck = 1
		} else {
			structDayRangeInfo.WorkCheck = 0
		}

		structDayRangeInfo.Days = structRequst.Days

		structDayRangeInfo.ID, _, err = my_db.AddDayRange(structDayRangeInfo)
		if err != nil {
			beego.Error(err)
			return
		}
	}

	if structRequst.Checktype > 2 {
		structDayRangeInfo.LateTime = int64(structRequst.Range3["late_time"].(float64))
		structDayRangeInfo.LeftEarly = int64(structRequst.Range3["left_early"].(float64))
		structDayRangeInfo.OffworkBegin = structRequst.Range3["offwork_begin"].(string)
		structDayRangeInfo.OffworkEnd = structRequst.Range3["offwork_end"].(string)
		if structRequst.Range3["offworkcheck"].(bool) {
			structDayRangeInfo.OffworkCheck = 1
		} else {
			structDayRangeInfo.OffworkCheck = 0
		}

		structDayRangeInfo.Time1 = structRequst.Range3["time1"].(string)
		structDayRangeInfo.Time2 = structRequst.Range3["time2"].(string)
		structDayRangeInfo.WorkBegin = structRequst.Range3["work_begin"].(string)
		structDayRangeInfo.WorkEnd = structRequst.Range3["work_end"].(string)
		if structRequst.Range3["workcheck"].(bool) {
			structDayRangeInfo.WorkCheck = 1
		} else {
			structDayRangeInfo.WorkCheck = 0
		}

		structDayRangeInfo.Days = structRequst.Days

		structDayRangeInfo.ID, _, err = my_db.AddDayRange(structDayRangeInfo)
		if err != nil {
			beego.Error(err)
			return
		}
	}

	//修改用户关联规则ID
	for _, ValueIndex := range structRequst.Person {
		switch ValueIndex.(type) {
		case float64:
			{
				nEmployeeID := int64(ValueIndex.(float64))
				err = my_db.UpdateEmployeeRuleID(nEmployeeID, structDayRangeInfo.Father)
				if err != nil {
					beego.Error(err)
					return
				}
			}
		}

	}

	structRequst.Type = 0
	if len(structRequst.Range1) == 0 {
		structRequst.Range1 = nil
	}

	if len(structRequst.Range2) == 0 {
		structRequst.Range2 = nil
	}

	if len(structRequst.Range3) == 0 {
		structRequst.Range3 = nil
	}

	var structReply RuleAddInfoReplyJSON
	structReply.Code = 20000
	structReply.Data = structRequst
	structReply.Message = ""

	byteReply, err := json.Marshal(structReply)
	if err != nil {
		beego.Error(err)
		return
	}
	c.Data["json"] = json.RawMessage(byteReply)
	c.ServeJSON()
}

///api/management/attendance_rule/selectedPersons
type RuleSelectPersonControl struct {
	beego.Controller
}

type RulseSelectReply_Json struct {
	Code    int64   `json:"code"`
	Message string  `json:"message"`
	Data    []int64 `json:"data"`
}

func (c *RuleSelectPersonControl) Get() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	Rule_Id := c.Input().Get("rule_id")

	//beego.Debug(Rule_Id)

	nRuleId, err := strconv.ParseInt(string(Rule_Id), 0, 64)
	if err != nil {
		beego.Error("获取选中的人员信息错误", err)
		return
	}

	//返回数据
	var structReply RulseSelectReply_Json
	structReply.Data = make([]int64, 0)

	//处理
	nEmployeeIdArray, err := my_db.GetEmployeeListByRuleID(nRuleId)
	if err != nil {
		beego.Error("获取选中人员信息错误", err)
		return
	}

	structReply.Data = nEmployeeIdArray
	structReply.Code = 20000
	structReply.Message = confreader.GetValue("SucRequest")

	responseByte, err := json.Marshal(structReply)
	if err != nil {
		return
	}

	c.Data["json"] = json.RawMessage(string(responseByte))
	c.ServeJSON()
}

///api/management/attendance_rule/update
type RuleUpdateController struct {
	beego.Controller
}

type RuleUpdateRequstJSON struct {
	Checktype int64                  `json:"checktype"`
	Days      string                 `json:"days"`
	Name      string                 `json:"name"`
	Person    []interface{}          `json:"person"`
	Range1    map[string]interface{} `json:"range1"`
	Range2    map[string]interface{} `json:"range2"`
	Range3    map[string]interface{} `json:"range3"`
	Sync      bool                   `json:"sync"`
	ID        int64                  `json:"id"`
	Type      int64                  `json:"type"`
}

type RuleUpdateReplyInfo_Json struct {
	Code    int64                `json:"code"`
	Message string               `json:"message"`
	Data    RuleUpdateRequstJSON `json:"data"`
}

func (c *RuleUpdateController) Post() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	data := c.Ctx.Input.RequestBody
	var structRequst RuleUpdateRequstJSON
	err := json.Unmarshal(data, &structRequst)
	if err != nil {
		beego.Error("更新规则错误", err)
		return
	}

	nMaxRuleID, err := my_db.GetMaxRuleID()
	if err != nil {
		beego.Error(err)
		return
	}
	nFather := structRequst.ID - nMaxRuleID

	err = my_db.DeleteRangeByFather(nFather)
	if err != nil {
		beego.Error(err)
		return
	}

	err = my_db.RemoveEmployeeRuleID(nFather)
	if err != nil {
		beego.Error(err)
		return
	}

	var structDayRangeInfo my_db.DayRangeInfo
	if structRequst.Checktype > 0 {
		structDayRangeInfo.ID = -1
		structDayRangeInfo.Father = structRequst.ID
		structDayRangeInfo.LateTime = int64(structRequst.Range1["late_time"].(float64))
		structDayRangeInfo.LeftEarly = int64(structRequst.Range1["left_early"].(float64))
		structDayRangeInfo.OffworkBegin = structRequst.Range1["offwork_begin"].(string)
		structDayRangeInfo.OffworkEnd = structRequst.Range1["offwork_end"].(string)
		if structRequst.Range1["offworkcheck"].(bool) {
			structDayRangeInfo.OffworkCheck = 1
		} else {
			structDayRangeInfo.OffworkCheck = 0
		}

		structDayRangeInfo.Time1 = structRequst.Range1["time1"].(string)
		structDayRangeInfo.Time2 = structRequst.Range1["time2"].(string)
		structDayRangeInfo.WorkBegin = structRequst.Range1["work_begin"].(string)
		structDayRangeInfo.WorkEnd = structRequst.Range1["work_end"].(string)
		if structRequst.Range1["workcheck"].(bool) {
			structDayRangeInfo.WorkCheck = 1
		} else {
			structDayRangeInfo.WorkCheck = 0
		}

		structDayRangeInfo.Days = structRequst.Days
		structDayRangeInfo.FatherName = structRequst.Name

		structDayRangeInfo.ID, structDayRangeInfo.Father, err = my_db.AddDayRange(structDayRangeInfo)
		if err != nil {
			beego.Error(err)
			return
		}
	}

	if structRequst.Checktype > 1 {
		structDayRangeInfo.LateTime = int64(structRequst.Range2["late_time"].(float64))
		structDayRangeInfo.LeftEarly = int64(structRequst.Range2["left_early"].(float64))
		structDayRangeInfo.OffworkBegin = structRequst.Range2["offwork_begin"].(string)
		structDayRangeInfo.OffworkEnd = structRequst.Range2["offwork_end"].(string)
		if structRequst.Range2["offworkcheck"].(bool) {
			structDayRangeInfo.OffworkCheck = 1
		} else {
			structDayRangeInfo.OffworkCheck = 0
		}

		structDayRangeInfo.Time1 = structRequst.Range2["time1"].(string)
		structDayRangeInfo.Time2 = structRequst.Range2["time2"].(string)
		structDayRangeInfo.WorkBegin = structRequst.Range2["work_begin"].(string)
		structDayRangeInfo.WorkEnd = structRequst.Range2["work_end"].(string)
		if structRequst.Range2["workcheck"].(bool) {
			structDayRangeInfo.WorkCheck = 1
		} else {
			structDayRangeInfo.WorkCheck = 0
		}

		structDayRangeInfo.Days = structRequst.Days

		structDayRangeInfo.ID, _, err = my_db.AddDayRange(structDayRangeInfo)
		if err != nil {
			beego.Error(err)
			return
		}
	}

	if structRequst.Checktype > 2 {
		structDayRangeInfo.LateTime = int64(structRequst.Range3["late_time"].(float64))
		structDayRangeInfo.LeftEarly = int64(structRequst.Range3["left_early"].(float64))
		structDayRangeInfo.OffworkBegin = structRequst.Range3["offwork_begin"].(string)
		structDayRangeInfo.OffworkEnd = structRequst.Range3["offwork_end"].(string)
		if structRequst.Range3["offworkcheck"].(bool) {
			structDayRangeInfo.OffworkCheck = 1
		} else {
			structDayRangeInfo.OffworkCheck = 0
		}

		structDayRangeInfo.Time1 = structRequst.Range3["time1"].(string)
		structDayRangeInfo.Time2 = structRequst.Range3["time2"].(string)
		structDayRangeInfo.WorkBegin = structRequst.Range3["work_begin"].(string)
		structDayRangeInfo.WorkEnd = structRequst.Range3["work_end"].(string)
		if structRequst.Range3["workcheck"].(bool) {
			structDayRangeInfo.WorkCheck = 1
		} else {
			structDayRangeInfo.WorkCheck = 0
		}

		structDayRangeInfo.Days = structRequst.Days

		structDayRangeInfo.ID, _, err = my_db.AddDayRange(structDayRangeInfo)
		if err != nil {
			beego.Error(err)
			return
		}
	}

	//修改用户关联规则ID
	for _, ValueIndex := range structRequst.Person {
		switch ValueIndex.(type) {
		case float64:
			{
				nEmployeeID := int64(ValueIndex.(float64))
				err = my_db.UpdateEmployeeRuleID(nEmployeeID, structDayRangeInfo.Father)
				if err != nil {
					beego.Error(err)
					return
				}
			}
		}

	}

	var structReply RuleUpdateReplyInfo_Json
	structReply.Code = 20000
	structReply.Data = structRequst
	structReply.Message = ""

	byteReply, err := json.Marshal(structReply)
	if err != nil {
		beego.Error(err)
		return
	}
	c.Data["json"] = json.RawMessage(byteReply)
	c.ServeJSON()
}

///api/management/attendance_rule/delete
type RuleDeleteController struct {
	beego.Controller
}

//接收的
type RuleDeleteRequst_Json struct {
	Company_id    int64         `json:"company_id"`
	Days          string        `json:"days"`
	Device_id     int64         `json:"device_id"`
	Id            int64         `json:"id"`
	Late_time     int64         `json:"late_time"`
	Left_early    int64         `json:"left_early"`
	Name          string        `json:"name"`
	Offwork_begin string        `json:"offwork_begin"`
	Offwork_end   string        `json:"offwork_end"`
	Time1         string        `json:"time1"`
	Time2         string        `json:"time2"`
	Work_begin    string        `json:"work_begin"`
	Work_end      string        `json:"work_end"`
	Type          int64         `json:"type"`
	Tags          []TagInfoJSON `json:"tags"`
}

//返回的
type RuleDeleteReply_Json struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

func (c *RuleDeleteController) Post() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	data := c.Ctx.Input.RequestBody

	nType, err := jsonparser.GetInt([]byte(data), "type")
	if err != nil {
		beego.Error(err)
		return
	}

	//固定时间
	if nType == 0 {
		var structRequst RuleUpdateRequstJSON
		err = json.Unmarshal(data, &structRequst)
		if err != nil {
			beego.Error(err)
			return
		}

		nMaxRuleID, err := my_db.GetMaxRuleID()
		if err != nil {
			beego.Error(err)
			return
		}

		nFather := structRequst.ID - nMaxRuleID

		err = my_db.DeleteRangeByFather(nFather)
		if err != nil {
			beego.Error(err)
			return
		}

		err = my_db.RemoveEmployeeRuleID(nFather)
		if err != nil {
			beego.Error(err)
			return
		}

	}

	//轮班
	if nType == 1 {
		var structRequstInfo RuleDeleteRequst_Json
		err := json.Unmarshal(data, &structRequstInfo)
		if err != nil {
			beego.Error("删除规则错误", err)
			return
		}

		for _, structTagIndex := range structRequstInfo.Tags {
			structTagInfo, err := my_db.GetTagByID(structTagIndex.Id)
			if err != nil {
				beego.Error(err)
				return
			}

			err = my_db.RemoveEmployeeTag(structTagInfo.ID)
			if err != nil {
				beego.Error(err)
				return
			}

			if structTagInfo.Range1 > 0 {
				err = my_db.RemoveRangeInfo(structTagInfo.Range1)
				if err != nil {
					beego.Error(err)
					return
				}
				err = my_db.DeleteRangeAndEmployeeByRangeID(structTagInfo.Range1)
				if err != nil {
					beego.Error(err)
					return
				}
			}

			if structTagInfo.Range2 > 0 {
				err = my_db.RemoveRangeInfo(structTagInfo.Range2)
				if err != nil {
					beego.Error(err)
					return
				}

				err = my_db.DeleteRangeAndEmployeeByRangeID(structTagInfo.Range2)
				if err != nil {
					beego.Error(err)
					return
				}
			}

			if structTagInfo.Range3 > 0 {
				err = my_db.RemoveRangeInfo(structTagInfo.Range3)
				if err != nil {
					beego.Error(err)
					return
				}

				err = my_db.DeleteRangeAndEmployeeByRangeID(structTagInfo.Range3)
				if err != nil {
					beego.Error(err)
					return
				}
			}

			err = my_db.RemoveTag(structTagInfo.ID)
			if err != nil {
				beego.Error(err)
				return
			}

		}

		err = my_db.RemoveEmployeeRuleID(structRequstInfo.Id)
		if err != nil {
			beego.Error(err)
			return
		}

		err = my_db.DeleteRule(structRequstInfo.Id)
		if err != nil {
			beego.Error("删除规则错误", err)
			return
		}
	}

	var structReply RuleDeleteReply_Json
	structReply.Code = 20000
	structReply.Message = confreader.GetValue("SucRequest")
	structReply.Data = ""

	byteReply, err := json.Marshal(structReply)
	if err != nil {
		beego.Error(err)
		return
	}

	c.Data["json"] = json.RawMessage(string(byteReply))
	c.ServeJSON()
}

///api/management/attendance/list   获取考勤记录列表
type AttendanceListController struct {
	beego.Controller
}

//返回数据
type AttendanceInfo_Json struct {
	Id              int64  `json:"id"`
	Time            string `json:"time"`
	Attendance_day  string `json:"attendance_day"`
	Attendance_time string `json:"attendance_time"`
	Pic             string `json:"pic"`
	Device_id       int64  `json:"device_id"`
	Company_id      int64  `json:"company_id"`
	Result          string `json:"result"`
	Person_id       int64  `json:"person_id"`
	Name            string `json:"name"`
	Create_time     string `json:"create_time"`
	Temperature     string `json:"temperature"`
	Alarm           int64  `json:"alarm"`
	Filter          int64  `json:"filter"`
	Person          string `json:"person"`
	Device          string `json:"device"`
}

type AttendancesInfo_Json struct {
	Total int64                 `json:"total"`
	Items []AttendanceInfo_Json `json:"items"`
}

type AttendanceListReply_Json struct {
	Code    int64                `json:"code"`
	Message string               `json:"message"`
	Data    AttendancesInfo_Json `json:"data"`
}

func (c *AttendanceListController) Get() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	//Xtoken := c.Ctx.Input.Header("X-Token")
	//Account, err := my_aes.Dncrypt(Xtoken)

	Page := c.Input().Get("page")
	Limit := c.Input().Get("limit")
	Keyword := c.Input().Get("keyword")
	Unknow := c.Input().Get("unknow")
	//Alarm := c.Input().Get("alarm")

	Date := ""
	DateEnd := ""
	arrayDate, key := c.Ctx.Request.Form["date[]"]
	if key && len(arrayDate) > 0 {
		Date = c.Ctx.Request.Form["date[]"][0]
	}

	if key && len(arrayDate) > 1 {
		DateEnd = c.Ctx.Request.Form["date[]"][1]
	}

	//beego.Debug("北京时间:", strTimeBeijin)
	bSearchByContent := false
	var structSearch my_db.SearchStruct
	if len(Date) != 0 && len(DateEnd) != 0 {
		strArray := strings.Split(string(Date), "T")
		strArray_ := strings.Split(strArray[1], "Z")

		strTimeIndex := strArray[0] + " " + strArray_[0]
		timeData, err := time.Parse(rule_algorithm.TIME_LAYOUT, strTimeIndex)
		if err != nil {
			beego.Error("获取考勤记录错误", err)
			return
		}

		strArray = strings.Split(string(DateEnd), "T")
		strArray_ = strings.Split(strArray[1], "Z")

		strTimeIndex = strArray[0] + " " + strArray_[0]
		timeDateEnd, err := time.Parse(rule_algorithm.TIME_LAYOUT, strTimeIndex)
		if err != nil {
			beego.Error(err)
			return
		}

		jetLag, err := time.ParseDuration("8h")
		timeBeijin := timeData.Add(jetLag)
		strTimeBeijin := timeBeijin.Format(rule_algorithm.TIME_LAYOUT)

		strBeijinTimeArray := strings.Split(strTimeBeijin, " ")
		structSearch.TimeBegin = strBeijinTimeArray[0] + "T" + strBeijinTimeArray[1] + "Z"

		timeBeijin = timeDateEnd.Add(jetLag)
		strTimeBeijin = timeBeijin.Format(rule_algorithm.TIME_LAYOUT)
		strBeijinTimeArray = strings.Split(strTimeBeijin, " ")
		structSearch.TimeEnd = strBeijinTimeArray[0] + "T" + strBeijinTimeArray[1] + "Z"

		//structSearch.TimeEnd = strBeijinTimeArray[0] + "T" + "23:59:59Z"
		bSearchByContent = true
	}

	if len(Keyword) != 0 {
		structSearch.SearchContent = string(Keyword)
		bSearchByContent = true
	}

	bUnknow, err := strconv.ParseBool(string(Unknow))
	if err != nil {
		beego.Error("获取考勤列表错误", err)
		return
	}

	nPage, err := strconv.ParseInt(string(Page), 0, 64)
	if err != nil {
		beego.Error("获取考勤列表错误", err)
		return
	}

	nLimit, err := strconv.ParseInt(string(Limit), 0, 64)
	if err != nil {
		beego.Error("获取考勤列表错误", err)
		return
	}

	var structRecordArray []my_db.VerifyRecordInfo
	var err_ error

	var structReply AttendanceListReply_Json
	structReply.Data.Items = make([]AttendanceInfo_Json, 0)

	XToken := c.Ctx.Input.Header("X-Token")

	strAccount, err := my_aes.Dncrypt(string(XToken))
	if err != nil {
		beego.Error("获取设备地图错误", err)
		return
	}

	structAdminInfo, err := my_db.GetAdminByAccount(strAccount)
	if err != nil {
		beego.Error(err)
		return
	}

	//处理
	if bSearchByContent {
		structRecordArray, err_ = my_db.GetVerifyRecordBySearchStruct(structAdminInfo.CompanyId, structSearch, !bUnknow, nPage, nLimit)
		if err_ != nil {
			beego.Error("获取考勤记录列表错误", err_)
			return
		}

		nTotalIndex, err_ := my_db.GetTotalVerifyRecordBySearchStruct(structAdminInfo.CompanyId, structSearch, !bUnknow)
		if err_ != nil {
			beego.Error("获取考勤记录列表错误", err_)
			return
		}
		structReply.Data.Total = nTotalIndex
	} else {
		structRecordArray, err_ = my_db.GetVerifyRecordList(structAdminInfo.CompanyId, !bUnknow, nPage, nLimit)
		if err_ != nil {
			beego.Error("获取考勤记录列表错误", err_)
			return
		}
		nTotalIndex, err_ := my_db.GetTotalVerifyRecordList(structAdminInfo.CompanyId, !bUnknow)
		if err != nil {
			beego.Error("获取考勤记录列表错误", err_)
			return
		}
		structReply.Data.Total = nTotalIndex
	}

	for _, structRecordInfo := range structRecordArray {

		var structChildInfo AttendanceInfo_Json
		structChildInfo.Result = structRecordInfo.Result
		structChildInfo.Time = structRecordInfo.CreateTime
		structChildInfo.Pic = structRecordInfo.VerifyPic
		structChildInfo.Company_id = 3
		structChildInfo.Device = structRecordInfo.DeviceName
		structChildInfo.Person_id = structRecordInfo.CutomId
		structChildInfo.Alarm = 0
		structChildInfo.Person = structRecordInfo.Name
		structChildInfo.Temperature = structRecordInfo.Template
		timeSplitArray := strings.Split(structRecordInfo.CreateTime, "T")
		timeSplitArray_ := strings.Split(timeSplitArray[1], "Z")
		structRecordInfo.CreateTime = timeSplitArray[0] + " " + timeSplitArray_[0]
		structChildInfo.Filter = 0
		structChildInfo.Time = structRecordInfo.CreateTime
		structChildInfo.Attendance_day = timeSplitArray[0]
		structChildInfo.Id = structRecordInfo.Id
		structChildInfo.Attendance_time = structRecordInfo.AttendanceTime

		structReply.Data.Items = append(structReply.Data.Items, structChildInfo)

	}

	//返回数据
	structReply.Code = 20000
	structReply.Message = ""
	responseByte, err := json.Marshal(structReply)
	if err != nil {
		return
	}
	c.Data["json"] = json.RawMessage(string(responseByte))
	c.ServeJSON()
}

//智能电闸
type DeviceSwitchController struct {
	beego.Controller
}

type DevcieSwitchInfo_Json struct {
	Total int64   `json:"total"`
	Items []int64 `json:"items"`
}

type DeviceSwitchReply_Json struct {
	Code    int64                 `json:"code"`
	Message string                `json:"message"`
	Data    DevcieSwitchInfo_Json `json:"data"`
}

func (c *DeviceSwitchController) Get() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	var structReplyInfo DeviceSwitchReply_Json

	structReplyInfo.Code = 20000
	structReplyInfo.Message = "获取电闸列表成功"
	structReplyInfo.Data.Total = 0
	structReplyInfo.Data.Items = make([]int64, 0)

	byteReply, err := json.Marshal(structReplyInfo)
	if err != nil {
		beego.Error("获取电闸列表错误", err)
		return
	}

	c.Data["json"] = json.RawMessage(byteReply)
	c.ServeJSON()
}

type GetAttendanceStaticController struct {
	beego.Controller
}

type AttendanceStaticInof_Json struct {
	Names     []string  `json:"names"`
	Series    [][]int64 `json:"series"`
	XAxisData []string  `json:"xAxisData"`
}

type GetAttendanceStaticReply_Json struct {
	Code    int64                     `json:"code"`
	Message string                    `json:"message"`
	Data    AttendanceStaticInof_Json `json:"data"`
}

func (c *GetAttendanceStaticController) Get() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	var structAttendanceInfo AttendanceStaticInof_Json
	structAttendanceInfo.Names = make([]string, 0)
	structAttendanceInfo.XAxisData = make([]string, 0)
	structAttendanceInfo.Series = make([][]int64, 0)

	var arrayCount []int64
	arrayCount = make([]int64, 0)

	timeNow := time.Now()
	timeLastMonth := timeNow.AddDate(0, -1, 0)
	timeIndex := timeLastMonth

	timeLesSpan, err := time.ParseDuration("0s")
	if err != nil {
		beego.Error(err)
		return
	}

	XToken := c.Ctx.Input.Header("X-Token")

	strAccount, err := my_aes.Dncrypt(string(XToken))
	if err != nil {
		beego.Error("获取登录账户失败", err)
		return
	}

	structAdminInfo, err := my_db.GetAdminByAccount(strAccount)
	if err != nil {
		beego.Error(err)
		return
	}

	for {
		spanindex := timeNow.Sub(timeIndex)

		if spanindex < timeLesSpan {
			break
		}

		//beego.Debug("时间:", timeIndex.Format(rule_algorithm.TIME_LAYOUT))

		strTimeIndex := timeIndex.Format(rule_algorithm.TIME_LAYOUT)
		strArray := strings.Split(strTimeIndex, " ")
		strTimeBegin := strArray[0] + "T" + "00:00:00"
		strTimeEnd := strArray[0] + "T" + "23:59:59"

		strIndexArray := strings.Split(strArray[0], "-")
		strMonthDay := strIndexArray[1] + "-" + strIndexArray[2]

		nCount, err := my_db.GetVerifyRecordCountByDate(structAdminInfo.CompanyId, strTimeBegin, strTimeEnd)
		if err != nil {
			beego.Error("获取月份内考勤人数错误", err)
			return
		}

		structAttendanceInfo.XAxisData = append(structAttendanceInfo.XAxisData, strMonthDay)
		arrayCount = append(arrayCount, nCount)

		timeIndex = timeIndex.AddDate(0, 0, 1)
	}

	structAttendanceInfo.Series = append(structAttendanceInfo.Series, arrayCount)

	structCompanyInfo, err := my_db.GetCompnyById(1)
	if err != nil {
		beego.Error("获取公司统计数据错误", err)
		return
	}

	structAttendanceInfo.Names = append(structAttendanceInfo.Names, structCompanyInfo.Name)

	var structReply GetAttendanceStaticReply_Json
	structReply.Code = 20000
	structReply.Message = ""
	structReply.Data = structAttendanceInfo

	byteReply, err := json.Marshal(structReply)
	if err != nil {
		beego.Error("获取月份内考勤人数错误", err)
		return
	}

	c.Data["json"] = json.RawMessage(string(byteReply))
	c.ServeJSON()
}

///api/management/dashboard/deviceMap
type GetDeviceMapContoller struct {
	beego.Controller
}

type DeviceMapInfo_Json struct {
	Name  string `json:"name"`
	Value int64  `json:"value"`
}

type DeviceMapReply_Json struct {
	Code    int64                `json:"code"`
	Message string               `json:"message"`
	Data    []DeviceMapInfo_Json `json:"data"`
}

func (c *GetDeviceMapContoller) Get() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	XToken := c.Ctx.Input.Header("X-Token")

	strAccount, err := my_aes.Dncrypt(string(XToken))
	if err != nil {
		beego.Error("获取登录账户失败", err)
		return
	}

	structAdminInfo, err := my_db.GetAdminByAccount(strAccount)
	if err != nil {
		beego.Error(err)
		return
	}

	var structReply DeviceMapReply_Json
	structReply.Data = make([]DeviceMapInfo_Json, 0)

	arrayDevice, err := my_db.GetDeviceList(structAdminInfo.CompanyId)
	if err != nil {
		beego.Error("获取设备地图错误", err)
	}

	for _, _ = range arrayDevice {
		var structDeviceInfo_ DeviceMapInfo_Json

		structDeviceInfo_.Name = "安徽"
		structDeviceInfo_.Value = 4

		structReply.Data = append(structReply.Data, structDeviceInfo_)
	}

	structReply.Code = 20000
	structReply.Message = ""

	byteReply, err := json.Marshal(structReply)
	if err != nil {
		beego.Error(err)
		return
	}

	c.Data["json"] = json.RawMessage(string(byteReply))
	c.ServeJSON()
}

///api/management/dashboard/newIOData
type GetNewIODataController struct {
	beego.Controller
}

type NewIODataInfo_Json struct {
	Device      string `json:"device"`
	Person      string `json:"person"`
	Pic         string `json:"pic"`
	Time        string `json:"time"`
	Temperature string `json:"temperature"`
	Alarm       int64  `json:"alarm"`
}

type NewIODataReply_Json struct {
	Code    int64                `json:"code"`
	Message string               `json:"message"`
	Data    []NewIODataInfo_Json `json:"data"`
}

func (c *GetNewIODataController) Get() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	XToken := c.Ctx.Input.Header("X-Token")

	strAccount, err := my_aes.Dncrypt(string(XToken))
	if err != nil {
		beego.Error("获取登录账户失败", err)
		return
	}

	structAdminInfo, err := my_db.GetAdminByAccount(strAccount)
	if err != nil {
		beego.Error(err)
		return
	}

	arrayVerify, err := my_db.GetVerifyRecordList(structAdminInfo.CompanyId, true, 1, 20)
	if err != nil {
		beego.Error("获取新的抓拍记录错误", err)
		return
	}

	var structReply NewIODataReply_Json
	structReply.Data = make([]NewIODataInfo_Json, 0)

	for _, structVerifyIndex := range arrayVerify {
		var structIODataInfo NewIODataInfo_Json
		structIODataInfo.Person = structVerifyIndex.Name
		structIODataInfo.Pic = structVerifyIndex.VerifyPic
		arrayCreateTime := strings.Split(structVerifyIndex.CreateTime, "T")
		arrayCreateTime_ := strings.Split(arrayCreateTime[1], "Z")
		structVerifyIndex.CreateTime = arrayCreateTime[0] + " " + arrayCreateTime_[0]
		timeCreate, err := time.Parse(rule_algorithm.TIME_LAYOUT, structVerifyIndex.CreateTime)
		if err != nil {
			beego.Error(err)
			return
		}
		structIODataInfo.Time = timeCreate.Format(rule_algorithm.TIME_LAYOUT)
		structIODataInfo.Temperature = structVerifyIndex.Template
		if len(structVerifyIndex.Template) > 0 {
			fTemperature, err := strconv.ParseFloat(structVerifyIndex.Template, len(structVerifyIndex.Template))
			if err != nil {
				beego.Error(err)
				return
			}

			if fTemperature > 37.7 {
				structIODataInfo.Alarm = 1
			} else {
				structIODataInfo.Alarm = 0
			}
		}

		structDeviceInfo, err := my_db.GetDeviceByDeviceId(fmt.Sprintf("%d", structVerifyIndex.DeviceId))
		if err != nil {
			beego.Error("获取抓拍记录错误", err)
			return
		}
		structIODataInfo.Device = structDeviceInfo.Name
		structReply.Data = append(structReply.Data, structIODataInfo)
	}

	structReply.Code = 20000
	structReply.Message = ""

	byteReply, err := json.Marshal(structReply)
	if err != nil {
		beego.Error("获取抓拍记录错误", err)
		return
	}
	c.Data["json"] = json.RawMessage(string(byteReply))
	c.ServeJSON()

}

///api/management/dashboard/statusStatistics
type GetStatusStatisticContoller struct {
	beego.Controller
}

type StatusStatisticInfo_Json struct {
	Name  string `json:"name"`
	Value int64  `json:"value"`
}

type StatusStatisticReply_Json struct {
	Code    int64                      `json:"code"`
	Data    []StatusStatisticInfo_Json `json:"data"`
	Message string                     `json:"message"`
}

func (c *GetStatusStatisticContoller) Get() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	XToken := c.Ctx.Input.Header("X-Token")

	strAccount, err := my_aes.Dncrypt(string(XToken))
	if err != nil {
		beego.Error("获取登录账户失败", err)
		return
	}

	structAdminInfo, err := my_db.GetAdminByAccount(strAccount)
	if err != nil {
		beego.Error(err)
		return
	}

	arrayRecord, err := my_db.GetAllRecordListNoStranger(structAdminInfo.CompanyId)
	if err != nil {
		beego.Error("获取考勤统计数据错误", err)
		return
	}

	mapAttendanceResultAndNum := make(map[string]int64, 0)

	for _, structRecordInfo := range arrayRecord {
		strAttendanceResult := structRecordInfo.Result

		mapAttendanceResultAndNum[strAttendanceResult]++
	}

	var structReply StatusStatisticReply_Json
	structReply.Code = 20000
	structReply.Message = ""

	for keyRecord, valueRecord := range mapAttendanceResultAndNum {
		structReply.Data = append(structReply.Data, StatusStatisticInfo_Json{Name: keyRecord, Value: valueRecord})
	}

	byteReply, err := json.Marshal(structReply)
	if err != nil {
		beego.Error("获取考勤统计数据错误", err)
		return
	}

	c.Data["json"] = json.RawMessage(string(byteReply))
	c.ServeJSON()
}

///api/management/dashboard/statis
type GetDashboardController struct {
	beego.Controller
}

type DashboardInfo_Json struct {
	Alarm   int64 `json:"alarm"`
	Offline int64 `json:"offline"`
	Online  int64 `json:"online"`
	Total   int64 `json:"total"`
}

type DashboardReply_Json struct {
	Code    int64              `json:"code"`
	Data    DashboardInfo_Json `json:"data"`
	Message string             `json:"message"`
}

func (c *GetDashboardController) Get() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	XToken := c.Ctx.Input.Header("X-Token")
	strAccount, err := my_aes.Dncrypt(string(XToken))
	if err != nil {
		beego.Error("获取设备统计数据错误", err)
		return
	}

	structAdminInfo, err := my_db.GetAdminByAccount(strAccount)
	if err != nil {
		beego.Error(err)
		return
	}

	arrayDeviceInfo, err := my_db.GetDeviceList(structAdminInfo.CompanyId)
	if err != nil {
		beego.Error("获取设备统计数据错误", err)
		return
	}

	var structDashInfo DashboardInfo_Json
	structDashInfo.Alarm = 0
	structDashInfo.Offline = 0
	structDashInfo.Online = 0
	structDashInfo.Total = 0

	for _, structDeviceInfo := range arrayDeviceInfo {
		_, err := facedevice.GetDeviceInfo(structDeviceInfo)
		if err != nil {
			beego.Warning("获取设备统计数据警告", err)
			structDashInfo.Offline++
		} else {
			structDashInfo.Online++
		}

		structDashInfo.Total++
	}

	var structReply DashboardReply_Json
	structReply.Code = 20000
	structReply.Data = structDashInfo
	structReply.Message = ""

	byteReply, err := json.Marshal(structReply)
	if err != nil {
		beego.Error("获取考勤统计数据错误", err)
		return
	}

	c.Data["json"] = json.RawMessage(string(byteReply))
	c.ServeJSON()
}

///api/management/company/fetch
type GetCompnyFetchContoller struct {
	beego.Controller
}

type LocationInfo_Json struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type CompnyFetchInfo_Json struct {
	Address    string      `json:"address"`
	Name       string      `json:"name"`
	Phone      string      `json:"phone"`
	Loaction   interface{} `json:"location"`
	ID         interface{} `json:"id"`
	Multi_cert bool        `json:"multi_cert"`
}

type CompnyFetchReply struct {
	Code    int64                `json:"code"`
	Message string               `json:"message"`
	Data    CompnyFetchInfo_Json `json:"data"`
}

func (c *GetCompnyFetchContoller) Get() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	Xtoken := c.Ctx.Input.Header("X-Token")
	strAccount, err := my_aes.Dncrypt(string(Xtoken))
	if err != nil {
		beego.Error("获取公司信息错误", err)
		return
	}

	structAdminInfo, err := my_db.GetAdminByAccount(strAccount)
	if err != nil {
		beego.Error("获取公司信息错误", err)
		return
	}

	var structReply CompnyFetchReply
	structReply.Code = 20000

	if structAdminInfo.CompanyId > 0 {
		strID := c.Input().Get("id")

		nID := int64(0)
		if len(strID) > 0 {
			nID, err = strconv.ParseInt(strID, 0, 64)
			if err != nil {
				beego.Error(nID)
				return
			}
		} else {
			nID = structAdminInfo.CompanyId
		}

		structCompanyInfo, err := my_db.GetCompnyById(nID)
		if err != nil {
			beego.Error("获取公司信息错误", err)
			return
		}

		structReply.Data.Address = structCompanyInfo.Address
		structReply.Data.Name = structCompanyInfo.Name
		structReply.Data.Phone = structCompanyInfo.Phone
		if structCompanyInfo.MultiCer > 0 {
			structReply.Data.Multi_cert = true
		} else {
			structReply.Data.Multi_cert = false
		}

		structReply.Message = confreader.GetValue("SucRequest")

	} else {

		structReply.Message = confreader.GetValue("NoRegCompany")
	}

	btyeReply, err := json.Marshal(structReply)
	if err != nil {
		beego.Error("获取公司信息错误", err)
		return
	}
	c.Data["json"] = json.RawMessage(string(btyeReply))
	c.ServeJSON()

}

///api/management/company/update
type CompanyUpdateContoller struct {
	beego.Controller
}

type CompanyUpdateReply_Json struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

func (c *CompanyUpdateContoller) Post() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	var structRequst CompnyFetchInfo_Json
	data := c.Ctx.Input.RequestBody
	Xtoken := c.Ctx.Input.Header("X-Token")
	strAccount, err := my_aes.Dncrypt(string(Xtoken))
	if err != nil {
		beego.Error("更新公司信息错误", err)
		return
	}

	err = json.Unmarshal(data, &structRequst)
	if err != nil {
		beego.Error("更新公司信息错误", err)
		return
	}

	structAdminInfo, err := my_db.GetAdminByAccount(strAccount)
	if err != nil {
		beego.Error("更新公司信息错误", err)
		return
	}

	if structAdminInfo.ISRoot <= 0 {
		var structReplyError CompanyUpdateReply_Json
		structReplyError.Code = 50000
		structReplyError.Message = confreader.GetValue("NeedRoot")
		structReplyError.Data = ""

		byteReplyerr, err := json.Marshal(structReplyError)
		if err != nil {
			beego.Error("更新公司信息错误", err)
			return
		}
		c.Data["json"] = json.RawMessage(string(byteReplyerr))
		c.ServeJSON()

		return
	}

	var structUpdateInfo my_db.CompanyInfo
	structUpdateInfo.Address = structRequst.Address
	structUpdateInfo.Name = structRequst.Name
	structUpdateInfo.Phone = structRequst.Phone
	if structRequst.Multi_cert {
		structUpdateInfo.MultiCer = 1
	} else {
		structUpdateInfo.MultiCer = 0
	}

	//if len(structRequst.ID) > 0 {

	//}

	if structRequst.ID != nil {
		switch reflect.TypeOf(structRequst.ID).Kind() {
		case reflect.Float64:
			structUpdateInfo.Id = int64(structRequst.ID.(float64))
			break
		case reflect.String:
			{
				nID, err := strconv.ParseInt(structRequst.ID.(string), 0, 64)
				if err != nil {
					beego.Error(err)
					return
				}
				structUpdateInfo.Id = nID
			}
			break

		}

	}

	if err != nil {
		beego.Error(err)
		return
	}

	switch structRequst.Loaction.(type) {
	case string:
		break
	default:
		if structRequst.Loaction != nil {
			fLat, err := jsonparser.GetFloat(data, "location", "lat")
			if err != nil {
				beego.Error(err)
				return
			}
			fLng, err := jsonparser.GetFloat(data, "location", "lng")
			if err != nil {
				beego.Error(err)
				return
			}
			structUpdateInfo.Location = fmt.Sprintf("%f,%f", fLat, fLng)
		}
		break
	}

	mapReply := make(map[string]interface{}, 0)
	mapDataInfo := make(map[string]interface{}, 0)

	mapDataInfo["multi_cert"] = structRequst.Multi_cert

	if structUpdateInfo.Id <= 0 {
		nId, err := my_db.InsertCompanyInfo(structUpdateInfo)
		if err != nil {
			beego.Error("更新公司信息错误", err)
			return
		}
		if structAdminInfo.CompanyId <= 0 {
			structAdminInfo.CompanyId = nId
		}

		if len(structAdminInfo.Companys) > 0 {
			structAdminInfo.Companys += fmt.Sprintf(",%d", nId)
		} else {
			structAdminInfo.Companys += fmt.Sprintf("%d", nId)
		}

		err = my_db.UpdateAdminInfo(structAdminInfo)
		if err != nil {
			beego.Error("更新公司信息错误", err)
			return
		}

		mapDataInfo["id"] = nId
	} else {
		err := my_db.UpdateCompanyInfo(structUpdateInfo)
		if err != nil {
			beego.Error("更新公司信息错误", err)
			return
		}

		mapDataInfo["id"] = structRequst.ID
	}

	mapReply["code"] = 20000
	mapReply["message"] = confreader.GetValue("SucRequest")
	mapReply["data"] = mapDataInfo

	byteReply, err := json.Marshal(mapReply)
	if err != nil {
		beego.Error(err)
		return
	}

	c.Data["json"] = json.RawMessage(byteReply)
	c.ServeJSON()
}

///api/management/attendance/export
type AttendanceExport struct {
	beego.Controller
}

type AttendanceExportInfo_Json struct {
	Url string `json:"url"`
}

type AttendanceExportReply_Json struct {
	Code    int64                     `json:"code"`
	Data    AttendanceExportInfo_Json `json:"data"`
	Message string                    `json:"message"`
}

func (c *AttendanceExport) Get() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	//Xtoken := c.Ctx.Input.Header("X-Token")
	//Account, err := my_aes.Dncrypt(Xtoken)

	Page := c.Input().Get("page")
	Limit := c.Input().Get("limit")
	Keyword := c.Input().Get("keyword")
	Unknow := c.Input().Get("unknow")
	//Alarm := c.Input().Get("alarm")
	Date := ""
	DateEnd := ""
	arrayDate, key := c.Ctx.Request.Form["date[]"]
	if key && len(arrayDate) > 0 {
		Date = c.Ctx.Request.Form["date[]"][0]
	}

	if key && len(arrayDate) > 1 {
		DateEnd = c.Ctx.Request.Form["date[]"][1]
	}
	//beego.Debug("北京时间:", strTimeBeijin)
	bSearchByContent := false
	var structSearch my_db.SearchStruct
	if len(Date) != 0 && len(DateEnd) != 0 {
		strArray := strings.Split(string(Date), "T")
		strArray_ := strings.Split(strArray[1], "Z")

		strTimeIndex := strArray[0] + " " + strArray_[0]
		timeData, err := time.Parse(rule_algorithm.TIME_LAYOUT, strTimeIndex)
		if err != nil {
			beego.Error("获取考勤记录错误", err)
			return
		}

		strArray = strings.Split(string(DateEnd), "T")
		strArray_ = strings.Split(strArray[1], "Z")

		strTimeIndex = strArray[0] + " " + strArray_[0]
		timeDateEnd, err := time.Parse(rule_algorithm.TIME_LAYOUT, strTimeIndex)
		if err != nil {
			beego.Error(err)
			return
		}

		jetLag, err := time.ParseDuration("8h")
		timeBeijin := timeData.Add(jetLag)
		strTimeBeijin := timeBeijin.Format(rule_algorithm.TIME_LAYOUT)

		strBeijinTimeArray := strings.Split(strTimeBeijin, " ")
		structSearch.TimeBegin = strBeijinTimeArray[0] + "T" + strBeijinTimeArray[1] + "Z"

		timeBeijin = timeDateEnd.Add(jetLag)
		strTimeBeijin = timeBeijin.Format(rule_algorithm.TIME_LAYOUT)
		strBeijinTimeArray = strings.Split(strTimeBeijin, " ")
		structSearch.TimeEnd = strBeijinTimeArray[0] + "T" + strBeijinTimeArray[1] + "Z"

		//structSearch.TimeEnd = strBeijinTimeArray[0] + "T" + "23:59:59Z"
		bSearchByContent = true
	}

	if len(Keyword) != 0 {
		structSearch.SearchContent = string(Keyword)
		bSearchByContent = true
	}

	bUnknow, err := strconv.ParseBool(string(Unknow))
	if err != nil {
		beego.Error("获取考勤列表错误", err)
		return
	}

	nPage, err := strconv.ParseInt(string(Page), 0, 64)
	if err != nil {
		beego.Error("获取考勤列表错误", err)
		return
	}

	nLimit, err := strconv.ParseInt(string(Limit), 0, 64)
	if err != nil {
		beego.Error("获取考勤列表错误", err)
		return
	}

	var structRecordArray []my_db.VerifyRecordInfo
	var err_ error

	var structReply AttendanceListReply_Json
	structReply.Data.Items = make([]AttendanceInfo_Json, 0)

	XToken := c.Ctx.Input.Header("X-Token")

	strAccount, err := my_aes.Dncrypt(string(XToken))
	if err != nil {
		beego.Error("获取登录账户失败", err)
		return
	}

	structAdminInfo, err := my_db.GetAdminByAccount(strAccount)
	if err != nil {
		beego.Error(err)
		return
	}

	if len(structSearch.TimeBegin) == 0 || len(structSearch.TimeEnd) == 0 {
		structSearch.TimeBegin, err = my_db.GetEarlyDateFromRecord(structAdminInfo.CompanyId)
		if err != nil {
			beego.Error(err)
			return
		}

		strTimeEnd := time.Now().Format(rule_algorithm.TIME_LAYOUT)
		arrayEndIndex := strings.Split(strTimeEnd, " ")
		structSearch.TimeEnd = arrayEndIndex[0] + "T" + arrayEndIndex[1] + "Z"
	}
	arrayIndex := strings.Split(structSearch.TimeBegin, "T")
	timeBegin, err := time.Parse(rule_algorithm.TIME_LAYOUT, arrayIndex[0]+" "+"00:00:00")
	if err != nil {
		beego.Error(err)
		return
	}

	arrayIndex = strings.Split(structSearch.TimeEnd, "T")
	timeEnd, err := time.Parse(rule_algorithm.TIME_LAYOUT, arrayIndex[0]+" "+"23:59:59")

	//处理
	if bSearchByContent {
		structRecordArray, err_ = my_db.GetVerifyRecordBySearchStruct(structAdminInfo.CompanyId, structSearch, !bUnknow, nPage, nLimit)
		if err_ != nil {
			beego.Error("获取考勤记录列表错误", err_)
			return
		}

		nTotalIndex, err_ := my_db.GetTotalVerifyRecordBySearchStruct(structAdminInfo.CompanyId, structSearch, !bUnknow)
		if err_ != nil {
			beego.Error("获取考勤记录列表错误", err_)
			return
		}
		structReply.Data.Total = nTotalIndex
	} else {
		structRecordArray, err_ = my_db.GetVerifyRecordList(structAdminInfo.CompanyId, !bUnknow, nPage, nLimit)
		if err_ != nil {
			beego.Error("获取考勤记录列表错误", err_)
			return
		}
		nTotalIndex, err_ := my_db.GetTotalVerifyRecordList(structAdminInfo.CompanyId, !bUnknow)
		if err != nil {
			beego.Error("获取考勤记录列表错误", err_)
			return
		}
		structReply.Data.Total = nTotalIndex
	}

	mapDateRecord := make(map[string]map[string][]my_db.VerifyRecordInfo, 0)

	for _, valueRecord := range structRecordArray {
		_, keyIndex := mapDateRecord[valueRecord.Name]
		if !keyIndex {
			mapDateRecord[valueRecord.Name] = make(map[string][]my_db.VerifyRecordInfo, 0)
		}

		if mapDateRecord[valueRecord.Name][valueRecord.CreateDate] == nil {
			mapDateRecord[valueRecord.Name][valueRecord.CreateDate] = make([]my_db.VerifyRecordInfo, 0)
		}

		mapDateRecord[valueRecord.Name][valueRecord.CreateDate] = append(mapDateRecord[valueRecord.Name][valueRecord.CreateDate], valueRecord)
	}

	t := make([]string, 0)
	t = append(t, "ID")
	t = append(t, confreader.GetValue("Name"))
	t = append(t, confreader.GetValue("AttendanceDay"))
	t = append(t, confreader.GetValue("AttendanceTime"))
	t = append(t, confreader.GetValue("AttendanceResult"))
	t = append(t, confreader.GetValue("RecordTime"))
	t = append(t, confreader.GetValue("Device"))
	file := xlsx.NewFile()
	sheet, err := file.AddSheet(confreader.GetValue("OriginRecord"))
	if err != nil {
		beego.Error("导出xls文件错误", err)
		return
	}
	titleRow := sheet.AddRow()
	xlsRow := xlsxtool.NewRow(titleRow, t)
	err = xlsRow.SetRowTitle()

	for _, structRecordInfo := range structRecordArray {
		currentRow := sheet.AddRow()
		tmp := make([]string, 0)
		tmp = append(tmp, fmt.Sprintf("%d", structRecordInfo.Id))
		tmp = append(tmp, structRecordInfo.Name)

		arrayIndex := strings.Split(structRecordInfo.CreateTime, "T")

		tmp = append(tmp, arrayIndex[0])
		tmp = append(tmp, structRecordInfo.AttendanceTime)
		tmp = append(tmp, structRecordInfo.Result)
		tmp = append(tmp, structRecordInfo.CreateTime)
		tmp = append(tmp, structRecordInfo.DeviceName)

		xlsRow := xlsxtool.NewRow(currentRow, tmp)
		err := xlsRow.GenerateRow()
		if err != nil {
			beego.Error("导出xls文件错误", err)
			return
		}
	}

	sheetPersonDayAttendance, err := file.AddSheet(confreader.GetValue("PunchTime"))
	if err != nil {
		beego.Error(err)
		return
	}

	tPersonDay := make([]string, 0)

	if bSearchByContent {
		strMainTittle := confreader.GetValue("TablePunchTime") + fmt.Sprintf(", %s - %s", timeBegin.Format(rule_algorithm.TIME_LAYOUT), timeEnd.Format(rule_algorithm.TIME_LAYOUT))
		tPersonDay = append(tPersonDay, strMainTittle)
	} else {
		tPersonDay = append(tPersonDay, confreader.GetValue("TablePunchTime"))
	}

	tPersonDayRow := sheetPersonDayAttendance.AddRow()
	xlsPersonDayRow := xlsxtool.NewRow(tPersonDayRow, tPersonDay)
	styleTitle := xlsx.NewStyle()
	styleTitle.Font.Size = 30
	err = xlsPersonDayRow.GenerateRowByStyle(styleTitle)
	if err != nil {
		beego.Error(err)
		return
	}

	arraySecondTitle := make([]string, 0)
	strSecondTitle := confreader.GetValue("ReportGenerationDate") + fmt.Sprintf(": %s", time.Now().Format(rule_algorithm.TIME_LAYOUT))
	arraySecondTitle = append(arraySecondTitle, strSecondTitle)

	tSencondTitleRow := sheetPersonDayAttendance.AddRow()
	xlsSecondRow := xlsxtool.NewRow(tSencondTitleRow, arraySecondTitle)
	styleTitle.Font.Size = 20
	err = xlsSecondRow.GenerateRowByStyle(styleTitle)

	arrayDataTitle := make([]string, 0)
	arrayDataTitle = append(arrayDataTitle, confreader.GetValue("Name"))
	arrayDataTitle = append(arrayDataTitle, confreader.GetValue("Organization"))

	arrayArrayData := make([][]string, 0)

	timeIndex := timeBegin
	for timeIndex.Before(timeEnd) {
		strTimeForm := timeIndex.Format(rule_algorithm.TIME_LAYOUT)
		arrayIndex := strings.Split(strTimeForm, " ")
		arrayIndex_ := strings.Split(arrayIndex[0], "-")
		strTimeIndex := arrayIndex_[1] + "-" + arrayIndex_[2]
		if timeIndex.Weekday() == 6 {
			arrayDataTitle = append(arrayDataTitle, strTimeIndex+"("+confreader.GetValue("Saturday")+")")
		} else if timeIndex.Weekday() == 0 {
			arrayDataTitle = append(arrayDataTitle, strTimeIndex+"("+confreader.GetValue("Sunday")+")")
		} else {
			arrayDataTitle = append(arrayDataTitle, strTimeIndex)
		}

		timeIndex = timeIndex.AddDate(0, 0, 1)
	}

	for _, valueMapData := range mapDateRecord {
		arrayDataIndex := make([]string, 0)
		for _, valueRecordArray := range valueMapData {
			for _, valueRecordInfo := range valueRecordArray {
				arrayDataIndex = append(arrayDataIndex, valueRecordInfo.Name)
				arrayGroup, err := my_db.GetGroupByEmployee(valueRecordInfo.CutomId)
				if err != nil {
					beego.Error(err)
					return
				}

				strGroups := ""
				for _, valueGroup := range arrayGroup {
					structGroupInfo, err := my_db.GetGroupById(valueGroup)
					if err != nil {
						beego.Error(err)
						return
					}
					strGroups += structGroupInfo.Name + "\r\n"
				}
				arrayDataIndex = append(arrayDataIndex, strGroups)
				break
			}
			break
		}

		timeIndex = timeBegin

		for timeIndex.Before(timeEnd) {
			strTimeFormt := timeIndex.Format(rule_algorithm.TIME_LAYOUT)
			arrayIndex := strings.Split(strTimeFormt, " ")

			_, keyDateRecord := valueMapData[arrayIndex[0]]
			if keyDateRecord {
				strCreateTimes := ""
				for _, valueDateRecordInfo := range valueMapData[arrayIndex[0]] {
					arrayCreateTime := strings.Split(valueDateRecordInfo.CreateTime, "T")
					arrayCreateTime_ := strings.Split(arrayCreateTime[1], "Z")

					strCreateTimes += (arrayCreateTime_[0] + " \r\n")
					//arrayDataIndex = append(arrayDataIndex, arrayCreateTime_[0])
				}
				arrayDataIndex = append(arrayDataIndex, strCreateTimes)
			} else {
				arrayDataIndex = append(arrayDataIndex, "")
			}

			timeIndex = timeIndex.AddDate(0, 0, 1)
		}
		arrayArrayData = append(arrayArrayData, arrayDataIndex)
	}

	tDataRow := sheetPersonDayAttendance.AddRow()
	xlsDataRow := xlsxtool.NewRow(tDataRow, arrayDataTitle)
	err = xlsDataRow.GenerateRow()
	if err != nil {
		beego.Error(err)
		return
	}

	for _, valueArrayIndex := range arrayArrayData {
		tRecordDataRow := sheetPersonDayAttendance.AddRow()
		tRecordDataRow.SetHeightCM(3)
		xlsRecorDataRow := xlsxtool.NewRow(tRecordDataRow, valueArrayIndex)
		style := xlsx.NewStyle()
		style.Alignment.WrapText = true
		err = xlsRecorDataRow.GenerateRowByStyle(style)
		if err != nil {
			beego.Error(err)
			return
		}
	}

	strPath := my_db.GetRootPath() + "/manager/考勤记录.xlsx"
	err = file.Save(strPath)
	if err != nil {
		beego.Error("导出xls文件错误", err)
		return
	}

	var structReply_ AttendanceExportReply_Json
	structReply_.Code = 20000
	structReply_.Message = "导出成功"
	structReply_.Data.Url = c.Ctx.Input.Referer() + "manager/考勤记录.xlsx"

	byteReply, err := json.Marshal(structReply_)
	if err != nil {
		beego.Error("导出考勤记录错误", err)
		return
	}

	c.Data["json"] = json.RawMessage(byteReply)
	c.ServeJSON()
}

//RuleSaveTakeTurnsController __
type RuleSaveTakeTurnsController struct {
	beego.Controller
}

//ruleTakeTurnsRequst_Json __
type ruleTakeTurnsRequstJSON struct {
	Begin   string                `json:"begin"`
	ID      int64                 `json:"id"`
	Name    string                `json:"name"`
	Persons map[int64]interface{} `json:"persons"` //personid  和 index
	Tags    []TagInfoJSON         `json:"tags"`
}

// structReply.Data.Tags = structRequest.Tags
// structReply.Data.Type = 1
// structReply.Data.Name = structRequest.Name
// structReply.Data.Id = nId
// structReply.Data.Company_id = 0

type ruleTakeTurnsDataJSON struct {
	Tags       []TagInfoJSON `json:"tags"`
	Type       int64         `json:"type"`
	Name       string        `json:"name"`
	Id         int64         `json:"id"`
	Company_id int64         `json:"company_id"`
}

type ruleTakeTurnsReplyJSON struct {
	Code    int64                 `json:"code"`
	Message string                `json:"message"`
	Data    ruleTakeTurnsDataJSON `json:"data"`
}

func (c *RuleSaveTakeTurnsController) Post() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	Xtoken := c.Ctx.Input.Header("X-Token")
	strAdmin, err := my_aes.Dncrypt(string(Xtoken))
	if err != nil {
		beego.Error("添加规则错误", err)
		return
	}

	data := c.Ctx.Input.RequestBody

	var structRequest ruleTakeTurnsRequstJSON
	err = json.Unmarshal(data, &structRequest)
	if err != nil {
		beego.Error("添加轮班规则错误", err)
		return
	}

	var errReply ruleTakeTurnsReplyJSON

	if len(structRequest.Name) <= 0 {
		errReply.Code = 40000
		errReply.Message = "未输入名称"
		byteReply, err := json.Marshal(errReply)
		if err != nil {
			beego.Error(err)
			return
		}

		c.Data["json"] = json.RawMessage(string(byteReply))
		c.ServeJSON()
		return
	}

	if len(structRequest.Begin) <= 0 {
		errReply.Code = 40000
		errReply.Message = "未设置开始时间"
		byteReply, err := json.Marshal(errReply)
		if err != nil {
			beego.Error(err)
			return
		}

		c.Data["json"] = json.RawMessage(string(byteReply))
		c.ServeJSON()

		return
	}

	if len(structRequest.Tags) <= 0 {
		errReply.Code = 40000
		errReply.Message = confreader.GetValue("NoShiftDay")
		byteReply, err := json.Marshal(errReply)
		if err != nil {
			beego.Error(err)
			return
		}

		c.Data["json"] = json.RawMessage(string(byteReply))
		c.ServeJSON()
		return
	}

	if structRequest.ID > 0 {
		for _, structTagIndex := range structRequest.Tags {
			structTagInfo, err := my_db.GetTagByID(structTagIndex.Id)
			if err != nil {
				beego.Error(err)
				return
			}

			err = my_db.RemoveEmployeeTag(structTagIndex.Id)
			if err != nil {
				beego.Error(err)
				return
			}

			if structTagInfo.Range1 > 0 {
				err = my_db.RemoveRangeInfo(structTagInfo.Range1)
				if err != nil {
					beego.Error(err)
					return
				}

				err = my_db.DeleteRangeAndEmployeeByRangeID(structTagInfo.Range1)
				if err != nil {
					beego.Error(err)
					return
				}
			}

			if structTagInfo.Range2 > 0 {
				err = my_db.RemoveRangeInfo(structTagInfo.Range2)
				if err != nil {
					beego.Error(err)
					return
				}

				err = my_db.DeleteRangeAndEmployeeByRangeID(structTagInfo.Range2)
				if err != nil {
					beego.Error(err)
					return
				}
			}

			if structTagInfo.Range3 > 0 {
				err = my_db.RemoveRangeInfo(structTagInfo.Range3)
				if err != nil {
					beego.Error(err)
					return
				}

				err = my_db.DeleteRangeAndEmployeeByRangeID(structTagInfo.Range3)
				if err != nil {
					beego.Error(err)
					return
				}
			}

			err = my_db.RemoveTag(structTagInfo.ID)
			if err != nil {
				beego.Error(err)
				return
			}

		}

		err = my_db.DeleteRule(structRequest.ID)
		if err != nil {
			beego.Error("删除规则错误", err)
			return
		}

		err = my_db.RemoveEmployeeRuleID(structRequest.ID)
		if err != nil {
			beego.Error(err)
			return
		}
	}

	//处理
	var structRuleAddInfo my_db.RuleInfo

	structRuleAddInfo.Offwork_begin = "00:00:00"
	structRuleAddInfo.Offwork_end = "00:00:00"
	structRuleAddInfo.Work_begin = "00:00:00"
	structRuleAddInfo.Work_end = "00:00:00"
	structRuleAddInfo.Time1 = "00:00:00"
	structRuleAddInfo.Time2 = "00:00:00"

	structRuleAddInfo.Begin = structRequest.Begin
	structRuleAddInfo.Name = structRequest.Name
	structRuleAddInfo.OwnerAccount = strAdmin
	structRuleAddInfo.Type = 1

	mapTagAndEmployee := make(map[int64][]int64, 0)

	for nPersonID, nTagIndex := range structRequest.Persons {
		if nTagIndex != nil {
			mapTagAndEmployee[int64(nTagIndex.(float64))] = append(mapTagAndEmployee[int64(nTagIndex.(float64))], nPersonID)
		}
	}

	for nIndex, structTagIndex := range structRequest.Tags {
		var structTagInfo my_db.TagInfo
		nRangeID1 := int64(0)
		nRangeID2 := int64(0)
		nRangeID3 := int64(0)

		if len(structTagIndex.From.Range1) > 0 && structTagIndex.From.Type > 0 {
			var structRangeInfo my_db.RangeInfo
			structRangeInfo.ID = -1
			structRangeInfo.LateTime = int64(structTagIndex.From.Range1["late_time"].(float64))
			structRangeInfo.LeftEarly = int64(structTagIndex.From.Range1["left_early"].(float64))
			structRangeInfo.OffWorkBegin = structTagIndex.From.Range1["offwork_begin"].(string)
			structRangeInfo.OffWorkEnd = structTagIndex.From.Range1["offwork_end"].(string)
			if structTagIndex.From.Range1["offworkcheck"].(bool) {
				structRangeInfo.OffWorkCheck = 1
			} else {
				structRangeInfo.OffWorkCheck = 0
			}

			structRangeInfo.Time1 = structTagIndex.From.Range1["time1"].(string)
			structRangeInfo.Time2 = structTagIndex.From.Range1["time2"].(string)
			structRangeInfo.WorkBegin = structTagIndex.From.Range1["work_begin"].(string)
			structRangeInfo.WorkEnd = structTagIndex.From.Range1["work_end"].(string)
			if structTagIndex.From.Range1["workcheck"].(bool) {
				structRangeInfo.WorkCheck = 1
			} else {
				structRangeInfo.WorkCheck = 0
			}

			strArray := strings.Split(structRequest.Begin, "T")
			strArray_ := strings.Split(strArray[1], "Z")

			strTimeIndex := strArray[0] + " " + strArray_[0]
			timeData, err := time.Parse(rule_algorithm.TIME_LAYOUT, strTimeIndex)
			if err != nil {
				beego.Error("获取考勤记录错误", err)
				return
			}

			jetLag, err := time.ParseDuration("8h")
			timeBeijin := timeData.Add(jetLag)
			strTimeBeijin := timeBeijin.Format(rule_algorithm.TIME_LAYOUT)
			structRangeInfo.Begin = strTimeBeijin
			nRangeID1, err = my_db.AddRangeInfo(structRangeInfo)
			if err != nil {
				beego.Error(err)
				return
			}
		} else {
			if structTagIndex.From.Type == 1 {
				errReply.Code = 40000
				errReply.Message = confreader.GetValue("NoTimeAttendance")
				byteReply, err := json.Marshal(errReply)
				if err != nil {
					beego.Error(err)
					return
				}

				c.Data["json"] = json.RawMessage(string(byteReply))
				c.ServeJSON()
				return
			}
		}

		if len(structTagIndex.From.Range2) > 0 && structTagIndex.From.Type > 1 {
			var structRangeInfo my_db.RangeInfo
			structRangeInfo.ID = -1
			structRangeInfo.LateTime = int64(structTagIndex.From.Range2["late_time"].(float64))
			structRangeInfo.LeftEarly = int64(structTagIndex.From.Range2["left_early"].(float64))
			structRangeInfo.OffWorkBegin = structTagIndex.From.Range2["offwork_begin"].(string)
			structRangeInfo.OffWorkEnd = structTagIndex.From.Range2["offwork_end"].(string)
			if structTagIndex.From.Range2["offworkcheck"].(bool) {
				structRangeInfo.OffWorkCheck = 1
			} else {
				structRangeInfo.OffWorkCheck = 0
			}

			structRangeInfo.Time1 = structTagIndex.From.Range2["time1"].(string)
			structRangeInfo.Time2 = structTagIndex.From.Range2["time2"].(string)
			structRangeInfo.WorkBegin = structTagIndex.From.Range2["work_begin"].(string)
			structRangeInfo.WorkEnd = structTagIndex.From.Range2["work_end"].(string)
			if structTagIndex.From.Range2["workcheck"].(bool) {
				structRangeInfo.WorkCheck = 1
			} else {
				structRangeInfo.WorkCheck = 0
			}
			strArray := strings.Split(structRequest.Begin, "T")
			strArray_ := strings.Split(strArray[1], "Z")

			strTimeIndex := strArray[0] + " " + strArray_[0]
			timeData, err := time.Parse(rule_algorithm.TIME_LAYOUT, strTimeIndex)
			if err != nil {
				beego.Error("获取考勤记录错误", err)
				return
			}

			jetLag, err := time.ParseDuration("8h")
			timeBeijin := timeData.Add(jetLag)
			strTimeBeijin := timeBeijin.Format(rule_algorithm.TIME_LAYOUT)
			structRangeInfo.Begin = strTimeBeijin
			nRangeID2, err = my_db.AddRangeInfo(structRangeInfo)
			if err != nil {
				beego.Error(err)
				return
			}
		} else {
			if structTagIndex.From.Type == 2 {
				errReply.Code = 40000
				errReply.Message = confreader.GetValue("NoTimeAttendance")
				byteReply, err := json.Marshal(errReply)
				if err != nil {
					beego.Error(err)
					return
				}

				c.Data["json"] = json.RawMessage(string(byteReply))
				c.ServeJSON()
				return
			}
		}

		if len(structTagIndex.From.Range3) > 0 && structTagIndex.From.Type > 2 {
			var structRangeInfo my_db.RangeInfo
			structRangeInfo.ID = -1
			structRangeInfo.LateTime = int64(structTagIndex.From.Range3["late_time"].(float64))
			structRangeInfo.LeftEarly = int64(structTagIndex.From.Range3["left_early"].(float64))
			structRangeInfo.OffWorkBegin = structTagIndex.From.Range3["offwork_begin"].(string)
			structRangeInfo.OffWorkEnd = structTagIndex.From.Range3["offwork_end"].(string)
			if structTagIndex.From.Range3["offworkcheck"].(bool) {
				structRangeInfo.OffWorkCheck = 1
			} else {
				structRangeInfo.OffWorkCheck = 0
			}

			structRangeInfo.Time1 = structTagIndex.From.Range3["time1"].(string)
			structRangeInfo.Time2 = structTagIndex.From.Range3["time2"].(string)
			structRangeInfo.WorkBegin = structTagIndex.From.Range3["work_begin"].(string)
			structRangeInfo.WorkEnd = structTagIndex.From.Range3["work_end"].(string)
			if structTagIndex.From.Range3["workcheck"].(bool) {
				structRangeInfo.WorkCheck = 1
			} else {
				structRangeInfo.WorkCheck = 0
			}
			strArray := strings.Split(structRequest.Begin, "T")
			strArray_ := strings.Split(strArray[1], "Z")

			strTimeIndex := strArray[0] + " " + strArray_[0]
			timeData, err := time.Parse(rule_algorithm.TIME_LAYOUT, strTimeIndex)
			if err != nil {
				beego.Error("获取考勤记录错误", err)
				return
			}

			jetLag, err := time.ParseDuration("8h")
			timeBeijin := timeData.Add(jetLag)
			strTimeBeijin := timeBeijin.Format(rule_algorithm.TIME_LAYOUT)
			structRangeInfo.Begin = strTimeBeijin
			nRangeID3, err = my_db.AddRangeInfo(structRangeInfo)
			if err != nil {
				beego.Error(err)
				return
			}
		} else {
			if structTagIndex.From.Type == 3 {
				errReply.Code = 40000
				errReply.Message = confreader.GetValue("NoTimeAttendance")
				byteReply, err := json.Marshal(errReply)
				if err != nil {
					beego.Error(err)
					return
				}

				c.Data["json"] = json.RawMessage(string(byteReply))
				c.ServeJSON()
				return
			}
		}

		if nRangeID1 > 0 {
			structTagInfo.Range1 = nRangeID1
		}

		if nRangeID2 > 0 {
			structTagInfo.Range2 = nRangeID2
		}

		if nRangeID3 > 0 {
			structTagInfo.Range3 = nRangeID3
		}

		structTagInfo.ID = -1
		structTagInfo.Index = structTagIndex.Index
		if structTagIndex.Selected {
			structTagInfo.Selected = 1
		} else {
			structTagInfo.Selected = 0
		}
		if structTagIndex.Active {
			structTagInfo.Active = 1
		} else {
			structTagInfo.Active = 0
		}
		structTagInfo.Color = structTagIndex.From.Color
		structTagInfo.Name = structTagIndex.From.Name

		for _, nPersonIndex := range mapTagAndEmployee[int64(nIndex)] {
			if nRangeID1 > 0 {
				err = my_db.AddRangeAndEmployee(nRangeID1, nPersonIndex)
				if err != nil {
					beego.Error(err)
					return
				}
			}

			if nRangeID2 > 0 {
				err = my_db.AddRangeAndEmployee(nRangeID2, nPersonIndex)
				if err != nil {
					beego.Error(err)
					return
				}
			}

			if nRangeID3 > 0 {
				err = my_db.AddRangeAndEmployee(nRangeID3, nPersonIndex)
				if err != nil {
					beego.Error(err)
					return
				}
			}
		}

		structTagInfo.Type = structTagIndex.From.Type
		tagIDIndex, err := my_db.AddTagInfo(structTagInfo)
		if err != nil {
			beego.Error(err)
			return
		}

		for _, nPersonIndex := range mapTagAndEmployee[int64(nIndex)] {
			err = my_db.UpdateEmployeeTagID(nPersonIndex, tagIDIndex)
			if err != nil {
				beego.Error(err)
				return
			}
		}

		if nIndex == 0 {
			structRuleAddInfo.Tags += fmt.Sprintf("%d", tagIDIndex)
		} else {
			structRuleAddInfo.Tags += ","
			structRuleAddInfo.Tags += fmt.Sprintf("%d", tagIDIndex)
		}

	}

	nId, err := my_db.AddRule(structRuleAddInfo)
	if err != nil {
		beego.Error("添加考勤规则失败:", err)
		return
	}

	for nPersonID := range structRequest.Persons {
		err = my_db.UpdateEmployeeRuleID(nPersonID, nId)
		if err != nil {
			beego.Error(err)
			return
		}
	}

	//返回结果
	var structReply ruleTakeTurnsReplyJSON
	structReply.Code = 20000
	structReply.Message = ""
	structReply.Data.Tags = structRequest.Tags
	structReply.Data.Type = 1
	structReply.Data.Name = structRequest.Name
	structReply.Data.Id = nId
	structReply.Data.Company_id = 0

	responseByte, err := json.Marshal(structReply)
	if err != nil {
		return
	}

	c.Data["json"] = json.RawMessage(string(responseByte))
	c.ServeJSON()
}

///api/management/attendance/snap
type SnapGetListController struct {
	beego.Controller
}

type SnapeInfo_Json struct {
	Pic  string `json:"pic"`
	Time string `json:"time"`
}

type SnapeGetListReplyInfo_Json struct {
	Items []SnapeInfo_Json `json:"items"`
	Total int64            `json:"total"`
}

type SnapeGetListReply_Json struct {
	Code    int64                      `json:"code"`
	Data    SnapeGetListReplyInfo_Json `json:"data"`
	Message string                     `json:"message"`
}

func (c *SnapGetListController) Get() {

	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	XToken := c.Ctx.Input.Header("X-Token")

	strAccount, err := my_aes.Dncrypt(string(XToken))
	if err != nil {
		beego.Error("获取登录账户失败", err)
		return
	}

	structAdminInfo, err := my_db.GetAdminByAccount(strAccount)
	if err != nil {
		beego.Error(err)
		return
	}

	arrayStranger, err := my_db.GetAllRecordListStranger(structAdminInfo.CompanyId)
	if err != nil {
		beego.Error("获取陌生人列表错误", err)
		return
	}

	var structReply SnapeGetListReply_Json
	structReply.Data.Items = make([]SnapeInfo_Json, 0)
	structReply.Data.Total, err = my_db.GetTotalVerifyRecordList(structAdminInfo.CompanyId, true)
	structReply.Code = 20000
	structReply.Message = ""
	if err != nil {
		beego.Error("获取陌生人列表错误", err)
		return
	}
	for _, structStrangeInfo := range arrayStranger {
		var structStangeIndex SnapeInfo_Json
		structStangeIndex.Pic = structStrangeInfo.VerifyPic
		structStangeIndex.Time = structStrangeInfo.CreateTime

		structReply.Data.Items = append(structReply.Data.Items, structStangeIndex)
	}

	byteReply, err := json.Marshal(structReply)
	if err != nil {
		beego.Error("获取陌生人列表错误", err)
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
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	Xtoken := c.Ctx.Input.Header("X-Token")
	strAccount, err := my_aes.Dncrypt(string(Xtoken))
	if err != nil {
		beego.Error("清空设备人员数据错误", err)
		return
	}

	structAdminInfo, err := my_db.GetAdminByAccount(strAccount)
	if err != nil {
		beego.Error(err)
		return
	}

	data := c.Ctx.Input.RequestBody

	mapRequest := make(map[string]interface{}, 0)

	err = json.Unmarshal(data, &mapRequest)

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
			arrayEmoloyeId, err := my_db.GetDeviceAndEmployee(structDeviceInfo.Id)
			if err != nil {
				beego.Error("清空设备人员数据错误", err)
				return
			}

			for _, nEmployeeId := range arrayEmoloyeId.EmployeeId {
				structEmployeeInfoIndex, err := my_db.GetEmployeeById(nEmployeeId)
				if err != nil {
					beego.Error("清空设备人员数据错误", err)
					return
				}

				err = facedevice.RemoveDeviceEmployee(structDeviceInfo, structEmployeeInfoIndex)
				if err != nil {
					beego.Warning("清空设备人员数据警告", err)
				}

				err = my_db.RemoveDeviceAndEmployeeByEmployeeId(nEmployeeId)
				if err != nil {
					beego.Error("清空设备人员数据错误", err)
					return
				}

				structEmployeeInfoIndex.Status = 0
				err = my_db.UpdateEmployee(structEmployeeInfoIndex)
				if err != nil {
					beego.Error("清空设备人员数据错误", err)
					return
				}
			}
		}

		if structDeviceInfo.Type == 1 {
			var structXMRemoveInfo facedevicexiongmai.RemoveEmployeeDataJSON
			structXMRemoveInfo.EmployeeInfo.UserID = "ALL"
			structXMRemoveInfo.DeviceInfo.DeviceID = structDeviceInfo.DevcieId
			err = facedevicexiongmai.RemoveEmployeeWSS(structXMRemoveInfo)
			if err != nil {
				beego.Warning("删除人员错误", err)
			}

		}

		err = my_db.RemoveDeviceAndEmployee(nID)
		if err != nil {
			beego.Error("清空设备人员数据错误", err)
			return
		}

	} else {
		arrayDevice, err := my_db.GetDeviceList(structAdminInfo.CompanyId)
		if err != nil {
			beego.Error("清空人员数据错误", err)
			return
		}

		for _, structDeviceInfo := range arrayDevice {
			if structDeviceInfo.Type == 0 {
				arrayEmoloyeId, err := my_db.GetDeviceAndEmployee(structDeviceInfo.Id)
				if err != nil {
					beego.Error("清空设备人员数据错误", err)
					return
				}

				for _, nEmployeeId := range arrayEmoloyeId.EmployeeId {
					structEmployeeInfoIndex, err := my_db.GetEmployeeById(nEmployeeId)
					if err != nil {
						beego.Error("清空设备人员数据错误", err)
						return
					}

					err = facedevice.RemoveDeviceEmployee(structDeviceInfo, structEmployeeInfoIndex)
					if err != nil {
						beego.Warning("清空设备人员数据警告", err)
					}
				}
			}

			if structDeviceInfo.Type == 1 {
				var structXMRemoveInfo facedevicexiongmai.RemoveEmployeeDataJSON
				structXMRemoveInfo.EmployeeInfo.UserID = "ALL"
				structXMRemoveInfo.DeviceInfo.DeviceID = structDeviceInfo.DevcieId
				err = facedevicexiongmai.RemoveEmployeeWSS(structXMRemoveInfo)
				if err != nil {
					beego.Warning("删除人员错误", err)
				}

			}

			err = my_db.RemoveDeviceAndEmployee(structDeviceInfo.Id)
			if err != nil {
				beego.Error("清空设备人员数据错误", err)
				return
			}

		}
	}

	var structReply DeviceTruncateReply_Json
	structReply.Code = 20000
	structReply.Data = ""
	structReply.Message = confreader.GetValue("SucRequest")

	byteReply, err := json.Marshal(structReply)
	if err != nil {
		beego.Error("清空设备人员数据错误", err)
		return
	}
	c.Data["json"] = json.RawMessage(string(byteReply))
	c.ServeJSON()
}

//GetAdminRightController 获取管理员权限
type GetAdminRightController struct {
	beego.Controller
}

//AdminRightInfoJSON 管理员权限信息
type AdminRightInfoJSON struct {
	Attendance        int64  `json:"attendance"`
	Circuitbreaker    int64  `json:"circuitbreaker"`
	Face              int64  `json:"face"`
	ID                int64  `json:"id"`
	IsAdmin           int64  `json:"is_admin"`
	IsInit            int64  `json:"is_init"`
	OrganizationLimit int64  `json:"organization_limit"`
	Person            int64  `json:"person"`
	PersonID          int64  `json:"person_id"`
	Remotedoor        string `json:"remotedoor"`
}

//GetAdminRightReplyJSON 获取管理员指令
type GetAdminRightReplyJSON struct {
	Code    int64              `json:"code"`
	Data    AdminRightInfoJSON `json:"data"`
	Message string             `json:"message"`
}

//Post  获取管理员权限
func (c *GetAdminRightController) Post() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	Xtoken := c.Ctx.Input.Header("X-Token")
	strAccount, err := my_aes.Dncrypt(string(Xtoken))
	if err != nil {
		beego.Error("清空设备人员数据错误", err)
		return
	}

	structAdminInfo, err := my_db.GetAdminByAccount(strAccount)
	if err != nil {
		beego.Error(err)
		return
	}

	structEmployeeInfo, err := my_db.GetEmployeeByAdminAccount(structAdminInfo.Account, structAdminInfo.CompanyId)
	if err != nil {
		beego.Error(err)
		return
	}

	var structReply GetAdminRightReplyJSON

	arrayRight := strings.Split(structEmployeeInfo.Rights, ",")
	for _, strRight := range arrayRight {
		if strRight == "person" {
			structReply.Data.Person = 1
		}

		if strRight == "attendance" {
			structReply.Data.Attendance = 1
		}

		if strRight == "face" {
			structReply.Data.Face = 1
		}

		if strRight == "circuitbreaker" {
			structReply.Data.Circuitbreaker = 1
		}
	}

	if structAdminInfo.ISRoot > 0 {
		structReply.Data.Person = 1
		structReply.Data.Attendance = 1
		structReply.Data.Face = 1
		structReply.Data.Circuitbreaker = 1
	}

	structReply.Data.IsAdmin = 1

	structReply.Code = 20000
	structReply.Message = ""

	structReply.Data.ID = 3
	structReply.Data.IsInit = 0
	structReply.Data.OrganizationLimit = 0

	structReply.Data.PersonID = structEmployeeInfo.Id
	structReply.Data.Remotedoor = structEmployeeInfo.RemoteDevice

	byteReply, err := json.Marshal(structReply)
	if err != nil {
		beego.Error(err)
		return
	}

	c.Data["json"] = json.RawMessage(string(byteReply))
	c.ServeJSON()
}

//fetchTakeTurns
type FetchTakeTurnsController struct {
	beego.Controller
}

//FetchTakeTurnsInfo __
type FetchTakeTurnsInfo struct {
	AdminID      int64           `json:"admin_id"`
	Begin        string          `json:"begin"`
	CompanyID    int64           `json:"company_id"`
	DeviceID     int64           `json:"device_id"`
	ID           int64           `json:"id"`
	LateTime     int64           `json:"late_time"`
	LeftEarly    int64           `json:"left_early"`
	Name         string          `json:"name"`
	OffworkBegin string          `json:"offwork_begin"`
	OffworkEnd   string          `json:"offwork_end"`
	Persons      map[int64]int64 `json:"persons"`
	Type         int64           `json:"type"`
	WorkBegin    string          `json:"work_begin"`
	WorkEnd      string          `json:"work_end"`
	Tags         []TagInfoJSON   `json:"tags"`
}

//FetchTakeTurnsReply _
type FetchTakeTurnsReply struct {
	Code    int64              `json:"code"`
	Data    FetchTakeTurnsInfo `json:"data"`
	Message string             `json:"message"`
}

//Get __
func (c *FetchTakeTurnsController) Get() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	strID := c.Input().Get("id")

	nID, err := strconv.ParseInt(strID, 0, 64)
	if err != nil {
		beego.Error(err)
		return
	}

	structRuleInfo, err := my_db.GetRuleByID(nID)
	if err != nil {
		beego.Error(err)
		return
	}

	var structReply FetchTakeTurnsReply
	structReply.Code = 20000
	structReply.Message = ""

	structReply.Data.Begin = structRuleInfo.Begin

	mapPersonAndTag := make(map[int64]int64, 0)

	arrayTags := strings.Split(structRuleInfo.Tags, ",")
	for nTagIndex, strTagID := range arrayTags {
		var structTagjsonInfo TagInfoJSON

		nTagID, err := strconv.ParseInt(strTagID, 0, 64)
		if err != nil {
			beego.Error(err)
			return
		}
		structTagInfo, err := my_db.GetTagByID(nTagID)
		if err != nil {
			beego.Error(err)
			return
		}

		arrayEmployeeID, err := my_db.GetEmployeeIDByTagID(nTagID)
		if err != nil {
			beego.Error(err)
			return
		}

		for _, nEmployeeID := range arrayEmployeeID {
			mapPersonAndTag[nEmployeeID] = int64(nTagIndex)
		}

		if structTagInfo.Active > 0 {
			structTagjsonInfo.Active = true
		} else {
			structTagjsonInfo.Active = false
		}

		structTagjsonInfo.Id = structTagInfo.ID
		structTagjsonInfo.Index = structTagInfo.Index
		structTagjsonInfo.Name = structTagInfo.Name
		if structTagInfo.Selected > 0 {
			structTagjsonInfo.Selected = true
		} else {
			structTagjsonInfo.Selected = false
		}

		structTagjsonInfo.From.Color = structTagInfo.Color
		structTagjsonInfo.From.Name = structTagInfo.Name
		structTagjsonInfo.From.Type = structTagInfo.Type

		structTagjsonInfo.From.Range1 = make(map[string]interface{}, 0)
		structTagjsonInfo.From.Range2 = make(map[string]interface{}, 0)
		structTagjsonInfo.From.Range3 = make(map[string]interface{}, 0)

		if structTagInfo.Range1 > 0 {
			structRangeInfo, err := my_db.GetRangeByID(structTagInfo.Range1)
			if err != nil {
				beego.Error(err)
				return
			}
			structTagjsonInfo.From.Range1["late_time"] = structRangeInfo.LateTime
			structTagjsonInfo.From.Range1["left_early"] = structRangeInfo.LeftEarly
			structTagjsonInfo.From.Range1["offwork_begin"] = structRangeInfo.OffWorkBegin
			structTagjsonInfo.From.Range1["offwork_end"] = structRangeInfo.OffWorkEnd
			if structRangeInfo.OffWorkCheck > 0 {
				structTagjsonInfo.From.Range1["offworkcheck"] = true
			} else {
				structTagjsonInfo.From.Range1["offworkcheck"] = false
			}

			structTagjsonInfo.From.Range1["time1"] = structRangeInfo.Time1
			structTagjsonInfo.From.Range1["time2"] = structRangeInfo.Time2
			structTagjsonInfo.From.Range1["work_begin"] = structRangeInfo.WorkBegin
			structTagjsonInfo.From.Range1["work_end"] = structRangeInfo.WorkEnd
			if structRangeInfo.WorkCheck > 0 {
				structTagjsonInfo.From.Range1["workcheck"] = true
			} else {
				structTagjsonInfo.From.Range1["workcheck"] = false
			}

		}

		if structTagInfo.Range2 > 0 {
			structRangeInfo, err := my_db.GetRangeByID(structTagInfo.Range2)
			if err != nil {
				beego.Error(err)
				return
			}
			structTagjsonInfo.From.Range2["late_time"] = structRangeInfo.LateTime
			structTagjsonInfo.From.Range2["left_early"] = structRangeInfo.LeftEarly
			structTagjsonInfo.From.Range2["offwork_begin"] = structRangeInfo.OffWorkBegin
			structTagjsonInfo.From.Range2["offwork_end"] = structRangeInfo.OffWorkEnd
			if structRangeInfo.OffWorkCheck > 0 {
				structTagjsonInfo.From.Range2["offworkcheck"] = true
			} else {
				structTagjsonInfo.From.Range2["offworkcheck"] = false
			}

			structTagjsonInfo.From.Range2["time1"] = structRangeInfo.Time1
			structTagjsonInfo.From.Range2["time2"] = structRangeInfo.Time2
			structTagjsonInfo.From.Range2["work_begin"] = structRangeInfo.WorkBegin
			structTagjsonInfo.From.Range2["work_end"] = structRangeInfo.WorkEnd
			if structRangeInfo.WorkCheck > 0 {
				structTagjsonInfo.From.Range2["workcheck"] = true
			} else {
				structTagjsonInfo.From.Range2["workcheck"] = false
			}
		}

		if structTagInfo.Range3 > 0 {
			structRangeInfo, err := my_db.GetRangeByID(structTagInfo.Range3)
			if err != nil {
				beego.Error(err)
				return
			}
			structTagjsonInfo.From.Range3["late_time"] = structRangeInfo.LateTime
			structTagjsonInfo.From.Range3["left_early"] = structRangeInfo.LeftEarly
			structTagjsonInfo.From.Range3["offwork_begin"] = structRangeInfo.OffWorkBegin
			structTagjsonInfo.From.Range3["offwork_end"] = structRangeInfo.OffWorkEnd
			if structRangeInfo.OffWorkCheck > 0 {
				structTagjsonInfo.From.Range3["offworkcheck"] = true
			} else {
				structTagjsonInfo.From.Range3["offworkcheck"] = false
			}

			structTagjsonInfo.From.Range3["time1"] = structRangeInfo.Time1
			structTagjsonInfo.From.Range3["time2"] = structRangeInfo.Time2
			structTagjsonInfo.From.Range3["work_begin"] = structRangeInfo.WorkBegin
			structTagjsonInfo.From.Range3["work_end"] = structRangeInfo.WorkEnd
			if structRangeInfo.WorkCheck > 0 {
				structTagjsonInfo.From.Range3["workcheck"] = true
			} else {
				structTagjsonInfo.From.Range3["workcheck"] = false
			}
		}
		structReply.Data.Tags = append(structReply.Data.Tags, structTagjsonInfo)
	}

	structReply.Data.Type = 1
	structReply.Data.WorkBegin = "00:00:00:00"
	structReply.Data.WorkEnd = "00:00:00:00"
	structReply.Data.OffworkBegin = "00:00:00:00"
	structReply.Data.OffworkEnd = "00:00:00:00"
	structReply.Data.Persons = mapPersonAndTag
	structReply.Data.AdminID = 0
	structReply.Data.CompanyID = 1
	structReply.Data.ID = nID
	structReply.Data.Name = structRuleInfo.Name

	byteReply, err := json.Marshal(structReply)
	if err != nil {
		beego.Error(err)
		return
	}
	c.Data["json"] = json.RawMessage(byteReply)
	c.ServeJSON()
	//structReply.Data.Persons
}

//SheduleContoller /api/management/attendance_rule/schedule
type SheduleContoller struct {
	beego.Controller
}

type PersonDateSheduleInfoJSON struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type SheduleInfoJSON struct {
	Dates     []string               `json:"dates"`
	Schedules map[string]interface{} `json:"schedules"`
}

type SheduleReplyJSON struct {
	Code    int64           `json:"code"`
	Data    SheduleInfoJSON `json:"data"`
	Message string          `json:"message"`
}

//Post __
func (c *SheduleContoller) Post() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	Xtoken := c.Ctx.Input.Header("X-Token")
	strAccount, err := my_aes.Dncrypt(string(Xtoken))
	if err != nil {
		beego.Error("清空设备人员数据错误", err)
		return
	}

	structAdminInfo, err := my_db.GetAdminByAccount(strAccount)
	if err != nil {
		beego.Error(err)
		return
	}

	arrayEmployeeInfo, err := my_db.GetEmployeeList(structAdminInfo.CompanyId, "")
	if err != nil {
		beego.Error(err)
		return
	}

	var structSheduleInfo SheduleInfoJSON
	structSheduleInfo.Dates = make([]string, 0)
	structSheduleInfo.Schedules = make(map[string]interface{}, 0)

	for _, structEmployeeInfo := range arrayEmployeeInfo {
		structRuleInfo, err := my_db.GetRuleByID(structEmployeeInfo.RuleID)
		if err != nil {
			beego.Error(err)
			return
		}
		if structRuleInfo.Type != 1 {
			continue
		}

		nSite := -1

		arraynTagID := make([]int64, 0)
		arrayTagID := strings.Split(structRuleInfo.Tags, ",")
		for nIndex, strTagID := range arrayTagID {
			nTagID, err := strconv.ParseInt(strTagID, 0, 64)
			arraynTagID = append(arraynTagID, nTagID)
			if err != nil {
				beego.Error(err)
				return
			}
			if nTagID == structEmployeeInfo.TagID {
				nSite = nIndex
			}
		}

		if nSite < 0 {
			continue
		}

		nStaticSite := nSite
		arrayPersonTagID := make([]int64, 0)
		for ; ; nSite++ {
			nSiteIndex := nSite % len(arraynTagID)
			if nSite > len(arraynTagID)-1 && nSiteIndex == nStaticSite {
				break
			}

			arrayPersonTagID = append(arrayPersonTagID, arraynTagID[nSiteIndex])
		}

		etLag, err := time.ParseDuration("8h")
		// 	timeBeijin := timeData.Add(jetLag)
		// 	strTimeBeijin := timeBeijin.Format(rule_algorithm.TIME_LAYOUT)
		// 	structRangeInfo.Begin = strTimeBeijin
		// 	nRangeID2, err = my_db.AddRangeInfo(structRangeInfo)
		arrayBegin := strings.Split(structRuleInfo.Begin, "T")
		arrayBegin_ := strings.Split(arrayBegin[1], "Z")
		strBegin := arrayBegin[0] + " " + arrayBegin_[0]
		timeBegin, err := time.Parse(rule_algorithm.TIME_LAYOUT, strBegin)
		if err != nil {
			beego.Error(err)
			return
		}
		timeBegin = timeBegin.Add(etLag)
		//timeBegin = timeBegin.AddDate(0, 0, nStaticSite)

		timeIndex := timeBegin

		timeEnd := timeBegin.AddDate(0, 1, 0)

		mapDateAndShedule := make(map[string]PersonDateSheduleInfoJSON, 0)

		for timeIndex.Before(timeEnd) {
			for _, nPersonTagID := range arrayPersonTagID {
				structTagInfo, err := my_db.GetTagByID(nPersonTagID)
				if err != nil {
					beego.Error(err)
					return
				}

				strTimeIndex := timeIndex.Format(rule_algorithm.TIME_LAYOUT)
				arrayTimeIndex := strings.Split(strTimeIndex, " ")
				arrayTimeIndex_ := strings.Split(arrayTimeIndex[0], "-")
				strMonthDay := arrayTimeIndex_[1] + "-" + arrayTimeIndex_[2]

				structSheduleInfo.Dates = append(structSheduleInfo.Dates, strMonthDay)

				mapDateAndShedule[strMonthDay] = PersonDateSheduleInfoJSON{Name: structTagInfo.Name, Color: structTagInfo.Color}

				timeIndex = timeIndex.AddDate(0, 0, 1)
			}

		}

		structSheduleInfo.Schedules[structEmployeeInfo.Name] = mapDateAndShedule
	}

	var structReply SheduleReplyJSON
	structReply.Code = 20000
	structReply.Data = structSheduleInfo
	structReply.Message = ""

	byteReply, err := json.Marshal(structReply)
	if err != nil {
		beego.Error(err)
		return
	}

	c.Data["json"] = json.RawMessage(string(byteReply))
	c.ServeJSON()
}

//OpenDoorControol __
type OpenDoorControol struct {
	beego.Controller
}

//OpenDoorReply __
type OpenDoorReply struct {
	Message string `json:"message"`
	Code    int64  `json:"code"`
	Data    string `json:"data"`
}

//Post __
func (c *OpenDoorControol) Post() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	data := c.Ctx.Input.RequestBody

	nID, err := jsonparser.GetInt([]byte(data), "id")
	if err != nil {
		beego.Error("添加抓拍内容错误", err)
		return
	}

	structDeviceInfo, err := my_db.GetDeviceById(nID)
	if err != nil {
		beego.Error(err)
		return
	}

	if structDeviceInfo.Type == 0 {
		var structRequstInfo facedevice.OpenDoorInfoJSON
		structRequstInfo.DeviceID = structDeviceInfo.DevcieId
		structRequstInfo.Status = 1
		structRequstInfo.Msg = "请通行"
		err = facedevice.OpenDoor(structDeviceInfo, structRequstInfo)
		if err != nil {
			beego.Error(err)
			return
		}
	}

	if structDeviceInfo.Type == 1 {
		var structDoorInfo facedevicexiongmai.OpenDoorInfoJSON
		structDoorInfo.DeviceID = structDeviceInfo.DevcieId
		structDoorInfo.CtrlType = 1
		structDoorInfo.Duration = 5
		err = facedevicexiongmai.OpenDoor(structDoorInfo)
		if err != nil {
			beego.Error(err)
			return
		}
	}

	var structReply OpenDoorReply
	structReply.Code = 20000
	structReply.Message = confreader.GetValue("SucRequest")
	structReply.Data = ""

	byteReply, err := json.Marshal(structReply)
	if err != nil {
		beego.Error(err)
		return
	}

	c.Data["json"] = json.RawMessage(byteReply)
	c.ServeJSON()
}

type DeviceGroupController struct {
	beego.Controller
}

type DeviceChildInfo struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Status int64  `json:"status"`
	Online int64  `json:"online"`
}

type DeviceGroupInfoJSON struct {
	ID    int64             `json:"id"`
	Name  string            `json:"name"`
	Child []DeviceChildInfo `json:"child"`
}

type DeviceGroupReplyJSON struct {
	Code    int64                 `json:"code"`
	Message string                `json:"message"`
	Data    []DeviceGroupInfoJSON `json:"data"`
}

func (c *DeviceGroupController) Get() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	Xtoken := c.Ctx.Input.Header("X-Token")
	Account, err := my_aes.Dncrypt(Xtoken)

	structAdminInfo, err := my_db.GetAdminByAccount(Account)
	if err != nil {
		beego.Error(err)
		return
	}

	mapGroupDevice, err := my_db.GetGroupDevice()
	if err != nil {
		beego.Error(err)
		return
	}
	var structReply DeviceGroupReplyJSON
	structReply.Data = make([]DeviceGroupInfoJSON, 0)
	for _, arrayGroupDevice := range mapGroupDevice {
		var structDeviceGroupInfo DeviceGroupInfoJSON
		structDeviceGroupInfo.Child = make([]DeviceChildInfo, 0)
		for _, structDeviceValue := range arrayGroupDevice {
			structDeviceGroupInfo.ID = structDeviceValue.GroupID
			structDeviceGroupInfo.Name = structDeviceValue.GroupName
			structDeviceInfo, err := my_db.GetDeviceById(structDeviceValue.ID)
			if err != nil {
				beego.Error(err)
				return
			}

			if structDeviceInfo.CompanyID != structAdminInfo.CompanyId {
				continue
			}

			var structChildDeviceInfo DeviceChildInfo
			structChildDeviceInfo.Name = structDeviceInfo.Name
			structChildDeviceInfo.ID = structDeviceInfo.Id
			structChildDeviceInfo.Online = 1
			structChildDeviceInfo.Status = 0
			if structDeviceInfo.Type == 0 {
				structGroupDeviceInfo, err := facedevice.GetDoorConditionParam(structDeviceInfo)
				if err != nil {
					structChildDeviceInfo.Online = 0
				}
				structChildDeviceInfo.Status = structGroupDeviceInfo.IOType
			}

			if structDeviceInfo.Type == 1 {
				structChildDeviceInfo.Status, err = facedevicexiongmai.GetDoorStatus(structDeviceInfo.DevcieId)
				if err != nil {
					structChildDeviceInfo.Online = 0
				}

				// _, err := facedevicexiongmai.GetDeviceInfoByIP(structDeviceInfo.Ip)
				// if err != nil {
				// 	structChildDeviceInfo.Online = 0
				// }
			}

			structDeviceGroupInfo.Child = append(structDeviceGroupInfo.Child, structChildDeviceInfo)
		}

		structReply.Data = append(structReply.Data, structDeviceGroupInfo)
	}

	structReply.Code = 20000
	structReply.Message = ""
	byteReply, err := json.Marshal(structReply)
	if err != nil {
		beego.Error(err)
		return
	}

	// byteReply := []byte(`{"code":20000,"message":"","data":[{"id":1,"name":"aa",
	// "child":[{"id":0,"name":"0楼大门","status":0,"online":0},{"id":1,"name":"1楼大门","status":1,"online":0},
	// {"id":2,"name":"2楼大门","status":0,"online":1},{"id":3,"name":"3楼大门","status":0,"online":1},{"id":4,"name":
	// "4楼大门","status":0,"online":1},{"id":5,"name":"5楼大门","status":0,"online":1},{"id":6,"name":"6楼大门","status":1,"online":1},
	// {"id":7,"name":"7楼大门","status":1,"online":0},{"id":8,"name":"8楼大门","status":1,"online":0},{"id":9,"name":"9楼大门","status":0,"online":1}]}]}
	// `)

	c.Data["json"] = json.RawMessage(byteReply)
	c.ServeJSON()
}

type DeviceStatusController struct {
	beego.Controller
}

type GetDeviceStatusRequst struct {
	IDS []int64 `json:"ids"`
}

type DeviceStatusInfoJSON struct {
	ID     int64 `json:"id"`
	Online int64 `json:"online"`
	Status int64 `json:"status"`
}

type DeviceStatusReplyJSON struct {
	Code    int64                  `json:"code"`
	Message string                 `json:"message"`
	Data    []DeviceStatusInfoJSON `json:"data"`
}

func (c *DeviceStatusController) Post() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	data := c.Ctx.Input.RequestBody
	var structRequst GetDeviceStatusRequst
	err := json.Unmarshal(data, &structRequst)
	if err != nil {
		beego.Error(err)
		return
	}

	var structReply DeviceStatusReplyJSON
	structReply.Code = 20000
	structReply.Data = make([]DeviceStatusInfoJSON, 0)
	structReply.Message = ""
	for _, nID := range structRequst.IDS {
		var structDeviceStatusInfo DeviceStatusInfoJSON
		structDeviceInfo, err := my_db.GetDeviceById(nID)
		if err != nil {
			beego.Error(err)
			return
		}

		structDeviceStatusInfo.ID = structDeviceInfo.Id
		structDeviceStatusInfo.Status = 0
		structDeviceStatusInfo.Online = 1
		if structDeviceInfo.Type == 0 {
			structGroupDeviceInfo, err := facedevice.GetDoorConditionParam(structDeviceInfo)
			if err != nil {
				structDeviceStatusInfo.Online = 0
			}
			structDeviceStatusInfo.Status = structGroupDeviceInfo.IOType
		}
		if structDeviceInfo.Type == 1 {
			structDeviceStatusInfo.Status, err = facedevicexiongmai.GetDoorStatus(structDeviceInfo.DevcieId)
			if err != nil {
				structDeviceStatusInfo.Online = 0
			}

			// _, err := facedevicexiongmai.GetDeviceInfoByIP(structDeviceInfo.Ip)
			// if err != nil {
			// 	structDeviceStatusInfo.Online = 0
			// }
		}

		structReply.Data = append(structReply.Data, structDeviceStatusInfo)
	}

	byteReply, err := json.Marshal(structReply)
	if err != nil {
		beego.Error(err)
		return
	}

	c.Data["json"] = json.RawMessage(byteReply)
	c.ServeJSON()
}

type KeepDoorOpenController struct {
	beego.Controller
}

type KeepDoorOpenRequstJSON struct {
	ID   int64 `json:"id"`
	Open bool  `json:"open"`
}

type KeepDoorOpenReplyJSON struct {
	Code    int64                `json:"code"`
	Message string               `json:"message"`
	Data    DeviceStatusInfoJSON `json:"data"`
}

func (c *KeepDoorOpenController) Post() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	data := c.Ctx.Input.RequestBody

	var structRequst KeepDoorOpenRequstJSON
	err := json.Unmarshal(data, &structRequst)
	if err != nil {
		beego.Error(err)
		return
	}

	structDeviceInfo, err := my_db.GetDeviceById(structRequst.ID)
	if err != nil {
		beego.Error(err)
		return
	}

	if structDeviceInfo.Type == 0 {
		structDoorOpenParam, err := facedevice.GetDoorConditionParam(structDeviceInfo)
		if err != nil {
			beego.Error(err)
			return
		}
		if structRequst.Open {
			structDoorOpenParam.IOType = 0
		} else {
			structDoorOpenParam.IOType = 1
		}

		err = facedevice.SetDoorConditionParam(structDeviceInfo, structDoorOpenParam)
		if err != nil {
			beego.Error(err)
			return
		}

	}

	if structDeviceInfo.Type == 1 {
		var structOpenDoorInfo facedevicexiongmai.OpenDoorInfoJSON
		structOpenDoorInfo.DeviceID = structDeviceInfo.DevcieId
		structOpenDoorInfo.Duration = 0
		if structRequst.Open {
			structOpenDoorInfo.CtrlType = 1
		} else {
			structOpenDoorInfo.CtrlType = 0
		}

		err = facedevicexiongmai.OpenDoor(structOpenDoorInfo)
		if err != nil {
			beego.Error(err)
			return
		}
	}

	var structReply KeepDoorOpenReplyJSON
	structReply.Code = 20000
	structReply.Data.ID = structDeviceInfo.Id
	structReply.Data.Online = 1
	if structRequst.Open {
		structReply.Data.Status = 0
	} else {
		structReply.Data.Status = 1
	}

	structReply.Message = ""

	byteReply, err := json.Marshal(structReply)
	if err != nil {
		beego.Error(err)
		return
	}

	c.Data["json"] = json.RawMessage(byteReply)
	c.ServeJSON()
}

type SaveGroupDeviceController struct {
	beego.Controller
}

type SaveGroupInfoJSON struct {
	ID   int64   `json:"id"`
	Name string  `json:"name"`
	IDS  []int64 `json:"ids"`
}

type SaveGroupInfoReplyJSON struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

func (c *SaveGroupDeviceController) Post() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	data := c.Ctx.Input.RequestBody
	var structRequst []SaveGroupInfoJSON
	err := json.Unmarshal(data, &structRequst)
	if err != nil {
		beego.Error(err)
		return
	}

	for _, structGroupInfoVlaue := range structRequst {
		if structGroupInfoVlaue.ID > 0 {
			err = my_db.RemoveGroupDeviceByGroupID(structGroupInfoVlaue.ID)
			if err != nil {
				beego.Error(err)
				return
			}
			var addGrouDeviceInfo my_db.GroupDeviceInfo
			addGrouDeviceInfo.GroupName = structGroupInfoVlaue.Name
			addGrouDeviceInfo.GroupID = structGroupInfoVlaue.ID
			for _, nDeviceID := range structGroupInfoVlaue.IDS {
				err = my_db.RemoveGroupDevice(nDeviceID)
				if err != nil {
					beego.Error(err)
					return
				}

				addGrouDeviceInfo.ID = nDeviceID
				structDeviceInfo, err := my_db.GetDeviceById(nDeviceID)
				if err != nil {
					beego.Error(err)
					return
				}
				addGrouDeviceInfo.Name = structDeviceInfo.Name

				_, err = my_db.AddGroupDevice(addGrouDeviceInfo)
				if err != nil {
					beego.Error(err)
					return
				}
			}
		}

		if structGroupInfoVlaue.ID == 0 && len(structGroupInfoVlaue.IDS) > 0 {
			var structGroupDeviceInfo my_db.GroupDeviceInfo
			structGroupDeviceInfo.GroupName = structGroupInfoVlaue.Name
			structGroupDeviceInfo.GroupID = structGroupInfoVlaue.ID
			for _, nDeviceID := range structGroupInfoVlaue.IDS {
				err = my_db.RemoveGroupDevice(nDeviceID)
				if err != nil {
					beego.Error(err)
					return
				}
				structDeviceInfo, err := my_db.GetDeviceById(nDeviceID)
				if err != nil {
					beego.Error(err)
					return
				}

				structGroupDeviceInfo.ID = nDeviceID
				structGroupDeviceInfo.Name = structDeviceInfo.Name
				nGroupID, err := my_db.AddGroupDevice(structGroupDeviceInfo)
				if err != nil {
					beego.Error(err)
					return
				}
				if structGroupDeviceInfo.GroupID <= 0 && nGroupID > 0 {
					structGroupDeviceInfo.GroupID = nGroupID
				}
			}

		}
	}

	var structReply SaveGroupInfoReplyJSON
	structReply.Code = 20000
	structReply.Message = confreader.GetValue("SucRequest")
	structReply.Data = ""

	byteReply, err := json.Marshal(structReply)
	if err != nil {
		beego.Error(err)
		return
	}

	c.Data["json"] = json.RawMessage(byteReply)
	c.ServeJSON()
}

type GetCompanyInfoController struct {
	beego.Controller
}

type GetCompanyInfoDataJSON struct {
	Address    string `json:"address"`
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Phone      string `json:"phone"`
	Is_manager int64  `json:"is_manager"`
	Is_default int64  `json:"is_default"`
	Location   string `json:"location"`
	Multi_cert bool   `json:"multi_cert"`
}

type GetCompanyInfoReplyJSON struct {
	Code    int64                    `json:"code"`
	Message string                   `json:"message"`
	Data    []GetCompanyInfoDataJSON `json:"data"`
}

func (c *GetCompanyInfoController) Post() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	Xtoken := c.Ctx.Input.Header("X-Token")
	strAccount, err := my_aes.Dncrypt(string(Xtoken))
	if err != nil {
		beego.Error("获取用户信息错误", err)
		return
	}

	var responseStruct InfoResult_Json
	responseStruct.Data.Companys = make([]InfoCompny_Json, 0)

	structAdminInfo, err := my_db.GetAdminByAccount(strAccount)
	if err != nil {
		beego.Error("获取用户信息错误", err)
		return
	}

	var structReply GetCompanyInfoReplyJSON
	structReply.Code = 20000
	structReply.Message = ""

	structReply.Data = make([]GetCompanyInfoDataJSON, 0)

	if len(structAdminInfo.Companys) > 0 {
		arraySplite := strings.Split(structAdminInfo.Companys, ",")
		for _, varIndex := range arraySplite {
			nCompanyIDIndex, err := strconv.ParseInt(varIndex, 0, 64)
			structCompanyInfoIndex, err := my_db.GetCompnyById(nCompanyIDIndex)
			if err != nil {
				beego.Error(err)
				return
			}

			var structCompanyInfo GetCompanyInfoDataJSON
			structCompanyInfo.Address = structCompanyInfoIndex.Address
			structCompanyInfo.ID = structCompanyInfoIndex.Id
			structCompanyInfo.Name = structCompanyInfoIndex.Name
			structCompanyInfo.Phone = structCompanyInfoIndex.Phone
			structCompanyInfo.Is_manager = 1
			structCompanyInfo.Location = structCompanyInfoIndex.Location
			if structCompanyInfo.ID == structAdminInfo.CompanyId {
				structCompanyInfo.Is_default = 1
			} else {
				structCompanyInfo.Is_default = 0
			}

			if structCompanyInfoIndex.MultiCer > 0 {
				structCompanyInfo.Multi_cert = true
			} else {
				structCompanyInfo.Multi_cert = false
			}

			structReply.Data = append(structReply.Data, structCompanyInfo)
		}
	}

	// for _, varIndex := range arrayCompany {
	// 	var structCompanyInfo GetCompanyInfoDataJSON
	// 	structCompanyInfo.Address = varIndex.Address
	// 	structCompanyInfo.ID = varIndex.Id
	// 	structCompanyInfo.Name = varIndex.Name
	// 	structCompanyInfo.Phone = varIndex.Phone
	// 	structCompanyInfo.Is_manager = 1
	// 	structCompanyInfo.Location = varIndex.Location
	// 	structCompanyInfo.Is_default = 0

	// 	structReply.Data = append(structReply.Data, structCompanyInfo)
	// }

	byteReply, err := json.Marshal(structReply)
	if err != nil {
		beego.Error(err)
		return
	}

	c.Data["json"] = json.RawMessage(string(byteReply))
	c.ServeJSON()
}

type SetDefaultCompanyController struct {
	beego.Controller
}

type SetDefaultCompanyReplyJSON struct {
	Message string `json:"message"`
	Code    int64  `json:"code"`
}

func (c *SetDefaultCompanyController) Post() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	Data := c.Ctx.Input.RequestBody

	//beego.Warning(string(Data))

	nCompanyID, err := jsonparser.GetInt([]byte(Data), "company_id")
	if err != nil {
		beego.Error(err)
		return
	}

	Xtoken := c.Ctx.Input.Header("X-Token")
	strAccount, err := my_aes.Dncrypt(string(Xtoken))
	if err != nil {
		beego.Error("获取用户信息错误", err)
		return
	}

	var responseStruct InfoResult_Json
	responseStruct.Data.Companys = make([]InfoCompny_Json, 0)

	structAdminInfo, err := my_db.GetAdminByAccount(strAccount)
	if err != nil {
		beego.Error("获取用户信息错误", err)
		return
	}

	structAdminInfo.CompanyId = nCompanyID

	err = my_db.UpdateAdminInfo(structAdminInfo)
	if err != nil {
		beego.Error(err)
		return
	}

	var structReply SetDefaultCompanyReplyJSON
	structReply.Code = 20000
	structReply.Message = ""

	byteReply, err := json.Marshal(structReply)
	if err != nil {
		beego.Error(err)
		return
	}
	c.Data["json"] = json.RawMessage(byteReply)
	c.ServeJSON()
}

type CheckPersonController struct {
	beego.Controller
}

type CheckPersonRequstJSON struct {
	Checktype int64                  `json:"checktype"`
	Days      string                 `json:"days"`
	Name      string                 `json:"name"`
	Person    []interface{}          `json:"person"`
	Range1    map[string]interface{} `json:"range1"`
	Range2    map[string]interface{} `json:"range2"`
	Range3    map[string]interface{} `json:"range3"`
	Sync      bool                   `json:"sync"`
}

type CheckPersonReplyJSON struct {
	Code    int64               `json:"code"`
	Message string              `json:"message"`
	Data    map[string][]string `json:"data"`
}

func (c *CheckPersonController) Post() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	Data := c.Ctx.Input.RequestBody

	var structRequst CheckPersonRequstJSON
	err := json.Unmarshal(Data, &structRequst)
	if err != nil {
		beego.Error(err)
		return
	}
	var structReply CheckPersonReplyJSON
	structReply.Code = 20000
	structReply.Message = ""
	structReply.Data = make(map[string][]string, 0)
	structReply.Data["names"] = make([]string, 0)

	for _, indexValue := range structRequst.Person {
		switch indexValue.(type) {
		case float64:
			{
				nEmployeeID := int64(indexValue.(float64))

				arrayResult, err := my_db.GetRangeAndEmployeeByEmployeeID(nEmployeeID)
				if err != nil {
					beego.Error(err)
					return
				}
				if len(arrayResult) > 0 {
					structEmployeeInfo, err := my_db.GetEmployeeById(nEmployeeID)
					if err != nil {
						beego.Error(err)
						return
					}

					structReply.Data["names"] = append(structReply.Data["names"], structEmployeeInfo.Name)
				}
			}

		}

	}

	byteReply, err := json.Marshal(structReply)
	if err != nil {
		beego.Error(err)
		return
	}

	c.Data["json"] = json.RawMessage(byteReply)
	c.ServeJSON()
}

type RemoveCompanyContoller struct {
	beego.Controller
}

type RemoveCompanyReplyJSON struct {
	Code int64  `json:"code"`
	Msg  string `json:"msg"`
	Data string `json:"data"`
}

func (c *RemoveCompanyContoller) Post() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	Data := c.Ctx.Input.RequestBody

	nCompanyID, err := jsonparser.GetInt(Data, "company_id")

	err = my_db.RemoveCompanyByID(nCompanyID)
	if err != nil {
		beego.Error(err)
		return
	}

	Xtoken := c.Ctx.Input.Header("X-Token")
	strAccount, err := my_aes.Dncrypt(string(Xtoken))
	if err != nil {
		beego.Error("获取用户信息错误", err)
		return
	}

	var responseStruct InfoResult_Json
	responseStruct.Data.Companys = make([]InfoCompny_Json, 0)

	structAdminInfo, err := my_db.GetAdminByAccount(strAccount)
	if err != nil {
		beego.Error("获取用户信息错误", err)
		return
	}

	strCompanys := ""
	bSet := false
	arrayCompany := strings.Split(structAdminInfo.Companys, ",")
	for _, valueCompanyID := range arrayCompany {
		nCompanyIDIndex, err := strconv.ParseInt(valueCompanyID, 0, 64)
		if err != nil {
			beego.Error(err)
			return
		}

		if nCompanyIDIndex != nCompanyID {
			if len(strCompanys) > 0 {
				strCompanys += fmt.Sprintf(",%d", nCompanyIDIndex)
			} else {
				strCompanys += fmt.Sprintf("%d", nCompanyIDIndex)
			}

			if structAdminInfo.CompanyId == nCompanyID && !bSet {
				structAdminInfo.CompanyId = nCompanyIDIndex
				bSet = true
			}
		}
	}

	structAdminInfo.Companys = strCompanys

	if len(strCompanys) == 0 {
		structAdminInfo.CompanyId = -1
	}

	err = my_db.UpdateAdminInfo(structAdminInfo)
	if err != nil {
		beego.Error(err)
		return
	}

	var structReply RemoveCompanyReplyJSON
	structReply.Code = 20000
	structReply.Msg = confreader.GetValue("SucRequest")
	structReply.Data = ""

	byteReply, err := json.Marshal(structReply)
	if err != nil {
		beego.Error(err)
		return
	}

	c.Data["json"] = json.RawMessage(byteReply)
	c.ServeJSON()
}

type CommonReplyJSON struct {
	Code    int64         `json:"code"`
	Data    []interface{} `json:"data"`
	Message string        `json:"message"`
}

type DashhoardController struct {
	beego.Controller
}

func (c *DashhoardController) Post() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	data := c.Ctx.Input.RequestBody

	nCompanyID, err := jsonparser.GetInt(data, "company")
	if err != nil {
		beego.Error("获取公司ID失败", err)
		return
	}

	//beego.Debug("ncompanyid:", nCompanyID)

	structRecordArray, err := my_db.GetVerifyRecordList(nCompanyID, true, 1, 20)
	if err != nil {
		beego.Error("获取考勤记录列表错误", err)
		return
	}

	var structReply CommonReplyJSON
	structReply.Data = make([]interface{}, 0)
	for _, valueRecord := range structRecordArray {
		mapIndex := make(map[string]interface{}, 0)
		mapIndex["device"] = valueRecord.DeviceName
		mapIndex["person"] = valueRecord.Name
		mapIndex["pic"] = valueRecord.VerifyPic
		mapIndex["time"] = valueRecord.CreateTime
		mapIndex["temperature"] = valueRecord.Template
		if len(valueRecord.Template) > 0 {
			fTemperature, err := strconv.ParseFloat(valueRecord.Template, len(valueRecord.Template))
			if err != nil {
				beego.Error(err)
				return
			}

			if fTemperature > 37.7 {
				mapIndex["alarm"] = 1
			} else {
				mapIndex["alarm"] = 0
			}
		}

		structReply.Data = append(structReply.Data, mapIndex)
	}

	// var structReply NewIODataReply_Json
	// structReply.Data = make([]NewIODataInfo_Json, 0)

	// for _, structVerifyIndex := range arrayVerify {
	// 	var structIODataInfo NewIODataInfo_Json
	// 	structIODataInfo.Person = structVerifyIndex.Name
	// 	structIODataInfo.Pic = structVerifyIndex.VerifyPic
	// 	arrayCreateTime := strings.Split(structVerifyIndex.CreateTime, "T")
	// 	arrayCreateTime_ := strings.Split(arrayCreateTime[1], "Z")
	// 	structVerifyIndex.CreateTime = arrayCreateTime[0] + " " + arrayCreateTime_[0]
	// 	timeCreate, err := time.Parse(rule_algorithm.TIME_LAYOUT, structVerifyIndex.CreateTime)
	// 	if err != nil {
	// 		beego.Error(err)
	// 		return
	// 	}
	// 	structIODataInfo.Time = timeCreate.Format(rule_algorithm.TIME_LAYOUT)
	// 	structIODataInfo.Temperature = structVerifyIndex.Template
	// 	if len(structVerifyIndex.Template) > 0 {
	// 		fTemperature, err := strconv.ParseFloat(structVerifyIndex.Template, len(structVerifyIndex.Template))
	// 		if err != nil {
	// 			beego.Error(err)
	// 			return
	// 		}

	// 		if fTemperature > 37.7 {
	// 			structIODataInfo.Alarm = 1
	// 		} else {
	// 			structIODataInfo.Alarm = 0
	// 		}
	// 	}

	// 	structDeviceInfo, err := my_db.GetDeviceByDeviceId(fmt.Sprintf("%d", structVerifyIndex.DeviceId))
	// 	if err != nil {
	// 		beego.Error("获取抓拍记录错误", err)
	// 		return
	// 	}
	// 	structIODataInfo.Device = structDeviceInfo.Name
	// 	structReply.Data = append(structReply.Data, structIODataInfo)
	// }

	structReply.Code = 20000
	structReply.Message = ""

	byteReply, err := json.Marshal(structReply)
	if err != nil {
		beego.Error(err)
		return
	}

	c.Data["json"] = json.RawMessage(byteReply)
	c.ServeJSON()
}

type ChangeXMPswController struct {
	beego.Controller
}

func (c *ChangeXMPswController) Post() {
	byteLan := c.Ctx.Input.Header("X-Lang")
	confreader.SetLanguage(string(byteLan))

	data := c.Ctx.Input.RequestBody

	mapRequst := make(map[string]interface{}, 0)

	err := json.Unmarshal(data, &mapRequst)
	if err != nil {
		beego.Error(err)
		return
	}

	nID := mapRequst["id"].(float64)
	strPsw := mapRequst["password"].(string)

	structDeviceInfo, err := my_db.GetDeviceById(int64(nID))
	if err != nil {
		beego.Error(err)
		return
	}

	err = facedevicexiongmai.DeviceSetServer(structDeviceInfo.DevcieId, strPsw)
	if err != nil {
		beego.Error(err)
		return
	}

	// err = facedevicexiongmai.RestartXMDevice(structDeviceInfo.DevcieId)
	// if err != nil {
	// 	beego.Error(err)
	// 	return
	// }

	mapReply := make(map[string]interface{}, 0)

	mapReply["message"] = confreader.GetValue("NeedReboot")
	mapReply["code"] = 20000
	mapReply["data"] = ""

	byteReply, err := json.Marshal(mapReply)
	if err != nil {
		beego.Error(err)
		return
	}

	c.Data["json"] = json.RawMessage(byteReply)
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
		mapReply["msg"] = confreader.GetValue("SucRequest")
	}

	byteReply, err := json.Marshal(mapReply)
	if err != nil {
		beego.Error(err)
		return
	}

	c.Data["json"] = json.RawMessage(byteReply)
	c.ServeJSON()
}

type GetLogController struct {
	beego.Controller
}

func (c *GetLogController) Get() {

	mapReply := make(map[string]interface{}, 0)

	mapReply["code"] = 50000
	mapReply["message"] = "该功能还未开放"

	byteReply, err := json.Marshal(mapReply)
	if err != nil {
		beego.Error(err)
		return
	}

	c.Data["json"] = json.RawMessage(byteReply)
	c.ServeJSON()
}

type GetAlarmLogController struct {
	beego.Controller
}

func (c *GetAlarmLogController) Get() {

	mapReply := make(map[string]interface{}, 0)
	mapReply["code"] = 50000
	mapReply["message"] = confreader.GetValue("NoFunction")

	byteReply, err := json.Marshal(mapReply)
	if err != nil {
		beego.Error(err)
		return
	}
	c.Data["json"] = json.RawMessage(byteReply)
	c.ServeJSON()
}
