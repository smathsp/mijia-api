package device

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Do1e/mijia-api/internal/api"
	"github.com/Do1e/mijia-api/internal/errors"
	"github.com/Do1e/mijia-api/internal/logger"
)

// Device represents a Xiaomi Mijia device.
type Device struct {
	Client     *api.Client
	DID        string
	Name       string
	Model      string
	PropList   map[string]*Property
	ActionList map[string]*Action
	SleepTime  time.Duration
}

// NewDevice creates a new device instance.
func NewDevice(client *api.Client, did, devName string, sleepTime float64) (*Device, error) {
	if did == "" && devName == "" {
		return nil, fmt.Errorf("必须提供 did 或 dev_name 参数之一")
	}

	if did != "" && devName != "" {
		logger.Warning("同时提供了 did 和 dev_name 参数，将忽略 dev_name")
	}

	// Get all devices
	devices, err := client.GetDevicesListAll()
	if err != nil {
		return nil, err
	}

	// Add shared devices
	shared, err := client.GetSharedDevicesList()
	if err == nil {
		devices = append(devices, shared...)
	}

	var matchedDevice map[string]interface{}

	if did == "" {
		// Match by name
		var matches []map[string]interface{}
		for _, d := range devices {
			if name, ok := d["name"].(string); ok && name == devName {
				matches = append(matches, d)
			}
		}
		if len(matches) == 0 {
			return nil, &errors.DeviceNotFoundError{DID: devName}
		}
		if len(matches) > 1 {
			return nil, &errors.MultipleDevicesFoundError{
				Message: fmt.Sprintf("找到多个 dev_name 为 '%s' 的设备，请使用 did 参数指定具体设备或者修改设备名称以区分", devName),
			}
		}
		matchedDevice = matches[0]
		did, _ = matchedDevice["did"].(string)
	} else {
		// Match by DID
		for _, d := range devices {
			if dDID, ok := d["did"].(string); ok && dDID == did {
				matchedDevice = d
				break
			}
		}
		if matchedDevice == nil {
			return nil, &errors.DeviceNotFoundError{DID: did}
		}
		devName, _ = matchedDevice["name"].(string)
	}

	model, _ := matchedDevice["model"].(string)

	// Get device info/spec
	cachePath := ""
	authPath := client.AuthDataPath
	if idx := strings.LastIndex(authPath, "/"); idx > 0 {
		cachePath = authPath[:idx]
	}
	if idx := strings.LastIndex(authPath, "\\"); idx > 0 {
		cachePath = authPath[:idx]
	}

	devInfo, err := GetDeviceInfo(model, cachePath)
	if err != nil {
		return nil, err
	}

	dev := &Device{
		Client:     client,
		DID:        did,
		Name:       devName,
		Model:      model,
		PropList:   make(map[string]*Property),
		ActionList: make(map[string]*Action),
		SleepTime:  time.Duration(sleepTime * float64(time.Second)),
	}

	// Build property list with aliases
	for i := range devInfo.Properties {
		prop := &devInfo.Properties[i]
		dev.PropList[prop.Name] = prop
		if strings.Contains(prop.Name, "-") {
			alias := strings.ReplaceAll(prop.Name, "-", "_")
			dev.PropList[alias] = prop
		}
	}

	// Build action list
	for i := range devInfo.Actions {
		action := &devInfo.Actions[i]
		dev.ActionList[action.Name] = action
	}

	return dev, nil
}

// Get gets a device property value.
func (d *Device) Get(name string) (interface{}, error) {
	prop, ok := d.PropList[name]
	if !ok {
		return nil, fmt.Errorf("不支持的属性: %s", name)
	}

	if !strings.Contains(prop.RW, "r") {
		return nil, fmt.Errorf("属性 %s 不可读取", name)
	}

	method := map[string]interface{}{
		"siid": prop.Method.SIID,
		"piid": prop.Method.PIID,
		"did":  d.DID,
	}

	result, err := d.Client.GetDevicesProp(method)
	if err != nil {
		return nil, err
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}

	code := 0
	if c, ok := resultMap["code"].(float64); ok {
		code = int(c)
	}

	if code != 0 {
		return nil, &errors.DeviceGetError{DevName: d.Name, Name: name, Code: code}
	}

	time.Sleep(d.SleepTime)
	logger.Debug("获取属性: %s -> %s, 结果: %v", d.Name, name, resultMap)
	return resultMap["value"], nil
}

// Set sets a device property value.
func (d *Device) Set(name string, value interface{}) error {
	prop, ok := d.PropList[name]
	if !ok {
		return fmt.Errorf("不支持的属性: %s", name)
	}

	if !strings.Contains(prop.RW, "w") {
		return fmt.Errorf("属性 %s 不可写入", name)
	}

	// Validate and coerce value
	coercedValue, err := validateValue(prop, value)
	if err != nil {
		return err
	}

	method := map[string]interface{}{
		"siid":  prop.Method.SIID,
		"piid":  prop.Method.PIID,
		"did":   d.DID,
		"value": coercedValue,
	}

	result, err := d.Client.SetDevicesProp(method)
	if err != nil {
		return err
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected response format")
	}

	code := 0
	if c, ok := resultMap["code"].(float64); ok {
		code = int(c)
	}

	if code == 1 {
		logger.Warning("网关已经接收指令，无法判断是否设置成功: %s -> %s, 值: %v", d.Name, name, coercedValue)
	} else if code != 0 {
		return &errors.DeviceSetError{DevName: d.Name, Name: name, Code: code}
	}

	time.Sleep(d.SleepTime)
	logger.Debug("设置属性: %s -> %s, 值: %v, 结果: %v", d.Name, name, coercedValue, result)
	return nil
}

// RunAction executes a device action.
func (d *Device) RunAction(name string, value interface{}, kwargs map[string]interface{}) error {
	action, ok := d.ActionList[name]
	if !ok {
		return fmt.Errorf("不支持的动作: %s", name)
	}

	method := map[string]interface{}{
		"siid": action.Method.SIID,
		"aiid": action.Method.AIID,
		"did":  d.DID,
	}

	if value != nil {
		method["value"] = value
	}

	// Add kwargs
	for k, v := range kwargs {
		if strings.HasPrefix(k, "_") {
			k = k[1:]
		}
		if _, exists := method[k]; exists {
			return fmt.Errorf("无效的参数: %s. 请勿使用以下参数", k)
		}
		method[k] = v
	}

	result, err := d.Client.RunAction(method)
	if err != nil {
		return err
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected response format")
	}

	code := 0
	if c, ok := resultMap["code"].(float64); ok {
		code = int(c)
	}

	if code == 1 {
		logger.Warning("网关已经接收指令，无法判断是否执行成功: %s -> %s", d.Name, name)
	} else if code != 0 {
		return &errors.DeviceActionError{DevName: d.Name, Name: name, Code: code}
	}

	time.Sleep(d.SleepTime)
	logger.Debug("执行动作: %s -> %s, 结果: %v", d.Name, name, result)
	return nil
}

// String returns a human-readable representation of the device.
func (d *Device) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s (%s)\n", d.Name, d.Model))

	sb.WriteString("Properties:\n")
	for name, prop := range d.PropList {
		if strings.Contains(name, "_") {
			continue // Skip aliases
		}
		sb.WriteString(fmt.Sprintf("  %s: %s\n", prop.Name, prop.Desc))
		sb.WriteString(fmt.Sprintf("    valuetype: %s, rw: %s, range: %v\n", prop.Type, prop.RW, prop.Range))
		for _, vl := range prop.ValueList {
			sb.WriteString(fmt.Sprintf("    %v: %s\n", vl.Value, vl.Description))
		}
	}

	sb.WriteString("Actions:\n")
	for _, action := range d.ActionList {
		sb.WriteString(fmt.Sprintf("  %s: %s\n", action.Name, action.Desc))
	}

	return sb.String()
}

func validateValue(prop *Property, value interface{}) (interface{}, error) {
	switch prop.Type {
	case "bool":
		return validateBool(value)
	case "int", "uint":
		return validateInt(prop, value)
	case "float":
		return validateFloat(prop, value)
	case "string":
		if s, ok := value.(string); ok {
			return s, nil
		}
		return nil, fmt.Errorf("无效字符串值: %v", value)
	default:
		return nil, fmt.Errorf("不支持的类型: %s", prop.Type)
	}
}

func validateBool(value interface{}) (interface{}, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		switch strings.ToLower(v) {
		case "true":
			return true, nil
		case "false":
			return false, nil
		case "0":
			return false, nil
		case "1":
			return true, nil
		default:
			return nil, fmt.Errorf("无效布尔值: %s", v)
		}
	case int:
		if v == 0 {
			return false, nil
		}
		if v == 1 {
			return true, nil
		}
		return nil, fmt.Errorf("无效布尔值: %d", v)
	case int64:
		if v == 0 {
			return false, nil
		}
		if v == 1 {
			return true, nil
		}
		return nil, fmt.Errorf("无效布尔值: %d", v)
	case float64:
		if v == 0 {
			return false, nil
		}
		if v == 1 {
			return true, nil
		}
		return nil, fmt.Errorf("无效布尔值: %v", v)
	default:
		return nil, fmt.Errorf("无效布尔值: %v", value)
	}
}

func validateInt(prop *Property, value interface{}) (int, error) {
	var v int
	switch val := value.(type) {
	case int:
		v = val
	case int64:
		v = int(val)
	case float64:
		v = int(val)
	case string:
		var err error
		v, err = strconv.Atoi(val)
		if err != nil {
			return 0, fmt.Errorf("无效整数值: %s", val)
		}
	default:
		return 0, fmt.Errorf("无效整数值: %v", value)
	}

	if len(prop.Range) >= 2 {
		if v < int(prop.Range[0]) || v > int(prop.Range[1]) {
			return 0, fmt.Errorf("%d 超出数值范围, 应该在 [%.0f, %.0f] 之间", v, prop.Range[0], prop.Range[1])
		}
		if len(prop.Range) >= 3 && int(prop.Range[2]) != 1 {
			if (v-int(prop.Range[0]))%int(prop.Range[2]) != 0 {
				return 0, fmt.Errorf("无效的值: %d, 应该在范围 [%.0f, %.0f] 内且步长为 %.0f", v, prop.Range[0], prop.Range[1], prop.Range[2])
			}
		}
	}

	if prop.ValueList != nil {
		valid := false
		for _, item := range prop.ValueList {
			if iv, ok := item.Value.(float64); ok && int(iv) == v {
				valid = true
				break
			}
		}
		if !valid {
			return 0, fmt.Errorf("无效值: %d, 请使用有效的枚举值", v)
		}
	}

	return v, nil
}

func validateFloat(prop *Property, value interface{}) (float64, error) {
	var v float64
	switch val := value.(type) {
	case float64:
		v = val
	case int:
		v = float64(val)
	case int64:
		v = float64(val)
	case string:
		var err error
		v, err = strconv.ParseFloat(val, 64)
		if err != nil {
			return 0, fmt.Errorf("无效浮点值: %s", val)
		}
	default:
		return 0, fmt.Errorf("无效浮点值: %v", value)
	}

	if len(prop.Range) >= 2 {
		if v < prop.Range[0] || v > prop.Range[1] {
			return 0, fmt.Errorf("%.2f 超出数值范围, 应该在 [%.2f, %.2f] 之间", v, prop.Range[0], prop.Range[1])
		}
		if len(prop.Range) >= 3 && int(prop.Range[2]) != 0 {
			if int(v-prop.Range[0])%int(prop.Range[2]) != 0 {
				return 0, fmt.Errorf("无效的值: %.2f, 应该在范围 [%.2f, %.2f] 内且步长为 %.2f", v, prop.Range[0], prop.Range[1], prop.Range[2])
			}
		}
	}

	return v, nil
}
