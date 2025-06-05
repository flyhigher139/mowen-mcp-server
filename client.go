package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	// 墨问API基础URL
	MowenAPIBaseURL = "https://api.mowen.cn"
	// API端点
	NoteCreateEndpoint = "/api/open/api/v1/note/create"
	NoteEditEndpoint   = "/api/open/api/v1/note/edit"
	NoteSetEndpoint    = "/api/open/api/v1/note/set"
	KeyResetEndpoint   = "/api/open/api/v1/auth/key/reset"
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
