package routers

import (
	"../controllers"
	"github.com/astaxie/beego"
)

func init() {
	beego.Router("/", &controllers.MainController{})
	beego.Router("/Subscribe/heartbeat", &controllers.HeartController{})
	beego.Router("/Subscribe/Snap", &controllers.SnapController{})
	beego.Router("/Subscribe/Verify", &controllers.VerifyController{})
	beego.Router("/api/device/xm/upload", &controllers.XMUpdateContoller{})
	beego.Router("/qiyeClient/GetServerConfig", &controllers.GetServerConfController{})
	beego.Router("/api/management/device/bind", &controllers.AddDeviceController{})
	beego.Router("/api/management/device/list", &controllers.FaceDeviceListController{})
	beego.Router("/api/management/device/unbind", &controllers.RemoveDeviceController{})
	beego.Router("/api/management/device/version", &controllers.DeviceGetVersionController{})
	beego.Router("/api/management/device/truncate", &controllers.DeviceTruncateController{})
	beego.Router("/api/management/device/reboot", &controllers.RebootXMDeviceController{})
	beego.Router("/api/management/Employee/Sync", &controllers.SyncEmployeeListController{})
	beego.Router("/api/management/Employee/list", &controllers.GetEmployeeListController{})
	beego.Router("/api/management/Employee/SyncOneByOne", &controllers.SyncEmployeeController{})
}
