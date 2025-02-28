package logger

import (
	"fmt"
	"github.com/faelmori/logz/internal/services"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

// Mapping dos níveis de log para um valor numérico de severidade.
var logLevels = map[LogLevel]int{
	DEBUG: 1,
	INFO:  2,
	WARN:  3,
	ERROR: 4,
	FATAL: 5,
}

// Logger orquestra a criação da entrada de log, sua escrita e o envio para notifiers.
type Logger struct {
	level     LogLevel
	writer    LogWriter
	notifiers []Notifier
	metadata  map[string]interface{}
}

// NewLogger cria uma nova instância de Logger com base nos parâmetros fornecidos.
func NewLogger(level LogLevel, format string, outputPath, externalURL, zmqEndpoint, discordWebhook string) *Logger {
	var out *os.File
	if outputPath == "stdout" {
		out = os.Stdout
	} else {
		var err error
		out, err = os.OpenFile(outputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println(fmt.Sprintf("Erro ao abrir arquivo de log: %v\nRedirecionando para stdout...", err))
			// Se não for possível abrir o arquivo de log, redireciona para stdout.
			out = os.Stdout
		}
	}

	var formatter LogFormatter
	if format == "json" {
		formatter = &JSONFormatter{}
	} else {
		formatter = &TextFormatter{}
	}
	writer := NewDefaultWriter(out, formatter)

	var notifiers []Notifier
	notifiers = make([]Notifier, 0)

	if externalURL != "" || zmqEndpoint != "" {
		extNotifier := NewExternalNotifier(externalURL, zmqEndpoint)
		notifiers = append(notifiers, extNotifier)
	}

	return &Logger{
		level:     level,
		writer:    writer,
		notifiers: notifiers,
		metadata:  make(map[string]interface{}),
	}
}

// ParseLogLevel converte uma string para LogLevel; valores inválidos retornam INFO.
func ParseLogLevel(level string) LogLevel {
	switch level {
	case "debug":
		return DEBUG
	case "info":
		return INFO
	case "warn":
		return WARN
	case "error":
		return ERROR
	case "fatal":
		return FATAL
	default:
		return INFO
	}
}

// SetMetadata adiciona um metadado global que será mesclado com o contexto de cada log.
func (l *Logger) SetMetadata(key string, value interface{}) {
	l.metadata[key] = value
}

func (l *Logger) shouldLog(level LogLevel) bool {
	return logLevels[level] >= logLevels[l.level]
}

// getCallerInfo captura informações do caller usando runtime.Caller.
// O parâmetro skip indica quantos níveis pular (ajuste conforme a estrutura de chamadas).
func getCallerInfo(skip int) string {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown"
	}
	funcName := runtime.FuncForPC(pc).Name()
	return fmt.Sprintf("%s:%d %s", trimFilePath(file), line, funcName)
}

// trimFilePath reduz o caminho do arquivo para os dois últimos componentes.
func trimFilePath(filePath string) string {
	parts := strings.Split(filePath, "/")
	if len(parts) > 2 {
		return strings.Join(parts[len(parts)-2:], "/")
	}
	return filePath
}

// log cria uma entrada de log, mescla os metadados e delega a escrita e envio.
func (l *Logger) log(level LogLevel, msg string, ctx map[string]interface{}) {
	if !l.shouldLog(level) {
		return
	}
	// Captura o timestamp e as informações do caller automaticamente.
	timestamp := time.Now().UTC()
	caller := getCallerInfo(3)

	// Cria a entrada de log e preenche os campos necessários,
	// incluindo Caller, que é obrigatório para rastreabilidade.
	entry := NewLogEntry().
		WithLevel(level).
		WithMessage(msg).
		WithSeverity(logLevels[level])
	// Define os campos obrigatórios automaticamente.
	entry.Timestamp = timestamp
	entry.Caller = caller

	// Adiciona os metadados globais e específicos (se houver).
	finalContext := mergeContext(l.metadata, ctx)
	for k, v := range finalContext {
		entry.AddMetadata(k, v)
	}

	// Escreve o log utilizando o writer.
	if err := l.writer.Write(entry); err != nil {
		log.Printf("Erro ao escrever log: %v", err)
	}

	// Notifica os notifiers configurados.
	for _, notifier := range l.notifiers {
		notifier.Notify(entry)
	}

	// Integração com métricas: atualização automática
	pm := services.GetPrometheusManager()
	if pm.IsEnabled() {
		pm.IncrementMetric("logs_total", 1)
		pm.IncrementMetric("logs_total_"+string(level), 1)
	}

	if level == FATAL {
		os.Exit(1)
	}
}

func (l *Logger) Debug(msg string, ctx map[string]interface{}) { l.log(DEBUG, msg, ctx) }
func (l *Logger) Info(msg string, ctx map[string]interface{})  { l.log(INFO, msg, ctx) }
func (l *Logger) Warn(msg string, ctx map[string]interface{})  { l.log(WARN, msg, ctx) }
func (l *Logger) Error(msg string, ctx map[string]interface{}) { l.log(ERROR, msg, ctx) }
func (l *Logger) Fatal(msg string, ctx map[string]interface{}) { l.log(FATAL, msg, ctx) }

// mergeContext une os metadados globais e o contexto específico.
func mergeContext(global, local map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	for k, v := range global {
		merged[k] = v
	}
	for k, v := range local {
		merged[k] = v
	}
	return merged
}
