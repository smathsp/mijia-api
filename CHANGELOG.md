# Changelog

## [4.0.0] - 2026-06-15

### 重构

- **完全用 Go 重写**，从 Python 迁移到 Go
- 零外部依赖，全部使用 Go 标准库
- 单二进制文件，静态编译

### 性能

- 冷启动内存：15-30 MB → 2-5 MB（降低 5-10 倍）
- 运行时内存：20-40 MB → 3-6 MB（降低 5-10 倍）
- 二进制体积：~6 MB（含所有依赖）

### 新增

- GitHub Actions 自动构建发布
- 路由器一键安装/更新脚本
- 支持 Linux/macOS/Windows 多平台
- 支持 aarch64/armv7/amd64 多架构

### 移除

- Python 运行时依赖
- pip/uv 安装方式

### 迁移指南

从 Python 版本迁移：

```bash
# 安装 Go 版本
curl -sL https://raw.githubusercontent.com/smathsp/mijia-api/main/scripts/update.sh | bash

# 认证文件兼容，无需重新登录
# auth.json 格式完全一致
```

CLI 命令变化：

| Python | Go |
|--------|-----|
| `python -m mijiaAPI` | `./mijia-api` |
| `--prop_name` | `--prop-name` |
| `--dev_name` | `--dev-name` |
| `--list_devices` | `--list-devices` |
| `--list_homes` | `--list-homes` |
| `True`/`False` | `true`/`false` |

---

## Python 版本历史

以下为 Python 版本的更新记录，Go 版本从 v4.0.0 开始。

### [3.2.0](https://github.com/smathsp/mijia-api/compare/v3.1.0...v3.2.0) - 2026-06-09

- 将 `--run` 参数重构为独立的 `run` 子命令

### [3.1.0](https://github.com/smathsp/mijia-api/compare/v3.0.5...v3.1.0) - 2026-05-27

- 删除设备属性的单位 `unit` 属性
- 适配 home.miot-spec.com 新的规格页格式

### [3.0.5](https://github.com/smathsp/mijia-api/compare/v3.0.4...v3.0.5) - 2026-01-24

- 蓝牙设备控制返回 code 为 1 时视为成功

### [3.0.4](https://github.com/smathsp/mijia-api/compare/v3.0.3...v3.0.4) - 2026-01-12

- 修复刷新 token 后依然提示不可用的问题

### [3.0.3](https://github.com/smathsp/mijia-api/compare/v3.0.2...v3.0.3) - 2026-01-02

- 新增 `MIJIA_LOG_LEVEL` 环境变量支持

### [3.0.2](https://github.com/smathsp/mijia-api/compare/v3.0.1...v3.0.2) - 2026-01-01

- 为 `available` 属性添加缓存机制

### [3.0.1](https://github.com/smathsp/mijia-api/compare/v3.0.0...v3.0.1) - 2025-12-09

- 新增 `get_shared_devices_list()` API
- 修复 alpine 下 locale 问题

### [3.0.0](https://github.com/smathsp/mijia-api/compare/v2.0.2...v3.0.0) - 2025-11-28

- 切换到新 API 接口 `api.mijia.tech`
- 彻底移除账号密码登录，仅支持二维码登录
- 实现自动 token 刷新

### [2.0.2](https://github.com/smathsp/mijia-api/compare/v2.0.1...v2.0.2) - 2025-09-23

- 修复 `set` 方法类型检查问题

### [2.0.1](https://github.com/smathsp/mijia-api/compare/v2.0.0...v2.0.1) - 2025-06-29

- 处理超过 200 个设备的情况

### [2.0.0](https://github.com/smathsp/mijia-api/compare/v1.5.0...v2.0.0) - 2025-06-27

- 新增 `get_statistics` API
- 新增解密工具
- **破坏性变更**：移除 `mijiaDevices`，请使用 `mijiaDevice`

### [1.5.0](https://github.com/smathsp/mijia-api/compare/v1.4.5...v1.5.0) - 2025-06-19

- 重命名 `mijiaDevices` 为 `mijiaDevice`

### [1.4.0](https://github.com/smathsp/mijia-api/compare/v1.3.14...v1.4.0) - 2025-05-19

- 新增 CLI 支持
