package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/MrBolas/MarketScrapper/manager"
	"github.com/go-redis/redis/v8"
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

	interactiveMode := flag.Bool("i", false, "interactive mode")
	targetId := flag.String("t", "", "target id")
	flag.Parse()

	// Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisHost + ":" + redisPort,
		Password: "",
		DB:       dBNumber,
	})

	// Create Manager
	mng := manager.NewCrawlerManager(rdb)
	go mng.StartManager()

	if !*interactiveMode {
		if *targetId == "" {
			fmt.Println("please specify a target with -t <target>")
		}
		err = mng.StartCrawlerById(*targetId)
		if err != nil {
			fmt.Println(err)
		}
		return
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("----------Market Scrapper Shell-----------")

	for {
		fmt.Print("-> ")
		text, _ := reader.ReadString('\n')
		// convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)

		// help
		if strings.HasPrefix(text, "help") {
			fmt.Printf("Available commands:\n help -> shows available commands\n ls -> lists available crawlers\n stats -> lists statistics from crawlers\n repeat -> sets crawler to repeat after it's finished\n start -> starts crawler\n")
			continue
		}

		// list crawlers
		if strings.HasPrefix(text, "ls") {
			mng.ListCrawlers()
			continue
		}

		// list crawlers stats
		if strings.HasPrefix(text, "stats") {
			mng.CrawlersStatistics()
			continue
		}

		// repeat crawlers
		if strings.HasPrefix(text, "repeat") {
			args, err := SanitizeInputs(text)
			if err != nil {
				fmt.Println(err)
				continue
			}
			rpt, err := strconv.ParseBool(args[0])
			if err != nil {
				fmt.Println("repeat <true/false> <crawler Id>")
				continue
			}
			err = mng.SetRepeatById(args[1], rpt)
			if err != nil {
				fmt.Println(err)
				continue
			}
			continue
		}

		// Start crawlers
		if strings.HasPrefix(text, "start") {
			args, err := SanitizeInputs(text)
			if err != nil {
				fmt.Println(err)
				fmt.Printf("start <all/crawlerId>\n")
				continue
			}
			if args[0] == "all" {
				mng.StartAllCrawlers()
			} else {
				err = mng.StartCrawlerByIdInBackground(args[0])
				if err != nil {
					fmt.Println(err)
					continue
				}
			}
			continue
		}
		fmt.Println("Unknown command")
	}
}

func SanitizeInputs(command string) ([]string, error) {

	arguments := strings.Split(command, " ")
	if len(arguments) < 2 {
		return []string{}, errors.New("not enough arguments")
	}

	return arguments[1:], nil
}
