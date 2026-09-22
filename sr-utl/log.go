package utl

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LogLevel 日志级别
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
)

func (l LogLevel) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// LogConfig 日志配置，所有字段均支持热更新（Dir、MaxAgeDays 除外）
type LogConfig struct {
	EnableConsole bool     // 是否输出终端
	ConsoleLevel  LogLevel // 终端输出最低级别

	EnableFile bool     // 是否写入文件
	FileLevel  LogLevel // 文件输出最低级别
	FilePath   string   // 日志文件完整路径（Dir 为空时生效，旧式单文件模式）
	Dir        string   // 日志目录（新模式：按日期自动切分，文件名 stargo-2006-01-02.log）
	MaxSize    int64    // 单个日志最大字节，超过轮转（Dir 为空时按大小轮转；支持热更新）
	MaxAgeDays int      // 超过 N 天的日志自动删除（仅 Dir 模式生效，启动时和跨天时清理）
}

// LogEntry 一条日志实体
type LogEntry struct {
	Time    time.Time
	Level   LogLevel
	Message string
}

// Logger 日志对象
type logger struct {
	rwMu sync.RWMutex // 保护配置、运行状态
	cfg  LogConfig

	fileMu      sync.Mutex // 保护文件句柄与写入操作
	file        *os.File
	currentDate string // 当前文件对应日期 "2006-01-02"（Dir 模式专用）

	queue    chan LogEntry  // 异步写文件队列
	fileStop chan struct{}  // 文件写协程停止信号
	wg       sync.WaitGroup // 协程等待组
	closed   bool           // 日志器是否已关闭
}

// NewLogger 创建日志实例
func NewLogger(cfg LogConfig) (*logger, error) {
	l := &logger{
		cfg:    cfg,
		queue:  make(chan LogEntry, 1000),
		closed: false,
	}

	// 初始启用文件日志则打开文件并启动协程
	if cfg.EnableFile {
		if err := l.openFileLocked(); err != nil {
			return nil, err
		}
		l.fileStop = make(chan struct{})
		l.wg.Add(1)
		go l.fileWriterLoop()
		// 启动时清理一次旧日志
		go l.cleanOldLogs()
	}

	return l, nil
}

// openFileLocked 打开日志文件（调用方需保证 rwMu 写锁持有）
func (l *logger) openFileLocked() error {
	path := l.cfg.FilePath
	if l.cfg.Dir != "" {
		// 新模式：按日期生成文件名
		if err := os.MkdirAll(l.cfg.Dir, 0751); err != nil {
			return fmt.Errorf("create log dir failed: %w", err)
		}
		l.currentDate = time.Now().Format("2006-01-02")
		path = filepath.Join(l.cfg.Dir, "stargo-"+l.currentDate+".log")
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	l.fileMu.Lock()
	l.file = f
	l.fileMu.Unlock()
	return nil
}

// UpdateConfig 热更新日志配置，原子生效
// 支持动态修改：EnableConsole、EnableFile、ConsoleLevel、FileLevel、MaxSize
// 不支持热更 FilePath、Dir、MaxAgeDays，如需修改请先关闭文件日志再重新开启
func (l *logger) UpdateConfig(newCfg LogConfig) error {
	l.rwMu.Lock()
	defer l.rwMu.Unlock()

	if l.closed {
		return errors.New("logger is closed")
	}

	oldCfg := l.cfg
	l.cfg = newCfg

	// 处理文件日志开关状态变化
	switch {
	case oldCfg.EnableFile && !newCfg.EnableFile:
		// 开启 -> 关闭：发停止信号，等待队列排空，关闭文件
		close(l.fileStop)
		l.wg.Wait()

		l.fileMu.Lock()
		if l.file != nil {
			_ = l.file.Close()
			l.file = nil
		}
		l.fileMu.Unlock()

	case !oldCfg.EnableFile && newCfg.EnableFile:
		// 关闭 -> 开启：打开文件，启动写协程
		if err := l.openFileLocked(); err != nil {
			l.cfg = oldCfg // 失败回滚配置
			return fmt.Errorf("open log file failed: %w", err)
		}
		l.fileStop = make(chan struct{})
		l.wg.Add(1)
		go l.fileWriterLoop()
		go l.cleanOldLogs()

	case oldCfg.EnableFile && newCfg.EnableFile:
		// 持续开启：配置直接生效（级别、MaxSize 等），无需重启协程
	}

	return nil
}

// rotateBySizeLocked 旧式按大小轮转（调用方需保证 fileMu 锁持有）
// 仅 Dir 为空时使用
func (l *logger) rotateBySizeLocked() error {
	if l.file == nil || l.cfg.MaxSize <= 0 {
		return nil
	}

	stat, err := l.file.Stat()
	if err != nil {
		return err
	}
	if stat.Size() < l.cfg.MaxSize {
		return nil
	}

	// 关闭旧文件并重命名
	_ = l.file.Close()
	ext := filepath.Ext(l.cfg.FilePath)
	base := l.cfg.FilePath[:len(l.cfg.FilePath)-len(ext)]
	backupName := fmt.Sprintf("%s.%s%s", base, time.Now().Format("20060102_150405"), ext)
	_ = os.Rename(l.cfg.FilePath, backupName)

	// 创建新文件
	f, err := os.OpenFile(l.cfg.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	l.file = f
	return nil
}

// rotateByDateLocked 按日期切分（调用方需保证 fileMu 锁持有）
// 仅 Dir 模式使用：关闭旧文件，按今日日期开新文件
func (l *logger) rotateByDateLocked() error {
	if l.file != nil {
		_ = l.file.Close()
	}
	l.currentDate = time.Now().Format("2006-01-02")
	path := filepath.Join(l.cfg.Dir, "stargo-"+l.currentDate+".log")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	l.file = f
	return nil
}

// cleanOldLogs 清理超过 MaxAgeDays 天的日志文件（仅 Dir 模式生效）
func (l *logger) cleanOldLogs() {
	l.rwMu.RLock()
	dir := l.cfg.Dir
	maxAge := l.cfg.MaxAgeDays
	l.rwMu.RUnlock()

	if dir == "" || maxAge <= 0 {
		return
	}
	cutoff := time.Now().AddDate(0, 0, -maxAge)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			_ = os.Remove(filepath.Join(dir, e.Name()))
		}
	}
}

// fileWriterLoop 异步文件写协程
func (l *logger) fileWriterLoop() {
	defer l.wg.Done()

	for {
		select {
		case entry := <-l.queue:
			l.writeEntryToFile(entry)
		case <-l.fileStop:
			// 退出前排空队列，保证日志不丢失
			for {
				select {
				case entry := <-l.queue:
					l.writeEntryToFile(entry)
				default:
					return
				}
			}
		}
	}
}

// writeEntryToFile 写入单条日志到文件
func (l *logger) writeEntryToFile(entry LogEntry) {
	l.fileMu.Lock()
	defer l.fileMu.Unlock()

	if l.file == nil {
		return
	}

	// 路径模式选择：Dir 模式按日期切分，否则按大小轮转
	l.rwMu.RLock()
	dirMode := l.cfg.Dir != ""
	l.rwMu.RUnlock()

	if dirMode {
		today := time.Now().Format("2006-01-02")
		if today != l.currentDate {
			if err := l.rotateByDateLocked(); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "[WARN] rotate log by date failed: %v\n", err)
			} else {
				go l.cleanOldLogs()
			}
		}
	} else {
		_ = l.rotateBySizeLocked()
	}

	line := fmt.Sprintf("[%s] [%s] %s\n",
		entry.Time.Format("2006-01-02 15:04:05"),
		entry.Level,
		entry.Message,
	)
	_, _ = l.file.WriteString(line)
}

// log 内部日志入口
func (l *logger) log(level LogLevel, msg string) {
	l.rwMu.RLock()
	if l.closed {
		l.rwMu.RUnlock()
		return
	}
	cfg := l.cfg
	l.rwMu.RUnlock()

	entry := LogEntry{
		Time:    time.Now(),
		Level:   level,
		Message: msg,
	}

	// 终端输出：同步写，独立开关+级别过滤
	if cfg.EnableConsole && level >= cfg.ConsoleLevel {
		line := fmt.Sprintf("[%s] [%s] %s\n",
			entry.Time.Format("2006-01-02 15:04:05"),
			entry.Level,
			entry.Message,
		)
		_, _ = os.Stdout.WriteString(line)
	}

	// 文件输出：入队异步写，独立开关+级别过滤
	if cfg.EnableFile && level >= cfg.FileLevel {
		select {
		case l.queue <- entry:
		default:
			_, _ = fmt.Fprintf(os.Stderr, "[WARN] log queue full, drop log: %s\n", msg)
		}
	}
}

// Debug/Info/Warn/Error 对外方法，nil-safe：未初始化时降级到 stderr
func (l *logger) Debug(msg string) {
	if l == nil {
		fmt.Fprintf(os.Stderr, "[DEBUG] %s\n", msg)
		return
	}
	l.log(LevelDebug, msg)
}

func (l *logger) Info(msg string) {
	if l == nil {
		fmt.Fprintf(os.Stderr, "[INFO] %s\n", msg)
		return
	}
	l.log(LevelInfo, msg)
}

func (l *logger) Warn(msg string) {
	if l == nil {
		fmt.Fprintf(os.Stderr, "[WARN] %s\n", msg)
		return
	}
	l.log(LevelWarn, msg)
}

func (l *logger) Error(msg string) {
	if l == nil {
		fmt.Fprintf(os.Stderr, "[ERROR] %s\n", msg)
		return
	}
	l.log(LevelError, msg)
}

// Close 关闭日志器，排空队列后释放资源
func (l *logger) Close() error {
	l.rwMu.Lock()
	defer l.rwMu.Unlock()

	if l.closed {
		return nil
	}
	l.closed = true

	// 停止文件写协程并关闭文件
	if l.cfg.EnableFile {
		close(l.fileStop)
		l.wg.Wait()

		l.fileMu.Lock()
		if l.file != nil {
			_ = l.file.Close()
			l.file = nil
		}
		l.fileMu.Unlock()
	}

	close(l.queue)
	return nil
}

// ===== 包级单例 =====

// Logger 包级默认日志器，调用方用 utl.Logger.Info(msg) 等。
// 未通过 InitLogger 初始化时为 nil，方法调用降级到 stderr 输出，不会 panic。
var Logger *logger

// InitLogger 初始化包级默认日志器
func InitLogger(cfg LogConfig) error {
	l, err := NewLogger(cfg)
	if err != nil {
		return err
	}
	Logger = l
	return nil
}

// CloseLogger 关闭包级默认日志器
func CloseLogger() {
	if Logger != nil {
		_ = Logger.Close()
		Logger = nil
	}
}

// ===== XDG 路径辅助 =====

// DefaultLogDir 返回默认日志目录 ~/.local/state/stargo/logs
// 优先使用 $XDG_STATE_HOME
func DefaultLogDir() string {
	if dir := os.Getenv("XDG_STATE_HOME"); dir != "" {
		return filepath.Join(dir, "stargo", "logs")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "stargo", "logs")
	}
	return filepath.Join(home, ".local", "state", "stargo", "logs")
}

// DefaultLogConfig 返回生产环境默认日志配置
// 终端 INFO + 文件 DEBUG，日志按日期切分到 DefaultLogDir()，保留 7 天
func DefaultLogConfig() LogConfig {
	return LogConfig{
		EnableConsole: true,
		ConsoleLevel:  LevelInfo,
		EnableFile:    true,
		FileLevel:     LevelDebug,
		Dir:           DefaultLogDir(),
		MaxAgeDays:    7,
	}
}

// ===== 测试入口 =====

func TestLogger() {
	// 初始配置：开发模式，终端+文件双输出
	cfg := LogConfig{
		EnableConsole: true,
		ConsoleLevel:  LevelInfo,
		EnableFile:    true,
		FileLevel:     LevelDebug,
		FilePath:      "./stargo_hot.log",
		MaxSize:       1024 * 1024, // 1MB 轮转
	}

	logger, err := NewLogger(cfg)
	if err != nil {
		fmt.Printf("init logger failed: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	fmt.Println("=== 初始状态：终端INFO + 文件DEBUG ===")
	logger.Debug("debug msg 1") // 终端不显示，文件有
	logger.Info("service start")
	logger.Warn("memory high")

	// 热更新1：关闭终端，只保留文件日志，级别提升到 INFO
	fmt.Println("\n=== 热更新：关闭终端，文件仅输出INFO以上 ===")
	_ = logger.UpdateConfig(LogConfig{
		EnableConsole: false,
		ConsoleLevel:  LevelInfo,
		EnableFile:    true,
		FileLevel:     LevelInfo,
		FilePath:      "./stargo_hot.log",
		MaxSize:       1024 * 1024,
	})
	logger.Debug("debug msg 2") // 文件也不写
	logger.Info("config updated")
	logger.Error("db connect failed")

	// 热更新2：关闭文件日志，重新开启终端，级别设为 WARN
	fmt.Println("\n=== 热更新：关闭文件，终端仅输出WARN以上 ===")
	_ = logger.UpdateConfig(LogConfig{
		EnableConsole: true,
		ConsoleLevel:  LevelWarn,
		EnableFile:    false,
		FilePath:      "./app_hot.log",
	})
	logger.Info("this info will not show") // 终端不显示
	logger.Warn("disk space low")
	logger.Error("request timeout")

	// 热更新3：重新开启文件日志，恢复双输出
	fmt.Println("\n=== 热更新：恢复终端+文件双输出 ===")
	_ = logger.UpdateConfig(LogConfig{
		EnableConsole: true,
		ConsoleLevel:  LevelDebug,
		EnableFile:    true,
		FileLevel:     LevelDebug,
		FilePath:      "./app_hot.log",
		MaxSize:       1024 * 1024,
	})
	logger.Debug("final debug log")
	logger.Info("demo finish")
}
