package errors

import "fmt"

// ErrorCode maps Xiaomi IoT error codes to Chinese descriptions.
// Source: https://github.com/kekeandzeyu/ha_xiaomi_home
var ErrorCode = map[int]string{
	-10000:    "未知错误",
	-10001:    "服务不可用",
	-10002:    "参数无效",
	-10003:    "资源不足",
	-10004:    "内部错误",
	-10005:    "权限不足",
	-10006:    "执行超时",
	-10007:    "设备离线或者不存在",
	-10020:    "未授权OAuth2",
	-10030:    "无效的token（HTTP）",
	-10040:    "无效的消息格式",
	-10050:    "无效的证书",
	-704000000: "未知错误",
	-704010000: "未授权（设备可能被删除）",
	-704014006: "没找到设备描述",
	-704030013: "Property不可读",
	-704030023: "Property不可写",
	-704030033: "Property不可订阅",
	-704040002: "Service不存在",
	-704040003: "Property不存在",
	-704040004: "Event不存在",
	-704040005: "Action不存在",
	-704040999: "功能未上线",
	-704042001: "Device不存在",
	-704042011: "设备离线",
	-704053036: "设备操作超时",
	-704053100: "设备在当前状态下无法执行此操作",
	-704083036: "设备操作超时",
	-704090001: "Device不存在",
	-704220008: "无效的ID",
	-704220025: "Action参数个数不匹配",
	-704220035: "Action参数错误",
	-704220043: "Property值错误",
	-704222034: "Action返回值错误",
	-705004000: "未知错误",
	-705004501: "未知错误",
	-705201013: "Property不可读",
	-705201015: "Action执行错误",
	-705201023: "Property不可写",
	-705201033: "Property不可订阅",
	-706012000: "未知错误",
	-706012013: "Property不可读",
	-706012015: "Action执行错误",
	-706012023: "Property不可写",
	-706012033: "Property不可订阅",
	-706012043: "Property值错误",
	-706014006: "没找到设备描述",
}

// LookupErrorCode returns the Chinese description for a given error code.
func LookupErrorCode(code int) string {
	if msg, ok := ErrorCode[code]; ok {
		return msg
	}
	return "未知错误"
}

// LoginError represents an authentication error.
type LoginError struct {
	Code    int
	Message string
}

func (e *LoginError) Error() string {
	return fmt.Sprintf("code: %d, message: %s", e.Code, e.Message)
}

// APIError represents an API request error.
type APIError struct {
	Code    int
	Message string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("code: %d, message: %s", e.Code, e.Message)
}

// DeviceNotFoundError is raised when a device is not found by DID.
type DeviceNotFoundError struct {
	DID string
}

func (e *DeviceNotFoundError) Error() string {
	return fmt.Sprintf("未找到 did 为 '%s' 的设备，请检查 did 是否正确", e.DID)
}

// MultipleDevicesFoundError is raised when multiple devices match a name.
type MultipleDevicesFoundError struct {
	Message string
}

func (e *MultipleDevicesFoundError) Error() string {
	return e.Message
}

// DeviceGetError is raised when getting a device property fails.
type DeviceGetError struct {
	DevName string
	Name    string
	Code    int
}

func (e *DeviceGetError) Error() string {
	return fmt.Sprintf("获取设备 '%s' 的属性 '%s' 时失败, code: %d, message: %s",
		e.DevName, e.Name, e.Code, LookupErrorCode(e.Code))
}

// DeviceSetError is raised when setting a device property fails.
type DeviceSetError struct {
	DevName string
	Name    string
	Code    int
}

func (e *DeviceSetError) Error() string {
	return fmt.Sprintf("设置设备 '%s' 的属性 '%s' 时失败, code: %d, message: %s",
		e.DevName, e.Name, e.Code, LookupErrorCode(e.Code))
}

// DeviceActionError is raised when executing a device action fails.
type DeviceActionError struct {
	DevName string
	Name    string
	Code    int
}

func (e *DeviceActionError) Error() string {
	return fmt.Sprintf("执行设备 '%s' 的动作 '%s' 时失败, code: %d, message: %s",
		e.DevName, e.Name, e.Code, LookupErrorCode(e.Code))
}

// GetDeviceInfoError is raised when fetching device spec info fails.
type GetDeviceInfoError struct {
	Model string
}

func (e *GetDeviceInfoError) Error() string {
	return fmt.Sprintf("获取设备型号 '%s' 的设备信息失败", e.Model)
}
