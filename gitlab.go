package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"gopkg.in/go-playground/webhooks.v5/gitlab"
)

// RepsonseData is ...
type RepsonseData struct {
	Response string
	Message  string
}

// Config is ...
type Config struct {
	ChatworkToken string
	RoomID        string
	ListenPort    string
	SecretToken   string
}

var (
	path     = "/"
	rs       = RepsonseData{Response: "", Message: ""}
	config   = &Config{}
	response string
)

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}

func loadConfig() *Config {
	conf := Config{}
	content, e := ioutil.ReadFile("./config.json")
	handleError(e)
	err := json.Unmarshal(content, &conf)
	handleError(err)
	return &conf
}

func sendMessageToChatwork(body string) {
	apiUrl := "https://api.chatwork.com/"
	resource := "/v2/rooms/" + config.RoomID + "/messages"

	u, _ := url.ParseRequestURI(apiUrl)
	u.Path = resource
	urlStr := fmt.Sprintf("%v", u)

	data := url.Values{}
	data.Set("body", body)
	fmt.Println(data.Encode())

	client := &http.Client{}
	r, _ := http.NewRequest("POST", urlStr, bytes.NewBufferString(data.Encode()))
	r.Header.Add("X-ChatWorkToken", config.ChatworkToken)
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(r)
	handleError(err)

	defer resp.Body.Close()
	contents, _ := ioutil.ReadAll(resp.Body)

	fmt.Println(resp.Status)
	fmt.Printf("result: %s\n", contents)
	fmt.Println(urlStr)
}

func main() {
	config = loadConfig()
	sendMessageToChatwork("[toall]")
	hook, _ := gitlab.New(gitlab.Options.Secret(config.SecretToken))
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		payload, err := hook.Parse(r, gitlab.MergeRequestEvents, gitlab.PipelineEvents)
		if err != nil {
			if err == gitlab.ErrEventNotFound {
				// ok event wasn;t one of the ones asked to be parsed
			}
		}
		switch payload.(type) {

		case gitlab.MergeRequestEventPayload:
			mergeRequest := payload.(gitlab.MergeRequestEventPayload)
			// Do whatever you want from here...
			fmt.Printf("%+v", mergeRequest)

		case gitlab.PipelineEventPayload:
			pipeline := payload.(gitlab.PipelineEventPayload)
			// Do whatever you want from here...
			fmt.Printf("%+v", pipeline)
		}
	})
	http.ListenAndServe(config.ListenPort, nil)
}
