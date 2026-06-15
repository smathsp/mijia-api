package api

import (
	"fmt"

	"github.com/smathsp/mijia-api/internal/errors"
	"github.com/smathsp/mijia-api/internal/logger"
)

// CheckNewMsg checks for new messages.
func (c *Client) CheckNewMsg(beginAt int, refreshToken bool) (interface{}, error) {
	uri := "/v2/message/v2/check_new_msg"
	data := map[string]interface{}{
		"begin_at": beginAt,
	}
	return c.DoRequest(uri, data, refreshToken)
}

// GetHomesList returns all homes for the user.
func (c *Client) GetHomesList() ([]map[string]interface{}, error) {
	uri := "/v2/homeroom/gethome_merged"
	data := map[string]interface{}{
		"fg":              true,
		"fetch_share":     true,
		"fetch_share_dev": true,
		"fetch_cariot":    true,
		"limit":           300,
		"app_ver":         7,
		"plat_form":       0,
	}

	result, err := c.DoRequest(uri, data, true)
	if err != nil {
		return nil, err
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}

	homelist, ok := resultMap["homelist"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("homelist not found")
	}

	homes := make([]map[string]interface{}, len(homelist))
	for i, h := range homelist {
		homes[i], _ = h.(map[string]interface{})
	}
	return homes, nil
}

// GetDevicesList returns devices for a specific home.
func (c *Client) GetDevicesList(homeID string) ([]map[string]interface{}, error) {
	ownerUID, err := c.getHomeOwner(homeID)
	if err != nil {
		return nil, err
	}

	uri := "/home/home_device_list"
	startDID := ""
	hasMore := true
	var devices []map[string]interface{}

	for hasMore {
		data := map[string]interface{}{
			"home_owner":        ownerUID,
			"home_id":           parseInt(homeID),
			"limit":             200,
			"start_did":         startDID,
			"get_split_device":  true,
			"support_smart_home": true,
			"get_cariot_device": true,
			"get_third_device":  true,
		}

		ret, err := c.DoRequest(uri, data, true)
		if err != nil {
			return nil, err
		}

		retMap, ok := ret.(map[string]interface{})
		if !ok || retMap == nil {
			break
		}

		deviceInfo, ok := retMap["device_info"].([]interface{})
		if !ok || len(deviceInfo) == 0 {
			break
		}

		for _, d := range deviceInfo {
			if dm, ok := d.(map[string]interface{}); ok {
				dm["home_id"] = homeID
				devices = append(devices, dm)
			}
		}

		startDID, _ = retMap["max_did"].(string)
		hasMore, _ = retMap["has_more"].(bool)
		if startDID == "" {
			hasMore = false
		}
	}

	return devices, nil
}

// GetDevicesListAll returns all devices across all homes.
func (c *Client) GetDevicesListAll() ([]map[string]interface{}, error) {
	homes, err := c.GetHomesList()
	if err != nil {
		return nil, err
	}

	var allDevices []map[string]interface{}
	for _, home := range homes {
		homeID := fmt.Sprintf("%v", home["id"])
		devices, err := c.GetDevicesList(homeID)
		if err != nil {
			logger.Warning("获取家庭 %s 设备失败: %v", homeID, err)
			continue
		}
		allDevices = append(allDevices, devices...)
	}
	return allDevices, nil
}

// GetSharedDevicesList returns shared devices.
func (c *Client) GetSharedDevicesList() ([]map[string]interface{}, error) {
	uri := "/v2/home/device_list_page"
	data := map[string]interface{}{
		"ssid":               "<unknown ssid>",
		"bssid":              "02:00:00:00:00:00",
		"getVirtualModel":    true,
		"getHuamiDevices":    1,
		"get_split_device":   true,
		"support_smart_home": true,
		"get_cariot_device":  true,
		"get_third_device":   true,
		"get_phone_device":   true,
		"get_miwear_device":  true,
	}

	result, err := c.DoRequest(uri, data, true)
	if err != nil {
		return nil, err
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}

	list, ok := resultMap["list"].([]interface{})
	if !ok {
		return nil, nil
	}

	var devices []map[string]interface{}
	for _, item := range list {
		d, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if owner, ok := d["owner"].(bool); ok && owner {
			d["home_id"] = "shared"
			devices = append(devices, d)
		}
	}
	return devices, nil
}

// GetDevicesProp gets device properties.
func (c *Client) GetDevicesProp(data interface{}) (interface{}, error) {
	var params []interface{}
	switch d := data.(type) {
	case map[string]interface{}:
		params = []interface{}{d}
	case []interface{}:
		params = d
	default:
		return nil, fmt.Errorf("unsupported data type")
	}

	uri := "/miotspec/prop/get"
	requestData := map[string]interface{}{
		"params":     params,
		"datasource": 1,
	}

	result, err := c.DoRequest(uri, requestData, true)
	if err != nil {
		return nil, err
	}

	resultList, ok := result.([]interface{})
	if !ok {
		return result, nil
	}

	// Unwrap single item
	if _, isDict := data.(map[string]interface{}); isDict && len(resultList) == 1 {
		return resultList[0], nil
	}

	return result, nil
}

// SetDevicesProp sets device properties.
func (c *Client) SetDevicesProp(data interface{}) (interface{}, error) {
	var params []interface{}
	switch d := data.(type) {
	case map[string]interface{}:
		params = []interface{}{d}
	case []interface{}:
		params = d
	default:
		return nil, fmt.Errorf("unsupported data type")
	}

	uri := "/miotspec/prop/set"
	requestData := map[string]interface{}{
		"params": params,
	}

	result, err := c.DoRequest(uri, requestData, true)
	if err != nil {
		return nil, err
	}

	resultList, ok := result.([]interface{})
	if !ok {
		return result, nil
	}

	// Post-process results
	for _, item := range resultList {
		if r, ok := item.(map[string]interface{}); ok {
			code := 0
			if c, ok := r["code"].(float64); ok {
				code = int(c)
			}
			if code != 0 && code != 1 {
				r["message"] = errors.LookupErrorCode(code)
			} else {
				r["message"] = "成功"
			}
		}
	}

	// Unwrap single item
	if _, isDict := data.(map[string]interface{}); isDict && len(resultList) == 1 {
		return resultList[0], nil
	}

	return result, nil
}

// RunAction executes a device action.
func (c *Client) RunAction(data interface{}) (interface{}, error) {
	var params []interface{}
	switch d := data.(type) {
	case map[string]interface{}:
		params = []interface{}{d}
	case []interface{}:
		params = d
	default:
		return nil, fmt.Errorf("unsupported data type")
	}

	uri := "/miotspec/action"
	var retData []interface{}

	for _, param := range params {
		requestData := map[string]interface{}{
			"params": param,
		}

		ret, err := c.DoRequest(uri, requestData, true)
		if err != nil {
			return nil, err
		}
		retData = append(retData, ret)
	}

	// Post-process results
	for _, item := range retData {
		if r, ok := item.(map[string]interface{}); ok {
			code := 0
			if c, ok := r["code"].(float64); ok {
				code = int(c)
			}
			if code != 0 && code != 1 {
				r["message"] = errors.LookupErrorCode(code)
			} else {
				r["message"] = "成功"
			}
		}
	}

	// Unwrap single item
	if _, isDict := data.(map[string]interface{}); isDict && len(retData) == 1 {
		return retData[0], nil
	}

	return retData, nil
}

// GetScenesList returns scenes for a specific home.
func (c *Client) GetScenesList(homeID string) ([]map[string]interface{}, error) {
	ownerUID, err := c.getHomeOwner(homeID)
	if err != nil {
		return nil, err
	}

	uri := "/appgateway/miot/appsceneservice/AppSceneService/GetSimpleSceneList"
	data := map[string]interface{}{
		"app_version": 12,
		"get_type":    2,
		"home_id":     homeID,
		"owner_uid":   ownerUID,
	}

	result, err := c.DoRequest(uri, data, true)
	if err != nil {
		return nil, err
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return nil, nil
	}

	sceneList, ok := resultMap["manual_scene_info_list"].([]interface{})
	if !ok {
		return nil, nil
	}

	var scenes []map[string]interface{}
	for _, s := range sceneList {
		if scene, ok := s.(map[string]interface{}); ok {
			scene["home_id"] = homeID
			scenes = append(scenes, scene)
		}
	}
	return scenes, nil
}

// RunScene executes a manual scene.
func (c *Client) RunScene(sceneID, homeID string) (interface{}, error) {
	ownerUID, err := c.getHomeOwner(homeID)
	if err != nil {
		return nil, err
	}

	uri := "/appgateway/miot/appsceneservice/AppSceneService/NewRunScene"
	data := map[string]interface{}{
		"scene_id":   sceneID,
		"scene_type": 2,
		"phone_id":   "null",
		"home_id":    homeID,
		"owner_uid":  ownerUID,
	}

	return c.DoRequest(uri, data, true)
}

// GetConsumableItems returns consumable items for a specific home.
func (c *Client) GetConsumableItems(homeID string) ([]map[string]interface{}, error) {
	ownerUID, err := c.getHomeOwner(homeID)
	if err != nil {
		return nil, err
	}

	uri := "/v2/home/standard_consumable_items"
	data := map[string]interface{}{
		"home_id":       parseInt(homeID),
		"owner_id":      ownerUID,
		"filter_ignore": true,
	}

	result, err := c.DoRequest(uri, data, true)
	if err != nil {
		return nil, err
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return nil, nil
	}

	items, ok := resultMap["items"].([]interface{})
	if !ok || len(items) == 0 {
		return nil, nil
	}

	firstItem, ok := items[0].(map[string]interface{})
	if !ok {
		return nil, nil
	}

	consumesData, ok := firstItem["consumes_data"].([]interface{})
	if !ok {
		return nil, nil
	}

	var resultItems []map[string]interface{}
	for _, item := range consumesData {
		if m, ok := item.(map[string]interface{}); ok {
			// Flatten single-element details
			if details, ok := m["details"].([]interface{}); ok && len(details) == 1 {
				m["details"] = details[0]
			}
			m["home_id"] = homeID
			resultItems = append(resultItems, m)
		}
	}

	return resultItems, nil
}

// GetStatistics returns device statistics.
func (c *Client) GetStatistics(data interface{}) (interface{}, error) {
	var params []interface{}
	switch d := data.(type) {
	case map[string]interface{}:
		params = []interface{}{d}
	case []interface{}:
		params = d
	default:
		return nil, fmt.Errorf("unsupported data type")
	}

	uri := "/v2/user/statistics"
	var retData []interface{}

	for _, param := range params {
		ret, err := c.DoRequest(uri, param, true)
		if err != nil {
			return nil, err
		}
		retData = append(retData, ret)
	}

	// Unwrap single item
	if _, isDict := data.(map[string]interface{}); isDict && len(retData) == 1 {
		return retData[0], nil
	}

	return retData, nil
}

// Helper functions

func (c *Client) getHomeOwner(homeID string) (int, error) {
	homes, err := c.GetHomesList()
	if err != nil {
		return 0, err
	}

	for _, home := range homes {
		if fmt.Sprintf("%v", home["id"]) == homeID {
			if uid, ok := home["uid"].(float64); ok {
				return int(uid), nil
			}
			return 0, fmt.Errorf("uid not found in home data")
		}
	}

	return 0, &errors.APIError{Code: -1, Message: fmt.Sprintf("未找到 home_id=%s 的家庭信息", homeID)}
}

func parseInt(s string) int {
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}
