package pkg

import (
	"errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"time"
)

type App struct {
	Port              int           `yaml:"port"`
	Salt              string        `yaml:"salt"`
	SessionExpiration time.Duration `yaml:"session_expire"`
	CookieDomain      string        `yaml:"cookie_domain"`
	CookieSecure      int           `yaml:"cookie_secure"`
	TemplateDirectory string        `yaml:"template_dir"`
	AssetDirectory    string        `yaml:"asset_dir"`
	Logo              struct {
		FavIcon string `yaml:"favicon"`
		Url     string `yaml:"url"`
		Alt     string `yaml:"alt"`
	} `yaml:"logo"`
	Contact struct {
		Email   string   `yaml:"email"`
		Phone   []string `yaml:"phone"`
		Address string   `yaml:"address"`
	} `yaml:"contact"`
}

type User struct {
	Username string `yaml:"uname"`
	Password string `yaml:"pass"`
}

type Link struct {
	Title    string `yaml:"title"`
	SubTitle string `yaml:"subtitle"`
	Icon     string `yaml:"icon"`
	Url      string `yaml:"url"`
	GitUrl   string `yaml:"git_url"`
}

type Group struct {
	Name  string `yaml:"name"`
	Slug  string `yaml:"-"`
	Links []Link `yaml:"links"`
}

type Configuration struct {
	App    App     `yaml:"app"`
	Users  []User  `yaml:"users"`
	Groups []Group `yaml:"groups"`
}

type config struct {
	filename []string
	data     *Configuration
}

type ConfigManager interface {
	GetConfig() *Configuration
	Initiate() error
	Reload() error
}

func NewConfig(filename ...string) ConfigManager {
	return &config{
		filename: filename,
		data:     nil,
	}
}

func (c *config) GetConfig() *Configuration {
	if c.data == nil {
		c.data, _ = c.read()
	}
	return c.data
}

func (c *config) GetApp() App {
	return c.data.App
}

func (c *config) GetUsers() []User {
	return c.data.Users
}

func (c *config) GetGroups() []Group {
	return c.data.Groups
}

func (c *config) Initiate() error {
	data, err := c.read()
	if err != nil {
		return err
	}
	c.data = data
	return nil
}

func (c *config) Reload() (err error) {
	c.data, err = c.read()
	return
}

// fileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func (c *config) fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func (c *config) read() (*Configuration, error) {
	fName := ""
	for _, f := range c.filename {
		if c.fileExists(f) {
			fName = f
			break
		}
	}
	if len(fName) < 1 {
		return nil, errors.New("no configuration file found")
	}
	cfg, err := ioutil.ReadFile(fName)
	if err != nil {
		return nil, err
	}
	var config Configuration
	err = yaml.Unmarshal(cfg, &config)
	if err != nil {
		return nil, err

	}
	return &config, nil
}
