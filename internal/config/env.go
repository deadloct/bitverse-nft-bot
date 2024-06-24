package config

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/joho/godotenv"
)

const Prefix = "BITVERSE_NFT_BOT"

func LoadEnvFiles() {
	env := os.Getenv(EnvKey("ENV"))
	if env == "" {
		env = "development"
	}

	godotenv.Load(envPath(fmt.Sprintf(".env.%s.local", env)))
	godotenv.Load(envPath(fmt.Sprintf(".env.%s", env)))
	godotenv.Load(envPath(".env"))
}

func GetenvStr(key string) string {
	return os.Getenv(EnvKey(key))
}

func EnvKey(str string) string {
	return fmt.Sprintf("%s_%s", Prefix, str)
}

func envPath(filename string) string {
	p, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	p = filepath.Dir(p)
	return path.Join(p, filename)
}
