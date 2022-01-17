package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

type BinanceCurrencyResponse struct {
	Code          string      `json:"code"`
	Message       interface{} `json:"message"`
	MessageDetail interface{} `json:"messageDetail"`
	Data          []struct {
		Pair     string  `json:"pair"`
		Rate     float64 `json:"rate"`
		Symbol   string  `json:"symbol"`
		FullName string  `json:"fullName"`
		ImageURL string  `json:"imageUrl"`
	} `json:"data"`
	Success bool `json:"success"`
}

var cache *BinanceCurrencyResponse = new(BinanceCurrencyResponse)

func main() {

	appEnv := os.Getenv("APP_ENV")
	if appEnv != "production" {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	exitChan := make(chan bool)
	// fetchBinanceCurrency(c)
	go routine(exitChan)

	app := fiber.New()
	app.Use(cors.New())

	app.Get("/api/currency", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(cache)
	})

	port := 7000

	portStr := os.Getenv("PORT")
	portInt, err := strconv.Atoi(portStr)
	if err != nil {
		fmt.Println("Could not convert port to int")
	}
	port = portInt

	app.Listen(fmt.Sprintf(":%d", port))

}

func routine(exitChan chan bool) {
	for {
		select {
		case <-exitChan:
			return
		case <-time.After(getTimeRemaining()):
			go fetchBinanceCurrency()
		}
	}
}

func fetchBinanceCurrency() {
	c := colly.NewCollector()
	c.OnResponse(func(r *colly.Response) {
		bCurrency := BinanceCurrencyResponse{}
		bodyBytes := r.Body
		err := json.Unmarshal(bodyBytes, &bCurrency)
		if err != nil {
			fmt.Println(err)
		}
		cache = &bCurrency
		fmt.Println("got currency from binance")
	})

	c.Visit("https://www.binance.com/bapi/asset/v1/public/asset-service/product/currency")
}

func getTimeRemaining() time.Duration {
	fetchEveryInSecod := 5
	fetchEveryStr := os.Getenv("FETCH_EVERY_IN_SECOND")
	fmt.Println(fetchEveryStr)

	fetchEvrInt, err := strconv.ParseInt(fetchEveryStr, 0, 64)
	if err != nil {
		fmt.Println("Could not convert to int64")
	}
	fetchEveryInSecod = int(fetchEvrInt)

	now := time.Now()
	nextHour := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second()+fetchEveryInSecod, 0, now.Location())
	timeToGo := nextHour.Sub(now)
	fmt.Printf("timeToGo: %v\n", timeToGo)
	return timeToGo
}
