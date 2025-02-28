package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

type ExchangeRates struct {
	TimeNextUpdate string             `json:"time_next_update_utc"`
	Rates          map[string]float64 `json:"conversion_rates"`
	Expire         time.Time
}

type CachedRates struct {
	sync.RWMutex
	Data map[string]*ExchangeRates
}

var cache = CachedRates{
	Data: make(map[string]*ExchangeRates),
}

func main() {
	for {
		fromCurrency, toCurrency, amount := getInput()
		apiKey := getAPI()
		rates := getRates(apiKey, fromCurrency)
		currencyCheck(rates, toCurrency)
		convertPrint(fromCurrency, toCurrency, rates.Rates[toCurrency], amount)
	}
}

// Считывание данных
func getInput() (string, string, float64) {
	var (
		amount                   float64
		fromCurrency, toCurrency string
	)

	fmt.Print("Исходная валюта: ")
	if _, err := fmt.Scan(&fromCurrency); err != nil {
		fmt.Println("Введена некорректная исходная валюта")
		os.Exit(1)
	}

	fmt.Print("Целевая валюта: ")
	if _, err := fmt.Scan(&toCurrency); err != nil {
		fmt.Println("Введена некорректная целевая валюта")
		os.Exit(1)
	}

	fmt.Print("Сумма для конвертации: ")
	if _, err := fmt.Scan(&amount); err != nil {
		fmt.Println("Введена некорректная сумма для конвертации")
		os.Exit(1)
	}
	if amount <= 0 {
		fmt.Println("Сумма должна быть положительной")
		os.Exit(1)
	}

	return strings.ToUpper(fromCurrency), strings.ToUpper(toCurrency), amount
}

// Получение API ключа из файла
func getAPI() string {
	if err := godotenv.Load(".env"); err != nil {
		fmt.Println("Файл .env не найден:", err)
		os.Exit(1)
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		fmt.Println("API-ключ не найден")
		os.Exit(1)
	}
	return apiKey
}

// Получение курса с кэшированием
func getRates(apiKey string, fromCurrency string) *ExchangeRates {
	cache.RLock()
	cacheFromCurrency, ok := cache.Data[fromCurrency]
	cache.RUnlock()

	if ok {
		if time.Now().UTC().Before(cacheFromCurrency.Expire) {
			fmt.Println("Данные считаны из кэша")
			return cacheFromCurrency
		}
	}

	exchangeUrl := fmt.Sprintf("https://v6.exchangerate-api.com/v6/%s/latest/%s", apiKey, fromCurrency)

	exchangeResp, err := http.Get(exchangeUrl)
	if err != nil {
		fmt.Println("Ошибка:", err)
		os.Exit(1)
	}
	defer exchangeResp.Body.Close()

	if exchangeResp.StatusCode != http.StatusOK {
		fmt.Println("Ошибка HTTP запроса")
		os.Exit(1)
	}

	exchangeBody, err := io.ReadAll(exchangeResp.Body)
	if err != nil {
		fmt.Println("Ошибка:", err)
		os.Exit(1)
	}

	var rates ExchangeRates

	if err := json.Unmarshal(exchangeBody, &rates); err != nil {
		fmt.Println("Введена некорректная исходная валюта")
		os.Exit(1)
	}

	if rates.Rates == nil {
		fmt.Println("Ошибка: неверный API-ключ или некорректный ответ от сервера")
		os.Exit(1)
	}

	expireTime, err := time.Parse(time.RFC1123Z, rates.TimeNextUpdate)
	if err != nil {
		fmt.Println("Ошибка парсинга времени:", err)
	} else {
		rates.Expire = expireTime
		cache.Lock()
		defer cache.Unlock()
		cache.Data[fromCurrency] = &rates
		fmt.Println("Данные закэшированы")
	}

	return cache.Data[fromCurrency]
}

// Проверка наличия валюты в мапе
func currencyCheck(rates *ExchangeRates, toCurrency string) {
	_, ok := rates.Rates[toCurrency]
	if !ok {
		fmt.Println("Целевая валюта не найдена")
		os.Exit(1)
	}
}

// Конвертация и вывод
func convertPrint(from string, to string, rate float64, amount float64) {
	result := amount * rate
	fmt.Printf("%.2f %s = %.2f %s\n", amount, from, result, to)
}
