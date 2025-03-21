.PHONY: all build clean run test ui dev

# 默认目标
all: ui build

# 构建前端
ui:
	@echo "正在构建前端..."
	cd ui && npm install && npm run build

# 构建后端
build:
	@echo "正在构建后端..."
	go mod tidy
	go build -o bin/app .

# 开发模式（前端）
ui-dev:
	@echo "启动前端开发服务器..."
	cd ui && npm run start

# 开发模式（后端）
dev:
	@echo "启动后端开发模式..."
	go run .

# 运行编译后的程序
run: build
	@echo "运行程序..."
	./bin/app

# 运行测试
test:
	@echo "运行测试..."
	go test -v ./...

# 清理构建产物
clean:
	@echo "清理构建文件..."
	rm -rf bin/
	rm -rf ui/dist/
	rm -rf ui/node_modules/

# 安装依赖
deps:
	@echo "安装依赖..."
	go mod download
	cd ui && npm install

# 检查代码格式
fmt:
	@echo "格式化代码..."
	go fmt ./...
	cd ui && npm run lint

# 显示帮助信息
help:
	@echo "可用的 make 命令："
	@echo "  make          - 构建整个项目（前端和后端）"
	@echo "  make ui      - 构建前端"
	@echo "  make build   - 构建后端"
	@echo "  make ui-dev  - 启动前端开发服务器"
	@echo "  make dev     - 启动后端开发模式"
	@echo "  make run     - 运行编译后的程序"
	@echo "  make test    - 运行测试"
	@echo "  make clean   - 清理构建文件"
	@echo "  make deps    - 安装依赖"
	@echo "  make fmt     - 格式化代码"
