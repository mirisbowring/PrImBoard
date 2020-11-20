package models

// TagMediaMap ist used to map an array of tags to an array of media
type TagMediaMap struct {
	IDs  []string `json:"ids,omitempty"`
	Tags []string `json:"tags,omitempty"`
}
