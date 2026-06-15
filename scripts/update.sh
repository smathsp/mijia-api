#!/bin/bash
# mijia-api 路由器自动更新脚本
# 用法: curl -sL https://raw.githubusercontent.com/smathsp/mijia-api/main/scripts/update.sh | bash

set -e

REPO="smathsp/mijia-api"
BINARY_NAME="mijia-api"
INSTALL_DIR="/usr/bin"
AUTH_DIR="$HOME/.config/mijia-api"

# 检测架构
detect_arch() {
    ARCH=$(uname -m)
    case $ARCH in
        x86_64|amd64)
            echo "amd64"
            ;;
        aarch64|arm64)
            echo "arm64"
            ;;
        armv7*|armv7l|armhf)
            echo "armv7"
            ;;
        *)
            echo "unsupported"
            exit 1
            ;;
    esac
}

# 获取最新版本
get_latest_version() {
    curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | cut -d '"' -f 4
}

# 下载二进制
download_binary() {
    local version=$1
    local arch=$2
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')

    # 构建文件名
    if [ "$arch" = "armv7" ]; then
        FILENAME="mijia-api-${os}-armv7"
    else
        FILENAME="mijia-api-${os}-${arch}"
    fi

    DOWNLOAD_URL="https://github.com/$REPO/releases/download/${version}/${FILENAME}"

    echo "下载 $DOWNLOAD_URL ..."
    curl -sL "$DOWNLOAD_URL" -o "/tmp/$BINARY_NAME"
    chmod +x "/tmp/$BINARY_NAME"
}

# 安装二进制
install_binary() {
    echo "安装到 $INSTALL_DIR/$BINARY_NAME ..."

    # 检查是否有写入权限
    if [ -w "$INSTALL_DIR" ]; then
        cp "/tmp/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
    else
        echo "需要 sudo 权限安装..."
        sudo cp "/tmp/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
    fi

    rm -f "/tmp/$BINARY_NAME"
}

# 检查认证文件
check_auth() {
    if [ ! -f "$AUTH_DIR/auth.json" ]; then
        echo ""
        echo "⚠️  未找到认证文件: $AUTH_DIR/auth.json"
        echo "首次使用需要登录:"
        echo "  $BINARY_NAME --list-devices"
        echo ""
    fi
}

# 主函数
main() {
    echo "=== mijia-api 更新脚本 ==="
    echo ""

    # 检测架构
    ARCH=$(detect_arch)
    echo "检测到架构: $ARCH"

    # 获取最新版本
    echo "获取最新版本..."
    VERSION=$(get_latest_version)
    if [ -z "$VERSION" ]; then
        echo "❌ 无法获取最新版本"
        exit 1
    fi
    echo "最新版本: $VERSION"

    # 检查当前版本
    if command -v "$BINARY_NAME" &> /dev/null; then
        CURRENT_VERSION=$("$BINARY_NAME" --version 2>&1 | awk '{print $2}')
        echo "当前版本: $CURRENT_VERSION"

        if [ "$CURRENT_VERSION" = "$VERSION" ]; then
            echo "✅ 已是最新版本，无需更新"
            check_auth
            exit 0
        fi
    fi

    # 下载并安装
    download_binary "$VERSION" "$ARCH"
    install_binary

    echo ""
    echo "✅ 更新完成!"
    echo "版本: $VERSION"
    echo "安装路径: $INSTALL_DIR/$BINARY_NAME"
    echo ""

    check_auth

    # 显示帮助
    echo "使用方法:"
    echo "  $BINARY_NAME --list-devices     # 列出设备"
    echo "  $BINARY_NAME --list-homes       # 列出家庭"
    echo "  $BINARY_NAME get --did DID --prop-name NAME  # 获取属性"
    echo "  $BINARY_NAME set --did DID --prop-name NAME --value VALUE  # 设置属性"
    echo ""
}

main "$@"
