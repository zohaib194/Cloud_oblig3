package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

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
	db := WebhookMongoDB{
		DatabaseURL:  "mongodb://localhost",
		DatabaseName: "Webhook",
		Collection:   "WebhookPayload",
	}

	var form Invoked
	var results []Subscriber

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
		f, ok := GetFixer(fixerURL)

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
	ticker := time.NewTicker(time.Second * 120)
	go func() {
		for t := range ticker.C {
			//call functions
			fmt.Printf("\n", t)
			InvokeWebhook()
			//GetFixerSevenDays(time.Now().AddDate(0, 0, -7), time.Now())
		}
	}()
	ticker.Stop()
}
