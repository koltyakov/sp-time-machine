package handlers

// Handlers base struct
type Handlers struct{}

// NewHandlers handlers constructor
func NewHandlers() *Handlers {
	return &Handlers{}
}

// InvokeResponse response struct
type InvokeResponse struct {
	Outputs     map[string]interface{}
	ReturnValue interface{}
	Logs        []string
}

// TimerInfo event struct
type TimerInfo struct {
	Data     map[string]interface{}
	Metadata map[string]interface{}
}
