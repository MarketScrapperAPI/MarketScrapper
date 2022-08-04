package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/MrBolas/MarketScrapper/crawler"
	"github.com/go-redis/redis"
	"github.com/joho/godotenv"
)

const ENV_REDIS_HOST = "REDIS_HOST"
const ENV_REDIS_PORT = "REDIS_PORT"
const ENV_REDIS_DB = "REDIS_DB"
const ENV_REDIS_PASSWORD = "REDIS_PASSWORD"

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	redisHost := os.Getenv(ENV_REDIS_HOST)
	if redisHost == "" {
		panic("missing env var: " + ENV_REDIS_HOST)
	}
	redisPort := os.Getenv(ENV_REDIS_PORT)
	if redisPort == "" {
		panic("missing env var: " + ENV_REDIS_PORT)
	}
	redisDB := os.Getenv(ENV_REDIS_DB)
	if redisDB == "" {
		panic("missing env var: " + ENV_REDIS_DB)
	}

	dBNumber, err := strconv.Atoi(redisDB)
	if err != nil {
		panic("invalid Redis DB number: " + redisDB)
	}

	// Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisHost + ":" + redisPort,
		Password: "",
		DB:       dBNumber,
	})

	// CrawlerOptions
	options := &crawler.Options{
		Delay: time.Millisecond,
	}

	ccrawler := crawler.NewContinenteCrawler(rdb, []string{}, options)
	ccrawler.Crawl()
}
