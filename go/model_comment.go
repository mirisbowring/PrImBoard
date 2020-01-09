package primboard

// Comment type that is beeing referenced by multiple other types
type Comment struct {
	Timestamp int64  `json:"timestamp,omitempty"`
	Username  string `json:"username,omitempty"`
	Comment   string `json:"comment,omitempty"`
}
