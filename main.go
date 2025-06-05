package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// 检查环境变量
	if os.Getenv("MOWEN_API_KEY") == "" {
		log.Fatal("错误：未设置MOWEN_API_KEY环境变量")
	}

	// 创建MCP服务器
	server, err := NewMowenMCPServer()
	if err != nil {
		log.Fatalf("创建MCP服务器失败: %v", err)
	}

	// 设置优雅关闭
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 监听系统信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 在goroutine中启动服务器
	go func() {
		if err := server.Run(); err != nil {
			log.Printf("服务器运行错误: %v", err)
			cancel()
		}
	}()

	// 等待关闭信号
	select {
	case <-sigChan:
		log.Println("收到关闭信号，正在关闭服务器...")
	case <-ctx.Done():
		log.Println("服务器上下文已取消")
	}

	// 优雅关闭服务器
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("关闭服务器时出错: %v", err)
	} else {
		log.Println("服务器已成功关闭")
	}
}
