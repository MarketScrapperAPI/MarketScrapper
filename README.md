# MarketScrapper
Web crawlers with pre set configurations to scrap prices of Supermarket websites


### Monitor Docker Compose Redis
To validate incoming notifications we can use:
```shell
docker exec -it marketscrapper_queue_1 redis-cli -h localhost subscribe items
```