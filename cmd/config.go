package cmd

import "fmt"

type config struct {
	DatabaseUser     string `env:"DB_USER"`
	DatabaseHost     string `env:"DB_HOST"`
	DatabasePassword string `env:"DB_PASSWORD"`
	DatabaseName     string `env:"DB_NAME"`
	DatabasePort     int    `env:"DB_PORT"`
	SignDirectory    string `env:"LOCAL_IMAGE_DIRECTORY"`
	EagleApiUrl      string `env:"EAGLE_URL"`
	EagleApiToken    string `env:"EAGLE_TOKEN"`
	BaseWebsite      string `env:"BASE_WEBSITE"`
}

func (c *config) GetDatabaseConn() string {
	if c.DatabasePassword == "" {
		return fmt.Sprintf("postgres://%s@%s:%v/%s", c.DatabaseUser, c.DatabaseHost, c.DatabasePort, c.DatabaseName)
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%v/%s", c.DatabaseUser, c.DatabasePassword, c.DatabaseHost, c.DatabasePort, c.DatabaseName)
}
