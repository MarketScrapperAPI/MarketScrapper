# MarketScrapper
Web crawlers with pre set configurations to scrap prices of Supermarket websites

# What is MarketScrapper
## Architecture
```mermaid
flowchart TD
A ----> B;
```
## Summary
MarketScrapper is a predefined set of web crawlers/scrappers that collect article information form specific market websites.

### Monitor Docker Compose Redis
To validate incoming notifications we can use:
```shell
docker exec -it marketscrapper_queue_1 redis-cli -h localhost subscribe items
```