package cmd

type Config struct {
	Environment     string `arg:"env:ENVIRONMENT"`
	ZipkinServerUrl string `arg:"env:ZIPKIN_SERVER_URL"`
    PostgresUrl string
	ServerConfig
	AmqpConfig
}

type ServerConfig struct {
	Port string `arg:"env:SERVER_PORT"`
	Name string `arg:"env:SERVICE_NAME"`
}

type AmqpConfig struct {
	ServerUrl string `arg:"env:AMQP_SERVER_URL"`
}

func DefaultConfiguration() *Config {
	return &Config{
		Environment:     "dev",
		ZipkinServerUrl: "http://zipkin:9411",
		PostgresUrl: "postgres://data:password@db:5432/data?sslmode=disable",

		ServerConfig: ServerConfig{
			Name: "image-service",
			Port: "7777",
		},
	}
}
