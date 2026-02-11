package config

type Config struct {
	DB     DBConfig
	Server ServerConfig
	JWT    JWTConfig
	Crypto CryptoConfig
	SMTP   SMTPConfig
}

type DBConfig struct {
	User     string
	Password string
	Address  string
	Port     string
	Name     string
}

type ServerConfig struct {
	Address string
}

type JWTConfig struct {
	JWTSecret string
}

type CryptoConfig struct {
	AESKey string // 32bytes
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	BaseURL  string
}
