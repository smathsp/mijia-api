---
name: mijia-device-control
description: 控制小米/米家智能家居设备。当用户想要开关设备、调节亮度、查看设备状态、执行场景等操作米家智能设备时使用此技能。
---

# 米家设备控制

通过 mijia-api CLI 控制小米/米家智能家居设备。

## 安装

```bash
curl -sL https://raw.githubusercontent.com/smathsp/mijia-api/main/scripts/update.sh | bash
```

## 首次登录

```bash
mijia-api --list-devices
```

终端会显示二维码链接，用米家 APP 扫描登录。登录成功后认证自动保存。

## 常用命令

### 查看设备

```bash
# 列出所有设备
mijia-api --list-devices

# 列出家庭
mijia-api --list-homes
```

### 控制设备

```bash
# 开灯
mijia-api set --did "设备ID" --prop-name "on" --value true

# 关灯
mijia-api set --did "设备ID" --prop-name "on" --value false

# 设置亮度 (1-100)
mijia-api set --did "设备ID" --prop-name "brightness" --value 50

# 设置色温 (2600-5100)
mijia-api set --did "设备ID" --prop-name "color-temperature" --value 4000
```

### 查看状态

```bash
# 获取开关状态
mijia-api get --did "设备ID" --prop-name "on"

# 获取亮度
mijia-api get --did "设备ID" --prop-name "brightness"
```

### 获取设备属性

```bash
# 查看设备支持哪些属性
mijia-api --get-device-info "设备型号"
```

设备型号可通过 `--list-devices` 获取，例如 `yeelink.light.lamp28`。

## 工作流程

1. **先列出设备** → 获取 `did` 和 `model`
2. **查看设备属性** → 用 `--get-device-info model` 确认可用属性
3. **执行操作** → 用 `get`/`set` 命令控制设备

## 常见设备属性

| 设备类型 | 属性 | 说明 | 值范围 |
|---------|------|------|--------|
| 灯 | `on` | 开关 | `true`/`false` |
| 灯 | `brightness` | 亮度 | 1-100 |
| 灯 | `color-temperature` | 色温 | 2600-5100 |
| 插座 | `on` | 开关 | `true`/`false` |
| 鱼缸 | `on` | 开关 | `true`/`false` |
| 养生壶 | `on` | 开关 | `true`/`false` |

**注意**：不同设备支持的属性不同，请先用 `--get-device-info` 确认。

## 故障排除

### 登录失败

删除认证文件重新登录：
```bash
rm ~/.config/mijia-api/auth.json
mijia-api --list-devices
```

### 找不到设备

确认设备已在米家 APP 中添加，然后用 `--list-devices` 查看准确的设备名称和 DID。

### 属性设置失败

1. 确认设备在线
2. 用 `--get-device-info` 确认设备支持该属性
3. 检查值是否在有效范围内
