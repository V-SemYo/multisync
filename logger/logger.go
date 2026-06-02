package logger

import "fmt"

// Logger — логгер с уровнями логирования
type Logger struct {
	Level   string
	LogFile string
}

func NewLogger(level, logFile string) (*Logger, error) {
	return &Logger{
		Level:   level,
		LogFile: logFile,
	}, nil
}

// levelPriority — внутренняя функция, превращает строку уровня в число, чем выше число, тем важнее уровень
func levelPriority(level string) int {
	switch level {
	case "debug":
		return 10
	case "info":
		return 20
	case "warn":
		return 30
	case "error":
		return 40
	default:
		return 0
	}
}

// Debug — отладочные сообщения
func (l *Logger) Debug(format string, args ...any) {

	if levelPriority(l.Level) <= 10 {
		fmt.Printf("DEBUG "+format+"\n", args...)
	}
}

// Info — информационные сообщения
func (l *Logger) Info(format string, args ...any) {

	if levelPriority(l.Level) <= 20 {
		fmt.Printf("INFO "+format+"\n", args...)
	}
}

// Warn — предупреждения (что-то пошло не так, но программа продолжает)
func (l *Logger) Warn(format string, args ...any) {

	if levelPriority(l.Level) <= 30 {
		fmt.Printf("WARN "+format+"\n", args...)
	}
}

// Error — ошибки (что-то сломалось)
func (l *Logger) Error(format string, args ...any) {

	if levelPriority(l.Level) <= 40 {
		fmt.Printf("ERROR "+format+"\n", args...)
	}
}
