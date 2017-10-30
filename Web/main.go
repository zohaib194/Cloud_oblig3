package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	types "github.com/zohaib194/oblig2"
	//	clock "github.com/zohaib194/oblig2/Clock"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func validateCurrency(c string) bool {
	body, err := ioutil.ReadFile("currency.json")
	if err != nil {
		fmt.Printf("Error occured! %s", err.Error())
	}
	var f types.Fixer
	err = json.Unmarshal(body, &f)

	for key, _ := range f.Rates {
		if c == key {
			return true
		}
	}
	return false
}

func postReqHandler(w http.ResponseWriter, r *http.Request) {
	db := WebhookMongoDB{
		DatabaseURL:  "mongodb://localhost",
		DatabaseName: "Webhook",
		Collection:   "WebhookPayload",
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	defer r.Body.Close()
	var p Subscriber

	err = json.Unmarshal(body, &p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	if len(p.WebhookURL) != 0 && len(p.BaseCurrency) != 0 && len(p.TargetCurrency) != 0 && p.MaxTriggerValue > p.MinTriggerValue && p.MinTriggerValue < p.MaxTriggerValue {
		if !strings.Contains(p.BaseCurrency, "EUR") {
			http.Error(w, "Not implemented", http.StatusNotImplemented)
		} else {
			ok := validateCurrency(p.BaseCurrency)
			ok2 := validateCurrency(p.TargetCurrency)
			if !ok && !ok2 {
				http.Error(w, "Invalid base or target currency", http.StatusBadRequest)
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
	db := WebhookMongoDB{
		DatabaseURL:  "mongodb://localhost",
		DatabaseName: "Webhook",
		Collection:   "WebhookPayload",
	}

	id := strings.Split(r.URL.Path, "/")

	switch r.Method {
	case "GET":
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
		//fmt.Printf(r.Method, id[2])
		ok := db.Delete(id[2])
		if !ok {
			http.Error(w, "The id is incorrect", http.StatusBadRequest)
		}
	}
}

func retrivingLatest(w http.ResponseWriter, r *http.Request) {
	db := WebhookMongoDB{
		DatabaseURL:  "mongodb://localhost",
		DatabaseName: "Webhook",
		Collection:   "FixerPayload",
	}
	var l Latest

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
	var fixer Fixer
	dbErr := session.DB(db.DatabaseName).C(db.Collection).Find(bson.M{"date": time.Now().Format("2006-01-02")}).One(&fixer)
	if dbErr != nil && dbErr.Error() != "not found" {

		http.Error(w, dbErr.Error(), http.StatusInternalServerError)

	} else if dbErr != nil && dbErr.Error() == "not found" {

		LatestFixer()

		err2 := session.DB(db.DatabaseName).C(db.Collection).Find(bson.M{"date": time.Now().Format("2006-01-02")}).One(&fixer)

		if err2 != nil {
			http.Error(w, err2.Error(), http.StatusInternalServerError)
		}

		for key, value := range fixer.Rates {
			if key == l.TargetCurrency {
				fmt.Fprint(w, value)
			}
		}
	} else if dbErr == nil {

		for key, value := range fixer.Rates {

			if key == l.TargetCurrency {

				fmt.Fprint(w, value)

			}
		}
	}
}
func AverageRate(w http.ResponseWriter, r *http.Request) {
	db := WebhookMongoDB{
		DatabaseURL:  "mongodb://localhost",
		DatabaseName: "Webhook",
		Collection:   "FixerPayload",
	}

	var fixer []Fixer
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	//startDate := time.Now().AddDate(0, 0, -3).Format("2006-01-02")
	//endDate := time.Now().Format("2006-01-02")

	//GetFixerSevenDays(time.Now().AddDate(0, 0, -3), time.Now())
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
	var l Latest

	err = json.Unmarshal(body, &l)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	var count []float32
	var averageValue float32

	for _, value := range fixer {
		//fmt.Print(value)
		if l.BaseCurrency == value.Base {
			//fmt.Print(value)
			temp := value
			for k, v := range temp.Rates {
				if l.TargetCurrency == k {
					count = append(count, v)
					fmt.Println(count)
				}
			}
		}
	}

	for _, value := range count {
		averageValue = averageValue + value
	}
	averageValue = averageValue / 3
	fmt.Fprint(w, averageValue)

}

/*
	Take the first Fixer payload from the collections
*/
func GetFixerSevenDays(sd time.Time, ed time.Time) {
	db := WebhookMongoDB{
		DatabaseURL:  "mongodb://localhost",
		DatabaseName: "Webhook",
		Collection:   "FixerPayload",
	}

	//Connection to the database
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	var fixer *Fixer
	for ; sd.Unix() <= ed.Unix(); sd = sd.AddDate(0, 0, +1) {

		URL := "http://api.fixer.io/" + sd.Format("2006-01-02")
		//fmt.Print(URL + "\n")
		res, err := http.Get(URL)
		if err != nil {
			panic(err) //TODO
		}

		body, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			panic(err)
		}

		err = json.Unmarshal(body, &fixer)
		if err != nil {
			panic(err)
		}
		//fmt.Print(fixer)
		fixer.Date = sd.Format("2006-01-02")
		ok := SaveFixer(fixer)
		if !ok {
			fmt.Print("Error occured during saving the data in database")
		}
	}
}

func DropFixerCollection() {
	db := WebhookMongoDB{
		DatabaseURL:  "mongodb://localhost",
		DatabaseName: "Webhook",
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

func LatestFixer() {
	//Send request to Fixer.io
	fixerURL := "http://api.fixer.io/latest?base=EUR"
	f, ok := GetFixer(fixerURL)
	if !ok {
		fmt.Print("latestFixer()")
	}
	SaveFixer(f)
}

/*
	Get the json from Fixer.io
*/
func GetFixer(url string) (*Fixer, bool) {
	var f *Fixer
	/*
		res, err := http.Get(url)
		if err != nil {
			fmt.Printf(err.Error(), http.StatusBadRequest)
			return f, false
		}
		defer res.Body.Close()

		body, err := ioutil.ReadAll(res.Body)
	*/
	body, err := ioutil.ReadFile("base.json")
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
func SaveFixer(f *Fixer) bool {
	db := WebhookMongoDB{
		DatabaseURL:  "mongodb://localhost",
		DatabaseName: "Webhook",
		Collection:   "FixerPayload",
	}

	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	err = session.DB(db.DatabaseName).C(db.Collection).Insert(&f)
	if err != nil {
		fmt.Printf("error in SaveFixer(), %v", err.Error())
		return false
	}
	return true
}

func main() {
	os.Chdir("/home/zohaib/Desktop/Go/projects/cloud_oblig2")

	http.HandleFunc("/root", postReqHandler)
	http.HandleFunc("/root/", registeredWebhook)
	http.HandleFunc("/root/latest", retrivingLatest)
	http.HandleFunc("/root/average", AverageRate)
	http.ListenAndServe(":8080", nil)

}
