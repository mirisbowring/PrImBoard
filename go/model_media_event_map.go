package primboard

// MediaEventMap is used to map an array of events to an array of media
type MediaEventMap struct {
	Events   []Event  `json:"events,omitempty"`
	MediaIDs []string `json:"mediaIDs,omitempty"`
}
