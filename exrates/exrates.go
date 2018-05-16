package exrates

import (
	"io/ioutil"
	"net/http"
	"time"

	mgo "gopkg.in/mgo.v2"
)

// API основная структура данных связанная с функционалом bitcoinaverage
type API struct {
	Currency   string
	Mongo      *mgo.Session
	Expiration time.Duration
	timeBorder int64
}

// NewAPI конструктор для API
// TODO: создать возможность пользователю указывать параметры базовой валюты
func NewAPI(expirationTime int) *API {
	return &API{"USD", nil, time.Duration(expirationTime) * time.Second, 0}
}

// sendRequest подготавливает и отправляет запрос на серверы API
// и возвращает тело ответа
func (api *API) sendRequest() ([]byte, error) {

	req, err := api.prepareRequest()
	if err != nil {
		panic(err)
	}

	resp, err := api.send(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	return body, err
}

// prepareRequest Вспомогательная функция подготовки запроса для отправки на сервер API
func (api *API) prepareRequest() (*http.Request, error) {
	requestString := "https://apiv2.bitcoinaverage.com/indices/global/ticker/all?fiat=USD"
	req, err := http.NewRequest("GET", requestString, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("X-testing", "testing")
	return req, err
}

// send непосредственно отправляет запрос на сервер API и возвращает ответ
func (api *API) send(req *http.Request) (*http.Response, error) {

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	return resp, err
}

// Start основная функция запуска микросервиса, связанного с получением курсов валют.
// TODO: сделать кастомизацию настроек NATS и REDIS пользователем (избавиться от хардкода)
func (api *API) Start() {

	// Коннект к mongo
	// TODO: создать возможность пользователю самому задвать настройки соединения
	session, err := mgo.Dial("127.0.0.1")
	if err != nil {
		panic(err)
	}
	api.Mongo = session

	// Запуск регулярного даных с сайта bitcoinaverage и записи в базу
	api.GetAndImport()
}
