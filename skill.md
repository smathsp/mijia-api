---
name: mijia-device-control
description: 控制小米/米家智能家居设备。当用户想要开关设备、调节亮度、查看设备状态、执行场景等操作米家智能设备时使用此技能。你应该自己执行命令并返回结果给用户，不要让用户手动执行命令。
---

# 米家设备控制

通过 mijia-api CLI 控制小米/米家智能家居设备。

**重要：你应该自己执行所有命令，用户只需要看到结果。不要让用户手动运行命令。**

## 安装

如果 mijia-api 未安装，自动执行：

```bash
curl -sL https://raw.githubusercontent.com/smathsp/mijia-api/main/scripts/update.sh | bash
```

## 登录流程

认证文件路径：`~/.config/mijia-api/auth.json`

**如果认证文件不存在，需要用户扫码登录。按以下步骤操作：**

### 步骤 1：获取二维码链接

执行命令获取登录链接（会很快返回）：

```bash
mijia-api --list-devices 2>&1 | head -5
```

从输出中提取二维码链接，展示给用户：

```
请用米家 APP 扫描登录：
[二维码链接]
```

### 步骤 2：等待用户扫码

告诉用户：扫码完成后告诉我，我来验证登录是否成功。

**不要等待命令完成，直接告诉用户扫码。**

### 步骤 3：验证登录

用户说扫完了，执行命令验证：

```bash
mijia-api --list-devices
```

如果成功输出设备列表，说明登录成功。
如果失败，让用户重新扫码。

## 使用方式

### 查看设备

```bash
mijia-api --list-devices
```

### 控制设备

```bash
# 开灯
mijia-api set --did "设备ID" --prop-name "on" --value true

# 关灯
mijia-api set --did "设备ID" --prop-name "on" --value false

# 设置亮度 (1-100)
mijia-api set --did "设备ID" --prop-name "brightness" --value 50
```

### 查看状态

```bash
mijia-api get --did "设备ID" --prop-name "on"
```

### 获取设备属性

```bash
mijia-api --get-device-info "设备型号"
```

## 常见设备属性

| 设备类型 | 属性 | 说明 | 值范围 |
|---------|------|------|--------|
| 灯 | `on` | 开关 | `true`/`false` |
| 灯 | `brightness` | 亮度 | 1-100 |
| 灯 | `color-temperature` | 色温 | 2600-5100 |
| 插座 | `on` | 开关 | `true`/`false` |

**注意**：不同设备支持的属性不同，请先用 `--get-device-info` 确认。

## 故障排除

### 登录失败

删除认证文件重新登录：
```bash
rm ~/.config/mijia-api/auth.json
```

然后按登录流程重新操作。

### 找不到设备

用 `--list-devices` 查看准确的设备名称和 DID。

### 属性设置失败

用 `--get-device-info` 确认设备支持该属性。
