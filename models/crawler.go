package models

import (
	"time"
)

type ICrawler interface {
	Crawl() error
	GetControls() *CrawlerControl
}

type Options struct {
	Id             string
	Delay          time.Duration
	StartingUrl    string
	AllowedDomains []string
}

type CrawlerMessage struct {
	Id     string
	Status string
}

type CrawlerControl struct {
	Id      string
	Running bool
	Repeat  bool
}
