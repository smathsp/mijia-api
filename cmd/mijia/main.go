package main

import (
	"encoding/json"
	stderrors "errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/smathsp/mijia-api/internal/api"
	"github.com/smathsp/mijia-api/internal/device"
	"github.com/smathsp/mijia-api/internal/errors"
	"github.com/smathsp/mijia-api/internal/logger"
)

const version = "4.0.1"

func main() {
	// Global flags
	authPath := flag.String("auth-path", defaultAuthPath(), "认证文件保存路径")
	listHomes := flag.Bool("list-homes", false, "列出家庭列表")
	listDevices := flag.Bool("list-devices", false, "列出所有米家设备，包含共享设备")
	listScenes := flag.Bool("list-scenes", false, "列出场景列表")
	listConsumable := flag.Bool("list-consumable-items", false, "列出耗材列表")
	getDeviceInfo := flag.String("get-device-info", "", "获取设备信息，指定设备model")
	showVersion := flag.Bool("version", false, "显示版本信息并退出")
	login := flag.Bool("login", false, "获取登录二维码链接（不阻塞）")

	logLevel := os.Getenv("MIJIA_LOG_LEVEL")
	if logLevel != "" {
		logger.SetLevelByName(logLevel)
	}

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Mijia API CLI (v%s)\n\n", version)
		fmt.Fprintf(os.Stderr, "用法:\n")
		fmt.Fprintf(os.Stderr, "  mijia-api [flags] <command> [args]\n\n")
		fmt.Fprintf(os.Stderr, "命令:\n")
		fmt.Fprintf(os.Stderr, "  get          获取设备属性\n")
		fmt.Fprintf(os.Stderr, "  set          设置设备属性\n")
		fmt.Fprintf(os.Stderr, "  run          使用自然语言控制小爱音箱\n\n")
		fmt.Fprintf(os.Stderr, "标志:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *showVersion {
		fmt.Printf("mijia-api %s\n", version)
		return
	}

	// Handle --get-device-info (no auth needed)
	if *getDeviceInfo != "" {
		handleGetDeviceInfo(*getDeviceInfo)
		return
	}

	// Handle --login (no auth needed, just get QR link)
	if *login {
		handleLogin(*authPath)
		return
	}

	// Handle subcommands
	args := flag.Args()
	if len(args) == 0 {
		// Handle top-level flags that need auth
		if *listHomes || *listDevices || *listScenes || *listConsumable {
			api := initAPI(*authPath)
			deviceMapping := make(map[string]map[string]interface{})
			homeMapping := make(map[string]map[string]interface{})

			if *listDevices {
				deviceMapping = handleListDevices(api, true)
			}
			if *listHomes {
				homeMapping = handleListHomes(api, true, deviceMapping)
			}
			if *listScenes {
				handleListScenes(api, true, homeMapping)
			}
			if *listConsumable {
				handleListConsumable(api, homeMapping)
			}
			return
		}
		flag.Usage()
		return
	}

	cmd := args[0]
	cmdArgs := args[1:]

	switch cmd {
	case "get":
		handleGet(*authPath, cmdArgs)
	case "set":
		handleSet(*authPath, cmdArgs)
	case "run":
		handleRun(*authPath, cmdArgs)
	default:
		fmt.Fprintf(os.Stderr, "未知命令: %s\n", cmd)
		flag.Usage()
		os.Exit(1)
	}
}

func initAPI(authPath string) *api.Client {
	client, err := api.NewClient(authPath)
	if err != nil {
		logger.Fatal("创建客户端失败: %v", err)
	}

	if !client.Available() {
		_, err := client.RefreshToken()
		if err != nil {
			// Token refresh failed, need to re-login
			_, err = client.Login()
			if err != nil {
				logger.Fatal("登录失败: %v", err)
			}
		}
	}

	return client
}

func handleGetDeviceInfo(model string) {
	info, err := device.GetDeviceInfo(model, "")
	if err != nil {
		logger.Fatal("获取设备信息失败: %v", err)
	}
	data, _ := json.MarshalIndent(info, "", "  ")
	fmt.Println(string(data))
}

func handleLogin(authPath string) {
	client, err := api.NewClient(authPath)
	if err != nil {
		logger.Fatal("创建客户端失败: %v", err)
	}

	// Check if already logged in
	if client.Available() {
		fmt.Println("已登录，无需重新登录")
		return
	}

	// Try to refresh token first
	_, err = client.RefreshToken()
	if err == nil {
		fmt.Println("Token 刷新成功")
		return
	}

	// Run full QR login flow (prints QR link and waits for scan)
	_, err = client.Login()
	if err != nil {
		fmt.Printf("登录失败: %v\n", err)
		return
	}

	fmt.Println("登录成功！")
}

func handleListDevices(api *api.Client, verbose bool) map[string]map[string]interface{} {
	devices, _ := api.GetDevicesListAll()
	shared, _ := api.GetSharedDevicesList()
	devices = append(devices, shared...)

	deviceMapping := make(map[string]map[string]interface{})
	for _, d := range devices {
		did, _ := d["did"].(string)
		deviceMapping[did] = d
	}

	if verbose {
		fmt.Println("设备列表:")
		for _, d := range devices {
			name, _ := d["name"].(string)
			did, _ := d["did"].(string)
			model, _ := d["model"].(string)
			online, _ := d["isOnline"].(bool)
			fmt.Printf("  - %s\n    did: %s\n    model: %s\n    online: %v\n", name, did, model, online)
		}
	}

	return deviceMapping
}

func handleListHomes(api *api.Client, verbose bool, deviceMapping map[string]map[string]interface{}) map[string]map[string]interface{} {
	homes, _ := api.GetHomesList()
	homeMapping := make(map[string]map[string]interface{})

	for _, h := range homes {
		homeID := fmt.Sprintf("%v", h["id"])
		homeMapping[homeID] = h
	}

	if verbose {
		if deviceMapping == nil {
			deviceMapping = handleListDevices(api, false)
		}

		fmt.Println("家庭列表:")
		for _, h := range homes {
			name, _ := h["name"].(string)
			id := fmt.Sprintf("%v", h["id"])
			address, _ := h["address"].(string)
			roomlist, _ := h["roomlist"].([]interface{})
			createTime, _ := h["create_time"].(float64)

			fmt.Printf("  - %s\n    ID: %s\n    地址: %s\n    房间数量: %d\n    创建时间: %s\n",
				name, id, address, len(roomlist),
				time.Unix(int64(createTime), 0).Format("2006-01-02 15:04:05"))

			fmt.Println("    房间列表:")
			for _, r := range roomlist {
				room, ok := r.(map[string]interface{})
				if !ok {
					continue
				}
				roomName, _ := room["name"].(string)
				roomID := fmt.Sprintf("%v", room["id"])
				dids, _ := room["dids"].([]interface{})
				roomCreate, _ := room["create_time"].(float64)

				var deviceNames []string
				for _, did := range dids {
					didStr := fmt.Sprintf("%v", did)
					if dm, ok := deviceMapping[didStr]; ok {
						if n, ok := dm["name"].(string); ok {
							deviceNames = append(deviceNames, n)
						} else {
							deviceNames = append(deviceNames, didStr)
						}
					} else {
						deviceNames = append(deviceNames, didStr)
					}
				}

				fmt.Printf("    - %s\n      ID: %s\n      设备列表: %s\n      创建时间: %s\n",
					roomName, roomID, strings.Join(deviceNames, ", "),
					time.Unix(int64(roomCreate), 0).Format("2006-01-02 15:04:05"))
			}
		}
	}

	return homeMapping
}

func handleListScenes(api *api.Client, verbose bool, homeMapping map[string]map[string]interface{}) map[string]map[string]interface{} {
	if homeMapping == nil {
		homeMapping = handleListHomes(api, false, nil)
	}

	sceneMapping := make(map[string]map[string]interface{})
	for homeID, home := range homeMapping {
		scenes, _ := api.GetScenesList(homeID)
		if verbose {
			homeName, _ := home["name"].(string)
			fmt.Printf("%s (%s) 中的场景:\n", homeName, homeID)
			for _, s := range scenes {
				name, _ := s["name"].(string)
				sceneID, _ := s["scene_id"].(string)
				createTime, _ := s["create_time"].(string)
				ts, _ := time.Parse("2006-01-02T15:04:05", createTime)
				fmt.Printf("  - %s\n    ID: %s\n    创建时间: %s\n", name, sceneID, ts.Format("2006-01-02 15:04:05"))
			}
		}
		for _, s := range scenes {
			sceneID, _ := s["scene_id"].(string)
			sceneMapping[sceneID] = s
		}
	}

	return sceneMapping
}

func handleListConsumable(api *api.Client, homeMapping map[string]map[string]interface{}) {
	if homeMapping == nil {
		homeMapping = handleListHomes(api, false, nil)
	}

	for homeID, home := range homeMapping {
		items, _ := api.GetConsumableItems(homeID)
		homeName, _ := home["name"].(string)
		fmt.Printf("%s (%s) 中的耗材:\n", homeName, homeID)
		for _, item := range items {
			name, _ := item["name"].(string)
			did, _ := item["did"].(string)
			details, _ := item["details"].(map[string]interface{})
			if details != nil {
				desc, _ := details["description"].(string)
				value, _ := details["value"].(string)
				fmt.Printf("  - %s(%s) 中的 %s\n    值: %s\n", name, did, desc, value)
			}
		}
	}
}

func handleGet(authPath string, args []string) {
	fs := flag.NewFlagSet("get", flag.ExitOnError)
	did := fs.String("did", "", "设备did")
	devName := fs.String("dev-name", "", "设备名称")
	propName := fs.String("prop-name", "", "属性名称")
	fs.Parse(args)

	if *propName == "" {
		fmt.Fprintln(os.Stderr, "错误: --prop-name 参数是必需的")
		os.Exit(1)
	}

	api := initAPI(authPath)
	dev, err := device.NewDevice(api, *did, *devName, 0.5)
	if err != nil {
		logger.Fatal("创建设备失败: %v", err)
	}

	value, err := dev.Get(*propName)
	if err != nil {
		logger.Fatal("获取属性失败: %v", err)
	}

	fmt.Printf("%s (%s) 的 %s 值为 %v\n", dev.Name, dev.DID, *propName, value)
}

func handleSet(authPath string, args []string) {
	fs := flag.NewFlagSet("set", flag.ExitOnError)
	did := fs.String("did", "", "设备did")
	devName := fs.String("dev-name", "", "设备名称")
	propName := fs.String("prop-name", "", "属性名称")
	value := fs.String("value", "", "属性值")
	fs.Parse(args)

	if *propName == "" || *value == "" {
		fmt.Fprintln(os.Stderr, "错误: --prop-name 和 --value 参数是必需的")
		os.Exit(1)
	}

	api := initAPI(authPath)
	dev, err := device.NewDevice(api, *did, *devName, 0.5)
	if err != nil {
		logger.Fatal("创建设备失败: %v", err)
	}

	if err := dev.Set(*propName, *value); err != nil {
		logger.Fatal("设置属性失败: %v", err)
	}

	fmt.Printf("%s (%s) 的 %s 值已设置为 %s\n", dev.Name, dev.DID, *propName, *value)
}

func handleRun(authPath string, args []string) {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	devName := fs.String("dev-name", "", "小爱音箱名称")
	quiet := fs.Bool("quiet", false, "静默执行")
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "错误: 请提供自然语言指令")
		os.Exit(1)
	}

	prompt := fs.Arg(0)
	api := initAPI(authPath)

	// Find WiFi speaker
	var wifispeaker *device.Device
	if *devName != "" {
		var err error
		wifispeaker, err = device.NewDevice(api, "", *devName, 0.5)
		if err != nil {
			logger.Fatal("创建设备失败: %v", err)
		}
	} else {
		// Find first xiaomi.wifispeaker
		devices, _ := api.GetDevicesListAll()
		for _, d := range devices {
			model, _ := d["model"].(string)
			if strings.Contains(model, "xiaomi.wifispeaker") {
				name, _ := d["name"].(string)
				var err error
				wifispeaker, err = device.NewDevice(api, "", name, 0.5)
				if err != nil {
					logger.Fatal("创建设备失败: %v", err)
				}
				break
			}
		}
		if wifispeaker == nil {
			logger.Fatal("未找到小爱音箱设备")
		}
	}

	quietVal := 0
	if *quiet {
		quietVal = 1
	}

	kwargs := map[string]interface{}{
		"_in": []interface{}{prompt, quietVal},
	}

	if err := wifispeaker.RunAction("execute-text-directive", nil, kwargs); err != nil {
		var deviceErr *errors.DeviceActionError
		if stderrors.As(err, &deviceErr) {
			fmt.Printf("执行小爱音箱指令失败: %v\n", err)
		} else {
			logger.Fatal("执行指令失败: %v", err)
		}
	}
}

func defaultAuthPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "auth.json"
	}
	return home + "/.config/mijia-api/auth.json"
}
