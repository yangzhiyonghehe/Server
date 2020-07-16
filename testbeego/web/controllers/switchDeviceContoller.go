package controllers

import (
	"encoding/json"

	"github.com/astaxie/beego"
	"golang.org/x/net/websocket"
)

var origin = "192.168.1.203:443/"

type SwtichDeviceAddController struct {
	beego.Controller
}

type SwitchDeviceRequst_Json struct {
	Sort   int64  `json:"sort"`
	Uuid   string `json:"uuid"`
	Uuname string `json:"uuname"`
}

type SwitchDeviceInfo_Json struct {
	Alarm     bool   `json:"alarm"`
	Area      string `json:"area"`
	Id        int64  `json:"id"`
	Linenum   int64  `json:"linenum"`
	Olstatus  bool   `json:"olstatus"`
	Sort      int64  `json:"sort"`
	Status    bool   `json:"status"`
	Usestatus int64  `json:"usestatus"`
	Uuid      string `json:"uuid"`
	Uuname    string `json:"uuname"`
}

type SwitchDeviceInfoArray_Json struct {
	Items []SwitchDeviceInfo_Json `json:"items"`
}

type SwitchDeviceReply_Json struct {
	Code    int64                      `json:"code"`
	Message string                     `json:"message"`
	Data    SwitchDeviceInfoArray_Json `json:"data"`
}

func (c *SwtichDeviceAddController) Post() {
	data := c.Ctx.Input.RequestBody

	var structRequst SwitchDeviceRequst_Json
	err := json.Unmarshal(data, &structRequst)
	if err != nil {
		beego.Error("添加闸机设备错误", err)
		return
	}

	var url = "wss://192.168.1.203:443//get_face_cfg"

	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		beego.Error(err)
		return
	}
	// message := []byte("hello, world!你好")
	// _, err = ws.Write(message)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Printf("Send: %s\n", message)

	var msg = make([]byte, 512)
	m, err := ws.Read(msg)
	if err != nil {
		beego.Error(err)
		return
	}
	beego.Debug("Receive:", msg[:m])

	ws.Close() //关闭连接

}
