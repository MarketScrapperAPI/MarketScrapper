package crawler

import (
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gocolly/colly"
)

var PingoDoceOptions = Options{
	Delay:       time.Millisecond,
	StartingUrl: "www.mercadao.pt/store/pingo-doce",
}

type PingoDoceCrawler struct {
	queueClient *redis.Client
	collector   *colly.Collector
	options     *Options
}

func NewPingoDoceCrawler(queueClient *redis.Client, options *Options) PingoDoceCrawler {
	allowedDomains := append(options.AllowedDomains, options.StartingUrl)
	c := colly.NewCollector(
		colly.AllowedDomains(allowedDomains...),
		colly.Async(true),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: options.Delay,
	})

	return PingoDoceCrawler{
		queueClient: queueClient,
		collector:   c,
		options:     options,
	}
}

func (c *PingoDoceCrawler) Crawl() error {

	log.Println("Crawler started on:", c.options.StartingUrl)

	// Find and print all links
	c.collector.OnHTML("div._1fWZPnanFwvNlG_yfYe7z6", func(e *colly.HTMLElement) {
		n := e.ChildText("h2._3MDF8HVHJABdafDgo7eFwa") // product name
		b := e.ChildText("p._2liXNl7HBoC31Y08witUtG")  // product brand
		// p := e.ChildText("span.ct-pdp--unit")          // product packaging
		vu := e.ChildText("span.pdo-inline-block") // price per unit
		// uq := e.ChildText("span.ct-m-unit")              // quantity unit
		// vq := e.ChildText("span.ct-price-value")         // price per quantity
		vq := e.ChildText("p._1S9KERJEEqKNCMJcYbWg63")   // packaging / price per quantity/ quantity unit
		iu := e.ChildAttr("img.ct-product-image", "src") // image url

		log.Printf("%s : %s : %s : %s : %s", n, b, vu, vq, iu)
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

	c.collector.Visit("https://" + c.options.StartingUrl)
	c.collector.Wait()
	return nil
}
