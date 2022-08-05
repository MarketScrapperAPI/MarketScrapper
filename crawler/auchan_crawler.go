package crawler

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/MrBolas/MarketScrapper/models"
	"github.com/go-redis/redis/v8"
	"github.com/gocolly/colly"
)

type AuchanCrawler struct {
	queueClient *redis.Client
	collector   *colly.Collector
	channel     chan<- models.CrawlerMessage
	options     *models.Options
	Control     models.CrawlerControl
}

var AuchantOptions = models.Options{
	Id:          "Auchan",
	Delay:       time.Millisecond,
	StartingUrl: "www.auchan.pt",
}

func NewAuchanCrawler(queueClient *redis.Client, options *models.Options, crawlerChan chan<- models.CrawlerMessage) AuchanCrawler {
	allowedDomains := append(options.AllowedDomains, options.StartingUrl)
	c := colly.NewCollector(
		colly.AllowedDomains(allowedDomains...),
		colly.Async(true),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: options.Delay,
	})

	return AuchanCrawler{
		queueClient: queueClient,
		collector:   c,
		channel:     crawlerChan,
		options:     options,
		Control: models.CrawlerControl{
			Id:      options.Id,
			Running: false,
			Repeat:  false,
		},
	}
}

func (c *AuchanCrawler) Crawl() error {
	log.Println("Crawler started on:", c.options.StartingUrl)

	// Find and print all links
	c.collector.OnHTML("div#maincontent", func(e *colly.HTMLElement) {

		n := e.ChildText("h1.product-name") // product name

		// validates if it's a product page
		if n == "" {
			return
		}

		vu := e.ChildText("span.sales > span.value") // price per unit
		//uq := e.ChildText("span.ct-m-unit")          // quantity unit
		vq := e.ChildText("span.auc-measures--price-per-unit") // price per quantity
		//u := e.ChildText("div.ct-tile--price-secondary.ct-m-unit") //quantity unit
		iu := e.ChildAttr("div:nth-child(1) > picture > img", "src") // image url

		// build model
		unitValue := strings.Trim(strings.Split(vu, "€")[0], " ")
		unValue := strings.Replace(unitValue, ",", ".", -1)
		pricePerUnit, err := strconv.ParseFloat(unValue, 32)
		if err != nil {
			log.Println(err)
		}

		var pricePerQtt *float32 = nil
		if vq != "" {
			quantityString := strings.Trim(strings.Split(vq, "€")[0], " ")
			//quantityStringValue := strings.Replace(quantityString[1], ",", ".", -1)
			pricePerQuantity64, err := strconv.ParseFloat(quantityString, 32)
			if err != nil {
				log.Println(err)
			}

			pricePerQuantity32 := float32(pricePerQuantity64)
			pricePerQtt = &pricePerQuantity32
		}

		var unitQtt *string = nil
		if vq != "" {
			unitQuantityString := strings.Split(vq, "/")
			unitQuantityStringValue := unitQuantityString[len(unitQuantityString)-1]
			unitQtt = &unitQuantityStringValue
		}

		message := models.Message{
			Item: models.Item{
				Name:             n,
				Brand:            "",
				Package:          "",
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
		//log.Println(msg)
	})

	// Callback for links on scraped pages
	c.collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		// Extract the linked URL from the anchor tag
		link := e.Attr("href")
		// Have our crawler visit the linked URL
		c.collector.Visit(e.Request.AbsoluteURL(link))
	})

	c.collector.OnHTML("div.product-tile", func(e *colly.HTMLElement) {
		// Extract the linked URL from the anchor tag
		itemId := e.Attr("data-pid")
		// Have our crawler visit the linked URL
		c.collector.Visit(e.Request.AbsoluteURL("https://" + c.options.StartingUrl + "/pt/" + itemId + ".html"))
	})

	c.collector.OnRequest(func(r *colly.Request) {})

	c.collector.OnResponse(func(r *colly.Response) {
		log.Printf("[%d] <- %s \n", r.StatusCode, r.Request.URL)
	})

	//c.collector.Visit("https://" + c.options.StartingUrl)
	c.collector.Visit("https://www.auchan.pt/")
	c.collector.Wait()
	c.channel <- models.CrawlerMessage{Id: c.options.Id, Status: "Done"}
	return nil
}

func (c *AuchanCrawler) GetControls() *models.CrawlerControl {
	return &c.Control
}
