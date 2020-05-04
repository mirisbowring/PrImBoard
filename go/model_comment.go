package primboard

import (
	"errors"
	"strings"
	"time"
)

// Comment type that is beeing referenced by multiple other types
type Comment struct {
	Timestamp int64  `json:"timestamp,omitempty"`
	Username  string `json:"username,omitempty"`
	Comment   string `json:"comment,omitempty"`
}

//AddMetadata sets the passed username and the current timestamp for this comment
func (c *Comment) AddMetadata(username string) {
	c.Username = username
	c.Timestamp = int64(time.Now().Unix())
}

//IsValid verifies, that all values of that comment are valid and allowed
func (c *Comment) IsValid() error {
	if c.Timestamp == 0 {
		return errors.New("timestamp of comment not set")
	} else if c.Username == "" {
		return errors.New("username of comment not set")
	} else if len(strings.TrimSpace(c.Comment)) == 0 {
		return errors.New("comment cannot be empty")
	}
	return nil
}
