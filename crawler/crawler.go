package crawler

import "time"

type ICrawler interface {
	Crawl() error
}

type Options struct {
	Delay time.Duration
}
