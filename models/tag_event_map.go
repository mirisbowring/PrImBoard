package models

// TagEventMap ist used to map an array of tags to an array of media
type TagEventMap struct {
	Events []Event  `json:"events,omitempty"`
	Tags   []string `json:"tags,omitempty"`
}
