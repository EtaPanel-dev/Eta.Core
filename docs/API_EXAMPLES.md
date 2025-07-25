# EtaPanel API 使用示例

## 认证流程

### 1. 用户登录
```bash
curl -X POST http://localhost:8080/api/public/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "password123"
  }'
```

响应示例：
```json
{
  "status": 200,
  "message": "登录成功",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": 1640995200
  }
}
```

### 2. 使用Token访问受保护的API
```bash
curl -X GET http://localhost:8080/api/auth/system \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

## 系统监控API示例

### 获取系统信息
```bash
curl -X GET http://localhost:8080/api/auth/system \
  -H "Authorization: Bearer {token}"
```

### 获取进程列表
```bash
curl -X GET http://localhost:8080/api/auth/system/processes \
  -H "Authorization: Bearer {token}"
```

### 终止进程
```bash
curl -X POST http://localhost:8080/api/auth/system/process/kill \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{
    "pid": 1234,
    "signal": "TERM"
  }'
```

## 文件管理API示例

### 列出目录文件
```bash
curl -X GET "http://localhost:8080/api/auth/files?path=/home" \
  -H "Authorization: Bearer {token}"
```

### 上传文件
```bash
curl -X POST http://localhost:8080/api/auth/files/upload \
  -H "Authorization: Bearer {token}" \
  -F "path=/home/user" \
  -F "file=@example.txt"
```

### 下载文件
```bash
curl -X GET "http://localhost:8080/api/auth/files/download?path=/home/user/example.txt" \
  -H "Authorization: Bearer {token}" \
  -O
```

## Nginx管理API示例

### 获取Nginx状态
```bash
curl -X GET http://localhost:8080/api/auth/nginx/status \
  -H "Authorization: Bearer {token}"
```

### 创建网站
```bash
curl -X POST http://localhost:8080/api/auth/nginx/sites \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "example.com",
    "domain": "example.com",
    "root": "/var/www/example.com",
    "index": "index.html",
    "ssl": false,
    "enabled": true
  }'
```

### 重启Nginx
```bash
curl -X POST http://localhost:8080/api/auth/nginx/restart \
  -H "Authorization: Bearer {token}"
```

## SSL证书管理API示例

### 申请SSL证书
```bash
curl -X POST http://localhost:8080/api/auth/acme/ssl \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{
    "acme_client_id": 1,
    "domain": "example.com"
  }'
```

### 获取SSL证书列表
```bash
curl -X GET http://localhost:8080/api/auth/acme/ssl \
  -H "Authorization: Bearer {token}"
```

## 定时任务API示例

### 创建定时任务
```bash
curl -X POST http://localhost:8080/api/auth/crontab \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{
    "minute": "0",
    "hour": "2",
    "day": "*",
    "month": "*",
    "weekday": "*",
    "command": "/usr/bin/backup.sh",
    "comment": "每天凌晨2点执行备份",
    "enabled": true
  }'
```

### 获取定时任务列表
```bash
curl -X GET http://localhost:8080/api/auth/crontab \
  -H "Authorization: Bearer {token}"
```

## 错误处理

所有API都遵循统一的错误响应格式：

```json
{
  "status": 400,
  "message": "请求参数错误",
  "data": null
}
```

常见HTTP状态码：
- `200` - 成功
- `400` - 请求参数错误
- `401` - 未授权（Token无效或过期）
- `403` - 禁止访问
- `404` - 资源不存在
- `500` - 服务器内部错误

## 分页和过滤

部分API支持分页和过滤参数：

```bash
# 分页示例
curl -X GET "http://localhost:8080/api/auth/system/processes?page=1&limit=10" \
  -H "Authorization: Bearer {token}"

# 过滤示例
curl -X GET "http://localhost:8080/api/auth/files?path=/home&filter=*.txt" \
  -H "Authorization: Bearer {token}"
```

## WebSocket连接

终端连接使用WebSocket：

```javascript
const ws = new WebSocket('ws://localhost:8080/ws/pty');
ws.onopen = function() {
    console.log('Terminal connected');
};
ws.onmessage = function(event) {
    console.log('Terminal output:', event.data);
};
```