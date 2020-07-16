package check_device

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/astaxie/beego"
)

type DeviceCheckRequst_Json struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

type DeviceCheckReply_Json struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

var privateKey = []byte(`
-----BEGIN RSA PRIVATE KEY-----
MIICXAIBAAKBgQCWOXtvJRT8fJxdLGMFNCPhZaReDoHsMRF5qTpHd2QT5Kq9NLzn
WAH3BlHtJJvrz5guy9gsJCjToadl759h+LZqb3KmPrj/jMvTgYns7nHCkOVBw92S
8xrY4RFoYoubQyL15JehxW3KlTSyTl0i9lXKyFQ+KMKUGw2WKhAS+ySlSQIDAQAB
AoGACILPNHfUXY2tyjWWkpfmpIF+s3l88OXCyLLGw3/HIr1k0v1m6nB5BAbOo3Hc
h5qmU5hm8fFGgt74vfS6gfF2XDElIjLEZXfNn94c3gOTScKjUKBoMys23ZtlWek0
aG47MWKLAxKv/S0Kvo0+V7uPLTw72JVZoDtbnIzoTCNuBoUCQQDGu3+pEh9yOM1x
uLr+SK/HubmvYVFXll0cocXI4WWIjIvh4a/3OovRzJmkzv/3UuvJfPsnTlr89mLU
qU9JA9IzAkEAwYOOAweXfSl9FRNOc58WYEvrG2RcZjgaQiEVW8woK9MLgKZASa4M
kVtM2yKmam/SR+tE2o2xgAGOuLINJKNGkwJAeehGtWYCmESz8hDJ1HauLayGdUkT
ZtE8KPYrp8BsUkk0/ck98kCdyILjtS+t4P+i2CSsxD3Snt5dXerGUhnf9QJAHo+W
J+hVBlE9Dc0EwMHJGOAkeyj4ZrRJgVQUOXEejv0/fcvDr18rYPFYS+tG+Nw8C1ue
fh2OgLa+QXDtHnIivQJBAJfGINjcp2niEAB3SHF/Hqv/JGsxOiz31UGvDs2RAHLo
BKhU1hmcpKZUxWpgRaeFPLRiAd2kqDPRIMxcFXpk24Q=
-----END RSA PRIVATE KEY-----
`)

var publicKey = []byte(`
-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCWOXtvJRT8fJxdLGMFNCPhZaRe
DoHsMRF5qTpHd2QT5Kq9NLznWAH3BlHtJJvrz5guy9gsJCjToadl759h+LZqb3Km
Prj/jMvTgYns7nHCkOVBw92S8xrY4RFoYoubQyL15JehxW3KlTSyTl0i9lXKyFQ+
KMKUGw2WKhAS+ySlSQIDAQAB
-----END PUBLIC KEY-----
`)

func RsaDecrypt(ciphertext []byte) ([]byte, error) {
	block, _ := pem.Decode(privateKey)
	if block == nil {
		return nil, errors.New("private key error!")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return rsa.DecryptPKCS1v15(rand.Reader, priv, ciphertext)
}

func RsaEncrypt(origData []byte) ([]byte, error) {
	block, _ := pem.Decode(publicKey)
	if block == nil {
		return nil, errors.New("public key error")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	pub := pubInterface.(*rsa.PublicKey)
	return rsa.EncryptPKCS1v15(rand.Reader, pub, origData)
}

type DeviceInfo_Json struct {
	Time   int64  `json:"time"`
	Serial string `json:"serial"`
}

func DeviceCheck(strDeviceId string) error {
	var structRequst DeviceInfo_Json
	structRequst.Time = time.Now().Unix()
	structRequst.Serial = strDeviceId

	byteRequst, err := json.Marshal(structRequst)
	if err != nil {
		beego.Error(err)
		return err
	}

	//fmt.Println("发送数据:", string(byteRequst))
	//beego.Debug("发送数据:", string(byteRequst))

	data, err := RsaEncrypt(byteRequst)
	if err != nil {
		beego.Error(err)
		return err
	}

	strIndex := base64.StdEncoding.EncodeToString(data)

	strUrl := "http://zhikong.qidun.cn/api/management/device/check"
	var r http.Request
	r.ParseForm()
	r.Form.Add("key", strIndex)
	bodystr := strings.TrimSpace(r.Form.Encode())
	request, err := http.NewRequest("POST", strUrl, strings.NewReader(bodystr))
	if err != nil {
		beego.Error(err)
		return err
	}
	request.Header.Set("Content-Type", "application/form-data")
	request.Header.Set("Connection", "Keep-Alive")
	client := http.Client{}                                     //创建客户端
	resp, err := client.Do(request.WithContext(context.TODO())) //发送请求
	if err != nil {
		return errors.New("本操作需要联网，请打开网络")
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.New("本操作需要联网，请打开网络")
	}

	respBytes = bytes.TrimPrefix(respBytes, []byte("\xef\xbb\xbf"))

	//fmt.Println("返回数据:", string(respBytes))
	var unMarshJson DeviceCheckReply_Json
	err = json.Unmarshal(respBytes, &unMarshJson)
	if err != nil {
		beego.Error(err)
		return err
	}

	if unMarshJson.Code != 20000 {
		return errors.New(unMarshJson.Message)
	}

	indexData, err := base64.StdEncoding.DecodeString(unMarshJson.Data)
	if err != nil {
		beego.Error(err)
		return err
	}

	byteDecry, err := RsaDecrypt(indexData)

	if err != nil {
		return err
	}

	//fmt.Println("返回结果", string(byteDecry))
	//beego.Debug("检查设备接收数据:", string(byteDecry))

	var structDeviceInfo DeviceInfo_Json
	err = json.Unmarshal(byteDecry, &structDeviceInfo)
	if err != nil {
		return err
	}

	if structDeviceInfo.Serial != strDeviceId {
		return errors.New(unMarshJson.Message)
	}

	return nil
}
