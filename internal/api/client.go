package api

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	APIBaseURL      = "https://api.mijia.tech/app"
	LoginURL        = "https://account.xiaomi.com/longPolling/loginUrl"
	ServiceLoginURL = "https://account.xiaomi.com/pass/serviceLogin?_json=true&sid=mijia&_locale=%s"
)

// Client is the Xiaomi Mijia API client.
type Client struct {
	HTTPClient      *http.Client
	AuthData        *AuthData
	AuthDataPath    string
	Locale          string
	APIBaseURL      string
	LoginURL        string
	ServiceLoginURL string

	// Caching
	availableCache     bool
	availableCacheTime int64
	passO              string
	userAgent          string
	deviceID           string
}

// AuthData holds authentication credentials.
type AuthData struct {
	UA           string `json:"ua"`
	SSecurity    string `json:"ssecurity"`
	UserID       int64  `json:"userId"`
	CUserID      string `json:"cUserId"`
	ServiceToken string `json:"serviceToken"`
	PassToken    string `json:"passToken"`
	DeviceID     string `json:"deviceId"`
	PassO        string `json:"pass_o"`
	ExpireTime   int64  `json:"expireTime"`
	SaveTime     int64  `json:"saveTime"`
	CountryCode  string `json:"countryCode"`
}

// NewClient creates a new API client with the given auth data path.
func NewClient(authDataPath string) (*Client, error) {
	locale := detectLocale()

	if authDataPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		authDataPath = filepath.Join(home, ".config", "mijia-api", "auth.json")
	} else {
		info, err := os.Stat(authDataPath)
		if err == nil && info.IsDir() {
			authDataPath = filepath.Join(authDataPath, "auth.json")
		}
	}

	c := &Client{
		HTTPClient:      &http.Client{Timeout: 30 * time.Second},
		AuthData:        &AuthData{},
		AuthDataPath:    authDataPath,
		Locale:          locale,
		APIBaseURL:      APIBaseURL,
		LoginURL:        LoginURL,
		ServiceLoginURL: fmt.Sprintf(ServiceLoginURL, locale),
	}

	// Load existing auth data
	if data, err := os.ReadFile(authDataPath); err == nil {
		if err := json.Unmarshal(data, c.AuthData); err == nil {
			// Auth data loaded successfully
		}
	}

	return c, nil
}

// Available checks if the client has valid authentication.
func (c *Client) Available() bool {
	if c.AuthData.SSecurity == "" || c.AuthData.UserID == 0 || c.AuthData.CUserID == "" || c.AuthData.ServiceToken == "" {
		return false
	}

	now := time.Now().Unix()
	if now-c.availableCacheTime < 60 {
		return c.availableCache
	}

	_, err := c.CheckNewMsg(int(now)-3600, false)
	if err != nil {
		c.availableCache = false
		c.availableCacheTime = 0
		return false
	}

	c.availableCache = true
	c.availableCacheTime = now
	return true
}

// PassO returns or generates the pass_o value.
func (c *Client) PassO() string {
	if c.AuthData.PassO != "" {
		return c.AuthData.PassO
	}
	c.AuthData.PassO = randomHex(16)
	return c.AuthData.PassO
}

// UserAgent returns or generates the User-Agent string.
func (c *Client) UserAgent() string {
	if c.AuthData.UA != "" {
		return c.AuthData.UA
	}

	country := "CN"
	if c.Locale != "" && strings.Contains(c.Locale, "_") {
		country = strings.Split(c.Locale, "_")[1]
	}

	uaID1 := randomHexUpper(40)
	uaID2 := randomHexUpper(32)
	uaID3 := randomHexUpper(32)
	uaID4 := randomHexUpper(40)

	c.AuthData.UA = fmt.Sprintf(
		"Android-15-11.0.701-Xiaomi-23046RP50C-OS2.0.212.0.VMYCNXM-%s-%s-%s-%s-SmartHome-MI_APP_STORE-%s|%s|%s-64",
		uaID1, country, uaID3, uaID2, uaID1, uaID4, c.PassO(),
	)
	return c.AuthData.UA
}

// DeviceID returns or generates the device ID.
func (c *Client) DeviceID() string {
	if c.AuthData.DeviceID != "" {
		return c.AuthData.DeviceID
	}
	c.AuthData.DeviceID = randomAlphanum(16)
	return c.AuthData.DeviceID
}

// SaveAuthData saves the current auth data to disk.
func (c *Client) SaveAuthData() error {
	c.AuthData.SaveTime = time.Now().UnixMilli()

	dir := filepath.Dir(c.AuthDataPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c.AuthData, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(c.AuthDataPath, data, 0600)
}

// BuildCookie builds the Cookie header for API requests.
func (c *Client) BuildCookie() string {
	now := time.Now()
	_, offset := now.Zone()
	hours := offset / 3600
	minutes := (offset % 3600) / 60

	country := "CN"
	if c.Locale != "" && strings.Contains(c.Locale, "_") {
		country = strings.Split(c.Locale, "_")[1]
	}

	tzName := "Asia/Shanghai"
	if loc, err := time.LoadLocation(""); err == nil {
		tzName = loc.String()
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "cUserId=%s;", c.AuthData.CUserID)
	fmt.Fprintf(&sb, "yetAnotherServiceToken=%s;", c.AuthData.ServiceToken)
	fmt.Fprintf(&sb, "serviceToken=%s;", c.AuthData.ServiceToken)
	fmt.Fprintf(&sb, "timezone_id=%s;", tzName)
	fmt.Fprintf(&sb, "timezone=GMT%+02d:%02d;", hours, minutes)
	fmt.Fprintf(&sb, "is_daylight=%d;", boolToInt(time.Now().IsDST()))
	fmt.Fprintf(&sb, "dst_offset=%d;", boolToInt64(time.Now().IsDST())*60*60*1000)
	fmt.Fprintf(&sb, "channel=MI_APP_STORE;")
	fmt.Fprintf(&sb, "countryCode=%s;", country)
	fmt.Fprintf(&sb, "PassportDeviceId=%s;", c.DeviceID())
	fmt.Fprintf(&sb, "locale=%s", c.Locale)

	return sb.String()
}

// BuildHeaders returns the default headers for API requests.
func (c *Client) BuildHeaders() map[string]string {
	return map[string]string{
		"User-Agent":                 c.UserAgent(),
		"accept-encoding":           "identity",
		"Content-Type":              "application/x-www-form-urlencoded",
		"miot-accept-encoding":      "GZIP",
		"miot-encrypt-algorithm":    "ENCRYPT-RC4",
		"x-xiaomi-protocal-flag-cli": "PROTOCAL-HTTP2",
		"Cookie":                    c.BuildCookie(),
	}
}

// Helper functions

func detectLocale() string {
	// Try to detect from environment or system
	if lang := os.Getenv("LANG"); lang != "" {
		if strings.Contains(lang, "_") {
			return lang
		}
	}
	return "zh_CN"
}

func randomHex(n int) string {
	const hex = "0123456789abcdef"
	b := make([]byte, n)
	for i := range b {
		b[i] = hex[rand.Intn(len(hex))]
	}
	return string(b)
}

func randomHexUpper(n int) string {
	const hex = "0123456789ABCDEF"
	b := make([]byte, n)
	for i := range b {
		b[i] = hex[rand.Intn(len(hex))]
	}
	return string(b)
}

func randomAlphanum(n int) string {
	const chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_-"
	b := make([]byte, n)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func boolToInt64(b bool) int64 {
	if b {
		return 1
	}
	return 0
}
