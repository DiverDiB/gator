package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const configFileName = ".gatorconfig.json"

// Export a Config struct the represents the JSON file structure, including struct tags for JSON decoding.
type Config struct {
	DbURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

// Export a SetUser method on the Config struct
// that writes the config struct to the JSON file
// after setting the current_user_name field.
func (cfg *Config) SetUser(userName string) error {
	cfg.CurrentUserName = userName
	return write(*cfg)
}

// Export a Read function that reads the JSON file
// found at ~/.gatorconfig.json and returns a
// Config struct. It should read the file from the
// HOME directory, then decode the JSON string
// into a new Config struct.  Use os.UserHomeDir
// to get the location of HOME.
func Read() (Config, error) {
	fullpath, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}

	file, err := os.Open(fullpath)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	cfg := Config{}
	err = decoder.Decode(&cfg)
	if err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func getConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configPath := filepath.Join(homeDir, configFileName)
	return configPath, nil
}

func write(cfg Config) error {
	fullpath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	file, err := os.Create(fullpath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(cfg)
	if err != nil {
		return err
	}

	return nil
}
