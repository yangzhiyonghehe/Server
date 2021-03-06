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
	beego.Router("/api/management/user/login", &controllers.LoginController{})
	beego.Router("/api/management/user/logout", &controllers.LoginOutController{})
	beego.Router("/api/management/admin/register", &controllers.RegisterController{})
	beego.Router("/api/management/user/info", &controllers.InfoController{})
	beego.Router("/api/management/company/check", &controllers.CheckController{})
	beego.Router("/api/management/company/getCompanyInfo", &controllers.GetCompanyInfoController{})
	beego.Router("/api/management/admin/setDefaultCompany", &controllers.SetDefaultCompanyController{})
	beego.Router("/api/management/device/list", &controllers.FaceDeviceListController{})
	beego.Router("/api/management/device/deviceOptions", &controllers.DeviceOptionController{})
	beego.Router("/api/management/device/bind", &controllers.AddDeviceController{})
	beego.Router("/api/management/device/unbind", &controllers.RemoveDeviceController{})
	beego.Router("/api/management/device/selectedPersons", &controllers.DeviceSelectEmployeeController{})
	beego.Router("/api/management/device/update", &controllers.DeviceUpdateController{})
	beego.Router("/api/management/device/getExtend", &controllers.DeviceGetExtendController{})
	beego.Router("/api/management/device/version", &controllers.DeviceGetVersionController{})
	beego.Router("/api/management/device/truncate", &controllers.DeviceTruncateController{})
	beego.Router("/api/management/device/opendoor", &controllers.OpenDoorControol{})
	beego.Router("/api/management/device/devicesByGroup", &controllers.DeviceGroupController{})
	beego.Router("/api/management/device/status", &controllers.DeviceStatusController{})
	beego.Router("/api/management/device/keepOpendoor", &controllers.KeepDoorOpenController{})
	beego.Router("/api/management/device/saveGroup", &controllers.SaveGroupDeviceController{})
	beego.Router("/api/management/device/changePassword", &controllers.ChangeXMPswController{})
	beego.Router("/api/management/device/reboot", &controllers.RebootXMDeviceController{})
	beego.Router("/api/management/device/getLog", &controllers.GetLogController{})
	beego.Router("/api/management/device/getAlarmLog", &controllers.GetAlarmLogController{})
	beego.Router("/api/management/organization/getRootOrgOptions", &controllers.RootOrgOptiponsController{})
	beego.Router("/api/management/organization/list", &controllers.OrganizationlistController{})
	beego.Router("/api/management/organization/getAllOrgOptions", &controllers.GetAllOrgOptionsController{})
	beego.Router("/api/management/organization/create", &controllers.CreateGroupController{})
	beego.Router("/api/management/organization/update", &controllers.UpdateGroupController{})
	beego.Router("/api/management/organization/delete", &controllers.DeleteGroupController{})
	beego.Router("/api/management/person/personSelectedDevice", &controllers.GetPersonSelectDevcieController{})
	beego.Router("/api/management/person/delete", &controllers.PersonDeleteRequestController{})
	beego.Router("/api/management/person/update", &controllers.PersonUpdateController{})
	beego.Router("/api/management/person/list", &controllers.PersonlistController{})
	beego.Router("/api/management/person/personOrgOptions", &controllers.PerSonOrgOptionsController{})
	beego.Router("/api/management/person/create", &controllers.AddEmployeeContoller{})
	beego.Router("/api/management/person/sync", &controllers.SyncEmployeeController{})
	beego.Router("/api/management/person/unbind", &controllers.RemoveCompanyContoller{})
	beego.Router("/api/management/attendance_rule/list", &controllers.RuleGetListContrller{})
	beego.Router("/api/management/attendance_rule/create", &controllers.RuleAddController{})
	beego.Router("/api/management/attendance_rule/selectedPersons", &controllers.RuleSelectPersonControl{})
	beego.Router("/api/management/attendance_rule/update", &controllers.RuleUpdateController{})
	beego.Router("/api/management/attendance_rule/delete", &controllers.RuleDeleteController{})
	beego.Router("/api/management/attendance_rule/saveTakeTurns", &controllers.RuleSaveTakeTurnsController{})
	beego.Router("/api/management/attendance_rule/fetchTakeTurns", &controllers.FetchTakeTurnsController{})
	beego.Router("/api/management/attendance_rule/schedule", &controllers.SheduleContoller{})
	beego.Router("/api/management/attendance_rule/check_person", &controllers.CheckPersonController{})
	beego.Router("/api/management/attendance/list", &controllers.AttendanceListController{})
	beego.Router("/api/management/attendance/export", &controllers.AttendanceExport{})
	beego.Router("/api/management/attendance/snap", &controllers.SnapGetListController{})
	beego.Router("/api/management/circuit_breaker/lists", &controllers.DeviceSwitchController{})
	beego.Router("/api/management/dashboard/attendanceStatistics", &controllers.GetAttendanceStaticController{})
	beego.Router("/api/management/dashboard/deviceMap", &controllers.GetDeviceMapContoller{})
	beego.Router("/api/management/dashboard/newIOData", &controllers.GetNewIODataController{})
	beego.Router("/api/management/dashboard/statusStatistics", &controllers.GetStatusStatisticContoller{})
	beego.Router("/api/management/dashboard/statis", &controllers.GetDashboardController{})
	beego.Router("/api/management/company/fetch", &controllers.GetCompnyFetchContoller{})
	beego.Router("/api/management/company/update", &controllers.CompanyUpdateContoller{})
	beego.Router("/api/device/xm/upload", &controllers.XMUpdateContoller{})
	beego.Router("/api/management/admin/get_adminright", &controllers.GetAdminRightController{})
	beego.Router("/api/management/dashboard/io_data", &controllers.DashhoardController{})

}
