VERSION := $(shell git describe --tags --always 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X main.version=$(VERSION)
BINARY := mijia-api

# 本地构建
.PHONY: build
build:
	go build -ldflags="$(LDFLAGS)" -o $(BINARY) ./cmd/mijia

# 交叉编译到 Linux aarch64（路由器）
.PHONY: build-router
build-router:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 \
		go build -ldflags="$(LDFLAGS)" -o $(BINARY)-linux-arm64 ./cmd/mijia

# 交叉编译到 Linux armv7（旧路由器）
.PHONY: build-armv7
build-armv7:
	GOOS=linux GOARCH=arm CGO_ENABLED=0 GOARM=7 \
		go build -ldflags="$(LDFLAGS)" -o $(BINARY)-linux-armv7 ./cmd/mijia

# 交叉编译到 Linux amd64
.PHONY: build-linux
build-linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
		go build -ldflags="$(LDFLAGS)" -o $(BINARY)-linux-amd64 ./cmd/mijia

# 测试
.PHONY: test
test:
	go test ./...

# 部署到路由器（需要修改 USER 和 ROUTER_IP）
USER ?= root
ROUTER_IP ?= 192.168.1.1
deploy: build-router
	scp $(BINARY)-linux-arm64 $(USER)@$(ROUTER_IP):/usr/bin/mijia-api

# 清理
.PHONY: clean
clean:
	rm -f $(BINARY) $(BINARY)-linux-*

# 代码检查
.PHONY: lint
lint:
	go vet ./...

# 格式化
.PHONY: fmt
fmt:
	go fmt ./...

# 查看帮助
.PHONY: help
help:
	@echo "可用目标:"
	@echo "  build         - 本地构建"
	@echo "  build-router  - 交叉编译到 Linux aarch64（路由器）"
	@echo "  build-armv7   - 交叉编译到 Linux armv7"
	@echo "  build-linux   - 交叉编译到 Linux amd64"
	@echo "  test          - 运行测试"
	@echo "  deploy        - 部署到路由器 (USER=root ROUTER_IP=192.168.1.1)"
	@echo "  clean         - 清理构建产物"
	@echo "  lint          - 代码检查"
	@echo "  fmt           - 格式化代码"
