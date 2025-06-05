package main

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/ThinkInAIXYZ/go-mcp/protocol"
	"github.com/ThinkInAIXYZ/go-mcp/server"
	"github.com/ThinkInAIXYZ/go-mcp/transport"
)

// MowenMCPServer 墨问MCP服务器
type MowenMCPServer struct {
	mcpServer   *server.Server
	mowenClient *MowenClient
}

// NewMowenMCPServer 创建新的墨问MCP服务器
func NewMowenMCPServer() (*MowenMCPServer, error) {
	// 创建墨问API客户端
	mowenClient, err := NewMowenClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create mowen client: %w", err)
	}

	// 创建SSE传输服务器
	transportServer, err := transport.NewSSEServerTransport("127.0.0.1:8080")
	if err != nil {
		return nil, fmt.Errorf("failed to create transport server: %w", err)
	}

	// 创建MCP服务器
	mcpServer, err := server.NewServer(transportServer)
	if err != nil {
		return nil, fmt.Errorf("failed to create MCP server: %w", err)
	}

	mowenMCPServer := &MowenMCPServer{
		mcpServer:   mcpServer,
		mowenClient: mowenClient,
	}

	// 注册工具
	if err := mowenMCPServer.registerTools(); err != nil {
		return nil, fmt.Errorf("failed to register tools: %w", err)
	}

	return mowenMCPServer, nil
}

// registerTools 注册所有MCP工具
func (s *MowenMCPServer) registerTools() error {
	// 注册创建笔记工具
	createNoteTool, err := protocol.NewTool(
		"create_note",
		"创建一篇新的墨问笔记，使用统一的富文本格式",
		CreateNoteArgs{},
	)
	if err != nil {
		return fmt.Errorf("failed to create create_note tool: %w", err)
	}
	s.mcpServer.RegisterTool(createNoteTool, s.handleCreateNote)

	// 注册编辑笔记工具
	editNoteTool, err := protocol.NewTool(
		"edit_note",
		"编辑已存在的笔记内容，使用统一的富文本格式",
		EditNoteArgs{},
	)
	if err != nil {
		return fmt.Errorf("failed to create edit_note tool: %w", err)
	}
	s.mcpServer.RegisterTool(editNoteTool, s.handleEditNote)

	// 注册设置笔记隐私工具
	setPrivacyTool, err := protocol.NewTool(
		"set_note_privacy",
		"设置笔记的隐私权限",
		SetNotePrivacyArgs{},
	)
	if err != nil {
		return fmt.Errorf("failed to create set_note_privacy tool: %w", err)
	}
	s.mcpServer.RegisterTool(setPrivacyTool, s.handleSetNotePrivacy)

	// 注册重置API密钥工具
	resetKeyTool, err := protocol.NewTool(
		"reset_api_key",
		"重置墨问API密钥",
		ResetAPIKeyArgs{},
	)
	if err != nil {
		return fmt.Errorf("failed to create reset_api_key tool: %w", err)
	}
	s.mcpServer.RegisterTool(resetKeyTool, s.handleResetAPIKey)

	return nil
}

// handleCreateNote 处理创建笔记请求
func (s *MowenMCPServer) handleCreateNote(ctx context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	var args CreateNoteArgs
	if err := protocol.VerifyAndUnmarshal(req.RawArguments, &args); err != nil {
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

// handleEditNote 处理编辑笔记请求
func (s *MowenMCPServer) handleEditNote(ctx context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	var args EditNoteArgs
	if err := protocol.VerifyAndUnmarshal(req.RawArguments, &args); err != nil {
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

// handleSetNotePrivacy 处理设置笔记隐私请求
func (s *MowenMCPServer) handleSetNotePrivacy(ctx context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	var args SetNotePrivacyArgs
	if err := protocol.VerifyAndUnmarshal(req.RawArguments, &args); err != nil {
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

// handleResetAPIKey 处理重置API密钥请求
func (s *MowenMCPServer) handleResetAPIKey(ctx context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	var args ResetAPIKeyArgs
	if err := protocol.VerifyAndUnmarshal(req.RawArguments, &args); err != nil {
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

// Run 启动MCP服务器
func (s *MowenMCPServer) Run() error {
	log.Println("启动墨问MCP服务器...")
	log.Println("服务器地址: http://127.0.0.1:8080")
	log.Println("SSE端点: http://127.0.0.1:8080/sse")
	return s.mcpServer.Run()
}

// Shutdown 关闭MCP服务器
func (s *MowenMCPServer) Shutdown(ctx context.Context) error {
	return s.mcpServer.Shutdown(ctx)
}
