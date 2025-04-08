package webapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
)

type Currency struct {
	Ccy  string `json:"Ccy"`
	Rate string `json:"Rate"`
}

func GetUSDCourse() (float64, error) {
	url := "https://cbu.uz/uz/arkhiv-kursov-valyut/json"

	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch data: %w", err)
	}
	defer resp.Body.Close()

	var currencies []Currency
	if err := json.NewDecoder(resp.Body).Decode(&currencies); err != nil {
		return 0, fmt.Errorf("failed to decode JSON: %w", err)
	}

	var usdRate float64
	for _, cur := range currencies {
		if cur.Ccy == "USD" {
			usdRate, err = strconv.ParseFloat(cur.Rate, 64)
			if err != nil {
				return 0, fmt.Errorf("failed to parse USD rate: %w", err)
			}
			break
		}
	}

	if usdRate == 0 {
		return 0, errors.New("USD rate not found")
	}

	usdRate = math.Round(usdRate/10) * 10

	return usdRate, nil
}
