.PHONY: build-cli build-gva build-web run-gva run-gva-web test clean

## 构建 CLI
build-cli:
	cd cli && go build -o pmocker.exe .

## 构建 gva server 二进制
build-gva:
	cd gva/server && go build -o ../../bin/gva-server.exe .

## 构建前端
build-web:
	cd gva/web && npm install && npm run build

## 启动 gva 后端（开发模式）
run-gva:
	cd gva/server && go run main.go

## 启动 gva 前端（开发模式）
run-gva-web:
	cd gva/web && npm run dev

## 运行所有测试
test:
	go work sync
	cd cli && go test ./...
	cd pkg && go test ./...

## 清理构建产物
clean:
	rm -f cli/pmocker cli/pmocker.exe
	rm -f bin/gva-server bin/gva-server.exe
	find . -name "*.db" -delete
