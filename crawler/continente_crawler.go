package crawler

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"

	"github.com/MrBolas/MarketScrapper/models"
	"github.com/go-redis/redis"
	"github.com/gocolly/colly"
)

type ContinenteCrawler struct {
	queueClient *redis.Client
	collector   *colly.Collector
	baseUrl     string
}

func NewContinenteCrawler(queueClient *redis.Client, allowedDomains []string, options *Options) ContinenteCrawler {
	continenteUrl := "www.continente.pt"
	allowedDomains = append(allowedDomains, continenteUrl)
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
		baseUrl:     continenteUrl,
	}
}

func (c *ContinenteCrawler) Crawl() error {

	// Find and print all links
	c.collector.OnHTML("div.product-name-details--wrapper", func(e *colly.HTMLElement) {
		n := e.ChildText("h1.product-name")          // product name
		b := e.ChildText("a.ct-pdp--brand")          // product brand
		p := e.ChildText("span.ct-pdp--unit")        // product packaging
		vu := e.ChildText("span.ct-price-formatted") // price per unit
		uq := e.ChildText("span.ct-m-unit")          // quantity unit
		vq := e.ChildText("span.ct-price-value")     // price per quantity
		//u := e.ChildText("div.ct-tile--price-secondary.ct-m-unit") //quantity unit

		// build model
		unitValue := strings.SplitAfter(vu, "€")
		unValue := strings.Replace(unitValue[1], ",", ".", -1)
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

		item := models.Item{
			Name:             n,
			Brand:            b,
			Package:          p,
			PricePerItem:     float32(pricePerUnit),
			PricePerQuantity: pricePerQtt,
			QuantityUnit:     unitQtt,
		}

		message := models.Message{
			Item: item,
			Market: models.Market{
				Name:     "Continente",
				Location: "Online",
			},
		}

		// send with redis
		u, err := json.Marshal(message)
		if err != nil {
			log.Println(err)
		}
		//c.queueClient.Publish("items", u)
		log.Println(string(u))
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
		log.Printf("[%d] <- %s \n", r.StatusCode, r.Request.URL)
	})

	c.collector.Visit("https://" + c.baseUrl)
	c.collector.Wait()
	return nil
}
