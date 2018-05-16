package endpoint

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/nats-io/go-nats"
)

type API struct {
	Agents        int
	AgentsCounter int
	NATS          *nats.Conn
}

// NewAPI конструктор для API
// Для создания, требуется указать количество агентов
func NewAPI(agents int) *API {
	return &API{agents, 0, nil}
}

func (api *API) handler(w http.ResponseWriter, r *http.Request) {

	// Переключаем по кругу идентификатор очереди
	if api.AgentsCounter == api.Agents {
		api.AgentsCounter = 0
	}
	api.AgentsCounter++

	appids, ok := r.URL.Query()["appid"]

	if !ok || len(appids) < 1 {
		log.Println("Url Param 'id' is missing")
		return
	}

	appid := appids[0]

	msg, _ := api.NATS.Request(strconv.Itoa(api.AgentsCounter), []byte(appid), 10*time.Second)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(msg.Data)

}

func (api *API) Start() {
	// Коннект к NATS серверу
	ns, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		panic(err)
	}
	api.NATS = ns

	http.HandleFunc("/", api.handler)

	fmt.Println("starting server at :8080")
	http.ListenAndServe(":8080", nil)
}
