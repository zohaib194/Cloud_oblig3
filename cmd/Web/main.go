package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	database "github.com/zohaib194/oblig2/Database"
	types "github.com/zohaib194/oblig2/types"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func postReqHandler(w http.ResponseWriter, r *http.Request) {
	db := database.WebhookMongoDB{
		DatabaseURL:  "mongodb://admin:admin@ds245805.mlab.com:45805/webhook",
		DatabaseName: "webhook",
		Collection:   "WebhookPayload",
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	defer r.Body.Close()
	var p types.Subscriber

	err = json.Unmarshal(body, &p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	if len(p.WebhookURL) != 0 && len(p.BaseCurrency) != 0 && len(p.TargetCurrency) != 0 && p.MaxTriggerValue > p.MinTriggerValue && p.MinTriggerValue < p.MaxTriggerValue {
		if !strings.Contains(p.BaseCurrency, "EUR") {
			http.Error(w, "Not implemented", http.StatusNotImplemented)
		} else {
			ok2 := validateCurrency(p.TargetCurrency)
			if !ok2 {
				log.Printf("Invalid currency: %v", p.TargetCurrency)
				http.Error(w, "Invalid target currency", http.StatusBadRequest)
			} else {
				db.Init()
				id, ok := db.Add(p)

				if !ok {
					http.Error(w, "Not found in database", http.StatusInternalServerError)
				}
				fmt.Fprint(w, id)
			}
		}

	} else {
		http.Error(w, "Post request body is not correctly formed", http.StatusBadRequest)
	}
}

func registeredWebhook(w http.ResponseWriter, r *http.Request) {
	db := database.WebhookMongoDB{
		DatabaseURL:  "mongodb://admin:admin@ds245805.mlab.com:45805/webhook",
		DatabaseName: "webhook",
		Collection:   "WebhookPayload",
	}

	id := strings.Split(r.URL.Path, "/")

	switch r.Method {
	case "GET":
		db.Init()
		p, ok := db.Get(id[2])

		if !ok {
			http.Error(w, "The id is incorrect", http.StatusBadRequest)
		}
		bytes, err := json.Marshal(p)
		if err != nil {
			http.Error(w, "Error during marshaling", http.StatusInternalServerError)
		}
		fmt.Fprint(w, string(bytes))

	case "DELETE":
		db.Init()
		ok := db.Delete(id[2])
		if !ok {
			http.Error(w, "The id is incorrect", http.StatusBadRequest)
		}
	}
}

func retrivingLatest(w http.ResponseWriter, r *http.Request) {
	db := database.WebhookMongoDB{
		DatabaseURL:  "mongodb://admin:admin@ds245805.mlab.com:45805/webhook",
		DatabaseName: "webhook",
		Collection:   "FixerPayload",
	}
	var l types.Latest

	//Connection to the database
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, &l)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	if l.BaseCurrency != "EUR" {
		http.Error(w, "Base currency must be EUR", http.StatusBadRequest)
	}
	// Check if latest payload already exist is DB
	var fixer types.Fixer
	dbErr := session.DB(db.DatabaseName).C(db.Collection).Find(bson.M{"date": time.Now().Format("2006-01-02")}).One(&fixer)
	if dbErr != nil && dbErr.Error() != "not found" {

		http.Error(w, dbErr.Error(), http.StatusInternalServerError)

	} else if dbErr != nil && dbErr.Error() == "not found" {

		err2 := session.DB(db.DatabaseName).C(db.Collection).Find(bson.M{"date": time.Now().Format("2006-01-02")}).One(&fixer)

		if err2 != nil {
			http.Error(w, err2.Error(), http.StatusInternalServerError)
		}

		for key, value := range fixer.Rates {
			if key == l.TargetCurrency {
				fmt.Fprint(w, value)
			}
		}
	} else {

		for key, value := range fixer.Rates {

			if key == l.TargetCurrency {

				fmt.Fprint(w, value)

			}
		}
	}
}
func AverageRate(w http.ResponseWriter, r *http.Request) {
	db := database.WebhookMongoDB{
		DatabaseURL:  "mongodb://admin:admin@ds245805.mlab.com:45805/webhook",
		DatabaseName: "webhook",
		Collection:   "FixerPayload",
	}

	var fixer []types.Fixer
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// Remove all fixer payloads from the database
	err = session.DB(db.DatabaseName).C(db.Collection).Find(nil).All(&fixer)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// Read request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	defer r.Body.Close()
	var l types.Latest

	err = json.Unmarshal(body, &l)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	if len(l.BaseCurrency) != 3 && len(l.TargetCurrency) != 3 {
		http.Error(w, "length must be 3", http.StatusBadRequest)
	}
	if !strings.Contains(l.BaseCurrency, "EUR") {
		http.Error(w, "not implemented", http.StatusNotImplemented)
	}
	ok := validateCurrency(l.TargetCurrency)
	if !ok {
		http.Error(w, "Target currency is not implemented", http.StatusNotImplemented)
	}
	var count []float32
	var averageValue float32

	for _, value := range fixer {

		temp := value
		for k, v := range temp.Rates {
			if l.TargetCurrency == k {
				count = append(count, v)
			}
		}

	}

	for _, value := range count {
		averageValue = averageValue + value
	}
	averageValue = averageValue / 3
	fmt.Fprint(w, averageValue)

}

func evaluationTrigger(w http.ResponseWriter, r *http.Request) {
	db := database.WebhookMongoDB{
		DatabaseURL:  "mongodb://admin:admin@ds245805.mlab.com:45805/webhook",
		DatabaseName: "webhook",
		Collection:   "WebhookPayload",
	}

	var form types.Invoked
	var results []types.Subscriber
	var latestFixer types.Fixer
	//Connection to the database
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	defer session.Close()

	count, err := session.DB(db.DatabaseName).C(db.Collection).Count()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	err = session.DB(db.DatabaseName).C(db.Collection).Find(nil).All(&results)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	err = session.DB(db.DatabaseName).C("FixerPayload").Find(bson.M{"date": time.Now().Format("2006-01-02")}).One(&latestFixer)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	//Update the payload with currect rate
	for i := 0; i < count; i++ {

		form.BaseCurrency = results[i].BaseCurrency
		form.TargetCurrency = results[i].TargetCurrency
		form.MinTriggerValue = results[i].MinTriggerValue
		form.MaxTriggerValue = results[i].MaxTriggerValue

		// Run through all the rates
		for key, value := range latestFixer.Rates {
			// Checks if key"currency" matches a target currency
			if key == form.TargetCurrency {
				form.CurrentRate = value

				if strings.Contains(results[i].WebhookURL, "slack") {

					var slack types.SlackPayload
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
					fmt.Fprint(w, http.StatusOK)
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
					fmt.Fprint(w, http.StatusOK)
				}

			}
		}
	}
}

func DropFixerCollection() {
	db := database.WebhookMongoDB{
		DatabaseURL:  "mongodb://admin:admin@ds245805.mlab.com:45805/webhook",
		DatabaseName: "webhook",
		Collection:   "FixerPayload",
	}

	//Connection to the database
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	err = session.DB(db.DatabaseName).C(db.Collection).DropCollection()
	if err != nil {
		panic(err)
	}
}

func validateCurrency(c string) bool {
	abspath, err := filepath.Abs("./currency.json")
	if err != nil {
		log.Fatal("abs path not found")
	}
	body, err := ioutil.ReadFile(abspath)
	if err != nil {
		fmt.Printf("Error occured! %s", err.Error())
	}
	var f types.Fixer
	err = json.Unmarshal(body, &f)

	if f.Rates[c] == 0 || f.Base == c {
		return false
	} else {
		return true
	}

}

func main() {
	//os.Chdir("/home/zohaib/Desktop/Go/src/github.com/zohaib194/oblig2/Web")

	port := os.Getenv("PORT")
	if len(port) == 0 {
		log.Fatal("Port is not set")
	}

	http.HandleFunc("/root", postReqHandler)
	http.HandleFunc("/root/", registeredWebhook)
	http.HandleFunc("/root/latest", retrivingLatest)
	http.HandleFunc("/root/average", AverageRate)
	http.HandleFunc("/root/evaluationtrigger", evaluationTrigger)
	http.ListenAndServe(":"+port, nil)

}
