/*
	*----------------------------sources---------------------------------*
	https://lanreadelowo.com/blog/2017/04/08/testing-http-handlers-go/
*/
package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	database "github.com/zohaib194/oblig2/database"
	types "github.com/zohaib194/oblig2/types"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func Test_postReqHandler(t *testing.T) {
	db := database.WebhookMongoDB{
		DatabaseURL:  "mongodb://admin:admin@ds245805.mlab.com:45805/webhook",
		DatabaseName: "webhook",
		Collection:   "WebhookPayload",
	}

	count := db.Count()
	sub := types.Subscriber{
		WebhookURL:      "https://hooks.slack.com/services/T7E02MPH7/B7N4L3S75/IZpacPzX93B1YcIDSav4irO",
		BaseCurrency:    "EUR",
		TargetCurrency:  "NOK",
		MinTriggerValue: 1.50,
		MaxTriggerValue: 9.2,
	}

	// Marshalling the payload
	content, err := json.Marshal(sub)
	if err != nil {
		t.Errorf("Error occured! %v", err.Error())
	}

	// io.Reader of bytes
	body := ioutil.NopCloser(bytes.NewBufferString(string(content)))

	// Creating the POST request with payload for testing
	req, err := http.NewRequest("POST", "/root", body)

	if err != nil {
		t.Errorf("Error occured! %v", err.Error())
	}

	// Serve as ResponsWriter for testing
	respRec := httptest.NewRecorder()

	// ServeHTTP calls postReqHandler with respRec as ResponsWriter and req as Request
	http.HandlerFunc(postReqHandler).ServeHTTP(respRec, req)

	// Check if response status code is 200.
	if respRec.Code != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, respRec.Code)
	}

	// Bytes from response Body
	resbody, err := ioutil.ReadAll(respRec.Body)
	if err != nil {
		t.Errorf("Error occured! %v", err.Error())
	}

	// ID recieved in return from response Body
	id := string(resbody)

	// DB check
	_, ok := db.Get(id)
	if !ok {
		t.Error("Id does not exist in the DB")
	}

	if count+1 != db.Count() {
		t.Error("Error during adding Subscriber in DB")
	}

}

func Test_registerWebhook(t *testing.T) {
	db := database.WebhookMongoDB{
		DatabaseURL:  "mongodb://admin:admin@ds245805.mlab.com:45805/webhook",
		DatabaseName: "webhook",
		Collection:   "WebhookPayload",
	}

	// Get the existing payload of testSub from the DB
	count := db.Count()
	////////////////////  adding sub in db for Get and Delete requests TESTS /////////////////////////////////
	sub := types.Subscriber{
		WebhookURL:      "https://hooks.slack.com/services/T7E02MPH7/B7N4L3S75/IZpacPzX93B1YcIDSav4irO",
		BaseCurrency:    "EUR",
		TargetCurrency:  "NOK",
		MinTriggerValue: 1.50,
		MaxTriggerValue: 9.2,
	}

	// Marshalling the payload
	content, err := json.Marshal(sub)
	if err != nil {
		t.Errorf("Error occured! %v", err.Error())
	}

	// io.Reader of bytes
	body := ioutil.NopCloser(bytes.NewBufferString(string(content)))

	// Creating the POST request with payload for testing
	req, err := http.NewRequest("POST", "/root", body)

	if err != nil {
		t.Errorf("Error occured! %v", err.Error())
	}

	// Serve as ResponsWriter for testing
	respRec := httptest.NewRecorder()

	// ServeHTTP calls postReqHandler with respRec as ResponsWriter and req as Request
	http.HandlerFunc(postReqHandler).ServeHTTP(respRec, req)

	// Check if response status code is 200.
	if respRec.Code != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, respRec.Code)
	}
	// Bytes from response Body
	resbody, err := ioutil.ReadAll(respRec.Body)
	if err != nil {
		t.Errorf("Error occured! %v", err.Error())
	}

	// ID recieved in return from response Body
	id := string(resbody)

	////////////////////  GET REQUEST TEST /////////////////////////////////
	url := "/root/" + id
	// Creating the POST request with payload for testing
	req, err = http.NewRequest("GET", url, nil)

	if err != nil {
		t.Errorf("Error occured! %v", err.Error())
	}

	// Serve as ResponsWriter for testing
	respRec = httptest.NewRecorder()

	// ServeHTTP calls postReqHandler with respRec as ResponsWriter and req as Request
	http.HandlerFunc(registeredWebhook).ServeHTTP(respRec, req)

	// Check if response status code is 200.
	if respRec.Code != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, respRec.Code)
	}

	bytes, err := ioutil.ReadAll(respRec.Body)

	if err != nil {
		t.Errorf("Error during ioutil %v", err.Error())
	}
	var result types.Subscriber

	err = json.Unmarshal(bytes, &result)

	if err != nil {
		t.Errorf("Error during unmarshalling %v", err.Error())
	}

	if result.WebhookURL != sub.WebhookURL || result.BaseCurrency != sub.BaseCurrency || result.TargetCurrency != sub.TargetCurrency || result.MinTriggerValue != sub.MinTriggerValue || result.MaxTriggerValue != sub.MaxTriggerValue {
		t.Errorf("Error! expected %s, got %s", sub, result)
	}

	// Test for DELETE method
	count = db.Count() - 1
	req, err = http.NewRequest("DELETE", url, nil)

	if err != nil {
		t.Errorf("Error occured! %v", err.Error())
	}

	// ServeHTTP calls postReqHandler with respRec as ResponsWriter and req as Request
	http.HandlerFunc(registeredWebhook).ServeHTTP(respRec, req)

	// Check if response status code is 200.
	if respRec.Code != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, respRec.Code)
	}

	if count != db.Count() {
		t.Errorf("Subscriber with id %s, was not deleted", id)
	}
}

func Test_retriveLatest(t *testing.T) {
	db := database.WebhookMongoDB{
		DatabaseURL:  "mongodb://admin:admin@ds245805.mlab.com:45805/webhook",
		DatabaseName: "webhook",
		Collection:   "FixerPayload",
	}

	l := types.Latest{
		"EUR",
		"NOK",
	}
	// start session to get value expected
	session, err := mgo.Dial(db.DatabaseURL)
	defer session.Close()
	if err != nil {
		t.Error(err)
	}

	var expectedValue float32

	// Marshalling the payload
	content, err := json.Marshal(l)
	if err != nil {
		t.Errorf("Error occured! %v", err.Error())
	}

	// io.Reader of bytes
	body := ioutil.NopCloser(bytes.NewBufferString(string(content)))

	// Creating the POST request with payload for testing
	req, err := http.NewRequest("POST", "/root/latest", body)

	if err != nil {
		t.Errorf("Error occured! %v", err.Error())
	}

	// Serve as ResponsWriter for testing
	respRec := httptest.NewRecorder()

	// ServeHTTP calls postReqHandler with respRec as ResponsWriter and req as Request
	http.HandlerFunc(retrivingLatest).ServeHTTP(respRec, req)

	// Check if response status code is 200.
	if respRec.Code != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, respRec.Code)
	}

	bytes, err := ioutil.ReadAll(respRec.Body)

	if err != nil {
		t.Errorf("Error during ioutil %v", err.Error())
	}
	var result float32

	err = json.Unmarshal(bytes, &result)

	if err != nil {
		t.Errorf("Error during unmarshalling %v", err.Error())
	}

	///////////////////////////////////
	var expected types.Fixer
	err = session.DB(db.DatabaseName).C(db.Collection).Find(bson.M{"date": time.Now().Format("2006-01-02")}).One(&expected)

	if err != nil {
		t.Errorf("err during getting date %v", err.Error())
	}
	for key, value := range expected.Rates {
		if l.TargetCurrency == key {
			expectedValue = value
		}
	}

	// Check if result value is as expected
	if result != expectedValue {
		t.Errorf("Expected %g, Got %g", expectedValue, result)
	}

}

func Test_averageRate(t *testing.T) {
	db := database.WebhookMongoDB{
		DatabaseURL:  "mongodb://admin:admin@ds245805.mlab.com:45805/webhook",
		DatabaseName: "webhook",
		Collection:   "FixerPayload",
	}
	l := types.Latest{
		BaseCurrency:   "EUR",
		TargetCurrency: "NOK",
	}

	// start session to get value expected
	session, err := mgo.Dial(db.DatabaseURL)
	defer session.Close()
	if err != nil {
		t.Error(err)
	}

	// Marshalling the payload
	content, err := json.Marshal(l)
	if err != nil {
		t.Errorf("Error occured! %v", err.Error())
	}

	// io.Reader of bytes
	body := ioutil.NopCloser(bytes.NewBufferString(string(content)))

	// Creating the POST request with payload for testing
	req, err := http.NewRequest("POST", "/root/average", body)

	if err != nil {
		t.Errorf("Error occured! %v", err.Error())
	}

	// Serve as ResponsWriter for testing
	respRec := httptest.NewRecorder()
	mock = true
	// ServeHTTP calls postReqHandler with respRec as ResponsWriter and req as Request
	http.HandlerFunc(AverageRate).ServeHTTP(respRec, req)

	// Check if response status code is 200.
	if respRec.Code != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, respRec.Code)
	}

	bytes, err := ioutil.ReadAll(respRec.Body)

	if err != nil {
		t.Errorf("Error during ioutil %v", err.Error())
	}
	var result float32

	err = json.Unmarshal(bytes, &result)

	if err != nil {
		t.Errorf("Error during unmarshalling %v", err.Error())
	}

	/////////////////
	var fixers []types.Fixer

	err = session.DB(db.DatabaseName).C(db.Collection).Find(nil).All(&fixers)

	if err != nil {
		t.Errorf("err during getting date %v", err.Error())
	}

	var count []float32
	var averageValue float32

	for _, value := range fixers {
		temp := value
		if temp.Date >= time.Now().AddDate(0, 0, -3).Format("2006-01-02") {
			for k, v := range temp.Rates {
				if l.TargetCurrency == k {
					count = append(count, v)
				}
			}
		}
	}

	for _, value := range count {
		averageValue = averageValue + value
	}
	averageValue = averageValue / 3

	// Check if result value is as expected
	if result != averageValue {
		t.Errorf("Expected %g, Got %g", averageValue, result)
	}
}
