package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/smathsp/mijia-api/internal/crypto"
	"github.com/smathsp/mijia-api/internal/errors"
	"github.com/smathsp/mijia-api/internal/logger"
)

// DoRequest performs an encrypted API request.
func (c *Client) DoRequest(uri string, data interface{}, refreshToken bool) (interface{}, error) {
	logger.Debug("请求 URI: %s, 数据: %v", uri, data)

	if refreshToken {
		if _, err := c.RefreshToken(); err != nil {
			return nil, err
		}
	}

	fullURL := c.APIBaseURL + uri

	// Serialize data as compact JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	// Compact JSON (remove spaces)
	compact := strings.NewReplacer(" ", "", "\n", "", "\r", "", "\t", "").Replace(string(jsonData))

	params := map[string]string{
		"data": compact,
	}

	nonce, err := crypto.Nonce()
	if err != nil {
		return nil, err
	}

	signedNonce, err := crypto.SignedNonce(c.AuthData.SSecurity, nonce)
	if err != nil {
		return nil, err
	}

	params = crypto.GenerateEncParams(uri, "POST", signedNonce, nonce, params, c.AuthData.SSecurity)

	// Build form body
	form := url.Values{}
	for k, v := range params {
		form.Set(k, v)
	}

	// Create request
	req, err := http.NewRequest("POST", fullURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}

	// Set headers
	for k, v := range c.BuildHeaders() {
		req.Header.Set(k, v)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Try JSON parse first
	var retData map[string]interface{}
	if err := json.Unmarshal(body, &retData); err != nil {
		// Try decrypting the response
		decrypted, err := crypto.Decrypt(c.AuthData.SSecurity, nonce, string(body))
		if err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}
		if err := json.Unmarshal([]byte(decrypted), &retData); err != nil {
			return nil, fmt.Errorf("failed to parse decrypted response: %w", err)
		}
	}

	logger.Debug("响应数据: %v", retData)

	code := 0
	if c, ok := retData["code"].(float64); ok {
		code = int(c)
	}

	_, hasResult := retData["result"]

	if code != 0 || !hasResult {
		msg, _ := retData["message"].(string)
		if msg == "" {
			msg, _ = retData["desc"].(string)
		}
		if msg == "" {
			msg = "未知错误"
		}
		return nil, &errors.APIError{Code: code, Message: msg}
	}

	return retData["result"], nil
}
