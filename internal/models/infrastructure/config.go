package infrastructure

// Config holds all possible configurations about the framework
type Config struct {
	APIGatewayConfig APIGatewayConfig `json:"api_gateway"`
	NodeConfig       NodeConfig       `json:"node"`
}

// APIGatewayConfig struct that stores every api related settings
type APIGatewayConfig struct {
	Domain               string          `json:"domain"`
	Port                 int             `json:"port"`
	MongoURL             string          `json:"mongo_url"`
	DBName               string          `json:"database_name"`
	HTTP                 bool            `json:"http"`
	CaCert               string          `json:"ca-cert"`
	Certificates         string          `json:"certificates"`
	CookieDomain         string          `json:"cookie_domain"`
	CookiePath           string          `json:"cookie_path"`
	CookieHTTPOnly       bool            `json:"cookie_http_only"`
	CookieSameSite       int             `json:"cookie_same_site"`
	CookieSecure         bool            `json:"cookie_secure"`
	CookieTokenTitle     string          `json:"cookie_token_title"`
	AllowedOrigins       []string        `json:"allowed_origins"`
	TagPreviewLimit      int             `json:"tag_preview_limit"`
	TLSInsecure          bool            `json:"tls_insecure"`
	SessionRotation      bool            `json:"session_rotation"`
	DefaultMediaPageSize int             `json:"default_media_page_size"`
	InviteValidity       int             `json:"invite_validity"`
	Keycloak             *KeycloakConfig `json:"keycloak_config"`
}

// KeycloakConfig stores the information for the client to connect to a keycloak
type KeycloakConfig struct {
	URL      string `json:"url"`
	Realm    string `json:"realm"`
	ClientID string `json:"client_id"`
	Secret   string `json:"secret"`
}

// NodeConfig struct that stores every api related settings
type NodeConfig struct {
	BasePath       string          `json:"basePath"`
	CaCert         string          `json:"ca-cert"`
	Certificates   string          `json:"certificates"`
	TargetPath     string          `json:"targetPath"`
	AllowedOrigins []string        `json:"allowed_origins"`
	GatewayURL     string          `json:"gateway_url"`
	Keycloak       *KeycloakConfig `json:"keycloak_config"`
	Port           int             `json:"port"`
	TLSInsecure    bool            `json:"tls_insecure"`
	NodeAuth       *NodeAuth       `json:"node_auth"`
}

// NodeAuth represents the id / secret map for the current node deployment
type NodeAuth struct {
	ID     string `json:"id"`
	Secret string `json:"secret"`
}
