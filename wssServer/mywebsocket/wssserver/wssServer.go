package wssserver

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"time"

	"../echo"
	"../mysqlitemanger"
	"github.com/astaxie/beego"
	"github.com/gorilla/websocket"
)

//443
var addr = flag.String("addr", ":443", "http service address")

type ClientInfo struct {
	m_bSendHeart bool
	m_Conn       websocket.Conn
}

var mapClients = make(map[string]*ClientInfo, 0)

var upgrader = websocket.Upgrader{ // 解决跨域问题
	CheckOrigin: func(r *http.Request) bool {
		return true
	}} // use default options

type testWriteBackJSON struct {
	Cmd string `json:"cmd"`
	Ret int64  `json:"ret"`
	Msg string `json:"msg"`
}

// RecvTag 处理接收包
func RecvTag(conn *websocket.Conn) {
	conn_ := *conn
	for {
		_, key := echo.MapAddrAndConn[conn_.RemoteAddr().String()]
		if !key {
			break
		}

		if !echo.MapAddrAndConn[conn_.RemoteAddr().String()] {
			break
		}
		//mutexRecv.Lock()
		_, message, err := conn_.ReadMessage()
		//mutexRecv.Unlock()
		if err != nil {
			beego.Error(err)
			break
		}

		if len(message) < 1024 {
			beego.Debug("接收", string(message), "from:", conn_.RemoteAddr().String())
		}

		go echo.MessageParsing(message, conn_.RemoteAddr().String())

	}
}

//SendTag 发送数据
func SendTag(conn *websocket.Conn) {
	conn_ := *conn
	for {
		_, key := echo.MapAddrAndConn[conn_.RemoteAddr().String()]
		if !key {
			break
		}

		if !echo.MapAddrAndConn[conn_.RemoteAddr().String()] {
			break
		}

		byteReply := <-echo.MapIPSendChan[conn_.RemoteAddr().String()]

		var structCmmon echo.CommonRequstJSON
		err := json.Unmarshal(byteReply, &structCmmon)
		if err != nil {
			beego.Error(err)
			break
		}

		if len(byteReply) < 1024 {
			//			beego.Debug("发送数据", string(byteReply), "from:", conn_.RemoteAddr().String())
		}

		//mutexSend.Lock()
		err = conn_.WriteMessage(1, byteReply)
		if err != nil {
			beego.Error(err)
			break
		}

		if structCmmon.Cmd == "login_resp" {
			var structLoginResp echo.LoginReplyJSON
			err = json.Unmarshal(byteReply, &structLoginResp)
			if err != nil {
				beego.Error(err)
				break
			}
			if structLoginResp.Ret == 200 {
				//登录成功开启心跳包通讯的goroutine
				go SendHeartTag(&conn_)

				//设置数据服务地址
				byteRequst, err := echo.GetSetDataServerCmd()
				if err != nil {
					beego.Error(err)
					break
				}

				echo.MapIPSendChan[conn_.RemoteAddr().String()] <- byteRequst

				byteRequst, err = echo.GetAddGroupCmd()
				if err != nil {
					beego.Error(err)
					break
				}
				echo.MapIPSendChan[conn_.RemoteAddr().String()] <- byteRequst

				byteRequst, err = echo.GetAddForbidGroupCmd()
				if err != nil {
					beego.Error(err)
					break
				}
				echo.MapIPSendChan[conn_.RemoteAddr().String()] <- byteRequst

				byteRequst, err = echo.GetSetTimeCmd()
				if err != nil {
					beego.Error(err)
					break
				}
				echo.MapIPSendChan[conn_.RemoteAddr().String()] <- byteRequst
				//beego.Debug("添加到发送队列的数据:", string(byteRequst))
			} else {
				echo.MapAddrAndConn[conn_.RemoteAddr().String()] = false
				delete(echo.MapAddrAndConn, conn_.RemoteAddr().String())
			}

		}
		//mutexSend.Unlock()
	}

}

func SendMessage(conn *websocket.Conn, byteMessage []byte) error {
	conn_ := *conn

	err := conn_.WriteMessage(1, byteMessage)
	if err != nil {
		beego.Error(err)
		return err
	}

	return nil
}

func SendGoRoutine() {
	for {
		for strAddr, structClient := range mapClients {
			valueSend, keySend := echo.MapIPSendChan[strAddr]
			if !keySend {
				break
			}
			for {
				select {
				case byteReply := <-valueSend:
					{
						err := structClient.m_Conn.WriteMessage(1, byteReply)
						if err != nil {
							beego.Error(err)
							break
						}

						if len(byteReply) < 1024 {
							beego.Debug("发送数据", string(byteReply), "to:", strAddr)
						}

						mapReply := make(map[string]interface{}, 0)
						err = json.Unmarshal(byteReply, &mapReply)
						if mapReply["cmd"].(string) == "login_resp" {
							var structLoginResp echo.LoginReplyJSON
							err = json.Unmarshal(byteReply, &structLoginResp)
							if err != nil {
								beego.Error(err)
								break
							}
							if structLoginResp.Ret == 200 {
								//登录成功开启心跳包通讯的goroutine
								(mapClients[strAddr]).m_bSendHeart = true

								//设置数据服务地址
								byteRequst, err := echo.GetSetDataServerCmd()
								if err != nil {
									beego.Error(err)
									break
								}

								echo.MapIPSendChan[strAddr] <- byteRequst

								byteRequst, err = echo.GetAddGroupCmd()
								if err != nil {
									beego.Error(err)
									break
								}
								echo.MapIPSendChan[strAddr] <- byteRequst

								byteRequst, err = echo.GetAddForbidGroupCmd()
								if err != nil {
									beego.Error(err)
									break
								}
								echo.MapIPSendChan[strAddr] <- byteRequst

								byteRequst, err = echo.GetSetTimeCmd()
								if err != nil {
									beego.Error(err)
									break
								}
								echo.MapIPSendChan[strAddr] <- byteRequst
								//beego.Debug("添加到发送队列的数据:", string(byteRequst))
							} else {
								delete(echo.MapAddrAndConn, strAddr)
								delete(echo.MapIPSendChan, strAddr)
								delete(mapClients, strAddr)
							}
						}
					}
				default:
					{
						break
					}
				}
			}

			//mutexSend.Lock()

		}
	}
}

//SendHeartTag 发送心跳数据
func SendHeartTag(conn *websocket.Conn) {
	conn_ := *conn
	ticker := time.NewTicker(5 * time.Second)
	for {
		_, key := echo.MapAddrAndConn[conn_.RemoteAddr().String()]
		if !key {
			break
		}

		if !echo.MapAddrAndConn[conn_.RemoteAddr().String()] {
			break
		}

		<-ticker.C

		byteRequst, err := echo.GetHeartRequst()
		if err != nil {
			beego.Error("心跳包错误", err)
			break
		}
		//beego.Debug("发送心跳包:", string(byteRequst), "到:", conn.RemoteAddr().String())
		structDeviceInfo, err := mysqlitemanger.GetDeviceByAddr(conn_.RemoteAddr().String())
		if err != nil {
			beego.Error(err)
			return
		}
		if structDeviceInfo.Effect <= 0 {
			return
		}

		err = conn_.WriteMessage(1, byteRequst)

		if err != nil {
			beego.Error("发送心跳包错误:", err)
			echo.MapAddrAndConn[conn_.RemoteAddr().String()] = false
			delete(echo.MapAddrAndConn, conn_.RemoteAddr().String())
			return
		}

		var structHeartInfo echo.HeartDeviceInfo
		structHeartInfo.Addr = conn_.RemoteAddr().String()
		structHeartInfo.SendTime = time.Now()
		echo.PushHeart(structHeartInfo)
	}

}

func SendHeartGoRoutine() {
	span, err := time.ParseDuration("1s")
	if err != nil {
		beego.Error(err)
		return
	}
	for {
		for strAddr, structClient := range mapClients {
			if !structClient.m_bSendHeart {
				continue
			}

			byteRequst, err := echo.GetHeartRequst()
			if err != nil {
				beego.Error("心跳包错误", err)
				break
			}
			//beego.Debug("发送心跳包:", string(byteRequst), "到:", conn.RemoteAddr().String())
			structDeviceInfo, err := mysqlitemanger.GetDeviceByAddr(strAddr)
			if err != nil {
				beego.Error(err)
				return
			}
			if structDeviceInfo.Effect <= 0 {
				continue
			}

			err = structClient.m_Conn.WriteMessage(1, byteRequst)

			if err != nil {
				beego.Error("发送心跳包错误:", err)
				echo.MapAddrAndConn[strAddr] = false
				delete(echo.MapAddrAndConn, strAddr)
				continue
			}

			var structHeartInfo echo.HeartDeviceInfo
			structHeartInfo.Addr = strAddr
			structHeartInfo.SendTime = time.Now()
			echo.PushHeart(structHeartInfo)
		}

		time.Sleep(span)
	}

}

//Websocket  wss数据接收回调
func Websocket(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	beego.Debug("客户端：", c.RemoteAddr().String())

	echo.MapAddrAndConn[c.RemoteAddr().String()] = true
	echo.MapIPSendChan[c.RemoteAddr().String()] = make(chan []byte, 10)

	go RecvTag(c)
	go SendTag(c)

	// var pClientInfo *ClientInfo
	// pClientInfo.m_Conn = c
	// pClientInfo.m_bSendHeart = false

	// var clientInfo ClientInfo
	// clientInfo.m_Conn = *c
	// clientInfo.m_bSendHeart = false

	// mapClients[c.RemoteAddr().String()] = &clientInfo
	// mapClients[c.RemoteAddr().String()].m_bSendHeart = false
	//	defer c.Close()
}

func home(w http.ResponseWriter, r *http.Request) {

}

//CheckRemoveClient 检查关闭队列  并关闭设备连接
func CheckRemoveClient() {
	spanindex, err := time.ParseDuration("1s")
	if err != nil {
		beego.Error(err)
		return
	}
	for {
		for _, SerialValue := range echo.DeviceRemoveArray {
			//beego.Debug("一次循环")

			structDeviceInfo, err := mysqlitemanger.GetDeviceBySerial(SerialValue)
			if err != nil {
				beego.Error(err)
				return
			}

			beego.Debug("删除设备地址：", structDeviceInfo.Addr)
			_, key := echo.MapAddrAndConn[structDeviceInfo.Addr]
			if key {
				echo.MapAddrAndConn[structDeviceInfo.Addr] = false
				delete(echo.MapAddrAndConn, structDeviceInfo.Addr)
				//beego.Debug("关闭一个设备服务:", structDeviceInfo.Addr)
			}

			echo.DeviceRemoveArray = append([]string{})
		}

		time.Sleep(spanindex)
		//beego.Debug("echo.MapAddrAndConn：", echo.MapAddrAndConn)
	}

}

//StartServer  开启服务
func StartServer() {
	//go echo.CheckHearArray()
	//go CheckRemoveClient()
	//go SendGoRoutine()
	//go SendHeartGoRoutine()
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/websocket", Websocket)
	http.HandleFunc("/", home)
	beego.Info("开启服务")
	strRootPath := mysqlitemanger.GetRootPath()
	err := http.ListenAndServeTLS(*addr, strRootPath+"/ca.crt", strRootPath+"/ca.key", nil)
	//err := http.ListenAndServe(":7019", nil)
	if err != nil {
		beego.Error(err)
	}

}
