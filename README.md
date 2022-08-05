# MarketScrapper
Web crawlers with pre set configurations to scrap prices of Supermarket websites

# What is MarketScrapper
## Architecture
```mermaid
flowchart TD
A[Market1 Crawler] ----> E[queue];
B[Market2 Crawler] ----> E[queue];
C[Market3 Crawler] ----> E[queue];
D[Marketn Crawler] ----> E[queue];
E --> F[Item Handling Service];
```
## Summary
MarketScrapper is a predefined set of web crawlers/scrappers that collect article information form specific market websites.
This information is formated and published from the crawlers into a queue, that is then consumed by a Item Handling Service.

### Monitor Docker Compose Redis
To validate incoming notifications we can use:
```shell
docker exec -it marketscrapper_queue_1 redis-cli -h localhost subscribe items
```