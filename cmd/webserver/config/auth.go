package config

type AuthenticationConfig struct {
	Local AuthenticationLocalConfig           `mapstructure:"local"`
	Oidc  map[string]AuthenticationOidcConfig `mapstructure:"oidc"`
}

type AuthenticationLocalConfig struct {
	Enabled           bool `mapstructure:"enabled"`
	AllowRegistration bool `mapstructure:"allow_registration"`
}

type AuthenticationOidcConfig struct {
	Name              string   `mapstructure:"name"`
	AllowRegistration bool     `mapstructure:"allow_registration"`
	Issuer            string   `mapstructure:"issuer"`
	ClientID          string   `mapstructure:"client_id"`
	ClientSecret      string   `mapstructure:"client_secret"`
	Scopes            []string `mapstructure:"scopes"`
}
