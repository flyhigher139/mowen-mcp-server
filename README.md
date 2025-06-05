# 墨问笔记 MCP 服务器 (Go版本)

这是一个基于**模型上下文协议（MCP）**的Go语言服务器，用于与墨问笔记软件进行交互。通过此服务器，你可以在支持MCP的应用（如Cursor、Claude Desktop等）中直接创建、编辑和管理墨问笔记。

## ✨ 功能特性

- 🔗 **兼容MCP协议**：支持最新的MCP协议规范
- 📝 **创建笔记**：统一的富文本格式，支持段落、加粗、高亮、链接、引用和内链笔记
- ✏️ **编辑笔记**：统一的富文本格式，完全替换笔记内容
- 💬 **引用段落**：创建引用文本块，支持富文本格式
- 🔗 **内链笔记**：引用其他笔记，创建笔记间的关联
- 🔒 **隐私设置**：设置笔记的公开、私有或规则公开权限
- 🔄 **密钥管理**：重置API密钥功能
- ⚡ **高性能**：基于Go语言，具有出色的并发性能和低资源占用

## 🚀 快速开始

### 前提条件

- Go 1.21+
- 墨问Pro会员账号（API功能仅对Pro会员开放）
- 墨问API密钥（在墨问小程序中获取）

### 安装和运行

1. **克隆项目**：
```bash
git clone <repository-url>
cd mowen-v1
```

2. **安装依赖**：
```bash
go mod tidy
```

3. **设置环境变量**：

**macOS/Linux**:
```bash
export MOWEN_API_KEY="你的墨问API密钥"
```

**Windows PowerShell**:
```powershell
$env:MOWEN_API_KEY="你的墨问API密钥"
```

**持久化设置** - 创建 `.env` 文件：
```
MOWEN_API_KEY=你的墨问API密钥
```

4. **运行服务器**：
```bash
go run .
```

服务器将在 `http://127.0.0.1:8080` 启动，SSE端点为 `http://127.0.0.1:8080/sse`。

### 配置 MCP 客户端

在 Cursor 或 Claude Desktop 的设置中添加以下配置：

```json
{
  "mcpServers": {
    "mowen-mcp-server": {
      "command": "go",
      "args": ["run", "."],
      "cwd": "/path/to/mowen-v1",
      "env": {
        "MOWEN_API_KEY": "${env:MOWEN_API_KEY}"
      }
    }
  }
}
```

**注意**: 请将 `/path/to/mowen-v1` 替换为你的实际项目路径。

## 🛠️ 可用工具

### create_note
创建一篇新的墨问笔记，使用统一的富文本格式

**参数**：
- `paragraphs` (数组，必需)：富文本段落列表，每个段落包含文本节点
- `auto_publish` (布尔值，可选)：是否自动发布，默认为false
- `tags` (字符串数组，可选)：笔记标签列表

**支持的段落类型**：
- 普通段落（默认）：`{"texts": [...]}`
- 引用段落：`{"type": "quote", "texts": [...]}`
- 内链笔记：`{"type": "note", "note_id": "笔记ID"}`

**段落格式示例**：
```json
[
  {
    "texts": [
      {"text": "普通文本"},
      {"text": "加粗文本", "bold": true},
      {"text": "高亮文本", "highlight": true},
      {"text": "链接文本", "link": "https://example.com"}
    ]
  },
  {
    "type": "quote",
    "texts": [
      {"text": "这是引用段落"},
      {"text": "支持富文本", "bold": true}
    ]
  },
  {
    "type": "note",
    "note_id": "VPrWsE_-P0qwrFUOygxxx"
  }
]
```

### edit_note
编辑已存在的笔记内容，使用统一的富文本格式

**参数**：
- `note_id` (字符串，必需)：要编辑的笔记ID
- `paragraphs` (数组，必需)：富文本段落列表，将完全替换原有内容

**注意**：此操作会完全替换笔记的原有内容，而不是追加内容。

### set_note_privacy
设置笔记的隐私权限

**参数**：
- `note_id` (字符串，必需)：笔记ID
- `privacy_type` (字符串，必需)：隐私类型（public/private/rule）
- `no_share` (布尔值，可选)：是否禁止分享（仅rule类型有效）
- `expire_at` (整数，可选)：过期时间戳（仅rule类型有效，0表示永不过期）

### reset_api_key
重置墨问API密钥

**注意**：此操作会使当前密钥立即失效。

## 📁 项目结构

```
mowen-v1/
├── main.go          # 主程序入口
├── server.go        # MCP服务器实现
├── client.go        # 墨问API客户端
├── types.go         # 数据结构定义
├── go.mod           # Go模块定义
└── README.md        # 项目文档
```

## 🔧 技术栈

- **Go 1.21+**: 主要编程语言
- **go-mcp**: MCP协议实现库
- **net/http**: HTTP客户端用于API调用
- **encoding/json**: JSON序列化/反序列化

## 📝 使用示例

### 创建简单文本笔记
```json
{
  "paragraphs": [
    {
      "texts": [
        {"text": "今天学习了Go编程，重点是并发编程概念"}
      ]
    }
  ],
  "auto_publish": true,
  "tags": ["学习", "Go", "编程"]
}
```

### 创建富文本笔记
```json
{
  "paragraphs": [
    {
      "texts": [
        {"text": "重要提醒：", "bold": true},
        {"text": "明天的会议已改期"}
      ]
    },
    {
      "type": "quote",
      "texts": [
        {"text": "会议时间：", "bold": true},
        {"text": "下周三上午10点"}
      ]
    }
  ]
}
```

## 🤝 贡献

欢迎提交Issue和Pull Request来改进这个项目！

## 📄 许可证

本项目采用 MIT 许可证。

## 🙏 致谢

- [go-mcp](https://github.com/ThinkInAIXYZ/go-mcp) - 提供了优秀的Go语言MCP实现
- [墨问笔记](https://mowen.cn) - 提供了强大的笔记API服务
- Python版本的 [mowen-mcp-server](https://github.com/z4656207/mowen-mcp-server) - 提供了实现参考