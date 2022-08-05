package manager

import (
	"fmt"
	"strconv"

	"github.com/MrBolas/MarketScrapper/crawler"
	"github.com/MrBolas/MarketScrapper/models"
	"github.com/go-redis/redis/v8"
)

type Manager struct {
	crawlers map[string]models.ICrawler
	channel  chan models.CrawlerMessage
}

func NewCrawlerManager(queueClient *redis.Client) Manager {

	// Create crawler channel
	crawlerChan := make(chan models.CrawlerMessage)

	// Instatiate Crawlers
	c := crawler.NewContinenteCrawler(queueClient, &crawler.ContinentOptions, crawlerChan)
	a := crawler.NewAuchanCrawler(queueClient, &crawler.AuchantOptions, crawlerChan)

	mp := make(map[string]models.ICrawler)
	mp[a.Control.Id] = &a
	mp[c.Control.Id] = &c

	return Manager{
		crawlers: mp,
		channel:  crawlerChan,
	}
}

// Launch Manager: go routine
func (m *Manager) StartManager() error {
	for {
		// restarts crawler when it's done
		select {
		case res := <-m.channel:
			if res.Status == "Done" {
				if m.crawlers[res.Id].GetControls().Repeat {
					m.crawlers[res.Id].Crawl()
				} else {
					m.crawlers[res.Id].GetControls().Running = false
				}
			}
		}
	}
}

// Start by Id
func (m *Manager) StartCrawlerById(Id string) {
	m.crawlers[Id].GetControls().Running = true
	go m.crawlers[Id].Crawl()
	fmt.Println("Started WebCrawler with Id:", Id)
}

// Set repeat by Id
func (m *Manager) SetRepeatById(Id string, repeat bool) {
	m.crawlers[Id].GetControls().Repeat = repeat
}

// Start all
func (m *Manager) StartAllCrawlers() {
	for k, _ := range m.crawlers {
		m.StartCrawlerById(k)
	}
}

// List all Crawlers
func (m *Manager) ListCrawlers() error {
	fmt.Println("Crawlers:")
	for _, v := range m.crawlers {
		fmt.Printf("Id: %s 		Running: %s		Repeat: %s\n", v.GetControls().Id, strconv.FormatBool(v.GetControls().Running), strconv.FormatBool(v.GetControls().Repeat))
	}
	return nil
}
