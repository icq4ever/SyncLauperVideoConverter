package encoder

// EncodingProgress represents the current encoding progress
type EncodingProgress struct {
	Filename    string  `json:"filename"`
	Progress    float64 `json:"progress"`    // 0-100
	ETA         string  `json:"eta"`         // e.g., "00:05:32"
	CurrentFile int     `json:"currentFile"` // 1-based index
	TotalFiles  int     `json:"totalFiles"`
	Status      string  `json:"status"` // "waiting", "encoding", "completed", "error", "cancelled"
	PassNumber  int     `json:"passNumber"` // 1 or 2 for multi-pass encoding
	TotalPasses int     `json:"totalPasses"`
	Speed       string  `json:"speed"` // e.g., "1.5x"
}

// Status constants
const (
	StatusWaiting   = "waiting"
	StatusEncoding  = "encoding"
	StatusCompleted = "completed"
	StatusError     = "error"
	StatusCancelled = "cancelled"
)
