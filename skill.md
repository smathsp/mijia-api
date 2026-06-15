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

## 登录

如果认证文件不存在，自动执行登录命令，将二维码链接展示给用户扫码：

```bash
mijia-api --list-devices
```

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

## 工作流程

1. 检查是否已安装，未安装则自动安装
2. 检查是否已登录，未登录则自动触发登录
3. 根据用户需求执行相应命令
4. 将结果返回给用户

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
mijia-api --list-devices
```

### 找不到设备

用 `--list-devices` 查看准确的设备名称和 DID。

### 属性设置失败

用 `--get-device-info` 确认设备支持该属性。
