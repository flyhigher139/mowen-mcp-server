package main

// NoteAtom 笔记原子节点信息
type NoteAtom struct {
	Type    string            `json:"type"`              // 节点类型
	Text    string            `json:"text,omitempty"`    // 节点文本
	Content []NoteAtom        `json:"content,omitempty"` // 节点内容
	Marks   []NoteAtom        `json:"marks,omitempty"`   // 节点标记
	Attrs   map[string]string `json:"attrs,omitempty"`   // 节点属性
}

// NoteCreateRequestSettings 笔记创建设置
type NoteCreateRequestSettings struct {
	AutoPublish bool     `json:"autoPublish,omitempty"` // 自动发布
	Tags        []string `json:"tags,omitempty"`        // 标签列表
}

// NoteCreateRequest 笔记创建请求
type NoteCreateRequest struct {
	Body     NoteAtom                  `json:"body"`               // 笔记内容
	Settings NoteCreateRequestSettings `json:"settings,omitempty"` // 笔记设置
}

// NoteEditRequest 笔记编辑请求
type NoteEditRequest struct {
	NoteID string   `json:"noteId"` // 笔记ID
	Body   NoteAtom `json:"body"`   // 笔记内容
}

// NotePrivacySetRule 隐私规则
type NotePrivacySetRule struct {
	NoShare  bool   `json:"noShare,omitempty"`  // 是否禁止分享与转发
	ExpireAt string `json:"expireAt,omitempty"` // 公开截止时间
}

// NotePrivacySet 笔记隐私设置
type NotePrivacySet struct {
	Type string              `json:"type"`           // 隐私类型: public, private, rule
	Rule *NotePrivacySetRule `json:"rule,omitempty"` // 隐私规则
}

// NoteSettings 笔记设置项
type NoteSettings struct {
	Privacy *NotePrivacySet `json:"privacy,omitempty"` // 笔记隐私设置
}

// NoteSetRequest 笔记设置请求
type NoteSetRequest struct {
	NoteID   string        `json:"noteId"`   // 笔记ID
	Section  int           `json:"section"`  // 设置类别: 1-笔记隐私
	Settings *NoteSettings `json:"settings"` // 设置项
}

// KeyResetRequest API密钥重置请求
type KeyResetRequest struct{}

// MCP工具参数结构体

// CreateNoteArgs 创建笔记工具参数
type CreateNoteArgs struct {
	Paragraphs  []Paragraph `json:"paragraphs" description:"富文本段落列表，每个段落包含文本节点"`
	AutoPublish bool        `json:"auto_publish,omitempty" description:"是否自动发布，默认为false"`
	Tags        []string    `json:"tags,omitempty" description:"笔记标签列表"`
}

// EditNoteArgs 编辑笔记工具参数
type EditNoteArgs struct {
	NoteID     string      `json:"note_id" description:"要编辑的笔记ID"`
	Paragraphs []Paragraph `json:"paragraphs" description:"富文本段落列表，将完全替换原有内容"`
}

// SetNotePrivacyArgs 设置笔记隐私工具参数
type SetNotePrivacyArgs struct {
	NoteID      string `json:"note_id" description:"笔记ID"`
	PrivacyType string `json:"privacy_type" description:"隐私类型（public/private/rule）"`
	NoShare     *bool  `json:"no_share,omitempty" description:"是否禁止分享（仅rule类型有效）"`
	ExpireAt    *int64 `json:"expire_at,omitempty" description:"过期时间戳（仅rule类型有效，0表示永不过期）"`
}

// ResetAPIKeyArgs 重置API密钥工具参数
type ResetAPIKeyArgs struct {
}

// UploadFileArgs 本地文件上传参数
type UploadFileArgs struct {
	FilePath string `json:"file_path" description:"要上传的文件路径"`
	FileType int    `json:"file_type" description:"文件类型：1-图片，2-音频，3-PDF"`
	FileName string `json:"file_name" description:"文件名称"`
}

// UploadFileViaURLArgs 基于URL的文件上传参数
type UploadFileViaURLArgs struct {
	FileURL  string `json:"file_url" description:"要上传的文件URL"`
	FileType int    `json:"file_type" description:"文件类型：1-图片，2-音频，3-PDF"`
	FileName string `json:"file_name,omitempty" description:"文件名称（可选）"`
}

// FileNode 文件节点
type FileNode struct {
	FileType   string            `json:"file_type" description:"文件类型：image、audio、pdf"`
	SourceType string            `json:"source_type" description:"来源类型：local、url"`
	SourcePath string            `json:"source_path" description:"文件路径或URL"`
	Metadata   map[string]string `json:"metadata,omitempty" description:"文件元数据"`
}

// Paragraph 段落结构
type Paragraph struct {
	Type   string     `json:"type,omitempty" description:"段落类型：quote（引用段落）、note（内链笔记）、file（文件）"`
	Texts  []TextNode `json:"texts,omitempty" description:"文本节点列表"`
	NoteID string     `json:"note_id,omitempty" description:"内链笔记ID（仅当type为note时使用）"`
	File   *FileNode  `json:"file,omitempty" description:"文件节点（仅当type为file时使用）"`
}

// TextNode 文本节点
type TextNode struct {
	Text      string `json:"text" description:"文本内容"`
	Bold      bool   `json:"bold,omitempty" description:"是否加粗"`
	Highlight bool   `json:"highlight,omitempty" description:"是否高亮"`
	Link      string `json:"link,omitempty" description:"链接地址"`
}

// 转换函数：将MCP参数转换为墨问API格式

// ConvertParagraphsToNoteAtom 将段落列表转换为NoteAtom格式
func ConvertParagraphsToNoteAtom(paragraphs []Paragraph) NoteAtom {
	doc := NoteAtom{
		Type:    "doc",
		Content: make([]NoteAtom, 0, len(paragraphs)),
	}

	for _, para := range paragraphs {
		switch para.Type {
		case "quote":
			// 引用段落
			quotePara := NoteAtom{
				Type: "paragraph",
				Attrs: map[string]string{
					"blockquote": "true",
				},
				Content: convertTextsToContent(para.Texts),
			}
			doc.Content = append(doc.Content, quotePara)
		case "note":
			// 内链笔记
			notePara := NoteAtom{
				Type: "note",
				Attrs: map[string]string{
					"uuid": para.NoteID,
				},
			}
			doc.Content = append(doc.Content, notePara)
		case "file":
			// 文件段落
			if para.File != nil {
				fileAtom := NoteAtom{
					Type: para.File.FileType,
					Attrs: map[string]string{
						"uuid":       para.File.SourcePath, // 墨问API中文件UUID即为SourcePath
						"sourceType": para.File.SourceType,
					},
				}
				// 合并元数据
				for k, v := range para.File.Metadata {
					fileAtom.Attrs[k] = v
				}
				doc.Content = append(doc.Content, fileAtom)
			}
		default:
			// 普通段落
			normalPara := NoteAtom{
				Type:    "paragraph",
				Content: convertTextsToContent(para.Texts),
			}
			doc.Content = append(doc.Content, normalPara)
		}
	}

	return doc
}

// convertTextsToContent 将文本节点列表转换为内容
func convertTextsToContent(texts []TextNode) []NoteAtom {
	content := make([]NoteAtom, 0, len(texts))

	for _, text := range texts {
		textAtom := NoteAtom{
			Type: "text",
			Text: text.Text,
		}

		// 添加标记
		var marks []NoteAtom
		if text.Bold {
			marks = append(marks, NoteAtom{Type: "bold"})
		}
		if text.Highlight {
			marks = append(marks, NoteAtom{Type: "highlight"})
		}
		if text.Link != "" {
			marks = append(marks, NoteAtom{
				Type: "link",
				Attrs: map[string]string{
					"href": text.Link,
				},
			})
		}

		if len(marks) > 0 {
			textAtom.Marks = marks
		}

		content = append(content, textAtom)
	}

	return content
}
