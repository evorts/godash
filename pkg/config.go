package pkg

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"time"
)

type App struct {
	Port int `yaml:"port"`
	Salt string `yaml:"salt"`
	SessionExpiration time.Duration `yaml:"session_expire"`
	CookieDomain string `yaml:"cookie_domain"`
	CookieSecure int `yaml:"cookie_secure"`
	TemplateDirectory string `yaml:"template_dir"`
	AssetDirectory string `yaml:"asset_dir"`
	Logo struct {
		Url string `yaml:"url"`
		Alt string `yaml:"alt"`
	} `yaml:"logo"`
	Contact struct {
		Email string `yaml:"email"`
		Phone []string `yaml:"phone"`
		Address string `yaml:"address"`
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
	Links []Link `yaml:"links"`
}

type Configuration struct {
	App App `yaml:"app"`
	Users  []User `yaml:"users"`
	Groups []Group `yaml:"groups"`
}

type config struct {
	filename string
	data *Configuration
}

type ConfigManager interface {
	GetConfig() *Configuration
	Initiate() error
}

func NewConfig(filename string) ConfigManager {
	return &config{
		filename: filename,
		data:      nil,
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

func (c *config) Reload() {
	c.data, _ = c.read()
}

func (c *config) read() (*Configuration, error) {
	cfg, err := ioutil.ReadFile(c.filename)
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