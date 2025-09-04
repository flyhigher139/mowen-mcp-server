package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

const (
	// MowenAPIBaseURL 墨问API基础URL
	MowenAPIBaseURL = "https://open.mowen.cn"
	// API端点
	NoteCreateEndpoint    = "/api/open/api/v1/note/create"
	NoteEditEndpoint      = "/api/open/api/v1/note/edit"
	NoteSetEndpoint       = "/api/open/api/v1/note/set"
	KeyResetEndpoint      = "/api/open/api/v1/auth/key/reset"
	UploadPrepareEndpoint = "/api/open/api/v1/upload/prepare"
	UploadURLEndpoint     = "/api/open/api/v1/upload/url"
)

// MowenClient 墨问API客户端
type MowenClient struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

// NewMowenClient 创建新的墨问API客户端
func NewMowenClient() (*MowenClient, error) {
	apiKey := os.Getenv("MOWEN_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("MOWEN_API_KEY environment variable is required")
	}

	return &MowenClient{
		apiKey:  apiKey,
		baseURL: MowenAPIBaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// makeRequest 发送HTTP请求到墨问API
func (c *MowenClient) makeRequest(method, endpoint string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, c.baseURL+endpoint, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// CreateNote 创建笔记
func (c *MowenClient) CreateNote(req NoteCreateRequest) (map[string]interface{}, error) {
	respBody, err := c.makeRequest("POST", NoteCreateEndpoint, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create note: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result, nil
}

// UploadFileViaURL 通过URL上传文件到墨问
func (c *MowenClient) UploadFileViaURL(fileURL string, fileType int, fileName string) (map[string]interface{}, error) {
	req := map[string]interface{}{
		"url":       fileURL,
		"file_type": fileType,
	}

	if fileName != "" {
		req["file_name"] = fileName
	}

	respBody, err := c.makeRequest("POST", UploadURLEndpoint, req)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file via URL: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result, nil
}

// EditNote 编辑笔记
func (c *MowenClient) EditNote(req NoteEditRequest) (map[string]interface{}, error) {
	respBody, err := c.makeRequest("POST", NoteEditEndpoint, req)
	if err != nil {
		return nil, fmt.Errorf("failed to edit note: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result, nil
}

// SetNotePrivacy 设置笔记隐私
func (c *MowenClient) SetNotePrivacy(req NoteSetRequest) (map[string]interface{}, error) {
	respBody, err := c.makeRequest("POST", NoteSetEndpoint, req)
	if err != nil {
		return nil, fmt.Errorf("failed to set note privacy: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result, nil
}

// ResetAPIKey 重置API密钥
func (c *MowenClient) ResetAPIKey() (map[string]interface{}, error) {
	req := KeyResetRequest{}
	respBody, err := c.makeRequest("POST", KeyResetEndpoint, req)
	if err != nil {
		return nil, fmt.Errorf("failed to reset API key: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result, nil
}

// UploadFile 上传文件
// UploadFile 通过准备接口上传本地文件到墨问
func (c *MowenClient) UploadFile(filePath string, fileType int, fileName string) (map[string]interface{}, error) {
	// 第一步：获取上传准备信息
	prepareReq := map[string]interface{}{
		"file_type": fileType,
		"file_name": fileName,
	}

	prepareResp, err := c.makeRequest("POST", UploadPrepareEndpoint, prepareReq)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare upload: %w", err)
	}

	var prepareResult map[string]interface{}
	if err = json.Unmarshal(prepareResp, &prepareResult); err != nil {
		return nil, fmt.Errorf("failed to unmarshal prepare response: %w", err)
	}

	// 检查准备响应中的数据结构
	data, ok := prepareResult["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid prepare response format")
	}

	uploadURL, ok := data["upload_url"].(string)
	if !ok {
		return nil, fmt.Errorf("missing upload_url in prepare response")
	}

	formData, ok := data["form_data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing form_data in prepare response")
	}

	// 第二步：上传文件到指定的URL
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 添加表单数据字段
	for key, value := range formData {
		if strValue, ok := value.(string); ok {
			_ = writer.WriteField(key, strValue)
		}
	}

	// 添加文件字段
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err = io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("failed to copy file content: %w", err)
	}

	writer.Close()

	// 发送上传请求
	req, err := http.NewRequest("POST", uploadURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create upload request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send upload request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upload response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("upload request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal upload response: %w", err)
	}

	return result, nil
}
