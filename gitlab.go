package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

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

func main() {
	config = loadConfig()
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
