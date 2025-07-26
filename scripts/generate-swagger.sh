#!/bin/bash

# Swagger文档生成脚本
# 用于生成和更新EtaPanel API文档

echo "🚀 开始生成EtaPanel API文档..."

# 检查swag工具是否已安装
if ! command -v swag &> /dev/null; then
    echo "📦 安装swag工具..."
    go install github.com/swaggo/swag/cmd/swag@latest
fi

# 切换到项目根目录
cd "$(dirname "$0")/.."

# 生成Swagger文档
echo "📚 生成Swagger文档..."
swag init -g cmd/swagger-main.go -o cmd/api/docs --parseDependency --parseInternal

# 检查生成是否成功
if [ $? -eq 0 ]; then
    echo "✅ Swagger文档生成成功！"
    echo ""
    echo "📖 API文档访问地址："
    echo "   - Swagger UI: http://localhost:8080/swagger/index.html"
    echo "   - JSON文档: http://localhost:8080/swagger/doc.json"
    echo ""
    echo "🎯 AI工具链API端点："
    echo "   - POST /api/auth/ai/query - 自然语言数据库查询"
    echo "   - POST /api/auth/ai/execute - 直接执行工具调用"
    echo "   - GET /api/auth/ai/tools - 获取可用工具列表"
    echo "   - GET /api/auth/ai/tools/{name} - 获取特定工具信息"
    echo "   - GET /api/auth/ai/health - AI服务健康检查"
    echo ""
    echo "🔧 传统AI功能："
    echo "   - POST /api/auth/ai/log - 智能日志分析"
    echo "   - POST /api/auth/ai/files - 智能文件分析"
    echo ""
    echo "📋 所有API端点都包含完整的："
    echo "   ✓ 参数说明和示例"
    echo "   ✓ 响应格式定义"
    echo "   ✓ 错误代码说明"
    echo "   ✓ JWT认证要求"
    echo ""
    echo "🚀 启动带Swagger的开发服务器："
    echo "   go run cmd/swagger-main.go"
else
    echo "❌ Swagger文档生成失败！"
    exit 1
fi
