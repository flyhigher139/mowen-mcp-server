package main

import (
	"encoding/json"
	"fmt"
	"log"
)

// TestExample 测试示例函数，展示如何使用各种功能
func TestExample() {
	fmt.Println("=== 墨问MCP服务器测试示例 ===")

	// 测试段落转换功能
	testParagraphs := []Paragraph{
		{
			Texts: []TextNode{
				{Text: "这是一个"},
				{Text: "加粗的", Bold: true},
				{Text: "测试文本"},
			},
		},
		{
			Type: "quote",
			Texts: []TextNode{
				{Text: "这是引用段落", Highlight: true},
			},
		},
		{
			Type:   "note",
			NoteID: "example-note-id",
		},
	}

	// 转换为NoteAtom格式
	noteAtom := ConvertParagraphsToNoteAtom(testParagraphs)

	// 序列化为JSON查看结果
	jsonData, err := json.MarshalIndent(noteAtom, "", "  ")
	if err != nil {
		log.Printf("JSON序列化失败: %v", err)
		return
	}

	fmt.Println("\n转换后的NoteAtom结构:")
	fmt.Println(string(jsonData))

	// 测试创建笔记请求结构
	createReq := NoteCreateRequest{
		Body: noteAtom,
		Settings: NoteCreateRequestSettings{
			AutoPublish: true,
			Tags:        []string{"测试", "示例"},
		},
	}

	reqJSON, err := json.MarshalIndent(createReq, "", "  ")
	if err != nil {
		log.Printf("请求JSON序列化失败: %v", err)
		return
	}

	fmt.Println("\n创建笔记请求结构:")
	fmt.Println(string(reqJSON))

	fmt.Println("\n=== 测试完成 ===")
}

// 如果直接运行此文件，执行测试
// 注释掉main函数以避免与实际的main.go冲突
/*
func main() {
	TestExample()
}
*/
