package main

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// TypesTestSuite 数据类型测试套件
type TypesTestSuite struct {
	suite.Suite
}

// SetupTest 设置测试环境
func (suite *TypesTestSuite) SetupTest() {
	// 测试前的设置
}

// TearDownTest 清理测试环境
func (suite *TypesTestSuite) TearDownTest() {
	// 测试后的清理
}

// TestConvertParagraphsToNoteAtom 测试段落转换为NoteAtom格式
func (suite *TypesTestSuite) TestConvertParagraphsToNoteAtom() {
	// 测试普通段落
	paragraphs := []Paragraph{
		{
			Texts: []TextNode{
				{Text: "这是一个普通段落"},
				{Text: "加粗文本", Bold: true},
			},
		},
	}

	result := ConvertParagraphsToNoteAtom(paragraphs)

	// 验证结果结构
	assert.Equal(suite.T(), "doc", result.Type)
	assert.Len(suite.T(), result.Content, 1)
	assert.Equal(suite.T(), "paragraph", result.Content[0].Type)
	assert.Len(suite.T(), result.Content[0].Content, 2)
}

// TestConvertQuoteParagraph 测试引用段落转换
func (suite *TypesTestSuite) TestConvertQuoteParagraph() {
	paragraphs := []Paragraph{
		{
			Type: "quote",
			Texts: []TextNode{
				{Text: "这是一个引用段落"},
			},
		},
	}

	result := ConvertParagraphsToNoteAtom(paragraphs)

	// 验证引用段落结构
	assert.Equal(suite.T(), "doc", result.Type)
	assert.Len(suite.T(), result.Content, 1)
	assert.Equal(suite.T(), "paragraph", result.Content[0].Type)
	assert.Equal(suite.T(), "true", result.Content[0].Attrs["blockquote"])
}

// TestConvertNoteParagraph 测试内链笔记段落转换
func (suite *TypesTestSuite) TestConvertNoteParagraph() {
	paragraphs := []Paragraph{
		{
			Type:   "note",
			NoteID: "test-note-uuid-123",
		},
	}

	result := ConvertParagraphsToNoteAtom(paragraphs)

	// 验证内链笔记结构
	assert.Equal(suite.T(), "doc", result.Type)
	assert.Len(suite.T(), result.Content, 1)
	assert.Equal(suite.T(), "note", result.Content[0].Type)
	assert.Equal(suite.T(), "test-note-uuid-123", result.Content[0].Attrs["uuid"])
}

// TestConvertFileParagraph 测试文件段落转换
func (suite *TypesTestSuite) TestConvertFileParagraph() {
	paragraphs := []Paragraph{
		{
			Type: "file",
			File: &FileNode{
				FileType:   "image",
				SourcePath: "test-image-uuid-456",
				SourceType: "upload",
				Metadata: map[string]string{
					"alt":   "测试图片",
					"align": "center",
				},
			},
		},
	}

	result := ConvertParagraphsToNoteAtom(paragraphs)

	// 验证文件段落结构
	assert.Equal(suite.T(), "doc", result.Type)
	assert.Len(suite.T(), result.Content, 1)
	assert.Equal(suite.T(), "image", result.Content[0].Type)
	assert.Equal(suite.T(), "test-image-uuid-456", result.Content[0].Attrs["uuid"])
	assert.Equal(suite.T(), "upload", result.Content[0].Attrs["sourceType"])
	assert.Equal(suite.T(), "测试图片", result.Content[0].Attrs["alt"])
	assert.Equal(suite.T(), "center", result.Content[0].Attrs["align"])
}

// TestConvertTextsToContent 测试文本转换为内容
func (suite *TypesTestSuite) TestConvertTextsToContent() {
	texts := []TextNode{
		{Text: "普通文本"},
		{Text: "加粗文本", Bold: true},
		{Text: "高亮文本", Highlight: true},
		{Text: "链接文本", Link: "https://example.com"},
		{Text: "组合格式", Bold: true, Highlight: true, Link: "https://test.com"},
	}

	result := convertTextsToContent(texts)

	// 验证转换结果
	assert.Len(suite.T(), result, 5)

	// 验证普通文本
	assert.Equal(suite.T(), "text", result[0].Type)
	assert.Equal(suite.T(), "普通文本", result[0].Text)
	assert.Nil(suite.T(), result[0].Marks)

	// 验证加粗文本
	assert.Equal(suite.T(), "text", result[1].Type)
	assert.Equal(suite.T(), "加粗文本", result[1].Text)
	assert.Len(suite.T(), result[1].Marks, 1)
	assert.Equal(suite.T(), "bold", result[1].Marks[0].Type)

	// 验证高亮文本
	assert.Equal(suite.T(), "text", result[2].Type)
	assert.Equal(suite.T(), "高亮文本", result[2].Text)
	assert.Len(suite.T(), result[2].Marks, 1)
	assert.Equal(suite.T(), "highlight", result[2].Marks[0].Type)

	// 验证链接文本
	assert.Equal(suite.T(), "text", result[3].Type)
	assert.Equal(suite.T(), "链接文本", result[3].Text)
	assert.Len(suite.T(), result[3].Marks, 1)
	assert.Equal(suite.T(), "link", result[3].Marks[0].Type)
	assert.Equal(suite.T(), "https://example.com", result[3].Marks[0].Attrs["href"])

	// 验证组合格式
	assert.Equal(suite.T(), "text", result[4].Type)
	assert.Equal(suite.T(), "组合格式", result[4].Text)
	assert.Len(suite.T(), result[4].Marks, 3) // bold + highlight + link
}

// TestNoteCreateRequestSerialization 测试笔记创建请求序列化
func (suite *TypesTestSuite) TestNoteCreateRequestSerialization() {
	req := NoteCreateRequest{
		Body: NoteAtom{
			Type: "doc",
			Content: []NoteAtom{
				{
					Type: "paragraph",
					Content: []NoteAtom{
						{
							Type: "text",
							Text: "测试笔记",
						},
					},
				},
			},
		},
		Settings: NoteCreateRequestSettings{
			AutoPublish: true,
			Tags:        []string{"测试", "API"},
		},
	}

	// 序列化为JSON
	jsonData, err := json.Marshal(req)
	require.NoError(suite.T(), err)

	// 反序列化验证
	var decoded NoteCreateRequest
	err = json.Unmarshal(jsonData, &decoded)
	require.NoError(suite.T(), err)

	// 验证数据完整性
	assert.Equal(suite.T(), req.Body.Type, decoded.Body.Type)
	assert.Equal(suite.T(), req.Settings.AutoPublish, decoded.Settings.AutoPublish)
	assert.Equal(suite.T(), req.Settings.Tags, decoded.Settings.Tags)
}

// TestNotePrivacySetSerialization 测试笔记隐私设置序列化
func (suite *TypesTestSuite) TestNotePrivacySetSerialization() {
	// 测试完全公开
	publicPrivacy := &NotePrivacySet{
		Type: "public",
	}

	jsonData, err := json.Marshal(publicPrivacy)
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), string(jsonData), `"type":"public"`)

	// 测试规则公开
	rulePrivacy := &NotePrivacySet{
		Type: "rule",
		Rule: &NotePrivacySetRule{
			NoShare:  true,
			ExpireAt: "1640995200", // 2022-01-01 00:00:00 UTC
		},
	}

	jsonData, err = json.Marshal(rulePrivacy)
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), string(jsonData), `"type":"rule"`)
	assert.Contains(suite.T(), string(jsonData), `"noShare":true`)
	assert.Contains(suite.T(), string(jsonData), `"expireAt":"1640995200"`)
}

// TestUploadFileViaURLArgsSerialization 测试URL文件上传参数序列化
func (suite *TypesTestSuite) TestUploadFileViaURLArgsSerialization() {
	args := UploadFileViaURLArgs{
		FileURL:  "https://example.com/test.jpg",
		FileType: 1, // 图片
		FileName: "test.jpg",
	}

	// 序列化为JSON
	jsonData, err := json.Marshal(args)
	require.NoError(suite.T(), err)

	// 反序列化验证
	var decoded UploadFileViaURLArgs
	err = json.Unmarshal(jsonData, &decoded)
	require.NoError(suite.T(), err)

	// 验证数据完整性
	assert.Equal(suite.T(), args.FileURL, decoded.FileURL)
	assert.Equal(suite.T(), args.FileType, decoded.FileType)
	assert.Equal(suite.T(), args.FileName, decoded.FileName)
}

// TestTypesTestSuite 运行数据类型测试套件
func TestTypesTestSuite(t *testing.T) {
	suite.Run(t, new(TypesTestSuite))
}