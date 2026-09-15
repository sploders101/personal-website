package config

type AuthenticationConfig struct {
	Local AuthenticationLocalConfig           `json:"local"`
	Oidc  map[string]AuthenticationOidcConfig `json:"oidc"`
}

type AuthenticationLocalConfig struct {
	Enabled           bool `json:"enabled"`
	AllowRegistration bool `json:"allow_registration"`
}

type AuthenticationOidcConfig struct {
	Name              string   `json:"name"`
	AllowRegistration bool     `json:"allow_registration"`
	Issuer            string   `json:"issuer"`
	ClientID          string   `json:"client_id"`
	ClientSecret      string   `json:"client_secret"`
	Scopes            []string `json:"scopes"`
}
