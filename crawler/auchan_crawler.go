package crawler

import (
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gocolly/colly"
)

type AuchanCrawler struct {
	queueClient *redis.Client
	collector   *colly.Collector
	options     *Options
}

var AuchantOptions = Options{
	Delay:       time.Millisecond,
	StartingUrl: "www.auchan.pt",
}

func NewAuchanCrawler(queueClient *redis.Client, options *Options) AuchanCrawler {
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
		options:     options,
	}
}

func (c *AuchanCrawler) Crawl() error {

	log.Println("Crawler started on:", c.options.StartingUrl)

	// Find and print all links
	c.collector.OnHTML("div.auc-pdp__header", func(e *colly.HTMLElement) {
		n := e.ChildText("h1.product-name") // product name

		log.Printf("product name: %s \n ", n)
	})
	c.collector.OnHTML("div.product-detail", func(e *colly.HTMLElement) {
		vu := e.ChildText("span.value") // price per unit
		//uq := e.ChildText("span.ct-m-unit")          // quantity unit
		vq := e.ChildText("span.auc-measures--price-per-unit") // price per quantity
		//u := e.ChildText("div.ct-tile--price-secondary.ct-m-unit") //quantity unit
		iu := e.ChildAttr("img.zoomImg", "src") // image url

		log.Printf("price: %s \n price per quantity: %s \n imageUrl: %s \n", vu, vq, iu)
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

	c.collector.Visit("https://" + c.options.StartingUrl)
	c.collector.Wait()
	return nil
}
