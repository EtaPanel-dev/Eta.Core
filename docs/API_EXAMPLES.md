# AI数据库工具链 API 使用示例

## 新增的API端点

以下是新添加到路由系统中的AI数据库工具链API端点：

### 1. 自然语言查询
**POST** `/api/auth/ai/query`

使用自然语言与数据库进行交互。

**请求体：**
```json
{
  "query": "连接到数据库 ./test.db 并显示所有表"
}
```

**响应：**
```json
{
  "success": true,
  "data": {
    "response": "我已经连接到数据库并获取了表列表。数据库中包含以下表：users, products, orders..."
  }
}
```

**示例查询：**
- "连接到数据库 ./data.db"
- "创建一个用户表，包含id、姓名、邮箱字段"
- "查询用户表中的所有数据"
- "备份数据库到 ./backup/data_backup.db"

### 2. 直接工具调用
**POST** `/api/auth/ai/execute`

直接执行特定的数据库工具调用。

**请求体：**
```json
{
  "tool_calls": "[{\"id\":\"call_1\",\"type\":\"function\",\"function\":{\"name\":\"connect_sqlite_database\",\"arguments\":{\"database_path\":\"./test.db\"}}}]"
}
```

**响应：**
```json
{
  "success": true,
  "data": [
    {
      "tool_call_id": "call_1",
      "success": true,
      "result": {
        "message": "Successfully connected to SQLite database at ./test.db",
        "path": "./test.db"
      }
    }
  ]
}
```

### 3. 获取可用工具
**GET** `/api/auth/ai/tools`

获取所有可用的数据库操作工具列表。

**响应：**
```json
{
  "success": true,
  "data": {
    "tools": [
      {
        "type": "function",
        "function": {
          "name": "connect_sqlite_database",
          "description": "连接到SQLite数据库",
          "parameters": {
            "type": "object",
            "properties": {
              "database_path": {
                "type": "string",
                "description": "SQLite数据库文件路径"
              }
            },
            "required": ["database_path"]
          }
        }
      }
    ],
    "count": 10
  }
}
```

### 4. 获取特定工具信息
**GET** `/api/auth/ai/tools/{toolName}`

获取特定工具的详细信息。

**示例：** `GET /api/auth/ai/tools/connect_sqlite_database`

**响应：**
```json
{
  "success": true,
  "data": {
    "type": "function",
    "function": {
      "name": "connect_sqlite_database",
      "description": "连接到SQLite数据库",
      "parameters": {
        "type": "object",
        "properties": {
          "database_path": {
            "type": "string",
            "description": "SQLite数据库文件路径"
          }
        },
        "required": ["database_path"]
      }
    }
  }
}
```

### 5. 健康检查
**GET** `/api/auth/ai/health`

检查AI服务的健康状态。

**响应：**
```json
{
  "success": true,
  "data": {
    "status": "healthy"
  }
}
```

## 完整的工具列表

系统提供以下10个数据库操作工具：

1. **connect_sqlite_database** - 连接到SQLite数据库
2. **get_database_info** - 获取数据库基本信息
3. **list_tables** - 列出数据库中的所有表
4. **get_table_schema** - 获取指定表的结构信息
5. **execute_query** - 执行SQL查询语句（SELECT）
6. **execute_statement** - 执行SQL语句（INSERT, UPDATE, DELETE）
7. **create_table** - 创建新表
8. **drop_table** - 删除表
9. **backup_database** - 备份数据库
10. **vacuum_database** - 优化数据库

## 使用场景示例

### 场景1：数据库管理
```bash
# 1. 连接数据库
curl -X POST http://localhost:8080/api/auth/ai/query \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"query": "连接到数据库 ./data.db"}'

# 2. 查看数据库信息
curl -X POST http://localhost:8080/api/auth/ai/query \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"query": "显示数据库信息和所有表"}'
```

### 场景2：表操作
```bash
# 创建用户表
curl -X POST http://localhost:8080/api/auth/ai/query \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"query": "创建一个用户表，包含id、姓名、邮箱和创建时间字段"}'

# 插入测试数据
curl -X POST http://localhost:8080/api/auth/ai/query \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"query": "向用户表插入一些测试数据"}'
```

### 场景3：直接工具调用
```bash
# 批量操作示例
curl -X POST http://localhost:8080/api/auth/ai/execute \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tool_calls": "[
      {
        \"id\":\"1\",
        \"type\":\"function\",
        \"function\":{
          \"name\":\"connect_sqlite_database\",
          \"arguments\":{\"database_path\":\"./test.db\"}
        }
      },
      {
        \"id\":\"2\",
        \"type\":\"function\",
        \"function\":{
          \"name\":\"get_database_info\",
          \"arguments\":{}
        }
      }
    ]"
  }'
```

## 错误处理

所有API都使用统一的错误响应格式：

```json
{
  "success": false,
  "error": "错误描述信息"
}
```

常见错误：
- `400`: 请求参数错误
- `401`: 认证失败
- `404`: 工具不存在
- `500`: 服务器内部错误
- `503`: AI服务不可用

## 安全说明

1. 所有AI工具链API都需要JWT认证
2. 工具调用仅限于预定义的数据库操作
3. 参数会进行严格验证
4. 支持完整的错误处理和日志记录

通过这些API，您可以用自然语言或结构化的工具调用来管理SQLite数据库，大大简化了数据库操作的复杂性。
