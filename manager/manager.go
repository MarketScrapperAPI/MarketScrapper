package manager

import (
	"errors"
	"fmt"
	"strconv"
	"time"

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
func (m *Manager) StartCrawlerByIdInBackground(Id string) error {
	if _, ok := m.crawlers[Id]; ok {
		ctrl := m.crawlers[Id].GetControls()
		ctrl.Running = true
		if ctrl.StartedAt.IsZero() {
			ctrl.StartedAt = time.Now()
		}
		go m.crawlers[Id].Crawl()
		fmt.Println("Started WebCrawler with Id:", Id)
	} else {
		return errors.New("id does not exist")
	}
	return nil
}

func (m *Manager) StartCrawlerById(Id string) error {
	if _, ok := m.crawlers[Id]; ok {
		m.crawlers[Id].GetControls().Running = true
		fmt.Println("Started WebCrawler with Id:", Id)
		m.crawlers[Id].Crawl()
	} else {
		return errors.New("id does not exist")
	}
	return nil
}

// Set repeat by Id
func (m *Manager) SetRepeatById(Id string, repeat bool) error {
	if _, ok := m.crawlers[Id]; ok {
		m.crawlers[Id].GetControls().Repeat = repeat
	} else {
		return errors.New("id does not exist")
	}
	return nil
}

// Start all
func (m *Manager) StartAllCrawlers() {
	for k, v := range m.crawlers {
		if !v.GetControls().Running {
			m.StartCrawlerByIdInBackground(k)
		}
	}
}

// List all Crawlers
func (m *Manager) ListCrawlers() error {
	fmt.Println("Crawlers:")
	for _, v := range m.crawlers {
		ctrl := v.GetControls()

		strAt := ctrl.StartedAt.Format("15:04:05.000")
		elapsed := ctrl.StartedAt.Format("15:04:05.000")
		if ctrl.Running {
			elapsed = time.Since(ctrl.StartedAt).String()
		}

		fmt.Printf("Id: %s 	Running: %s	Repeat: %s	Scrapped Items: %d	StartedAt: %s	Elapsed Time: %s \n",
			ctrl.Id,
			strconv.FormatBool(ctrl.Running),
			strconv.FormatBool(ctrl.Repeat),
			ctrl.ScrappedAmt,
			strAt,
			elapsed)
	}
	return nil
}

// List all Crawlers
func (m *Manager) CrawlersStatistics() error {
	fmt.Println("Crawlers:")
	for _, v := range m.crawlers {
		ctrl := v.GetControls()

		startedAt := ctrl.StartedAt.Format("15:04:05.000")
		elapsedTime := ctrl.StartedAt.Format("15:04:05.000")
		var scrapsPerSecond float32 = 0.0
		if ctrl.Running {
			elapsedTimeDuration := time.Since(ctrl.StartedAt)
			scrapsPerSecond = float32(ctrl.ScrappedAmt) / float32(elapsedTimeDuration/time.Second)
			elapsedTime = elapsedTimeDuration.String()
		}

		fmt.Printf("Id: %s	Scrapped Items: %d	StartedAt: %s	Elapsed Time: %s Scraps/second: %f\n",
			ctrl.Id,
			ctrl.ScrappedAmt,
			startedAt,
			elapsedTime,
			scrapsPerSecond)
	}
	return nil
}
