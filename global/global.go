package g

import (
	"encoding/json"
	"fmt"
	"forum/common/custom_errs"
	"forum/persistence/database"
	"forum/server/core/config"
	"forum/server/core/handlers"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"time"
)

//g stands for global, this is a package meant to be accessed globally from the entire project
//should be read only, holding enums and configs

// Place your struct pointer here
type Configuration struct {
	Database       *database.DBConfig     `json:"database_configuration"`
	Server         *http.Server           `json:"server"`
	Certifications *config.Certifications `json:"certifications"`
	Handlers       *handlers.Config       `json:"handlers"`
	OAuth          *config.OAuthConfig    `json:"oauth"`
}

var Configs *Configuration

func Initialize() error {
	var err error
	Configs, err = SetUp()
	if err != nil {
		return fmt.Errorf("start: %w", err)
	}
	return nil
}

func SetUp() (*Configuration, error) {
	errorMsg := "SetUp: %w"
	data, err := os.ReadFile(filepath.Join("configs.json"))
	if err != nil {
		return nil, fmt.Errorf(errorMsg, err)
	}

	var cfg *Configuration
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Fatalf(errorMsg, err)
	}

	SetTimeouts(cfg.Server)

	if cfg.Database == nil || reflect.ValueOf(*cfg.Database).IsZero() {
		log.Fatalf(errorMsg, "database config: %w", custom_errs.ErrNilConfigStruct)
	}

	if cfg.Server == nil || reflect.ValueOf(*cfg.Server).IsZero() {
		log.Fatalf(errorMsg, "server config: %w", custom_errs.ErrNilConfigStruct)
	}

	if cfg.Certifications == nil || reflect.ValueOf(*cfg.Certifications).IsZero() {
		log.Fatalf(errorMsg, "certifications config: %w", custom_errs.ErrNilConfigStruct)
	}
	if cfg.Handlers == nil || reflect.ValueOf(*cfg.Handlers).IsZero() {
		log.Fatalf(errorMsg, "handlers config: %w", custom_errs.ErrNilConfigStruct)
	}
	return cfg, nil
}

func SetTimeouts(server *http.Server) {
	// Set server timeouts to prevent slowloris and other DDoS attacks
	server.ReadTimeout = 4 * time.Second
	server.WriteTimeout = 8 * time.Second
	server.IdleTimeout = 10 * time.Second
	server.ReadHeaderTimeout = 2 * time.Second
}
