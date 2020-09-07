package glog

const (
	PanicLevel Level = iota
	FatalLevel
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
	TraceLevel
)

// Level 日志级别
type Level int8

var levelNameMapping = []string{
	PanicLevel: "PANIC",
	FatalLevel: "FATAL",
	ErrorLevel: "ERROR",
	WarnLevel:  "WARN",
	InfoLevel:  "INFO",
	DebugLevel: "DEBUG",
	TraceLevel: "TRACE",
}

// String returns a ASCII representation of the log level.
func (l Level) String() string {
	return levelNameMapping[l]
}

var syslogLevelMapping = []SyslogLevel{
	PanicLevel: SLAlert,
	FatalLevel: SLCritical,
	ErrorLevel: SLError,
	WarnLevel:  SLWarning,
	InfoLevel:  SLInformational,
	DebugLevel: SLDebug,
	TraceLevel: SLDebug,
}

func (l Level) ToSyslogLevel() SyslogLevel {
	return syslogLevelMapping[l]
}

// SyslogLevel syslog日志级别
type SyslogLevel int8

const (
	SLEmergency SyslogLevel = iota
	SLAlert
	SLCritical
	SLError
	SLWarning
	SLNotice
	SLInformational
	SLDebug
)
