package agent

import (
	"encoding/json"
	"time"
)

// Products - структура для получения списка товаров из магазина Steam
type Products struct {
	Applist ApplistData `json:"applist"`
}

type ApplistData struct {
	Apps []AppsDataWrapper `json:"apps"`
}

type AppsDataWrapper struct {
	AppsData `json:"apps"`
}

type AppsData struct {
	Appid      uint64 `json:"appid"`
	Name       string `json:"name"`
	AddingTime int64
}

// GetProducts получает данные обо всех товарах в магазине Steam
func (api *API) GetProducts() (*Products, error) {

	body, err := api.sendRequest("list")
	if err != nil {
		panic(err)
	}
	products := Products{}
	err = json.Unmarshal(body, &products)
	if err != nil {
		panic(err)
	}

	return &products, err
}

// UnmarshalJSON - Кастомный анмаршаллер, добавление AddingTime в структуру для сохранения в бд
func (a *AppsDataWrapper) UnmarshalJSON(b []byte) error {
	if err := json.Unmarshal(b, &a.AppsData); err != nil {
		return err
	}
	a.AddingTime = time.Now().Unix()
	return nil
}
