package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type ExchangeRates struct {
	Result string             `json:"result"`
	Error  string             `json:"error-type"`
	Code   string             `json:"base_code"`
	Time   string             `json:"time_next_update_utc"`
	Rates  map[string]float64 `json:"conversion_rates"`
}

func main() {
	fromCurrency, toCurrency, amount := getInput()
	apiKey := getAPI()
	rates := getRates(apiKey, fromCurrency)
	currencyCheck(rates, toCurrency)
	convertPrint(fromCurrency, toCurrency, rates.Rates[toCurrency], amount)
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

// Получение курса
func getRates(apiKey string, fromCurrency string) *ExchangeRates {
	exchangeUrl := fmt.Sprintf("https://v6.exchangerate-api.com/v6/%s/latest/%s", apiKey, fromCurrency)

	exchangeResp, err := http.Get(exchangeUrl)
	if err != nil {
		fmt.Println("Ошибка:", err)
		os.Exit(1)
	}
	defer exchangeResp.Body.Close()

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

	if rates.Result != "success" {
		fmt.Println("Ошибка API запроса:", rates.Error)
		os.Exit(1)
	}

	if rates.Rates == nil {
		fmt.Println("Ошибка: неверный API-ключ или некорректный ответ от сервера")
		os.Exit(1)
	}

	return &rates
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
	fmt.Printf("%.2f %s = %.2f %s", amount, from, result, to)
}
