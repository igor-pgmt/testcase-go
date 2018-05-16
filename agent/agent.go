package agent

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/nats-io/go-nats"
	"gopkg.in/mgo.v2"
)

// API основная структура данных связанная с функционалом магазина Steam
type API struct {
	Currency   string
	AgentID    int
	AppID      string
	NATS       *nats.Conn
	Redis      *redis.Client
	Mongo      *mgo.Session
	Expiration time.Duration
	timeBorder int64
}

// NewAPI конструктор для API
// TODO: создать возможность пользователю указывать параметры
func NewAPI(agentID, expirationTime int) *API {
	return &API{"us", agentID, "", nil, nil, nil, time.Duration(expirationTime) * time.Second, 0}
}

// sendRequest подготавливает и отправляет запрос на серверы API
// и возвращает тело ответа
func (api *API) sendRequest(apiType string) ([]byte, error) {

	req, err := api.prepareRequest(apiType)
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
func (api *API) prepareRequest(apiType string) (*http.Request, error) {
	var requestString string
	values := url.Values{}

	switch apiType {
	case "list":
		requestString = "http://api.steampowered.com/ISteamApps/GetAppList/v0002/"
	case "store":
		requestString = "https://store.steampowered.com/api/appdetails?"
		values.Add("cc", api.Currency)
		values.Add("appids", api.AppID)
		requestString += values.Encode()
	}

	req, err := http.NewRequest("GET", requestString, nil)
	if err != nil {
		panic(err)
	}
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

// Start основная функция запуска микросервиса, связанного с магазином Steam.
// TODO: сделать кастомизацию настроек NATS и REDIS пользователем (избавиться от хардкода)
func (api *API) Start() {

	// Коннект к NATS
	ns, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		panic(err)
	}
	api.NATS = ns

	// Коннект к Redis
	// TODO: создать возможность пользователю самому задвать настройки соединения
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	pong, err := client.Ping().Result()
	log.Println(pong, err)
	if err != nil {
		panic(err)
	}

	api.Redis = client

	api.Mongo, err = api.mongoConnect()
	if err != nil {
		panic(err)
	}

	// Создание подписки на прослушивание NATS-сервера с ожиданием ID товара
	api.NATS.Subscribe(strconv.Itoa(api.AgentID), func(m *nats.Msg) {
		api.AppID = string(m.Data)
		log.Println(api.AppID)
		// Попытка получения данных о товаре из Redis
		fromRedis := api.Redis.HMGet("id."+api.AppID, "appid", "name", "price")
		result, err := fromRedis.Result()
		if err != nil {
			panic(err)
		}

		// Карта для полученных данных
		d := map[string]interface{}{}

		// Если данных в Redis не оказалось, агент должен их получить
		if result[0] == nil {

			priceInfo, err := api.GetPriceInfo()
			if err != nil {
				panic(err)
			}

			d["appid"] = api.AppID
			d["name"] = string(priceInfo[api.AppID].Data.Name)
			d["price"] = string(priceInfo[api.AppID].Data.PriceOverwiew.Final)

			// Если данные оказались в Redis, сохраняем их в карту
		} else {
			d["appid"] = result[0]
			d["name"] = result[1]
			d["price"] = result[2]
		}

		// Перевод данных в json для отправки серверу очередей
		b, err := json.Marshal(d)
		if err != nil {
			panic(err)
		}

		// Отправка данных на сервер очередей
		api.NATS.Publish(m.Reply, b)

		// Если данные получены с API магазина, сохраняем в Redis
		if result[0] == nil {
			err = api.Redis.HMSet("id."+api.AppID, d).Err()
			if err != nil {
				panic(err)
			}
		}

		// Назначаем срок хранения данных
		err = api.Redis.Expire("id."+api.AppID, api.Expiration).Err()
		if err != nil {
			panic(err)
		}
	})
	runtime.Goexit()
}

func (api *API) mongoConnect() (*mgo.Session, error) {
	// Коннект к mongo
	// TODO: создать возможность пользователю самому задвать настройки соединения
	session, err := mgo.Dial("127.0.0.1")

	return session, err
}
