package agent

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

// GetAndImport Периодически получает данные о товарах и сохраняет в бд
func (api *API) GetAndImport() {
	// Подключение к бд
	session, err := api.mongoConnect()
	if err != nil {
		panic(err)
	}
	api.Mongo = session

	for {
		collection := api.Mongo.DB("steam").C("products")
		api.timeBorder = time.Now().Unix()
		tcs, err := api.GetProducts()
		if err != nil {
			panic(err)
		}
		list := tcs.Applist.Apps

		// Конвертирование слайса в интерфейс https://golang.org/doc/faq#convert_slice_of_interface
		s := make([]interface{}, len(list))
		for i, v := range list {
			s[i] = v.AppsData
		}

		// Сохранение данных в базу
		err = collection.Insert(s...)
		if err != nil {
			panic(err)
		}

		// Удаление старых записей
		err = api.RemoveOld("products")
		if err != nil {
			panic(err)
		}

		time.Sleep(api.Expiration)
	}

}

// RemoveOld Удаляет неактуальные записи из таблиц
func (api *API) RemoveOld(coll string) error {
	collection := api.Mongo.DB("steam").C(coll)
	_, err := collection.RemoveAll(bson.M{"addingtime": bson.M{"$lt": api.timeBorder}})
	if err != nil {
		panic(err)
	}
	return err
}
