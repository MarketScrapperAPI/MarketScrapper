package crawler

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/MarketScrapperAPI/MarketScrapper/models"
	"github.com/go-redis/redis/v8"
	"github.com/gocolly/colly"
)

type ContinenteCrawler struct {
	queueClient *redis.Client
	collector   *colly.Collector
	channel     chan<- models.CrawlerMessage
	options     *models.Options
	Control     models.CrawlerControl
}

var ContinentOptions = models.Options{
	Id:          "Continente",
	Delay:       time.Millisecond,
	StartingUrl: "www.continente.pt",
}

func NewContinenteCrawler(queueClient *redis.Client, options *models.Options, crawlerChan chan<- models.CrawlerMessage) ContinenteCrawler {
	allowedDomains := append(options.AllowedDomains, options.StartingUrl)
	c := colly.NewCollector(
		colly.AllowedDomains(allowedDomains...),
		colly.Async(true),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: options.Delay,
	})

	return ContinenteCrawler{
		queueClient: queueClient,
		collector:   c,
		channel:     crawlerChan,
		options:     options,
		Control: models.CrawlerControl{
			Id:          options.Id,
			Running:     false,
			Repeat:      false,
			ScrappedAmt: 0,
			StartedAt:   time.Time{},
		},
	}
}

func (c *ContinenteCrawler) Crawl() error {
	//log.Println("Crawler started on:", c.options.StartingUrl)

	// Find and print all links
	c.collector.OnHTML("div.product-wrapper", func(e *colly.HTMLElement) {
		n := e.ChildText("h1.product-name")          // product name
		b := e.ChildText("a.ct-pdp--brand")          // product brand
		p := e.ChildText("span.ct-pdp--unit")        // product packaging
		vu := e.ChildText("span.ct-price-formatted") // price per unit
		uq := e.ChildText("span.ct-m-unit")          // quantity unit
		vq := e.ChildText("span.ct-price-value")     // price per quantity
		//u := e.ChildText("div.ct-tile--price-secondary.ct-m-unit") //quantity unit
		iu := e.ChildAttr("img.ct-product-image", "src") // image url

		// build model
		unitValue := strings.SplitAfter(vu, "€")
		unValue := strings.Replace(unitValue[1], ".", "", -1)
		unValue = strings.Replace(unValue, ",", ".", -1)
		pricePerUnit, err := strconv.ParseFloat(unValue, 32)
		if err != nil {
			log.Println(err)
		}

		var pricePerQtt *float32 = nil
		if vq != "" {
			quantityString := strings.SplitAfter(vq, "€")
			quantityStringValue := strings.Replace(quantityString[1], ",", ".", -1)
			pricePerQuantity64, err := strconv.ParseFloat(quantityStringValue, 32)
			if err != nil {
				log.Println(err)
			}

			pricePerQuantity32 := float32(pricePerQuantity64)
			pricePerQtt = &pricePerQuantity32
		}

		var unitQtt *string = nil
		if uq != "" {
			unitQuantityString := strings.SplitAfter(uq, "/")
			unitQuantityStringValue := unitQuantityString[len(unitQuantityString)-1]
			unitQtt = &unitQuantityStringValue
		}

		message := models.Message{
			Item: models.Item{
				Name:             n,
				Brand:            b,
				Package:          p,
				PricePerItem:     float32(pricePerUnit),
				PricePerQuantity: pricePerQtt,
				QuantityUnit:     unitQtt,
				Url:              e.Request.URL.String(),
				ImageUrl:         iu,
			},
			Market: models.Market{
				Name:     c.options.Id,
				Location: "Online",
			},
		}

		// send with redis
		u, err := json.Marshal(message)
		if err != nil {
			log.Println(err)
		}
		msg := string(string(u))
		c.queueClient.Publish(context.Background(), "items", msg)
		c.Control.ScrappedAmt = c.Control.ScrappedAmt + 1
		//log.Println(msg)
	})

	// Callback for links on scraped pages
	c.collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		// Extract the linked URL from the anchor tag
		link := e.Attr("href")
		// Have our crawler visit the linked URL
		c.collector.Visit(e.Request.AbsoluteURL(link))
	})

	c.collector.OnRequest(func(r *colly.Request) {})

	c.collector.OnResponse(func(r *colly.Response) {
		//log.Printf("[%d] <- %s \n", r.StatusCode, r.Request.URL)
	})

	c.collector.Visit("https://" + c.options.StartingUrl)
	c.collector.Wait()
	c.channel <- models.CrawlerMessage{Id: c.options.Id, Status: "Done"}
	return nil
}

func (c *ContinenteCrawler) GetControls() *models.CrawlerControl {
	return &c.Control
}
