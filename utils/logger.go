package utils

// Log default method
type Log interface {
	ERROR(msg string, tags map[string]any)
	INFO(msg string, tags map[string]any)
	WARN(msg string, tags map[string]any)
	DEBUG(msg string, tags map[string]any)
	STACK(msg ...string)
}
