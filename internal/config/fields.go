package config

type Config struct {
	DB     DBConfig
	Server ServerConfig
	JWT    JWTConfig
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
