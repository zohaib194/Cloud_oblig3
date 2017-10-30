package types

import "gopkg.in/mgo.v2/bson"

// Webhook mongodb stores the details of the DB connection.
type WebhookMongoDB struct {
	DatabaseURL  string
	DatabaseName string
	Collection   string
}

type Subscriber struct {
	//ID              bson.ObjectId `json:"_id, omitempty" bson:"_id"`
	WebhookURL      string  `json:"webhookurl" bson:"webhookurl"`
	BaseCurrency    string  `json:"basecurrency" bson:"basecurrency"`
	TargetCurrency  string  `json:"targetcurrency" bson:"targetcurrency"`
	MinTriggerValue float32 `json:"mintriggervalue" bson:"mintriggervalue"`
	MaxTriggerValue float32 `json:"maxtriggervalue" bson:"maxtriggervalue"`
}
type Id struct {
	ID bson.ObjectId `bson:"_id"`
}
type Invoked struct {
	BaseCurrency    string  `json:"basecurrency"`
	TargetCurrency  string  `json:"targetcurrency"`
	CurrentRate     float32 `json:"currentrate"`
	MinTriggerValue float32 `json:"mintriggervalue"`
	MaxTriggerValue float32 `json:"maxtriggervalue"`
}

type Fixer struct {
	Base  string             `json:"base"`
	Date  string             `json:"date"`
	Rates map[string]float32 `json:"rates"`
}

type Latest struct {
	BaseCurrency   string `json:"basecurrency"`
	TargetCurrency string `json:"targetcurrency"`
}

// SlackPayload (This payload is used if a webhook is from Slack)
type SlackPayload struct {
	Text string `json:"text"`
}
