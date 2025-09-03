package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/ThinkInAIXYZ/go-mcp/protocol"
)

// MockHTTPClient 模拟HTTP客户端
type MockHTTPClient struct {
	mock.Mock
}

// Do 模拟HTTP请求
func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

// MowenClientInterface 墨问客户端接口
type MowenClientInterface interface {
	CreateNote(req NoteCreateRequest) (map[string]interface{}, error)
	EditNote(req NoteEditRequest) (map[string]interface{}, error)
	SetNotePrivacy(req NoteSetRequest) (map[string]interface{}, error)
	ResetAPIKey() (map[string]interface{}, error)
	UploadFile(filePath string, fileType int, fileName string) (map[string]interface{}, error)
	UploadFileViaURL(fileURL string, fileType int, fileName string) (map[string]interface{}, error)
}

// MockMowenClient 模拟墨问客户端
type MockMowenClient struct {
	mock.Mock
}

// CreateNote 模拟创建笔记
func (m *MockMowenClient) CreateNote(req NoteCreateRequest) (map[string]interface{}, error) {
	args := m.Called(req)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

// EditNote 模拟编辑笔记
func (m *MockMowenClient) EditNote(req NoteEditRequest) (map[string]interface{}, error) {
	args := m.Called(req)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

// SetNotePrivacy 模拟设置笔记隐私
func (m *MockMowenClient) SetNotePrivacy(req NoteSetRequest) (map[string]interface{}, error) {
	args := m.Called(req)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

// ResetAPIKey 模拟重置API密钥
func (m *MockMowenClient) ResetAPIKey() (map[string]interface{}, error) {
	args := m.Called()
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

// UploadFile 模拟本地文件上传
func (m *MockMowenClient) UploadFile(filePath string, fileType int, fileName string) (map[string]interface{}, error) {
	args := m.Called(filePath, fileType, fileName)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

// UploadFileViaURL 模拟URL文件上传
func (m *MockMowenClient) UploadFileViaURL(fileURL string, fileType int, fileName string) (map[string]interface{}, error) {
	args := m.Called(fileURL, fileType, fileName)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

// TestMowenMCPServer 测试用的MCP服务器结构
type TestMowenMCPServer struct {
	mowenClient MowenClientInterface
}

// handleCreateNote 处理创建笔记请求（测试版本）
func (s *TestMowenMCPServer) handleCreateNote(ctx context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	var args CreateNoteArgs
	if err := json.Unmarshal(req.RawArguments, &args); err != nil {
		return nil, fmt.Errorf("invalid arguments: %v", err)
	}

	// 转换参数为墨问API格式
	noteBody := ConvertParagraphsToNoteAtom(args.Paragraphs)
	createReq := NoteCreateRequest{
		Body: noteBody,
		Settings: NoteCreateRequestSettings{
			AutoPublish: args.AutoPublish,
			Tags:        args.Tags,
		},
	}

	// 调用墨问API
	result, err := s.mowenClient.CreateNote(createReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create note: %w", err)
	}

	// 格式化响应
	responseText := fmt.Sprintf("笔记创建成功！\n\n响应详情：\n%+v", result)

	return &protocol.CallToolResult{
		Content: []protocol.Content{
			&protocol.TextContent{
				Type: "text",
				Text: responseText,
			},
		},
	}, nil
}

// handleEditNote 处理编辑笔记请求（测试版本）
func (s *TestMowenMCPServer) handleEditNote(ctx context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	var args EditNoteArgs
	if err := json.Unmarshal(req.RawArguments, &args); err != nil {
		return nil, fmt.Errorf("invalid arguments: %v", err)
	}

	// 转换参数为墨问API格式
	noteBody := ConvertParagraphsToNoteAtom(args.Paragraphs)
	editReq := NoteEditRequest{
		NoteID: args.NoteID,
		Body:   noteBody,
	}

	// 调用墨问API
	result, err := s.mowenClient.EditNote(editReq)
	if err != nil {
		return nil, fmt.Errorf("failed to edit note: %w", err)
	}

	// 格式化响应
	responseText := fmt.Sprintf("笔记编辑成功！\n\n响应详情：\n%+v", result)

	return &protocol.CallToolResult{
		Content: []protocol.Content{
			&protocol.TextContent{
				Type: "text",
				Text: responseText,
			},
		},
	}, nil
}

// handleSetNotePrivacy 处理设置笔记隐私请求（测试版本）
func (s *TestMowenMCPServer) handleSetNotePrivacy(ctx context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	var args SetNotePrivacyArgs
	if err := json.Unmarshal(req.RawArguments, &args); err != nil {
		return nil, fmt.Errorf("invalid arguments: %v", err)
	}

	// 构建隐私设置
	privacySet := &NotePrivacySet{
		Type: args.PrivacyType,
	}

	// 如果是规则公开，设置规则
	if args.PrivacyType == "rule" {
		rule := &NotePrivacySetRule{}
		if args.NoShare != nil {
			rule.NoShare = *args.NoShare
		}
		if args.ExpireAt != nil {
			rule.ExpireAt = strconv.FormatInt(*args.ExpireAt, 10)
		}
		privacySet.Rule = rule
	}

	// 构建请求
	setReq := NoteSetRequest{
		NoteID:  args.NoteID,
		Section: 1, // 1表示笔记隐私设置
		Settings: &NoteSettings{
			Privacy: privacySet,
		},
	}

	// 调用墨问API
	result, err := s.mowenClient.SetNotePrivacy(setReq)
	if err != nil {
		return nil, fmt.Errorf("failed to set note privacy: %w", err)
	}

	// 格式化响应
	responseText := fmt.Sprintf("笔记隐私设置成功！\n\n响应详情：\n%+v", result)

	return &protocol.CallToolResult{
		Content: []protocol.Content{
			&protocol.TextContent{
				Type: "text",
				Text: responseText,
			},
		},
	}, nil
}

// handleResetAPIKey 处理重置API密钥请求（测试版本）
func (s *TestMowenMCPServer) handleResetAPIKey(ctx context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	var args ResetAPIKeyArgs
	if err := json.Unmarshal(req.RawArguments, &args); err != nil {
		return nil, fmt.Errorf("invalid arguments: %v", err)
	}

	// 调用墨问API
	result, err := s.mowenClient.ResetAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to reset API key: %w", err)
	}

	// 格式化响应
	responseText := fmt.Sprintf("API密钥重置成功！\n\n⚠️ 注意：此操作会使当前密钥立即失效\n\n响应详情：\n%+v", result)

	return &protocol.CallToolResult{
		Content: []protocol.Content{
			&protocol.TextContent{
				Type: "text",
				Text: responseText,
			},
		},
	}, nil
}

// handleUploadFileViaURL 处理URL文件上传请求（测试版本）
func (s *TestMowenMCPServer) handleUploadFileViaURL(ctx context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	var args UploadFileViaURLArgs
	if err := json.Unmarshal(req.RawArguments, &args); err != nil {
		return nil, fmt.Errorf("invalid arguments: %v", err)
	}

	// 调用墨问API通过URL上传文件
	result, err := s.mowenClient.UploadFileViaURL(args.FileURL, args.FileType, args.FileName)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file via URL: %w", err)
	}

	// 格式化响应
	responseText := fmt.Sprintf("文件通过URL上传成功！\n\n响应详情：\n%+v", result)

	return &protocol.CallToolResult{
		Content: []protocol.Content{
			&protocol.TextContent{
				Type: "text",
				Text: responseText,
			},
		},
	}, nil
}

// MockTestSuite 模拟测试套件
type MockTestSuite struct {
	suite.Suite
	mockClient *MockMowenClient
	mcpServer  *TestMowenMCPServer
}

// SetupTest 设置测试环境
func (suite *MockTestSuite) SetupTest() {
	suite.mockClient = new(MockMowenClient)
	
	// 创建一个使用模拟客户端的测试MCP服务器
	suite.mcpServer = &TestMowenMCPServer{
		mowenClient: suite.mockClient,
	}
}

// TearDownTest 清理测试环境
func (suite *MockTestSuite) TearDownTest() {
	suite.mockClient.AssertExpectations(suite.T())
}

// TestMockCreateNote 测试模拟创建笔记
func (suite *MockTestSuite) TestMockCreateNote() {
	// 设置模拟期望
	expectedResponse := map[string]interface{}{
		"noteId": "mock-note-id-123",
		"status": "success",
	}
	
	suite.mockClient.On("CreateNote", mock.AnythingOfType("main.NoteCreateRequest")).Return(expectedResponse, nil)

	// 准备测试请求
	args := CreateNoteArgs{
		Paragraphs: []Paragraph{
			{
				Texts: []TextNode{
					{Text: "模拟测试笔记"},
				},
			},
		},
		AutoPublish: true,
		Tags:        []string{"测试", "模拟"},
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
	assert.Contains(suite.T(), textContent.Text, "mock-note-id-123")
}

// TestMockEditNote 测试模拟编辑笔记
func (suite *MockTestSuite) TestMockEditNote() {
	// 设置模拟期望
	expectedResponse := map[string]interface{}{
		"noteId": "mock-edit-note-id-456",
		"status": "updated",
	}
	
	suite.mockClient.On("EditNote", mock.AnythingOfType("main.NoteEditRequest")).Return(expectedResponse, nil)

	// 准备测试请求
	args := EditNoteArgs{
		NoteID: "mock-edit-note-id-456",
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
	assert.Contains(suite.T(), textContent.Text, "mock-edit-note-id-456")
}

// TestMockSetNotePrivacy 测试模拟设置笔记隐私
func (suite *MockTestSuite) TestMockSetNotePrivacy() {
	// 设置模拟期望
	expectedResponse := map[string]interface{}{
		"noteId": "mock-privacy-note-id-789",
		"privacy": "private",
	}
	
	suite.mockClient.On("SetNotePrivacy", mock.AnythingOfType("main.NoteSetRequest")).Return(expectedResponse, nil)

	// 准备测试请求
	args := SetNotePrivacyArgs{
		NoteID:      "mock-privacy-note-id-789",
		PrivacyType: "private",
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
	assert.Contains(suite.T(), textContent.Text, "mock-privacy-note-id-789")
}

// TestMockResetAPIKey 测试模拟重置API密钥
func (suite *MockTestSuite) TestMockResetAPIKey() {
	// 设置模拟期望
	expectedResponse := map[string]interface{}{
		"newApiKey": "mock-new-api-key-xyz",
		"status":   "reset",
	}
	
	suite.mockClient.On("ResetAPIKey").Return(expectedResponse, nil)

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
	assert.Contains(suite.T(), textContent.Text, "mock-new-api-key-xyz")
}

// TestMockUploadFileViaURL 测试模拟URL文件上传
func (suite *MockTestSuite) TestMockUploadFileViaURL() {
	// 设置模拟期望
	expectedResponse := map[string]interface{}{
		"fileId":   "mock-file-id-abc",
		"fileName": "mock-test.jpg",
		"status":   "uploaded",
	}
	
	suite.mockClient.On("UploadFileViaURL", "https://example.com/mock-test.jpg", 1, "mock-test.jpg").Return(expectedResponse, nil)

	// 准备测试请求
	args := UploadFileViaURLArgs{
		FileURL:  "https://example.com/mock-test.jpg",
		FileType: 1,
		FileName: "mock-test.jpg",
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
	assert.Contains(suite.T(), textContent.Text, "mock-file-id-abc")
}

// TestMockErrorHandling 测试模拟错误处理
func (suite *MockTestSuite) TestMockErrorHandling() {
	// 设置模拟期望返回错误
	suite.mockClient.On("CreateNote", mock.AnythingOfType("main.NoteCreateRequest")).Return(map[string]interface{}{}, assert.AnError)

	// 准备测试请求
	args := CreateNoteArgs{
		Paragraphs: []Paragraph{
			{
				Texts: []TextNode{
					{Text: "错误测试笔记"},
				},
			},
		},
	}
	
	argsJSON, err := json.Marshal(args)
	require.NoError(suite.T(), err)
	
	req := &protocol.CallToolRequest{
		RawArguments: argsJSON,
	}

	// 调用处理器，期望返回错误
	result, err := suite.mcpServer.handleCreateNote(context.Background(), req)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "failed to create note")
}

// TestMockInvalidJSON 测试模拟无效JSON处理
func (suite *MockTestSuite) TestMockInvalidJSON() {
	// 准备无效JSON请求
	req := &protocol.CallToolRequest{
		RawArguments: []byte(`{"invalid_json": `),
	}

	// 调用处理器，期望返回错误
	result, err := suite.mcpServer.handleCreateNote(context.Background(), req)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "invalid arguments")
}

// TestMockTestSuite 运行模拟测试套件
func TestMockTestSuite(t *testing.T) {
	suite.Run(t, new(MockTestSuite))
}