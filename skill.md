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
wget -qO- https://raw.githubusercontent.com/smathsp/mijia-api/main/scripts/update.sh | sh
```

## 登录流程

认证文件路径：`~/.config/mijia-api/auth.json`

**如果认证文件不存在，需要用户扫码登录。**

### 检查是否已登录

```bash
test -f ~/.config/mijia-api/auth.json && echo "已登录" || echo "未登录"
```

### 未登录时的处理

执行以下命令会输出二维码链接，然后等待扫码（会阻塞约 120 秒）：

```bash
mijia-api --list-devices
```

**处理方式：**
1. 执行命令
2. 从输出中找到 `也可以访问链接查看二维码图片:` 后面的 URL
3. 将 URL 展示给用户，让用户用米家 APP 扫描
4. 告诉用户：扫码完成后告诉我，我来验证登录是否成功
5. 如果命令超时，重新执行验证命令

### 验证登录

```bash
mijia-api --list-devices
```

如果成功输出设备列表，说明登录成功。

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
