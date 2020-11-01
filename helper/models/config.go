package models

// Config holds all possible configurations about the framework
type Config struct {
	APIGatewayConfig APIGatewayConfig `json:"api_gateway"`
	NodeConfig       NodeConfig       `json:"node"`
}

// APIGatewayConfig struct that stores every api related settings
type APIGatewayConfig struct {
	Domain               string   `json:"domain"`
	Port                 int      `json:"port"`
	MongoURL             string   `json:"mongo_url"`
	DBName               string   `json:"database_name"`
	CookiePath           string   `json:"cookie_path"`
	CookieHTTPOnly       bool     `json:"cookie_http_only"`
	CookieSecure         bool     `json:"cookie_secure"`
	CookieTokenTitle     string   `json:"cookie_token_title"`
	AllowedOrigins       []string `json:"allowed_origins"`
	TagPreviewLimit      int64    `json:"tag_preview_limit"`
	SessionRotation      bool     `json:"session_rotation"`
	DefaultMediaPageSize int      `json:"default_media_page_size"`
	InviteValidity       int      `json:"invite_validity"`
}

// NodeConfig struct that stores every api related settings
type NodeConfig struct {
	BasePath       string    `json:"basePath"`
	AllowedOrigins []string  `json:"allowed_origins"`
	GatewayURL     string    `json:"gateway_url"`
	Port           int       `json:"port"`
	NodeAuth       *NodeAuth `json:"node_auth"`
}
