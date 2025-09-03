package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ClientTestSuite 客户端测试套件
type ClientTestSuite struct {
	suite.Suite
	client     *MowenClient
	testServer *httptest.Server
	originalAPIKey string
}

// SetupSuite 测试套件初始化
func (suite *ClientTestSuite) SetupSuite() {
	// 保存原始环境变量
	suite.originalAPIKey = os.Getenv("MOWEN_API_KEY")
	
	// 设置测试用的API密钥
	os.Setenv("MOWEN_API_KEY", "test-api-key")
}

// TearDownSuite 测试套件清理
func (suite *ClientTestSuite) TearDownSuite() {
	// 恢复原始环境变量
	if suite.originalAPIKey != "" {
		os.Setenv("MOWEN_API_KEY", suite.originalAPIKey)
	} else {
		os.Unsetenv("MOWEN_API_KEY")
	}
}

// SetupTest 每个测试前的初始化
func (suite *ClientTestSuite) SetupTest() {
	// 创建测试服务器
	suite.testServer = httptest.NewServer(http.HandlerFunc(suite.mockHandler))
	
	// 创建客户端实例
	client, err := NewMowenClient()
	require.NoError(suite.T(), err)
	
	// 替换为测试服务器URL
	client.baseURL = suite.testServer.URL
	suite.client = client
}

// TearDownTest 每个测试后的清理
func (suite *ClientTestSuite) TearDownTest() {
	if suite.testServer != nil {
		suite.testServer.Close()
	}
}

// mockHandler 模拟HTTP处理器
func (suite *ClientTestSuite) mockHandler(w http.ResponseWriter, r *http.Request) {
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
func (suite *ClientTestSuite) handleMockNoteCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	
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
func (suite *ClientTestSuite) handleMockNoteEdit(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	
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
func (suite *ClientTestSuite) handleMockNoteSet(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	
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
func (suite *ClientTestSuite) handleMockKeyReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	
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
func (suite *ClientTestSuite) handleMockUploadPrepare(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	
	response := map[string]interface{}{
		"code": 0,
		"data": map[string]interface{}{
			"upload_url": suite.testServer.URL + "/upload/dynamic",
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
func (suite *ClientTestSuite) handleMockUploadURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	
	response := map[string]interface{}{
		"code": 0,
		"data": map[string]interface{}{
			"uuid": "test-url-file-uuid-999",
		},
		"message": "success",
	}
	json.NewEncoder(w).Encode(response)
}

// TestNewMowenClient 测试客户端创建
func (suite *ClientTestSuite) TestNewMowenClient() {
	// 测试正常创建
	client, err := NewMowenClient()
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), client)
	assert.Equal(suite.T(), "test-api-key", client.apiKey)
	assert.Equal(suite.T(), MowenAPIBaseURL, client.baseURL)
	assert.NotNil(suite.T(), client.httpClient)
	assert.Equal(suite.T(), 30*time.Second, client.httpClient.Timeout)
	
	// 测试缺少API密钥的情况
	os.Unsetenv("MOWEN_API_KEY")
	_, err = NewMowenClient()
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "MOWEN_API_KEY environment variable is required")
	
	// 恢复API密钥
	os.Setenv("MOWEN_API_KEY", "test-api-key")
}

// TestCreateNote 测试笔记创建
func (suite *ClientTestSuite) TestCreateNote() {
	req := NoteCreateRequest{
		Body: NoteAtom{
			Type: "doc",
			Content: []NoteAtom{
				{
					Type: "paragraph",
					Content: []NoteAtom{
						{
							Type: "text",
							Text: "测试笔记内容",
						},
					},
				},
			},
		},
		Settings: NoteCreateRequestSettings{
			AutoPublish: true,
			Tags:        []string{"测试"},
		},
	}
	
	result, err := suite.client.CreateNote(req)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	
	// 验证响应结构
	data, ok := result["data"].(map[string]interface{})
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), "test-note-id-123", data["note_id"])
}

// TestEditNote 测试笔记编辑
func (suite *ClientTestSuite) TestEditNote() {
	req := NoteEditRequest{
		NoteID: "test-note-id-123",
		Body: NoteAtom{
			Type: "doc",
			Content: []NoteAtom{
				{
					Type: "paragraph",
					Content: []NoteAtom{
						{
							Type: "text",
							Text: "编辑后的笔记内容",
						},
					},
				},
			},
		},
	}
	
	result, err := suite.client.EditNote(req)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	
	// 验证响应结构
	data, ok := result["data"].(map[string]interface{})
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), "test-note-id-123", data["note_id"])
}

// TestSetNotePrivacy 测试笔记隐私设置
func (suite *ClientTestSuite) TestSetNotePrivacy() {
	req := NoteSetRequest{
		NoteID:  "test-note-id-123",
		Section: 1,
		Settings: &NoteSettings{
			Privacy: &NotePrivacySet{
				Type: "public",
			},
		},
	}
	
	result, err := suite.client.SetNotePrivacy(req)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	
	// 验证响应结构
	data, ok := result["data"].(map[string]interface{})
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), "test-note-id-123", data["note_id"])
}

// TestResetAPIKey 测试API密钥重置
func (suite *ClientTestSuite) TestResetAPIKey() {
	result, err := suite.client.ResetAPIKey()
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	
	// 验证响应结构
	data, ok := result["data"].(map[string]interface{})
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), "new-test-api-key-456", data["api_key"])
}

// TestUploadFileViaURL 测试URL文件上传
func (suite *ClientTestSuite) TestUploadFileViaURL() {
	result, err := suite.client.UploadFileViaURL("https://example.com/test.jpg", 1, "test.jpg")
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	
	// 验证响应结构
	data, ok := result["data"].(map[string]interface{})
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), "test-url-file-uuid-999", data["uuid"])
}

// TestMakeRequestError 测试请求错误处理
func (suite *ClientTestSuite) TestMakeRequestError() {
	// 创建一个会返回错误的客户端
	client := &MowenClient{
		apiKey:     "test-key",
		baseURL:    "http://invalid-url-that-does-not-exist",
		httpClient: &http.Client{Timeout: 1 * time.Second},
	}
	
	_, err := client.makeRequest("POST", "/test", map[string]string{"test": "data"})
	assert.Error(suite.T(), err)
}

// TestClientTestSuite 运行客户端测试套件
func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}

// TestConstants 测试常量定义
func TestConstants(t *testing.T) {
	assert.Equal(t, "https://open.mowen.cn", MowenAPIBaseURL)
	assert.Equal(t, "/api/open/api/v1/note/create", NoteCreateEndpoint)
	assert.Equal(t, "/api/open/api/v1/note/edit", NoteEditEndpoint)
	assert.Equal(t, "/api/open/api/v1/note/set", NoteSetEndpoint)
	assert.Equal(t, "/api/open/api/v1/auth/key/reset", KeyResetEndpoint)
	assert.Equal(t, "/api/open/api/v1/upload/prepare", UploadPrepareEndpoint)
	assert.Equal(t, "/api/open/api/v1/upload/url", UploadURLEndpoint)
}