package exrates

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

func (api *API) GetAndImport() {

	for {
		// получение данных коллекции
		collection := api.Mongo.DB("steam").C("rates")
		// Текущий timestamp
		api.timeBorder = time.Now().Unix()
		// Получение цен с сайта bitcoinaverage
		tcs, err := api.GetRates()
		if err != nil {
			panic(err)
		}

		// Конвертирование слайса в интерфейс https://golang.org/doc/faq#convert_slice_of_interface
		s := make([]interface{}, 0, len(tcs))
		for i, v := range tcs {
			v.Pair = i
			s = append(s, v.PairData)
		}

		// Сохранение данных в базу
		err = collection.Insert(s...)
		if err != nil {
			panic(err)
		}

		// Удаление старых записей
		err = api.RemoveOld("rates")
		if err != nil {
			panic(err)
		}

		time.Sleep(api.Expiration)
	}

}

// RemoveOld Удаляет неактуальные записи из таблиц,
// полученные до скачивания новых.
func (api *API) RemoveOld(coll string) error {
	collection := api.Mongo.DB("steam").C(coll)
	_, err := collection.RemoveAll(bson.M{"addingtime": bson.M{"$lt": api.timeBorder}})
	if err != nil {
		panic(err)
	}
	return err
}
