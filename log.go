package main

import (
	"bufio"
	"os/exec"

	"github.com/gin-gonic/gin"
)

func streamLogs(c *gin.Context) {
	path := "/home/steam/pz-stdout.log" // 你的日志路径

	// 设置 SSE 标头
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*") // 解决跨域
	c.Writer.Header().Set("X-Accel-Buffering", "no")          // 解决 Nginx 缓冲
	c.Writer.Header().Set("Content-Encoding", "identity")     // 禁用压缩

	// 使用 tail -f -n 100 命令读取日志，这是最简单且性能最好的方式
	// -n 100 表示先输出最后100行，然后持续输出新内容
	cmd := exec.Command("tail", "-f", "-n", "100", path)

	// 获取 stdout 管道
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}

	if err := cmd.Start(); err != nil {
		return
	}

	// 创建 scanner 按行读取
	scanner := bufio.NewScanner(stdout)

	// 监听客户端断开连接 (Context Done)
	clientGone := c.Request.Context().Done()

	go func() {
		<-clientGone
		// 客户端断开后，杀掉 tail 进程，防止僵尸进程
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}()

	// 循环读取输出并发送给前端
	for scanner.Scan() {
		text := scanner.Text()
		// SSE 格式要求: "data: <content>\n\n"
		// 可以在这里简单过滤一下空行
		if text != "" {
			c.Writer.Write([]byte("data: " + text + "\n\n"))
			c.Writer.Flush()
		}
	}
}
