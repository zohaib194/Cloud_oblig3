package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	types "github.com/zohaib194/oblig2"
	database "github.com/zohaib194/oblig2/Database"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
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
		DatabaseURL:  "mongodb://<Webhook>:<123456789>@ds241065.mlab.com:41065/webhook",
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

func LatestFixer() {
	//Send request to Fixer.io
	fixerURL := "http://api.fixer.io/latest?base=EUR"
	f, ok := GetFixer(fixerURL)
	if !ok {
		fmt.Print("latestFixer()")
	}
	f.Date = time.Now().Format("2006-01-02")
	SaveFixer(f)
}

/*
	Get the json from Fixer.io
*/
func GetFixer(url string) (*types.Fixer, bool) {
	var f *types.Fixer

	res, err := http.Get(url)
	if err != nil {
		fmt.Printf(err.Error(), http.StatusBadRequest)
		return f, false
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		fmt.Printf(err.Error(), http.StatusNotFound)
		return f, false
	}
	err = json.Unmarshal(body, &f)
	if err != nil {
		fmt.Printf(err.Error(), http.StatusBadRequest)
		return f, false
	}
	return f, true
}

/*
	Save Fixer payload in the collection
*/
func SaveFixer(f *types.Fixer) bool {
	db := database.WebhookMongoDB{
		DatabaseURL:  "mongodb://<Webhook>:<123456789>@ds241065.mlab.com:41065/webhook",
		DatabaseName: "webhook",
		Collection:   "FixerPayload",
	}

	var found types.Fixer
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	err = session.DB(db.DatabaseName).C(db.Collection).Find(bson.M{"date": f.Date}).One(&found)
	if err != nil && err.Error() != "not found" {
		fmt.Printf("error finding existing doc in DB, %v", err.Error())
		return false
	} else if err != nil && err.Error() == "not found" {
		err2 := session.DB(db.DatabaseName).C(db.Collection).Insert(&f)
		if err2 != nil {
			fmt.Printf("error in SaveFixer(), %v", err2.Error())
			return false
		}
	} else {
		fmt.Print("Latest Fixer already exist in DB")
		return true
	}

	return true
}

func main() {
	for range time.NewTicker(1 * time.Second).C {
		//call functions
		LatestFixer()
		//InvokeWebhook()
	}
}
