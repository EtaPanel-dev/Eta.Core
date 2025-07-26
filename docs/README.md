# EtaPanel API 文档

## 概述

EtaPanel 是一个现代化的服务器管理面板，提供系统监控、文件管理、Nginx配置、SSL证书管理等功能。本文档描述了EtaPanel后端API的使用方法。

## 快速开始

### 1. 启动服务

```bash
cd Core
go run cmd/main.go
```

服务将在 `http://localhost:8080` 启动。

### 2. 访问Swagger文档

打开浏览器访问：`http://localhost:8080/swagger/index.html`

### 3. 认证

所有受保护的API都需要JWT认证。首先调用登录接口获取token：

```bash
curl -X POST http://localhost:8080/api/public/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "password123"}'
```

然后在后续请求中添加Authorization头：

```bash
curl -X GET http://localhost:8080/api/auth/system \
  -H "Authorization: Bearer {your-token}"
```

## API 分组

### 🔓 公共接口
- 健康检查
- 用户登录

### 🔐 认证接口
需要Bearer token认证的接口：

#### 📊 系统监控
- 系统信息查询
- CPU、内存、磁盘、网络监控
- 进程管理

#### 📁 文件管理
- 文件浏览、上传、下载
- 文件操作（复制、移动、删除）
- 权限管理
- 压缩解压

#### ⏰ 定时任务
- Crontab任务管理
- 任务启用/禁用

#### 🌐 Nginx管理
- Nginx状态监控
- 配置文件管理
- 网站配置
- 服务控制

#### 🔒 SSL证书管理
- 证书申请和管理
- ACME客户端配置
- DNS账号管理

## 数据格式

### 统一响应格式

```json
{
  "status": 200,
  "message": "操作成功",
  "data": {}
}
```

### 错误响应格式

```json
{
  "status": 400,
  "message": "错误描述",
  "data": null
}
```

## 状态码说明

| 状态码 | 说明 |
|--------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未授权（Token无效或过期） |
| 403 | 禁止访问 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

## 安全考虑

1. **JWT Token**: 所有受保护的API都需要有效的JWT token
2. **路径保护**: 系统关键路径受到保护，无法通过API访问
3. **进程保护**: 系统关键进程无法通过API终止
4. **权限验证**: 文件操作会验证用户权限

## 限制说明

1. **文件上传**: 默认最大文件大小限制
2. **并发连接**: WebSocket连接数限制
3. **API频率**: 部分API有频率限制
4. **路径访问**: 受保护的系统路径无法访问

## 开发指南

### 添加新的API

1. 在对应的handler包中创建处理函数
2. 添加Swagger注释
3. 在router中注册路由
4. 重新生成Swagger文档

```bash
swag init -g cmd/main.go -o cmd/api/docs
```

### Swagger注释格式

```go
// FunctionName 函数描述
// @Summary 简短描述
// @Description 详细描述
// @Tags 标签名
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param param_name param_type data_type true "参数描述"
// @Success 200 {object} ResponseType "成功描述"
// @Failure 400 {object} handler.Response "错误描述"
// @Router /path [method]
func FunctionName(c *gin.Context) {
    // 实现代码
}
```

## 部署说明

### 生产环境配置

1. 修改JWT密钥
2. 配置HTTPS
3. 设置CORS策略
4. 配置日志级别
5. 设置数据库连接

### Docker部署

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o etapanel cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/etapanel .
CMD ["./etapanel"]
```

## 故障排除

### 常见问题

1. **Token过期**: 重新登录获取新token
2. **权限不足**: 检查用户权限设置
3. **文件访问失败**: 检查文件路径和权限
4. **Nginx操作失败**: 检查Nginx服务状态

### 日志查看

```bash
# 查看应用日志
tail -f /var/log/etapanel/app.log

# 查看错误日志
tail -f /var/log/etapanel/error.log
```

## 更新日志

### v1.0.0
- 初始版本发布
- 完整的Swagger API文档
- JWT认证机制
- 系统监控功能
- 文件管理功能
- Nginx管理功能
- SSL证书管理功能

## 支持

- GitHub: https://github.com/EtaPanel-dev/EtaPanel
- 文档: https://docs.etapanel.com
- 问题反馈: https://github.com/EtaPanel-dev/EtaPanel/issues