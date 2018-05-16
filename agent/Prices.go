package agent

import (
	"encoding/json"
)

// Prices - структура данных для получаемой информации о свойствах товаров из API магазина Steam
type Prices map[string]PriceInfo

type PriceInfo struct {
	Success bool      `json:"success"`
	Data    PriceData `json:"data"`
}

type PriceData struct {
	Name          string            `json:"name"`
	PriceOverwiew PriceOverwiewData `json:"price_overview"`
}

type PriceOverwiewData struct {
	Currency        string      `json:"currency"`
	Initial         uint        `json:"initial"`
	Final           json.Number `json:"final,Number"`
	DiscountPercent float64     `json:"discount_percent"`
}

// GetPriceInfo получает информацию о ценах на товар магазина
func (api *API) GetPriceInfo() (Prices, error) {

	body, err := api.sendRequest("store")
	if err != nil {
		panic(err)
	}

	prices := make(Prices)

	err = json.Unmarshal(body, &prices)
	if err != nil {
		panic(err)
	}
	return prices, err
}

// GetPrice возвращает только значение цены из полученной информации
func (api *API) GetPrice() (string, error) {

	info, err := api.GetPriceInfo()

	return string(info[api.AppID].Data.PriceOverwiew.Final), err
}
