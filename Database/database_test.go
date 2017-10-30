package database

import (
	"testing"

	"gopkg.in/mgo.v2"
)

func setupDB(t *testing.T) *WebhookMongoDB {
	db := WebhookMongoDB{
		DatabaseURL:  "mongodb://localhost",
		DatabaseName: "testPayload",
		Collection:   "payload",
	}

	session, err := mgo.Dial(db.DatabaseURL)
	defer session.Close()
	if err != nil {
		t.Error(err)
	}
	return &db
}

func dropDB(t *testing.T, db *WebhookMongoDB) {
	session, err := mgo.Dial(db.DatabaseURL)
	defer session.Close()
	if err != nil {
		t.Error(err)
	}
	err = session.DB(db.DatabaseName).DropDatabase()
	if err != nil {
		t.Error(err)
	}

}

func TestPayloadMongoDB_Add(t *testing.T) {
	db := setupDB(t)
	defer dropDB(t, db)

	db.Init()
	if db.Count() != 0 {
		t.Error("Database not properly initialized, Subsrcribe count should be 0")
	}
	subs := Subscriber{
		WebhookURL:      "http://remoteUrl:8080/randomWebhookPath",
		BaseCurrency:    "EUR",
		TargetCurrency:  "NOK",
		MinTriggerValue: 1.50,
		MaxTriggerValue: 2.55,
	}
	db.Add(subs)
	if db.Count() != 1 {
		t.Error("Adding new Subscriber failed.")
	}

}

func TestPayloadMongoDB_Get(t *testing.T) {
	db := setupDB(t)
	defer dropDB(t, db)

	db.Init()
	if db.Count() != 0 {
		t.Error("Database not properly initialized, Subscriber count should be 0")
	}
	subs := Subscriber{
		WebhookURL:      "http://remoteUrl:8080/randomWebhookPath",
		BaseCurrency:    "EUR",
		TargetCurrency:  "NOK",
		MinTriggerValue: 1.50,
		MaxTriggerValue: 2.55,
	}
	id, ok := db.Add(subs)

	if !ok {
		t.Error("Adding new Subscriber failed")
	}

	if db.Count() != 1 {
		t.Error("Adding new Subscriber failed.")
	}

	newPayload, ok := db.Get(id)
	if !ok {
		t.Error("couldn't find " + subs.WebhookURL)
	}

	if newPayload.WebhookURL != subs.WebhookURL ||
		newPayload.BaseCurrency != subs.BaseCurrency ||
		newPayload.TargetCurrency != subs.TargetCurrency ||
		newPayload.MaxTriggerValue != subs.MaxTriggerValue ||
		newPayload.MinTriggerValue != subs.MinTriggerValue {
		t.Error("Subscriber do not match")

	}
}

func TestPayloadMongoDB_Delete(t *testing.T) {
	db := setupDB(t)
	defer dropDB(t, db)

	db.Init()
	if db.Count() != 0 {
		t.Error("Database not properly initialized, Subscriber count should be 0")
	}
	subs := Subscriber{
		WebhookURL:      "http://remoteUrl:8080/randomWebhookPath",
		BaseCurrency:    "EUR",
		TargetCurrency:  "NOK",
		MinTriggerValue: 1.50,
		MaxTriggerValue: 2.55,
	}
	id, ok := db.Add(subs)
	if !ok {
		t.Error("Adding new Subscriber failed")
	}

	count := db.Count()

	ok = db.Delete(id)

	if !ok {
		t.Error("Deleting the Subscriber failed")
	}

	if db.Count() == count {
		t.Error("Deleting the Subscriber failed")
	}

}
