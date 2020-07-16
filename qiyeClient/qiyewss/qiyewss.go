package qiyewss

import (
	"bytes"
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"os/signal"
	"sort"
	"strings"
	"sync"
	"time"

	"../facedevice"
	"../facedevicexiongmai"
	"../mysqlitemanger"

	"github.com/Unknwon/goconfig"
	"github.com/astaxie/beego"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
)

const TIME_LAYOUT = "2006-01-02 15:04:05"

var mapWSSCmd = make(map[string][]byte, 0)
var wssconn *websocket.Conn

var qiyeAddr = flag.String("addrWX", "openhw.work.weixin.qq.com", "http service address")

//StartWssCLient 开启wss客户端
func StartWssCLient() error {
	var addr = qiyeAddr
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

	beego.Info("企业服务开启")

	return nil
}

//CommonReplyJSON __
type CommonReplyJSON struct {
	Headers map[string]string      `json:"headers"`
	Body    map[string]interface{} `json:"body"`
	Errcode int64                  `json:"errcode"`
	Errmsg  string                 `json:"errmsg"`
	Cmd     string                 `json:"cmd"`
}

var mapRecv = make(map[string]CommonReplyJSON, 0)
var mapRecvEx = make(map[string][]byte, 0)
var mapPushUsrCmd = make(map[string]CommonReplyJSON, 0)
var mapSend = make(map[string]CommonCMDJSON, 0)

//WSSRecvGoroutine 接收数据的goroutine
func WSSRecvGoroutine() {
	done := make(chan struct{})

	defer wssconn.Close()
	defer close(done)
	for {
		_, message, err := wssconn.ReadMessage()
		if err != nil {
			beego.Error("wss接收goroutine关闭", err)
			return
		}

		var structCommoncmd CommonReplyJSON
		err = json.Unmarshal(message, &structCommoncmd)
		if err != nil {
			if len(message) > 10 {
				//图片数据
				structCommoncmd, err = DecodeImageInfo(message)
				if err != nil {
					continue
				}
			} else {
				beego.Error("wss接收goroutine关闭", err)
				return
			}
		}

		beego.Debug("接收到WSS服务的数据:", string(message))

		if structCommoncmd.Cmd == "push_user_face" {
			go GetPotoInfo(structCommoncmd)
			continue
		}

		if structCommoncmd.Cmd == "push_change_contact" {
			RemoveEmployeeFromDevice(structCommoncmd)
			continue
		}

		mapRecv[structCommoncmd.Headers["req_id"]] = structCommoncmd
		mapRecvEx[structCommoncmd.Headers["req_id"]] = message
	}

	beego.Info("企业服务接收Goroutine退出")
}

//RemoveEmployeeFromDevice __
func RemoveEmployeeFromDevice(structRequst CommonReplyJSON) error {
	arrayDevice, err := mysqlitemanger.GetDeviceList(0)
	if err != nil {
		beego.Error(err)
		return err
	}

	var structRemoveInfo facedevicexiongmai.RemoveEmployeeDataJSON
	for _, UpdateValue := range structRequst.Body["update_user"].([]interface{}) {
		structRemoveInfo.EmployeeInfo.UserID = UpdateValue.(map[string]interface{})["userid"].(string)
		for _, valueIndex := range arrayDevice {
			if valueIndex.Type == 0 {
				var StructUserInfoHQ facedevice.EmployeeInfo
				StructUserInfoHQ.UserID, err = mysqlitemanger.GetUserIDByXM(structRemoveInfo.EmployeeInfo.UserID)
				if err != nil {
					beego.Error(err)
					continue
				}

				if StructUserInfoHQ.UserID > 0 {
					err = facedevice.RemoveDeviceEmployee(valueIndex, StructUserInfoHQ)
					if err != nil {
						beego.Error(err)
						continue
					}
				}
			} else if valueIndex.Type == 1 {
				structRemoveInfo.DeviceInfo.DeviceID = valueIndex.DevcieId
				err = facedevicexiongmai.RemoveEmployeeWSS(structRemoveInfo)
				if err != nil {
					beego.Error(err)
					continue
				}
			}
		}

	}

	return nil
}

type atomicInt struct {
	value int
	lock  sync.Mutex // 互斥锁, 也就是我正在操作的时候，我排斥其他人操作
}

var mutextLock atomicInt

//var chanSendMsg = make(chan []byte, 0)

//SendPackageToWSS  发送数据到wss服务
func SendPackageToWSS(byteSendPack []byte) error {

	beego.Debug("发送数据", string(byteSendPack))
	//mutextLock.lock.Lock()
	//chanSendMsg <- byteSendPack
	//mutextLock.lock.Unlock()

	return wssconn.WriteMessage(websocket.TextMessage, byteSendPack)
}

//SendWSSMsgGoroutine __
// func SendWSSMsgGoroutine() {
// 	defer wssconn.Close()

// 	for {
// 		byteSendMsg := <-chanSendMsg
// 		err := wssconn.WriteMessage(websocket.TextMessage, byteSendMsg)
// 		if err != nil {
// 			beego.Error(err)
// 			return
// 		}
// 		beego.Debug("发送数据:", string(byteSendMsg))
// 		// select {
// 		// case byteSendMsg := <-chanSendMsg:
// 		// 	{
// 		// 		err := wssconn.WriteMessage(websocket.TextMessage, byteSendMsg)
// 		// 		if err != nil {
// 		// 			beego.Error(err)
// 		// 			return
// 		// 		}
// 		// 		beego.Debug("发送数据:", string(byteSendMsg))
// 		// 	}
// 		// default:
// 		// 	{
// 		// 		if !m_bQiyeServerRun {
// 		// 			beego.Info("企业服务发送Goroutine退出")
// 		// 			return
// 		// 		}
// 		// 	}
// 		// }

// 	}
// }

//CommonCMDJSON 公用的指令数据
type CommonCMDJSON struct {
	Cmd     string                 `json:"cmd"`
	Headers map[string]string      `json:"headers"`
	Body    map[string]interface{} `json:"body"`
}

//bytesToIntS __
func bytesToIntS(b []byte) (int, error) {
	if len(b) == 3 {
		b = append([]byte{0}, b...)
	}
	bytesBuffer := bytes.NewBuffer(b)
	switch len(b) {
	case 1:
		var tmp int8
		err := binary.Read(bytesBuffer, binary.LittleEndian, &tmp)
		return int(tmp), err
	case 2:
		var tmp int16
		err := binary.Read(bytesBuffer, binary.LittleEndian, &tmp)
		return int(tmp), err
	case 4:
		var tmp int32
		err := binary.Read(bytesBuffer, binary.LittleEndian, &tmp)
		return int(tmp), err
	default:
		return 0, fmt.Errorf("%s", "BytesToInt bytes lenth is invaild!")
	}
}

//DecodeImageInfo __
func DecodeImageInfo(byteImage []byte) (CommonReplyJSON, error) {
	var strcutResult CommonReplyJSON
	strcutResult.Body = make(map[string]interface{}, 0)
	strcutResult.Headers = make(map[string]string, 0)
	//byteFormatVer := byteImage[0:4]
	byteRqIDLength := byteImage[4:8]
	nReIDlen, err := bytesToIntS(byteRqIDLength)
	if err != nil {
		beego.Error(err)
		return strcutResult, err
	}
	byteRQID := byteImage[8 : 8+nReIDlen]
	//beego.Debug(string(byteRQID))
	byteDateLen := byteImage[8+nReIDlen : 12+nReIDlen]

	nDateLen, err := bytesToIntS(byteDateLen)
	if err != nil {
		beego.Error(err)
		return strcutResult, err
	}

	byteBianery := byteImage[12+nReIDlen : nDateLen+12+nReIDlen]

	strcutResult.Cmd = "download_photo"
	strcutResult.Headers["req_id"] = string(byteRQID)
	strcutResult.Body["binary"] = byteBianery
	//beego.Debug("图片数据:", byteBianery)

	return strcutResult, nil
}

//AddPerson2DeviceGoRoutine  __
func AddPerson2DeviceGoRoutine(structRequst CommonReplyJSON, byteImage []byte) error {
	strUserID := structRequst.Body["userid"].(string)
	strName := ""

	nUserID := rand.Int31()

	nUserIDIndex, err := mysqlitemanger.GetUserIDByXM(strUserID)
	if err != nil {
		beego.Error(err)
		return err
	}

	var structEmployeeHQ facedevice.EmployeeInfo
	structEmployeeHQ.Name = strName
	structEmployeeHQ.Pic = base64.StdEncoding.EncodeToString(byteImage)

	if nUserIDIndex <= 0 {
		err = mysqlitemanger.InsertUserID(mysqlitemanger.UserIDInfo{XMUserID: strUserID, HQUserID: int64(nUserID)})
		if err != nil {
			beego.Error(err)
			return err
		}

		structEmployeeHQ.UserID = int64(nUserID)
	} else {
		structEmployeeHQ.UserID = nUserIDIndex
	}

	arraySearchParam := make([]UserSearchParam, 0)
	arraySearchParam = append(arraySearchParam, UserSearchParam{UserID: strUserID, UserType: 0})
	structSearchUserResult, err := GetUserInfoByIDS(arraySearchParam)
	for _, value := range structSearchUserResult.Body["userinfo"].([]interface{}) {
		if value.(map[string]interface{})["userid"].(string) == strUserID {
			strName = value.(map[string]interface{})["name"].(string)
		}
	}

	arrayDevice, err := mysqlitemanger.GetDeviceList(0)
	if err != nil {
		return err
	}

	var structEmployeeXM facedevicexiongmai.AddEmployeeInfoJSON
	structEmployeeXM.EmployeInfo.UserID = strUserID
	structEmployeeXM.EmployeInfo.CardID = ""
	structEmployeeXM.EmployeInfo.GroupID = "default"
	structEmployeeXM.EmployeInfo.Name = strName
	structEmployeeXM.EmployeInfo.StartTime = time.Now().Format(TIME_LAYOUT)
	spanindex, err := time.ParseDuration("8760h")
	if err != nil {
		beego.Error(err)
		return err
	}
	structEmployeeXM.EmployeInfo.EndTime = time.Now().Add(spanindex).Format(TIME_LAYOUT)
	structEmployeeXM.EmployeInfo.FaceFeature = strings.ToUpper(hex.EncodeToString(byteImage))
	for _, valueDevide := range arrayDevice {
		if valueDevide.Type == 1 {
			//雄迈
			structEmployeeXM.DeviceInfo.DeviceID = valueDevide.DevcieId
			err = facedevicexiongmai.AddEmployeeWSS(structEmployeeXM)
			if err != nil {
				beego.Error(err)
				return err
			}
		} else if valueDevide.Type == 0 {
			//海清
			err := facedevice.AddDeviceEmployee(valueDevide, structEmployeeHQ)
			if err != nil {
				beego.Error(err)
				return err
			}
		}

	}

	return nil
}

//GetPotoInfo __
func GetPotoInfo(structRequst CommonReplyJSON) error {
	for _, value := range structRequst.Body["face_info"].([]interface{}) {
		strMediaID, _ := value.(map[string]interface{})["media_id"].(string)
		mapPushUsrCmd[strMediaID] = structRequst
		structPhotoInfo, err := DownloadPoto(strMediaID)
		if err != nil {
			return err
		}

		err = AddPerson2DeviceGoRoutine(structRequst, structPhotoInfo.Body["binary"].([]byte))

		if err != nil {
			beego.Error(err)
			return err
		}

		arrayUserInfo := make([]UserInfoParam, 0)
		arrayUserInfo = append(arrayUserInfo,
			UserInfoParam{UserID: structRequst.Body["userid"].(string),
				Image:      structPhotoInfo.Body["binary"].([]byte),
				OperatorID: structRequst.Body["oper_id"].(string)})
		UploadUserInfo(arrayUserInfo)

	}

	return nil
}

//DownloadPoto _
func DownloadPoto(strMediaID string) (CommonReplyJSON, error) {
	var structPhotoInfo CommonReplyJSON

	u1, err := uuid.NewV4()
	if err != nil {
		beego.Error(err)
		return structPhotoInfo, err
	}

	var structRequst CommonCMDJSON
	structRequst.Cmd = "download_photo"
	structRequst.Body = make(map[string]interface{}, 0)
	structRequst.Body["media_id"] = strMediaID
	structRequst.Body["format_version"] = 1
	structRequst.Headers = make(map[string]string, 0)
	structRequst.Headers["req_id"] = u1.String()

	mapSend[structRequst.Headers["req_id"]] = structRequst
	byteRequst, err := json.Marshal(structRequst)
	if err != nil {
		beego.Error(err)
		return structPhotoInfo, err
	}

	err = SendPackageToWSS(byteRequst)
	if err != nil {
		beego.Error(err)
		return structPhotoInfo, err
	}

	timeSpan, err := time.ParseDuration("1s")
	if err != nil {
		beego.Error(err)
		return structPhotoInfo, err
	}

	time.Sleep(timeSpan)
	bFind := false
	for i := 0; i < 10; i++ {
		valueIndex, key := mapRecv[u1.String()]
		if key {
			structPhotoInfo = valueIndex
			bFind = true
			break
		}
		time.Sleep(timeSpan)
	}

	if bFind {
		return structPhotoInfo, nil
	}

	beego.Error("接收消息超时, reqid:", u1.String())
	return structPhotoInfo, fmt.Errorf("接收消息超时, reqid: %s", u1.String())
}

//GetScretNo __
func GetScretNo() error {

	u1, err := uuid.NewV4()
	if err != nil {
		beego.Error(err)
		return err
	}

	//beego.Debug("uid:", u1)

	nRangeNum := rand.Uint64()

	var structRequst CommonCMDJSON
	structRequst.Cmd = "get_secret_no"
	structRequst.Headers = make(map[string]string, 0)
	structRequst.Headers["req_id"] = u1.String()
	structRequst.Body = make(map[string]interface{}, 0)
	structRequst.Body["device_signature"] = ""
	structRequst.Body["nonce"] = nRangeNum
	structRequst.Body["timestamp"] = time.Now()
	structRequst.Body["sn"] = "07C9ADD061C20CF5"

	byteRequst, err := json.Marshal(structRequst)
	err = SendPackageToWSS(byteRequst)
	if err != nil {
		beego.Error(err)
		return err
	}
	return nil
}

//Register __
func Register() (CommonReplyJSON, error) {
	confIndex, err := goconfig.LoadConfigFile(mysqlitemanger.GetRootPath() + "/conf/config.ini")
	if err != nil {
		return CommonReplyJSON{}, err
	}

	strSN, err := confIndex.GetValue("DeviceInfo", "SN")
	if err != nil {
		beego.Error(err)
		return CommonReplyJSON{}, err
	}

	strSecreteNo, err := confIndex.GetValue("DeviceInfo", "SECRETNO")
	if err != nil {
		beego.Error(err)
		return CommonReplyJSON{}, err
	}

	sha1Handle := sha1.New() // md5加密类似md5.New()

	u1, err := uuid.NewV4()
	if err != nil {
		beego.Error(err)
		return CommonReplyJSON{}, err
	}

	nRangeNum := rand.Uint64()
	ntimeStamp := time.Now().Unix()

	var structRequst CommonCMDJSON
	structRequst.Cmd = "register"
	structRequst.Headers = make(map[string]string, 0)
	structRequst.Headers["req_id"] = u1.String()
	structRequst.Body = make(map[string]interface{}, 0)
	structRequst.Body["nonce"] = nRangeNum
	structRequst.Body["timestamp"] = ntimeStamp
	structRequst.Body["sn"] = strSN

	//07C9ADD061C20CF5   773fef975b6c6296aabc4fdd827fa649
	arrayShalcode := []string{strSN, strSecreteNo, fmt.Sprintf("%d", ntimeStamp), fmt.Sprintf("%d", nRangeNum), "register"}
	sort.Strings(arrayShalcode)
	//beego.Debug(arrayShalcode)
	for _, valueStr := range arrayShalcode {
		sha1Handle.Write([]byte(valueStr))
	}

	bs := sha1Handle.Sum(nil)
	structRequst.Body["device_signature"] = fmt.Sprintf("%x", bs)

	return SendMessage(structRequst)
}

func SendMessageEx(structRequst CommonCMDJSON) ([]byte, error) {
	byteRequst, err := json.Marshal(structRequst)
	if err != nil {
		beego.Error(err)
		return []byte(""), err
	}

	err = SendPackageToWSS(byteRequst)
	if err != nil {
		beego.Error(err)
		return []byte(""), err
	}

	timeSpan, err := time.ParseDuration("1s")
	if err != nil {
		beego.Error(err)
		return []byte(""), err
	}

	time.Sleep(timeSpan)

	value, key := mapRecvEx[structRequst.Headers["req_id"]]
	if !key {
		beego.Error("没有返回数据")
		return []byte(""), errors.New("没有返回数据")
	}

	return value, nil
}

//SendMessage __
func SendMessage(structRequst CommonCMDJSON) (CommonReplyJSON, error) {
	byteRequst, err := json.Marshal(structRequst)
	if err != nil {
		beego.Error(err)
		return CommonReplyJSON{}, err
	}

	err = SendPackageToWSS(byteRequst)
	if err != nil {
		beego.Error(err)
		return CommonReplyJSON{}, err
	}

	timeSpan, err := time.ParseDuration("1s")
	if err != nil {
		beego.Error(err)
		return CommonReplyJSON{}, err
	}

	time.Sleep(timeSpan)

	value, key := mapRecv[structRequst.Headers["req_id"]]
	if !key {
		beego.Error("没有返回数据")
		return CommonReplyJSON{}, errors.New("没有返回数据")
	}

	delete(mapRecv, structRequst.Headers["req_id"])

	if value.Errcode != 0 {
		beego.Error(value.Errmsg)
		return value, errors.New(value.Errmsg)
	}

	return value, nil
}

//ActiveDevice _
func ActiveDevice(strActiveCode string) (CommonReplyJSON, error) {
	u1, err := uuid.NewV4()
	if err != nil {
		beego.Error(err)
		return CommonReplyJSON{}, err
	}

	var structRequst CommonCMDJSON
	structRequst.Headers = make(map[string]string, 0)
	structRequst.Headers["req_id"] = u1.String()
	structRequst.Cmd = "active"
	structRequst.Body = make(map[string]interface{}, 0)
	structRequst.Body["active_code"] = strActiveCode

	return SendMessage(structRequst)
}

//SubScribeCorp __
func SubScribeCorp(strSecret string) (CommonReplyJSON, error) {
	u1, err := uuid.NewV4()
	if err != nil {
		beego.Error(err)
		return CommonReplyJSON{}, err
	}

	var structRequst CommonCMDJSON
	structRequst.Headers = make(map[string]string, 0)
	structRequst.Headers["req_id"] = u1.String()
	structRequst.Cmd = "subscribe_corp"
	structRequst.Body = make(map[string]interface{}, 0)
	structRequst.Body["secret"] = strSecret
	structRequst.Body["firmware_version"] = "version1"

	return SendMessage(structRequst)
}

type UserInfoParam struct {
	UserID     string
	Image      []byte
	OperatorID string
}

//UploadUserInfo __
func UploadUserInfo(arrayUser []UserInfoParam) (CommonReplyJSON, error) {
	u1, err := uuid.NewV4()
	if err != nil {
		beego.Error(err)
		return CommonReplyJSON{}, err
	}

	var structRequst CommonCMDJSON
	structRequst.Headers = make(map[string]string, 0)
	structRequst.Headers["req_id"] = u1.String()
	structRequst.Cmd = "upload_userinfo"
	structRequst.Body = make(map[string]interface{}, 0)

	arrayUserInfo := make([]interface{}, 0)

	for nIndex, value := range arrayUser {
		mapUserInfo := make(map[string]interface{}, 0)
		mapUserInfo["userid"] = value.UserID

		if nIndex == 0 {
			structRequst.Body["oper_id"] = value.OperatorID
			structRequst.Body["errcode"] = 0
			structRequst.Body["errmsg"] = "ok"
		}

		arrayFace := make([]interface{}, 0)

		mapFaceInfo := make(map[string]interface{}, 0)
		mapFaceInfo["id"] = 0
		mapFaceInfo["data"] = base64.StdEncoding.EncodeToString(value.Image)
		arrayFace = append(arrayFace, mapFaceInfo)
		mapUserInfo["fa_list"] = arrayFace

		arrayUserInfo = append(arrayUserInfo, mapUserInfo)
	}

	structRequst.Body["userinfo"] = arrayUserInfo

	return SendMessage(structRequst)
}

//Checkin __
func Checkin(strUserID string, nUserType int64, nTimeStamp int64) (CommonReplyJSON, error) {
	u1, err := uuid.NewV4()
	if err != nil {
		beego.Error(err)
		return CommonReplyJSON{}, err
	}

	var structRequst CommonCMDJSON
	structRequst.Headers = make(map[string]string, 0)
	structRequst.Headers["req_id"] = u1.String()
	structRequst.Cmd = "checkin"
	structRequst.Body = make(map[string]interface{}, 0)

	arrayCheckinData := make([]map[string]interface{}, 0)
	mapCheckinData := make(map[string]interface{}, 0)
	mapCheckinData["userid"] = strUserID
	mapCheckinData["user_type"] = nUserType
	mapCheckinData["timestamp"] = nTimeStamp

	arrayCheckinData = append(arrayCheckinData, mapCheckinData)
	structRequst.Body["checkin_data"] = arrayCheckinData

	return SendMessage(structRequst)
}

//GetUserInfoByPage 全量拉取用户数据
func GetUserInfoByPage() ([]byte, error) {
	u1, err := uuid.NewV4()
	if err != nil {
		beego.Error(err)
		return []byte(""), err
	}

	var structRequst CommonCMDJSON
	structRequst.Headers = make(map[string]string, 0)
	structRequst.Headers["req_id"] = u1.String()
	structRequst.Cmd = "get_userinfo_by_page"
	structRequst.Body = make(map[string]interface{}, 0)
	structRequst.Body["offset"] = 0
	structRequst.Body["limit"] = 200
	structRequst.Body["is_req_fp_info"] = 1
	structRequst.Body["is_req_fa_info"] = 1

	return SendMessageEx(structRequst)
}

type UserSearchParam struct {
	UserID   string
	UserType int64
}

//GetUserInfoByIDS __
func GetUserInfoByIDS(arraySearchParam []UserSearchParam) (CommonReplyJSON, error) {
	u1, err := uuid.NewV4()
	if err != nil {
		beego.Error(err)
		return CommonReplyJSON{}, err
	}

	var structRequest CommonCMDJSON
	structRequest.Cmd = "get_userinfo_by_ids"
	structRequest.Headers = make(map[string]string, 0)
	structRequest.Headers["req_id"] = u1.String()
	structRequest.Body = make(map[string]interface{}, 0)
	arrayUserItem := make([]interface{}, 0)
	for _, value := range arraySearchParam {
		mapUserItem := make(map[string]interface{})
		mapUserItem["userid"] = value.UserID
		mapUserItem["user_type"] = value.UserType

		arrayUserItem = append(arrayUserItem, mapUserItem)
	}

	structRequest.Body["user_item"] = arrayUserItem

	return SendMessage(structRequest)
}

//HeartGoRoutine _
func HeartGoRoutine() {
	for {
		u1, err := uuid.NewV4()
		if err != nil {
			beego.Error(err)
			return
		}

		var structHeart CommonCMDJSON
		structHeart.Cmd = "ping"
		structHeart.Headers = make(map[string]string, 0)
		structHeart.Headers["req_id"] = u1.String()

		_, err = SendMessage(structHeart)
		if err != nil {
			beego.Error(err)
			return
		}

		timeSpan, err := time.ParseDuration("20s")
		time.Sleep(timeSpan)

	}

	beego.Info("企业服务心跳Goroutine退出")
}
