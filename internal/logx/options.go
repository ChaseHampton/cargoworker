package logx

type Options struct {
	ConsoleEnabled bool
	ConsoleFormat  Format // text|json (default text)
	ConsoleLevel   Level  // default info

	// file sink
	FileEnabled bool
	FilePath    string
	FileFormat  Format // usually json
	FileLevel   Level  // often debug to keep richer file logs

	Source    bool
	RunID     string
	Component string
	Quiet     bool
}
