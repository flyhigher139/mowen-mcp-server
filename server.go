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

// NewMowenMCPServer 创建并初始化一个新的墨问MCP服务器。
// 它会创建墨问API客户端，设置传输层，并注册所有MCP工具。
func NewMowenMCPServer() (*MowenMCPServer, error) {
	// 创建墨问API客户端
	mowenClient, err := NewMowenClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create mowen client: %w", err)
	}

	// 创建传输服务器
	//transportServer, err := transport.NewSSEServerTransport("127.0.0.1:8080")
	transportServer := transport.NewStreamableHTTPServerTransport(
		"127.0.0.1:8080",
		transport.WithStreamableHTTPServerTransportOptionStateMode(transport.Stateful),
	)

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

// registerTools 注册所有墨问MCP服务器支持的工具。
// 这些工具包括创建笔记、编辑笔记、设置笔记隐私、重置API密钥和文件上传。
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

	// 注册本地文件上传工具
	uploadFileTool, err := protocol.NewTool(
		"upload_file",
		"上传本地文件到墨问笔记，支持图片、音频和PDF",
		UploadFileArgs{},
	)
	if err != nil {
		return fmt.Errorf("failed to create upload_file tool: %w", err)
	}
	s.mcpServer.RegisterTool(uploadFileTool, s.handleUploadFile)

	// 注册基于URL的文件上传工具
	uploadFileViaURLTool, err := protocol.NewTool(
		"upload_file_via_url",
		"通过URL上传文件到墨问笔记，支持图片、音频和PDF",
		UploadFileViaURLArgs{},
	)
	if err != nil {
		return fmt.Errorf("failed to create upload_file_via_url tool: %w", err)
	}
	s.mcpServer.RegisterTool(uploadFileViaURLTool, s.handleUploadFileViaURL)

	return nil
}

// handleCreateNote 处理创建笔记的MCP工具请求。
// 它解析请求参数，将其转换为墨问API所需的格式，然后调用墨问API创建笔记。
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

// handleEditNote 处理编辑笔记的MCP工具请求。
// 它解析请求参数，将其转换为墨问API所需的格式，然后调用墨问API编辑笔记。
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

// handleSetNotePrivacy 处理设置笔记隐私的MCP工具请求。
// 它解析请求参数，构建隐私设置，然后调用墨问API更新笔记的隐私设置。
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

// handleResetAPIKey 处理重置API密钥的MCP工具请求。
// 它调用墨问API重置API密钥。
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

// handleUploadFile 处理文件上传的MCP工具请求。
// 它解析请求参数，然后调用墨问API上传文件。
func (s *MowenMCPServer) handleUploadFile(ctx context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	var args UploadFileArgs
	if err := protocol.VerifyAndUnmarshal(req.RawArguments, &args); err != nil {
		return nil, fmt.Errorf("invalid arguments: %v", err)
	}

	// 调用墨问API上传文件
	result, err := s.mowenClient.UploadFile(args.FilePath, args.FileType, args.FileName)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// 格式化响应
	responseText := fmt.Sprintf("文件上传成功！\n\n响应详情：\n%+v", result)

	return &protocol.CallToolResult{
		Content: []protocol.Content{
			&protocol.TextContent{
				Type: "text",
				Text: responseText,
			},
		},
	}, nil
}

// handleUploadFileViaURL 处理基于URL的文件上传请求
func (s *MowenMCPServer) handleUploadFileViaURL(ctx context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	var args UploadFileViaURLArgs
	if err := protocol.VerifyAndUnmarshal(req.RawArguments, &args); err != nil {
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

// Run 启动墨问MCP服务器，开始监听传入的MCP请求。
func (s *MowenMCPServer) Run() error {
	log.Println("启动墨问MCP服务器...")
	//log.Println("服务器地址: http://127.0.0.1:8080")
	//log.Println("SSE端点: http://127.0.0.1:8080/sse")
	return s.mcpServer.Run()
}

// Shutdown 关闭墨问MCP服务器。
// 它会优雅地关闭底层的MCP服务器。
func (s *MowenMCPServer) Shutdown(ctx context.Context) error {
	return s.mcpServer.Shutdown(ctx)
}
