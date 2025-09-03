package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/ThinkInAIXYZ/go-mcp/protocol"
)

// ServerTestSuite MCP服务器测试套件
type ServerTestSuite struct {
	suite.Suite
	mcpServer      *MowenMCPServer
	mockHTTPServer *httptest.Server
	originalAPIKey string
}

// SetupSuite 测试套件初始化
func (suite *ServerTestSuite) SetupSuite() {
	// 保存原始环境变量
	suite.originalAPIKey = os.Getenv("MOWEN_API_KEY")
	
	// 设置测试用的API密钥
	os.Setenv("MOWEN_API_KEY", "test-api-key")
}

// TearDownSuite 测试套件清理
func (suite *ServerTestSuite) TearDownSuite() {
	// 恢复原始环境变量
	if suite.originalAPIKey != "" {
		os.Setenv("MOWEN_API_KEY", suite.originalAPIKey)
	} else {
		os.Unsetenv("MOWEN_API_KEY")
	}
}

// SetupTest 每个测试前的初始化
func (suite *ServerTestSuite) SetupTest() {
	// 创建模拟HTTP服务器
	suite.mockHTTPServer = httptest.NewServer(http.HandlerFunc(suite.mockAPIHandler))
	
	// 创建MCP服务器实例
	mcpServer, err := NewMowenMCPServer()
	require.NoError(suite.T(), err)
	
	// 替换客户端的baseURL为测试服务器
	mcpServer.mowenClient.baseURL = suite.mockHTTPServer.URL
	suite.mcpServer = mcpServer
}

// TearDownTest 每个测试后的清理
func (suite *ServerTestSuite) TearDownTest() {
	if suite.mockHTTPServer != nil {
		suite.mockHTTPServer.Close()
	}
}

// mockAPIHandler 模拟墨问API处理器
func (suite *ServerTestSuite) mockAPIHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	switch r.URL.Path {
	case NoteCreateEndpoint:
		suite.handleMockNoteCreate(w, r)
	case NoteEditEndpoint:
		suite.handleMockNoteEdit(w, r)
	case NoteSetEndpoint:
		suite.handleMockNoteSet(w, r)
	case KeyResetEndpoint:
		suite.handleMockKeyReset(w, r)
	case UploadPrepareEndpoint:
		suite.handleMockUploadPrepare(w, r)
	case UploadURLEndpoint:
		suite.handleMockUploadURL(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "endpoint not found"})
	}
}

// handleMockNoteCreate 模拟笔记创建响应
func (suite *ServerTestSuite) handleMockNoteCreate(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"code": 0,
		"data": map[string]interface{}{
			"note_id": "test-note-id-123",
			"url": "https://mowen.cn/note/test-note-id-123",
		},
		"message": "success",
	}
	json.NewEncoder(w).Encode(response)
}

// handleMockNoteEdit 模拟笔记编辑响应
func (suite *ServerTestSuite) handleMockNoteEdit(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"code": 0,
		"data": map[string]interface{}{
			"note_id": "test-note-id-123",
		},
		"message": "success",
	}
	json.NewEncoder(w).Encode(response)
}

// handleMockNoteSet 模拟笔记设置响应
func (suite *ServerTestSuite) handleMockNoteSet(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"code": 0,
		"data": map[string]interface{}{
			"note_id": "test-note-id-123",
		},
		"message": "success",
	}
	json.NewEncoder(w).Encode(response)
}

// handleMockKeyReset 模拟密钥重置响应
func (suite *ServerTestSuite) handleMockKeyReset(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"code": 0,
		"data": map[string]interface{}{
			"api_key": "new-test-api-key-456",
		},
		"message": "success",
	}
	json.NewEncoder(w).Encode(response)
}

// handleMockUploadPrepare 模拟上传准备响应
func (suite *ServerTestSuite) handleMockUploadPrepare(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"code": 0,
		"data": map[string]interface{}{
			"upload_url": suite.mockHTTPServer.URL + "/upload/dynamic",
			"form_data": map[string]interface{}{
				"key": "test-file-key",
				"policy": "test-policy",
				"signature": "test-signature",
			},
			"uuid": "test-file-uuid-789",
		},
		"message": "success",
	}
	json.NewEncoder(w).Encode(response)
}

// handleMockUploadURL 模拟URL上传响应
func (suite *ServerTestSuite) handleMockUploadURL(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"code": 0,
		"data": map[string]interface{}{
			"uuid": "test-url-file-uuid-999",
		},
		"message": "success",
	}
	json.NewEncoder(w).Encode(response)
}

// TestNewMowenMCPServer 测试MCP服务器创建
func (suite *ServerTestSuite) TestNewMowenMCPServer() {
	server, err := NewMowenMCPServer()
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), server)
	assert.NotNil(suite.T(), server.mcpServer)
	assert.NotNil(suite.T(), server.mowenClient)
}

// TestHandleCreateNote 测试创建笔记处理器
func (suite *ServerTestSuite) TestHandleCreateNote() {
	// 准备测试请求
	args := CreateNoteArgs{
		Paragraphs: []Paragraph{
			{
				Texts: []TextNode{
					{Text: "测试笔记内容"},
				},
			},
		},
		AutoPublish: true,
		Tags:        []string{"测试"},
	}
	
	argsJSON, err := json.Marshal(args)
	require.NoError(suite.T(), err)
	
	req := &protocol.CallToolRequest{
		RawArguments: argsJSON,
	}
	
	// 调用处理器
	result, err := suite.mcpServer.handleCreateNote(context.Background(), req)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	
	// 验证结果
	assert.Len(suite.T(), result.Content, 1)
	textContent, ok := result.Content[0].(*protocol.TextContent)
	assert.True(suite.T(), ok)
	assert.Contains(suite.T(), textContent.Text, "test-note-id-123")
}

// TestHandleEditNote 测试编辑笔记处理器
func (suite *ServerTestSuite) TestHandleEditNote() {
	// 准备测试请求
	args := EditNoteArgs{
		NoteID: "test-note-id-123",
		Paragraphs: []Paragraph{
			{
				Texts: []TextNode{
					{Text: "编辑后的笔记内容"},
				},
			},
		},
	}
	
	argsJSON, err := json.Marshal(args)
	require.NoError(suite.T(), err)
	
	req := &protocol.CallToolRequest{
		RawArguments: argsJSON,
	}
	
	// 调用处理器
	result, err := suite.mcpServer.handleEditNote(context.Background(), req)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	
	// 验证结果
	assert.Len(suite.T(), result.Content, 1)
	textContent, ok := result.Content[0].(*protocol.TextContent)
	assert.True(suite.T(), ok)
	assert.Contains(suite.T(), textContent.Text, "test-note-id-123")
}

// TestHandleSetNotePrivacy 测试设置笔记隐私处理器
func (suite *ServerTestSuite) TestHandleSetNotePrivacy() {
	// 准备测试请求
	args := SetNotePrivacyArgs{
		NoteID:      "test-note-id-123",
		PrivacyType: "public",
	}
	
	argsJSON, err := json.Marshal(args)
	require.NoError(suite.T(), err)
	
	req := &protocol.CallToolRequest{
		RawArguments: argsJSON,
	}
	
	// 调用处理器
	result, err := suite.mcpServer.handleSetNotePrivacy(context.Background(), req)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	
	// 验证结果
	assert.Len(suite.T(), result.Content, 1)
	textContent, ok := result.Content[0].(*protocol.TextContent)
	assert.True(suite.T(), ok)
	assert.Contains(suite.T(), textContent.Text, "test-note-id-123")
}

// TestHandleResetAPIKey 测试重置API密钥处理器
func (suite *ServerTestSuite) TestHandleResetAPIKey() {
	// 准备测试请求
	args := ResetAPIKeyArgs{}
	
	argsJSON, err := json.Marshal(args)
	require.NoError(suite.T(), err)
	
	req := &protocol.CallToolRequest{
		RawArguments: argsJSON,
	}
	
	// 调用处理器
	result, err := suite.mcpServer.handleResetAPIKey(context.Background(), req)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	
	// 验证结果
	assert.Len(suite.T(), result.Content, 1)
	textContent, ok := result.Content[0].(*protocol.TextContent)
	assert.True(suite.T(), ok)
	assert.Contains(suite.T(), textContent.Text, "new-test-api-key-456")
}

// TestHandleUploadFileViaURL 测试URL文件上传处理器
func (suite *ServerTestSuite) TestHandleUploadFileViaURL() {
	// 准备测试请求
	args := UploadFileViaURLArgs{
		FileURL:  "https://example.com/test.jpg",
		FileType: 1,
		FileName: "test.jpg",
	}
	
	argsJSON, err := json.Marshal(args)
	require.NoError(suite.T(), err)
	
	req := &protocol.CallToolRequest{
		RawArguments: argsJSON,
	}
	
	// 调用处理器
	result, err := suite.mcpServer.handleUploadFileViaURL(context.Background(), req)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	
	// 验证结果
	assert.Len(suite.T(), result.Content, 1)
	textContent, ok := result.Content[0].(*protocol.TextContent)
	assert.True(suite.T(), ok)
	assert.Contains(suite.T(), textContent.Text, "test-url-file-uuid-999")
}

// TestInvalidArguments 测试无效参数处理
func (suite *ServerTestSuite) TestInvalidArguments() {
	// 测试无效的JSON参数
	req := &protocol.CallToolRequest{
		RawArguments: []byte(`{"invalid_json": `),
	}
	
	result, err := suite.mcpServer.handleCreateNote(context.Background(), req)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
}

// TestServerTestSuite 运行服务器测试套件
func TestServerTestSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}