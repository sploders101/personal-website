package config

type AuthenticationConfig struct {
	Local AuthenticationLocalConfig          `json:"local"`
	Oidc  map[string]AuthenticationOidcConfig `json:"oidc"`
}

type AuthenticationLocalConfig struct {
	Enabled bool `json:"enabled"`
}

type AuthenticationOidcConfig struct {
	Name         string   `json:"name"`
	Issuer       string   `json:"issuer"`
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	Scopes       []string `json:"scopes"`
}
