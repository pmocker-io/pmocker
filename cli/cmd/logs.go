package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/pmocker-io/pmocker/cli/internal/instance"
	"github.com/pmocker-io/pmocker/pkg/pmocker/image"
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs <name|id>",
	Short: "查看实例的服务日志",
	Long: `查看实例中 gva-server 或 MCP 服务的日志。

日志类型：
  默认       gva-server 业务日志
  --access   HTTP 访问日志（nginx 风格，从 gva-server.log 过滤）
  --mcp      MCP 服务日志

控制输出：
  -f, --follow  实时跟踪日志输出（类似 tail -f，Ctrl+C 退出）
  -n, --tail N  只显示最后 N 行（默认全部）`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := instance.InitDefaultStore()
		if err != nil {
			return err
		}
		defer store.Close()
		vols, err := instance.InitDefaultVolumes()
		if err != nil {
			return err
		}
		imgDir, _ := image.DefaultStoreDir()
		imgStore := image.NewStore(imgDir)
		mgr := instance.NewManager(store, vols, imgStore, "")

		inst, err := mgr.Get(args[0])
		if err != nil {
			return err
		}

		// access 日志从 gva-server.log 过滤；--mcp 单独看 MCP 日志
		logName := "gva-server.log"
		if logMcp {
			logName = "mcp-server.log"
		}
		logPath := filepath.Join(vols.Path(inst.VolumeID), logName)

		if _, err := os.Stat(logPath); err != nil {
			return fmt.Errorf("日志文件不存在: %s", logPath)
		}

		if logFollow {
			return followLog(logPath, logAccess)
		}
		return printLog(logPath, logTail, logAccess)
	},
}

// printLog 打印日志文件，access=true 时只输出格式化的访问日志
func printLog(path string, tail int, access bool) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1024*1024), 1024*1024)

	// 三种模式各自独立，互不干扰：
	//   access=true  只输出访问日志（格式化为 nginx 风格）
	//   access=false 默认业务日志模式，过滤掉访问日志行（含"请求完成"标记）避免刷屏
	processLine := func(line string) string {
		if access {
			return formatAccessLog(line) // 非访问日志行返回 "" 被跳过
		}
		if isAccessLogLine(line) {
			return "" // 业务日志模式跳过访问日志
		}
		return line
	}

	// tail<=0 全量输出，否则滑动窗口保留最后 N 行
	if tail <= 0 {
		for sc.Scan() {
			if out := processLine(sc.Text()); out != "" {
				fmt.Println(out)
			}
		}
		return sc.Err()
	}
	var lines []string
	for sc.Scan() {
		if out := processLine(sc.Text()); out != "" {
			lines = append(lines, out)
			if len(lines) > tail {
				lines = lines[1:]
			}
		}
	}
	if err := sc.Err(); err != nil {
		return err
	}
	for _, l := range lines {
		fmt.Println(l)
	}
	return nil
}

// followLog 实时跟踪日志输出，类似 tail -f，按 Ctrl+C 退出
func followLog(path string, access bool) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// 先打印已有内容
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1024*1024), 1024*1024)
	for sc.Scan() {
		emitLogLine(sc.Text(), access)
	}

	// 监听中断信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	r := bufio.NewReader(f)
	for {
		select {
		case <-sigCh:
			return nil
		case <-ticker.C:
			// 读取新增内容
			for {
				line, err := r.ReadString('\n')
				if line != "" {
					line = strings.TrimRight(line, "\n\r")
					emitLogLine(line, access)
				}
				if err != nil {
					break
				}
			}
		}
	}
}

// emitLogLine 按模式输出单行：access 模式只输出格式化的访问日志，业务模式跳过访问日志
func emitLogLine(line string, access bool) {
	if access {
		if out := formatAccessLog(line); out != "" {
			fmt.Println(out)
		}
		return
	}
	if isAccessLogLine(line) {
		return
	}
	fmt.Println(line)
}

// accessFields 访问日志 JSON 字段
type accessFields struct {
	HttpMethod string `json:"http_method"`
	HttpPath   string `json:"http_path"`
	HttpStatus int    `json:"http_status"`
	LatencyMs  int64  `json:"latency_ms"`
	BytesIn    int64  `json:"bytes_in"`
	BytesOut   int64  `json:"bytes_out"`
	ClientIP   string `json:"client_ip"`
}

// tsRe 提取日志行中的时间戳 YYYY-MM-DD HH:MM:SS.mmm
var tsRe = regexp.MustCompile(`\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d{3}`)

// isAccessLogLine 判断是否为访问日志行（AccessLog 中间件写入，消息为"请求完成"）
func isAccessLogLine(line string) bool {
	return strings.Contains(line, "请求完成")
}

// formatAccessLog 从日志行中提取访问日志并格式化为 nginx 风格单行。
// 非访问日志行返回空字符串（被调用方跳过）。
// 输出格式: 时间  CLIENT_IP  METHOD PATH  STATUS  LATENCYms  IN/OUT
func formatAccessLog(line string) string {
	// 快速过滤：访问日志行含 "请求完成" 标记
	if !strings.Contains(line, "请求完成") {
		return ""
	}
	// 提取 JSON 部分
	idx := strings.Index(line, "{")
	if idx < 0 {
		return ""
	}
	var af accessFields
	if err := json.Unmarshal([]byte(line[idx:]), &af); err != nil {
		return ""
	}
	// 提取时间戳
	ts := tsRe.FindString(line)
	if ts == "" {
		ts = "?"
	}
	return fmt.Sprintf("%s  %s  %s %s  %d  %dms  %d/%d",
		ts, af.ClientIP, af.HttpMethod, af.HttpPath,
		af.HttpStatus, af.LatencyMs, af.BytesIn, af.BytesOut)
}

func init() {
	rootCmd.AddCommand(logsCmd)
	logsCmd.Flags().BoolVarP(&logFollow, "follow", "f", false, "实时跟踪日志输出")
	logsCmd.Flags().BoolVar(&logMcp, "mcp", false, "查看 MCP 服务日志（默认 gva-server）")
	logsCmd.Flags().BoolVar(&logAccess, "access", false, "查看 HTTP 访问日志（nginx 风格）")
	logsCmd.Flags().IntVarP(&logTail, "tail", "n", 0, "只显示最后 N 行（默认全部）")
}

var (
	logFollow bool
	logMcp    bool
	logAccess bool
	logTail   int
)
