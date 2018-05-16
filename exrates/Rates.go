package exrates

import (
	"encoding/json"
	"time"
)

// Rates - структура для получения списка курсов валют
type Rates map[string]PairDataWrapper

type PairDataWrapper struct {
	PairData
}

type PairData struct {
	Pair             string
	AddingTime       int64
	Ask              float64      `json:"ask"`
	Bid              float64      `json:"bid"`
	Last             float64      `json:"last"`
	High             float64      `json:"high"`
	Low              float64      `json:"low"`
	Open             OpenData     `json:"open"`
	Averages         AveragesData `json:"averages"`
	Volume           float64      `json:"volume"`
	Changes          ChangesData  `json:"changes"`
	VolumePercent    float64      `json:"volume_percent"`
	Timestamp        uint64       `json:"timestamp"`
	DisplayTimestamp string       `json:"display_timestamp"`
}

type OpenData struct {
	Hour   ffloat64 `json:"hour"`
	Day    ffloat64 `json:"day"`
	Week   ffloat64 `json:"week"`
	Month  ffloat64 `json:"month"`
	Month3 ffloat64 `json:"month_3"`
	Month6 ffloat64 `json:"month_6"`
	Year   ffloat64 `json:"year"`
}

type AveragesData struct {
	Day   ffloat64 `json:"day"`
	Week  ffloat64 `json:"week"`
	Month ffloat64 `json:"month"`
}

type ChangesData struct {
	Price   PriceData   `json:"price"`
	Percent PercentData `json:"percent"`
}

type PriceData struct {
	Hour   ffloat64 `json:"hour"`
	Day    ffloat64 `json:"day"`
	Week   ffloat64 `json:"week"`
	Month  ffloat64 `json:"month"`
	Month3 ffloat64 `json:"month_3"`
	Month6 ffloat64 `json:"month_6"`
	Year   ffloat64 `json:"year"`
}

type PercentData struct {
	Hour   ffloat64 `json:"hour"`
	Day    ffloat64 `json:"day"`
	Week   ffloat64 `json:"week"`
	Month  ffloat64 `json:"month"`
	Month3 ffloat64 `json:"month_3"`
	Month6 ffloat64 `json:"month_6"`
	Year   ffloat64 `json:"year"`
}

// GetRates получает с сайта bitcoinaverage список валют и возвращает полученные данные
func (api *API) GetRates() (Rates, error) {

	body, err := api.sendRequest()
	if err != nil {
		panic(err)
	}
	response := make(Rates)
	err = json.Unmarshal(body, &response)
	if err != nil {
		panic(err)
	}

	return response, err
}

// Кастомный анмаршаллер, распознавание вариативного типпа данных в JSON
type ffloat64 float64

func (ff *ffloat64) UnmarshalJSON(b []byte) error {
	if b[0] != '"' {
		return json.Unmarshal(b, (*float64)(ff))
	}

	*ff = ffloat64(0)
	return nil
}

// Кастомный анмаршаллер, добавление AddingTime в структуру для сохранения в бд
func (a *PairDataWrapper) UnmarshalJSON(b []byte) error {
	if err := json.Unmarshal(b, &a.PairData); err != nil {
		return err
	}
	a.AddingTime = time.Now().Unix()
	return nil
}
