---
name: mijia-device-control
description: 控制小米/米家智能家居设备。当用户想要开关设备、调节亮度、查看设备状态、执行场景等操作米家智能设备时使用此技能。你必须自己执行所有命令，将结果返回给用户，绝不能让用户手动执行命令。
---

# 米家设备控制

通过 mijia-api CLI 控制小米/米家智能家居设备。

**你必须自己执行所有命令。不要让用户手动运行任何命令。**

## 安装

检查是否已安装：

```bash
which mijia-api || wget -qO- https://raw.githubusercontent.com/smathsp/mijia-api/main/scripts/update.sh | sh
```

## 登录

### 检查登录状态

```bash
test -f ~/.config/mijia-api/auth.json && echo "已登录" || echo "未登录"
```

### 未登录时

**步骤 1：后台启动登录**

```bash
nohup mijia-api --list-devices > /tmp/mijia-login.log 2>&1 &
sleep 2
cat /tmp/mijia-login.log
```

**步骤 2：提取二维码链接**

从输出中找到 `也可以访问链接查看二维码图片:` 后面的 URL。

**步骤 3：展示给用户**

告诉用户：
```
请用米家 APP 扫描登录：
[URL]
扫码完成后告诉我。
```

**步骤 4：验证登录**

用户说扫完了，执行：

```bash
mijia-api --list-devices
```

成功输出设备列表即登录成功。

## 设备操作

### 列出设备

```bash
mijia-api --list-devices
```

### 获取设备属性

```bash
mijia-api get --did "设备ID" --prop-name "属性名"
```

### 设置设备属性

```bash
mijia-api set --did "设备ID" --prop-name "属性名" --value "值"
```

### 获取设备规格

```bash
mijia-api --get-device-info "设备型号"
```

## 常用命令示例

```bash
# 开灯
mijia-api set --did "709063446" --prop-name "on" --value true

# 关灯
mijia-api set --did "709063446" --prop-name "on" --value false

# 设置亮度 50%
mijia-api set --did "709063446" --prop-name "brightness" --value 50

# 获取亮度
mijia-api get --did "709063446" --prop-name "brightness"

# 列出家庭
mijia-api --list-homes
```

## 常见属性

| 属性 | 说明 | 值 |
|------|------|-----|
| `on` | 开关 | `true`/`false` |
| `brightness` | 亮度 | 1-100 |
| `color-temperature` | 色温 | 2600-5100 |

不同设备支持不同属性，用 `--get-device-info` 确认。

## 故障排除

- 登录失败：`rm ~/.config/mijia-api/auth.json` 然后重新登录
- 设备离线：检查设备是否通电、网络是否正常
- 属性不支持：用 `--get-device-info` 查看设备支持的属性
