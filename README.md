# mijia-api

小米米家设备的 API，可以使用代码直接控制米家设备。Go 语言实现，单二进制文件，零依赖。

[![GitHub](https://img.shields.io/badge/GitHub-Do1e%2Fmijia--api-blue)](https://github.com/Do1e/mijia-api)
[![License: GPL-3.0](https://img.shields.io/badge/License-GPL--3.0-green.svg)](https://opensource.org/licenses/GPL-3.0)

常见问题见 [FAQ.md](FAQ.md)。

## 特性

- **零依赖** — 全部使用 Go 标准库，无需安装任何依赖
- **单二进制** — 编译后一个文件即可运行
- **低内存** — 运行时仅需 3-6MB，适合路由器等嵌入式设备
- **跨平台** — 支持 Linux/macOS/Windows，支持 ARM/aarch64/x86

## 安装

### 从源码构建

```bash
git clone https://github.com/Do1e/mijia-api.git
cd mijia-api
go build -o mijia-api ./cmd/mijia
```

### 交叉编译到路由器 (Linux aarch64)

```bash
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o mijia-api ./cmd/mijia
```

### 使用 Makefile

```bash
make build          # 本地构建
make build-router   # 交叉编译到 Linux aarch64
make build-armv7    # 交叉编译到 Linux armv7
make build-linux    # 交叉编译到 Linux amd64
```

## 使用

### 登录

首次使用需要通过二维码登录，认证数据将被保存以便后续使用：

```bash
# 运行任意命令即可触发登录
./mijia-api --list-devices
```

登录时会显示二维码链接，使用米家 APP 扫描即可完成身份验证。认证数据保存在 `~/.config/mijia-api/auth.json`。

### CLI 命令

#### 查看帮助

```bash
./mijia-api --help
```

#### 列出设备

```bash
# 列出所有设备
./mijia-api --list-devices

# 列出所有家庭
./mijia-api --list-homes

# 列出所有场景
./mijia-api --list-scenes

# 列出耗材
./mijia-api --list-consumable-items
```

#### 获取设备属性

```bash
# 通过设备名称获取
./mijia-api get --dev-name "台灯" --prop-name brightness

# 通过设备 DID 获取
./mijia-api get --did 709063446 --prop-name on
```

#### 设置设备属性

```bash
# 通过设备名称设置
./mijia-api set --dev-name "台灯" --prop-name brightness --value 60

# 打开设备
./mijia-api set --dev-name "台灯" --prop-name on --value true
```

#### 获取设备规格信息

```bash
# 获取设备支持的属性和动作
./mijia-api --get-device-info yeelink.light.lamp28
```

#### 控制小爱音箱

```bash
# 使用自然语言控制
./mijia-api run "打开卧室台灯"

# 指定小爱音箱
./mijia-api run --dev-name "客厅小爱" "播放音乐"

# 静默执行
./mijia-api run --quiet "关闭所有灯"
```

### 环境变量

| 环境变量 | 默认值 | 说明 |
|---------|--------|------|
| `MIJIA_LOG_LEVEL` | `INFO` | 日志级别：`DEBUG`, `INFO`, `WARNING`, `ERROR`, `FATAL` |

```bash
# 启用调试日志
MIJIA_LOG_LEVEL=DEBUG ./mijia-api --list-devices
```

## 部署到路由器

### 方式一：一键安装脚本（推荐）

在路由器上执行：

```bash
curl -sL https://raw.githubusercontent.com/Do1e/mijia-api/main/scripts/update.sh | bash
```

脚本会自动：
- 检测路由器架构（aarch64/armv7/amd64）
- 下载最新版本
- 安装到 `/usr/bin/mijia-api`

### 方式二：手动部署

#### 1. 交叉编译

```bash
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o mijia-api ./cmd/mijia
```

#### 2. 上传到路由器

```bash
scp mijia-api root@192.168.1.1:/usr/bin/mijia-api
```

#### 3. 复制认证文件

```bash
scp ~/.config/mijia-api/auth.json root@192.168.1.1:~/.config/mijia-api/auth.json
```

### 方式三：从 GitHub Release 下载

```bash
# 在路由器上执行
cd /tmp

# 下载最新版本（以 aarch64 为例）
VERSION=$(curl -s https://api.github.com/repos/Do1e/mijia-api/releases/latest | grep tag_name | cut -d '"' -f 4)
wget https://github.com/Do1e/mijia-api/releases/download/${VERSION}/mijia-api-linux-arm64

# 安装
chmod +x mijia-api-linux-arm64
mv mijia-api-linux-arm64 /usr/bin/mijia-api
```

### 更新

已安装后，运行更新脚本即可升级到最新版本：

```bash
curl -sL https://raw.githubusercontent.com/Do1e/mijia-api/main/scripts/update.sh | bash
```

### 在路由器上使用

```bash
ssh root@192.168.1.1

# 列出设备
mijia-api --list-devices

# 获取属性
mijia-api get --did 709063446 --prop-name brightness

# 设置属性
mijia-api set --did 709063446 --prop-name on --value true
```

## 认证文件兼容

Go 版本与 Python 版本的 `auth.json` 格式完全兼容，可以互换使用。

## 内存对比

| | Python | Go |
|---|---|---|
| 冷启动内存 | 15-30 MB | 2-5 MB |
| 运行时内存 | 20-40 MB | 3-6 MB |
| 二进制体积 | N/A | ~6 MB |

## 项目结构

```
cmd/mijia/main.go           # CLI 入口
internal/
  api/
    client.go               # 核心客户端
    auth.go                 # QR 登录、Token 刷新
    request.go              # 加密请求管道
    methods.go              # 所有 API 端点
  crypto/
    nonce.go                # nonce 生成
    rc4.go                  # RC4 加解密
    sign.go                 # 签名算法
  device/
    device.go               # 设备抽象层
    spec.go                 # 设备规格解析
  errors/errors.go          # 错误类型
  logger/logger.go          # 彩色日志
Makefile                    # 构建脚本
```

## 致谢

* [janzlan/mijia-api](https://gitee.com/janzlan/mijia-api/tree/master)
* [米家 APP 网络请求的抓包、加解密与构造的代码笔记](https://imkero.net/posts/mihome-app-api/)
* [al-one/hass-xiaomi-miot](https://github.com/al-one/hass-xiaomi-miot)
* [Squachen/micloud](https://github.com/Squachen/micloud) - 加密算法参考

## 开源许可

本项目采用 [GPL-3.0](LICENSE) 开源许可证。

**请注意：GPL-3.0 是具有"强传染性"的开源许可证。**  
如果您在您的项目中使用、修改或分发本项目的代码（包括作为库依赖），您的整个项目也**必须**以 GPL-3.0 或兼容许可证开源发布。

## 免责声明

* 本项目仅供学习交流使用，不得用于商业用途，如有侵权请联系删除
* 用户使用本项目所产生的任何后果，需自行承担风险
* 开发者不对使用本项目产生的任何直接或间接损失负责
