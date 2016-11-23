package main

import (
	"encoding/json"
	"fmt"
	"github.com/ChimeraCoder/anaconda"
	"github.com/levigross/grequests"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/url"
        "os"
)

type Issue struct {
	IssueId int64 `bson:"issueId"`
}

type BeginnerResult struct {
	Items []struct {
		Html_url string `json:"html_url"`
		Id       int64  `json:"id"`
		Title    string `json:"title"`
	} `json:"items"`
}

var Api *anaconda.TwitterApi

func init() {
	anaconda.SetConsumerKey(os.Getenv("CONSUMER_KEY"))
	anaconda.SetConsumerSecret(os.Getenv("CONSUMER_SECRET"))
	Api = anaconda.NewTwitterApi(os.Getenv("ACESSS_TOKEN"),os.Getenv("ACESSS_SECRET"))
}

func main() {
  Job()
  ticker := time.NewTicker(time.Minute * 30)

      for t := range ticker.C {
           Job()
      }
}

func Job()  {
  session, err := mgo.Dial(os.Getenv("MLAB"))
  defer session.Close()
  if err != nil {
    panic(err)
  }
  c := session.DB("begitwit").C("beginner")
  c.EnsureIndexKey("IssueId")
  res, _ := grequests.Get("https://api.github.com/search/issues?q=label:beginner+is:issue+is:open&sort=updated&order=desc", nil)
  var demo BeginnerResult
  json.Unmarshal(res.Bytes(), &demo)
  for _, item := range demo.Items {
		go func ()  {
			exist, _ := c.Find(bson.M{"issueId": item.Id}).Count()
			if exist == 0 {
				c.Insert(&Issue{item.Id})
				Tweet(item.Title, item.Html_url)
			} else {
				fmt.Print("already posted")
			}
		}()
  }
}

func Tweet(title string, Url string) {
	if len(title) > 130 {
		title = title[0:170]
	}
	Api.PostTweet("#github "+title+" "+Url, url.Values{})
}
