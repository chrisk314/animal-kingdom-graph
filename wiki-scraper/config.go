package main

import (
	"github.com/spf13/viper"
)

const (
	PhylumCollName  = "pyhlum"
	ClassCollName   = "class"
	OrderCollName   = "order"
	FamilyCollName  = "family"
	GenusCollName   = "genus"
	SpeciesCollName = "species"
)

type Config struct {
	CrawlerSeedURL             string `mapstructure:"CRAWLER_SEED_URL"`
	CrawlerAllowedDomain       string `mapstructure:"CRAWLER_ALLOWED_DOMAIN"`
	CrawlerRegexURLWikiNoFiles string `mapstructure:"CRAWLER_REGEX_URL_WIKI_NO_FILES"`
	CrawlerMaxTreeDepth        int    `mapstructure:"CRAWLER_MAX_TREE_DEPTH"`
	CrawlerAsync               bool   `mapstructure:"CRAWLER_ASYNC"`
	CrawlerParallelism         int    `mapstructure:"CRAWLER_PARALLELISM"`

	DatabaseUrl      string `mapstructure:"DATABASE_URL"`
	DatabaseUser     string `mapstructure:"DATABASE_USER"`
	DatabasePassword string `mapstructure:"DATABASE_PASSWORD"`
	DatabaseName     string `mapstructure:"DATABASE_NAME"`
}

// LoadConfig loads the config from the given path.
func LoadConfig(path string) (config Config, err error) {
	viper.SetConfigFile(path)
	viper.AutomaticEnv()
	err = viper.ReadInConfig()
	if err != nil {
		return
	}
	err = viper.Unmarshal(&config)
	return
}
