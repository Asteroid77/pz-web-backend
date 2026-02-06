package httpserver

import (
	"bufio"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (a App) handleStreamLogs(c *gin.Context) {
	path := a.LogPath
	if path == "" {
		path = "/home/steam/pz-stdout.log"
	}

	// 设置 SSE 标头
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*") // 解决跨域
	c.Writer.Header().Set("X-Accel-Buffering", "no")          // 解决 Nginx 缓冲
	c.Writer.Header().Set("Content-Encoding", "identity")     // 禁用压缩

	ctx := c.Request.Context()
	if a.LogTailer == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "log tailer not configured"})
		return
	}
	rc, err := a.LogTailer.Tail(ctx, path, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rc.Close()

	// 创建 scanner 按行读取
	scanner := bufio.NewScanner(rc)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

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

	if err := scanner.Err(); err != nil {
		c.Writer.Write([]byte("event: error\ndata: " + err.Error() + "\n\n"))
		c.Writer.Flush()
	}

	// Give the client a moment to receive buffered data before returning.
	time.Sleep(10 * time.Millisecond)
}
