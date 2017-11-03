package fix

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/zohaib194/oblig2/database"
	"github.com/zohaib194/oblig2/types"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

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
		DatabaseURL:  "mongodb://admin:admin@ds245805.mlab.com:45805/webhook",
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
