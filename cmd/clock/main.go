package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	database "github.com/zohaib194/oblig2/database"
	"github.com/zohaib194/oblig2/fixer"
	types "github.com/zohaib194/oblig2/types"
	mgo "gopkg.in/mgo.v2"
)

type SlackPayload struct {
	Text string `json:"text"`
}

/*
This function runs once automatically every 24hours
InvokeWebhook take out all the payloads from WebhookCollection,
get the current rate according to a certain payloads base currency and target currency
and send a notification if current rate trigger min or max value of the payload
*/
func InvokeWebhook() {
	db := database.WebhookMongoDB{
		DatabaseURL:  "mongodb://admin:admin@ds245805.mlab.com:45805/webhook",
		DatabaseName: "webhook",
		Collection:   "WebhookPayload",
	}

	var form types.Invoked
	var results []types.Subscriber

	//Connection to the database
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	count, err := session.DB(db.DatabaseName).C(db.Collection).Count()
	if err != nil {
		panic(err)
	}

	err = session.DB(db.DatabaseName).C(db.Collection).Find(nil).All(&results)
	if err != nil {
		panic(err)
	}

	//Update the payload with currect rate
	for i := 0; i < count; i++ {

		form.BaseCurrency = results[i].BaseCurrency
		form.TargetCurrency = results[i].TargetCurrency
		form.MinTriggerValue = results[i].MinTriggerValue
		form.MaxTriggerValue = results[i].MaxTriggerValue

		fixerURL := "http://api.fixer.io/latest?base=" + results[i].BaseCurrency
		f, ok := fix.GetFixer(fixerURL)

		if !ok {
			panic(err)
		}

		// Run through all the rates
		for key, value := range f.Rates {
			// Checks if key"currency" matches a target currency
			if key == form.TargetCurrency {
				form.CurrentRate = value

				if form.CurrentRate > form.MaxTriggerValue || form.CurrentRate < form.MinTriggerValue {

					if strings.Contains(results[i].WebhookURL, "slack") {

						var slack SlackPayload
						cr := strconv.FormatFloat(float64(form.CurrentRate), 'f', 3, 32)
						min := strconv.FormatFloat(float64(form.MinTriggerValue), 'f', 3, 32)
						max := strconv.FormatFloat(float64(form.MaxTriggerValue), 'f', 3, 32)

						slack.Text = "\nbaseCurrency: " + form.BaseCurrency + ",\ntargetCurrency: " + form.TargetCurrency + ",\ncurrentRate: " + cr + ",\nminTriggerValue: " + min + ",\nmaxTriggerValue: " + max

						postJSON, err := json.Marshal(slack)
						if err != nil {
							panic(err)
						}
						postContent := bytes.NewBuffer(postJSON)

						//Send notification to the webhookurl
						res, err := http.Post(results[i].WebhookURL, "application/json", postContent)
						if err != nil {
							panic(err)

						}
						//if recieved the 200 or 204 status code
						fmt.Printf("status: %s", res.Status)
					} else {
						//Trigger and send the notification
						postJSON, err := json.Marshal(form)
						if err != nil {
							panic(err)
						}
						postContent := bytes.NewBuffer(postJSON)

						//Send notification to the webhookurl
						res, err := http.Post(results[i].WebhookURL, "application/x-www-form-urlencoded", postContent)
						if err != nil {
							panic(err)
						}
						//if recieved the 200 or 204 status code
						fmt.Printf("status: %s", res.Status)
					}
				}
			}
		}
	}
}

func main() {
	for range time.NewTicker(24 * time.Hour).C {
		//call functions
		fix.LatestFixer()
		InvokeWebhook()
	}
}
