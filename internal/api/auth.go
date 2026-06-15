package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/smathsp/mijia-api/internal/crypto"
	"github.com/smathsp/mijia-api/internal/errors"
	"github.com/smathsp/mijia-api/internal/logger"
)

// Login performs QR code login.
func (c *Client) Login() (*AuthData, error) {
	return c.QRLogin()
}

// GetQRLink returns the QR login URL without blocking.
// Use this for agents that need to show the QR link to users.
func (c *Client) GetQRLink() (string, error) {
	// Step 1: Get login location
	locationData, err := c.getLocation()
	if err != nil {
		return "", err
	}

	// Check if token is still valid
	if code, ok := locationData["code"]; ok && code == "0" {
		if msg, ok := locationData["message"]; ok && msg == "刷新Token成功" {
			if err := c.SaveAuthData(); err != nil {
				return "", err
			}
			return "", nil // Already logged in
		}
	}

	// Step 2: Get QR code URL
	locationData["theme"] = ""
	locationData["bizDeviceType"] = ""
	locationData["_hasLogo"] = "false"
	locationData["_qrsize"] = "240"
	locationData["_dc"] = fmt.Sprintf("%d", time.Now().UnixMilli())

	qrURL := c.LoginURL + "?" + encodeValues(locationData)

	headers := map[string]string{
		"User-Agent":      c.UserAgent(),
		"Accept-Encoding": "identity",
		"Content-Type":    "application/x-www-form-urlencoded",
		"Connection":      "keep-alive",
	}

	resp, err := c.doGet(qrURL, headers)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	loginData, err := c.handleRet(resp, true)
	if err != nil {
		return "", err
	}

	qrImage, _ := loginData["qr"].(string)

	// Store lpURL and other data for later polling
	c.pendingLoginData = loginData

	return qrImage, nil
}

// PollLogin waits for the user to scan the QR code.
// Call this after GetQRLink and the user has scanned.
func (c *Client) PollLogin() (*AuthData, error) {
	if c.pendingLoginData == nil {
		return nil, &errors.LoginError{Code: -1, Message: "没有待处理的登录请求，请先调用 GetQRLink"}
	}

	loginData := c.pendingLoginData
	lpURL, _ := loginData["lp"].(string)

	headers := map[string]string{
		"User-Agent":      c.UserAgent(),
		"Accept-Encoding": "identity",
		"Content-Type":    "application/x-www-form-urlencoded",
		"Connection":      "keep-alive",
	}

	// Long poll for login
	lpResp, err := c.doGetWithTimeout(lpURL, headers, 120*time.Second)
	if err != nil {
		return nil, &errors.LoginError{Code: -1, Message: "超时，请重试"}
	}
	defer lpResp.Body.Close()

	lpData, err := c.handleRet(lpResp, true)
	if err != nil {
		return nil, err
	}

	// Extract auth keys
	authKeys := []string{"psecurity", "nonce", "ssecurity", "passToken", "userId", "cUserId"}
	for _, key := range authKeys {
		if val, ok := lpData[key]; ok {
			switch v := val.(type) {
			case string:
				setAuthField(c.AuthData, key, v)
			case float64:
				setAuthField(c.AuthData, key, fmt.Sprintf("%d", int64(v)))
			}
		}
	}

	// Get service token from cookies
	callbackURL, _ := lpData["location"].(string)
	if callbackURL != "" {
		cbResp, err := c.doGet(callbackURL, headers)
		if err == nil {
			defer cbResp.Body.Close()
			for _, cookie := range cbResp.Cookies() {
				if cookie.Name == "serviceToken" {
					c.AuthData.ServiceToken = cookie.Value
				}
			}
		}
	}

	c.AuthData.ExpireTime = time.Now().Add(30 * 24 * time.Hour).UnixMilli()
	c.pendingLoginData = nil

	if err := c.SaveAuthData(); err != nil {
		return nil, err
	}

	logger.Info("登录成功")
	return c.AuthData, nil
}

// QRLogin performs QR code login flow.
func (c *Client) QRLogin() (*AuthData, error) {
	// Step 1: Get login location
	locationData, err := c.getLocation()
	if err != nil {
		return nil, err
	}

	// Check if token is still valid
	if code, ok := locationData["code"]; ok && code == "0" {
		if msg, ok := locationData["message"]; ok && msg == "刷新Token成功" {
			if err := c.SaveAuthData(); err != nil {
				return nil, err
			}
			logger.Info("刷新Token成功，无需登录")
			return c.AuthData, nil
		}
	}

	// Step 2: Get QR code URL
	locationData["theme"] = ""
	locationData["bizDeviceType"] = ""
	locationData["_hasLogo"] = "false"
	locationData["_qrsize"] = "240"
	locationData["_dc"] = fmt.Sprintf("%d", time.Now().UnixMilli())

	qrURL := c.LoginURL + "?" + encodeValues(locationData)

	headers := map[string]string{
		"User-Agent":      c.UserAgent(),
		"Accept-Encoding": "identity",
		"Content-Type":    "application/x-www-form-urlencoded",
		"Connection":      "keep-alive",
	}

	resp, err := c.doGet(qrURL, headers)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	loginData, err := c.handleRet(resp, true)
	if err != nil {
		return nil, err
	}

	loginURL, _ := loginData["loginUrl"].(string)
	qrImage, _ := loginData["qr"].(string)
	lpURL, _ := loginData["lp"].(string)

	logger.Info("请使用米家APP扫描下方二维码")
	fmt.Printf("也可以访问链接查看二维码图片: %s\n", qrImage)
	fmt.Printf("登录链接: %s\n", loginURL)

	// Step 3: Long poll for login
	lpResp, err := c.doGetWithTimeout(lpURL, headers, 120*time.Second)
	if err != nil {
		return nil, &errors.LoginError{Code: -1, Message: "超时，请重试"}
	}
	defer lpResp.Body.Close()

	lpData, err := c.handleRet(lpResp, true)
	if err != nil {
		return nil, err
	}

	// Step 4: Extract auth keys
	authKeys := []string{"psecurity", "nonce", "ssecurity", "passToken", "userId", "cUserId"}
	for _, key := range authKeys {
		if val, ok := lpData[key]; ok {
			switch v := val.(type) {
			case string:
				setAuthField(c.AuthData, key, v)
			case float64:
				setAuthField(c.AuthData, key, fmt.Sprintf("%d", int64(v)))
			}
		}
	}

	// Get service token from cookies
	callbackURL, _ := lpData["location"].(string)
	if callbackURL != "" {
		cbResp, err := c.doGet(callbackURL, headers)
		if err == nil {
			defer cbResp.Body.Close()
			for _, cookie := range cbResp.Cookies() {
				if cookie.Name == "serviceToken" {
					c.AuthData.ServiceToken = cookie.Value
				}
			}
		}
	}

	c.AuthData.ExpireTime = time.Now().Add(30 * 24 * time.Hour).UnixMilli()

	if err := c.SaveAuthData(); err != nil {
		return nil, err
	}

	logger.Info("登录成功")
	return c.AuthData, nil
}

// RefreshToken refreshes the authentication token.
func (c *Client) RefreshToken() (*AuthData, error) {
	if c.Available() {
		logger.Debug("Token 有效，无需刷新")
		return c.AuthData, nil
	}

	locationData, err := c.getLocation()
	if err != nil {
		return nil, err
	}

	if code, ok := locationData["code"]; ok && code == "0" {
		if msg, ok := locationData["message"]; ok && msg == "刷新Token成功" {
			if err := c.SaveAuthData(); err != nil {
				return nil, err
			}
			logger.Debug("刷新Token成功")
			return c.AuthData, nil
		}
	}

	return nil, &errors.LoginError{Code: -1, Message: "刷新Token失败，请重新登录"}
}

func (c *Client) getLocation() (map[string]string, error) {
	cookies := fmt.Sprintf("deviceId=%s;pass_o=%s;passToken=%s;userId=%d;cUserId=%s;uLocale=%s",
		c.DeviceID(), c.PassO(), c.AuthData.PassToken, c.AuthData.UserID, c.AuthData.CUserID, c.Locale)

	headers := map[string]string{
		"User-Agent":      c.UserAgent(),
		"Connection":      "keep-alive",
		"Accept-Encoding": "identity",
		"Content-Type":    "application/x-www-form-urlencoded",
		"Cookie":          cookies,
	}

	resp, err := c.doGet(c.ServiceLoginURL, headers)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	serviceData, err := c.handleRet(resp, false)
	if err != nil {
		return nil, err
	}

	location, _ := serviceData["location"].(string)

	if code, ok := serviceData["code"]; ok && code == "0" {
		// Token refresh flow
		req, err := http.NewRequest("GET", location, nil)
		if err != nil {
			return nil, err
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		ret, err := c.HTTPClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer ret.Body.Close()

		if ret.StatusCode == 200 {
			body, _ := io.ReadAll(ret.Body)
			if string(body) == "ok" {
				// Update cookies and ssecurity
				for _, cookie := range ret.Cookies() {
					switch cookie.Name {
					case "serviceToken":
						c.AuthData.ServiceToken = cookie.Value
					case "cUserId":
						c.AuthData.CUserID = cookie.Value
					}
				}
				if ssecurity, ok := serviceData["ssecurity"].(string); ok {
					c.AuthData.SSecurity = ssecurity
				}
				return map[string]string{"code": "0", "message": "刷新Token成功"}, nil
			}
		}
	}

	// Parse location URL query params
	if location != "" {
		parsed, err := url.Parse(location)
		if err != nil {
			return nil, err
		}
		params := parsed.Query()
		result := make(map[string]string)
		for k, v := range params {
			if len(v) > 0 {
				result[k] = v[0]
			}
		}
		return result, nil
	}

	// Convert map[string]interface{} to map[string]string
	result := make(map[string]string)
	for k, v := range serviceData {
		if s, ok := v.(string); ok {
			result[k] = s
		}
	}
	return result, nil
}

func (c *Client) doGet(rawURL string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return c.HTTPClient.Do(req)
}

func (c *Client) doGetWithTimeout(rawURL string, headers map[string]string, timeout time.Duration) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", rawURL, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return c.HTTPClient.Do(req)
}

func (c *Client) handleRet(resp *http.Response, verifyCode bool) (map[string]interface{}, error) {
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, &errors.LoginError{Code: resp.StatusCode, Message: string(body)}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	text := strings.Replace(string(body), "&&&START&&&", "", 1)

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(text), &data); err != nil {
		return nil, err
	}

	if verifyCode {
		if code, ok := data["code"].(float64); ok && code != 0 {
			desc, _ := data["desc"].(string)
			if desc == "" {
				desc = "未知错误"
			}
			return nil, &errors.LoginError{Code: int(code), Message: desc}
		}
	}

	return data, nil
}

func setAuthField(auth *AuthData, key, value string) {
	switch key {
	case "psecurity":
		// Not stored in AuthData struct, skip
	case "nonce":
		// Not stored in AuthData struct, skip
	case "ssecurity":
		auth.SSecurity = value
	case "passToken":
		auth.PassToken = value
	case "userId":
		fmt.Sscanf(value, "%d", &auth.UserID)
	case "cUserId":
		auth.CUserID = value
	case "serviceToken":
		auth.ServiceToken = value
	}
}

func encodeValues(m map[string]string) string {
	values := make(url.Values)
	for k, v := range m {
		values.Set(k, v)
	}
	return values.Encode()
}

// Ensure crypto is used (imported via blank import in other files)
var _ = crypto.EncryptRC4
