package crawler

import (
	"github.com/go-redis/redis"
	"github.com/gocolly/colly"
)

type GenericCrawler struct {
	queueClient *redis.Client
	collector   *colly.Collector
}

func NewGenericCrawler(queueClient *redis.Client, allowedDomains []string) GenericCrawler {
	c := colly.NewCollector(
		colly.AllowedDomains(allowedDomains...),
		colly.Async(true),
	)

	return GenericCrawler{
		queueClient: queueClient,
		collector:   c,
	}
}

func (c *GenericCrawler) Crawl() error {
	return nil
}
