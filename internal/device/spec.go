package device

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/smathsp/mijia-api/internal/errors"
	"github.com/smathsp/mijia-api/internal/logger"
)

const deviceURL = "https://home.miot-spec.com/spec/"

// Regex for extracting JSON from HTML script tag (compiled once)
var scriptRegex = regexp.MustCompile(`<script data-page="app" type="application/json">(.*?)</script>`)

// DeviceInfo holds parsed device specification.
type DeviceInfo struct {
	Name       string       `json:"name"`
	Model      string       `json:"model"`
	Properties []Property   `json:"properties"`
	Actions    []Action     `json:"actions"`
}

// Property represents a device property.
type Property struct {
	Name      string     `json:"name"`
	Desc      string     `json:"description"`
	Type      string     `json:"type"`
	RW        string     `json:"rw"`
	Range     []float64  `json:"range"`
	ValueList []VLItem   `json:"value-list"`
	Method    Method     `json:"method"`
}

// VLItem represents a value list item.
type VLItem struct {
	Value       interface{} `json:"value"`
	Description string      `json:"description"`
	DescZhCN    string      `json:"desc_zh_cn,omitempty"`
}

// Action represents a device action.
type Action struct {
	Name   string `json:"name"`
	Desc   string `json:"description"`
	Method Method `json:"method"`
}

// Method holds siid/piid/aiid for API calls.
type Method struct {
	SIID int `json:"siid"`
	PIID int `json:"piid,omitempty"`
	AIID int `json:"aiid,omitempty"`
}

// GetDeviceInfo fetches and parses device spec from miot-spec.com.
func GetDeviceInfo(model string, cachePath string) (*DeviceInfo, error) {
	// Check cache
	if cachePath != "" {
		cacheFile := filepath.Join(cachePath, model+".json")
		if data, err := os.ReadFile(cacheFile); err == nil {
			var info DeviceInfo
			if err := json.Unmarshal(data, &info); err == nil {
				logger.Debug("从缓存加载设备信息: %s", cacheFile)
				return &info, nil
			}
		}
	}

	// Fetch from network
	resp, err := http.Get(deviceURL + model)
	if err != nil {
		return nil, &errors.GetDeviceInfoError{Model: model}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, &errors.GetDeviceInfoError{Model: model}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &errors.GetDeviceInfoError{Model: model}
	}

	// Extract JSON from HTML script tag
	matches := scriptRegex.FindSubmatch(body)
	if matches == nil {
		return nil, &errors.GetDeviceInfoError{Model: model}
	}

	var content map[string]interface{}
	if err := json.Unmarshal(matches[1], &content); err != nil {
		return nil, &errors.GetDeviceInfoError{Model: model}
	}

	props, ok := content["props"].(map[string]interface{})
	if !ok {
		return nil, &errors.GetDeviceInfoError{Model: model}
	}

	product, ok := props["product"].(map[string]interface{})
	if !ok {
		return nil, &errors.GetDeviceInfoError{Model: model}
	}

	name, _ := product["name"].(string)
	devModel, _ := product["model"].(string)

	// Get i18n translations
	i18n, _ := props["i18n"].(map[string]interface{})
	i18nZh, _ := i18n["zh_cn"].(map[string]interface{})

	// Get services tree
	tree, _ := props["tree"].(map[string]interface{})
	services, _ := tree["services"].([]interface{})

	result := &DeviceInfo{
		Name:       name,
		Model:      devModel,
		Properties: []Property{},
		Actions:    []Action{},
	}

	propertiesName := make(map[string]bool)
	actionsName := make(map[string]bool)

	for _, svc := range services {
		svcMap, ok := svc.(map[string]interface{})
		if !ok {
			continue
		}

		siidFloat, ok := svcMap["iid"].(float64)
		if !ok {
			continue
		}
		siid := int(siidFloat)
		svcType, _ := svcMap["type"].(string)

		// Process properties
		if props, ok := svcMap["properties"].([]interface{}); ok {
			for _, p := range props {
				propMap, ok := p.(map[string]interface{})
				if !ok {
					continue
				}

				piidFloat, ok := propMap["iid"].(float64)
				if !ok {
					continue
				}
				piid := int(piidFloat)
				propFormat, _ := propMap["format"].(string)
				propType := normalizeType(propFormat)

				// Build rw string
				access, _ := propMap["access"].([]interface{})
				rw := ""
				for _, a := range access {
					if s, ok := a.(string); ok {
						if s == "read" {
							rw += "r"
						}
						if s == "write" {
							rw += "w"
						}
					}
				}

				// Get i18n description
				i18nKey := fmt.Sprintf("service:%03d:property:%03d", siid, piid)
				zhDesc, _ := i18nZh[i18nKey].(string)

				propName, _ := propMap["type"].(string)
				propDesc, _ := propMap["description"].(string)
				if zhDesc != "" {
					propDesc = propDesc + " / " + zhDesc
				} else {
					propDesc = strings.TrimRight(propDesc, " / ")
				}

				// Get value range
				var valueRange []float64
				if vr, ok := propMap["valueRange"].([]interface{}); ok {
					for _, v := range vr {
						if f, ok := v.(float64); ok {
							valueRange = append(valueRange, f)
						}
					}
				}

				// Get value list
				var valueList []VLItem
				if vl, ok := propMap["valueList"].([]interface{}); ok {
					for _, v := range vl {
						vlMap, ok := v.(map[string]interface{})
						if !ok {
							continue
						}
						vlItem := VLItem{
							Value:       vlMap["value"],
							Description: vlMap["description"].(string),
						}
						if i18nKey, ok := vlMap["i18nKey"].(string); ok {
							if zh, ok := i18nZh[i18nKey].(string); ok {
								vlItem.DescZhCN = zh
							}
						}
						valueList = append(valueList, vlItem)
					}
				}

				// Handle duplicate names
				if propertiesName[propName] {
					propName = svcType + "-" + propName
				}
				propertiesName[propName] = true

				prop := Property{
					Name:      propName,
					Desc:      propDesc,
					Type:      propType,
					RW:        rw,
					Range:     valueRange,
					ValueList: valueList,
					Method: Method{
						SIID: siid,
						PIID: piid,
					},
				}
				result.Properties = append(result.Properties, prop)
			}
		}

		// Process actions
		if actions, ok := svcMap["actions"].([]interface{}); ok {
			for _, a := range actions {
				actMap, ok := a.(map[string]interface{})
				if !ok {
					continue
				}

				aiidFloat, ok := actMap["iid"].(float64)
				if !ok {
					continue
				}
				aiid := int(aiidFloat)

				i18nKey := fmt.Sprintf("service:%03d:action:%03d", siid, aiid)
				zhDesc, _ := i18nZh[i18nKey].(string)

				actName, _ := actMap["type"].(string)
				actDesc, _ := actMap["description"].(string)
				if zhDesc != "" {
					actDesc = actDesc + " / " + zhDesc
				} else {
					actDesc = strings.TrimRight(actDesc, " / ")
				}

				// Handle duplicate names
				if actionsName[actName] {
					actName = svcType + "-" + actName
				}
				actionsName[actName] = true

				action := Action{
					Name: actName,
					Desc: actDesc,
					Method: Method{
						SIID: siid,
						AIID: aiid,
					},
				}
				result.Actions = append(result.Actions, action)
			}
		}
	}

	// Cache to file
	if cachePath != "" {
		if err := os.MkdirAll(cachePath, 0755); err == nil {
			cacheFile := filepath.Join(cachePath, model+".json")
			if data, err := json.MarshalIndent(result, "", "  "); err == nil {
				os.WriteFile(cacheFile, data, 0644)
				logger.Debug("缓存设备信息到: %s", cacheFile)
			}
		}
	}

	return result, nil
}

func normalizeType(format string) string {
	if strings.HasPrefix(format, "int") {
		return "int"
	}
	if strings.HasPrefix(format, "uint") {
		return "uint"
	}
	return format
}
