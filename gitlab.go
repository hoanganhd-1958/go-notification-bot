package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/hoanganhd-1958/webhooks/gitlab"
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
	GGSheetAPI    string
}

type MemberInfo struct {
	Email    string `json:"email"`
	Chatwork string `json:"chatwork"`
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
	apiURL := "https://api.chatwork.com/"
	resource := "/v2/rooms/" + config.RoomID + "/messages"

	u, _ := url.ParseRequestURI(apiURL)
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

func findChatworkOfMember(slice []MemberInfo, val string) (int, string) {
	for i, item := range slice {
		if item.Email == val {
			return i, item.Chatwork
		}
	}
	return -1, ""
}

func fetchMemberInfoFromGGSheet() []MemberInfo {
	resp, err := http.Get(config.GGSheetAPI)
	if err != nil {
		// handle error
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	var memberInfo []MemberInfo
	json.Unmarshal(body, &memberInfo)

	return memberInfo
}

func makeMessage(chatwork string, pullRequestTitle string, mergeUrl string, status string) string {
	message := "[info][title]CI report " + chatwork + "[/title]" + pullRequestTitle + "\nPull request: " + mergeUrl + "\nStatus: " + status + "[/info]"
	return message
}

func main() {
	config = loadConfig()
	memberInfo := fetchMemberInfoFromGGSheet()

	hook, _ := gitlab.New(gitlab.Options.Secret(config.SecretToken))
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		payload, err := hook.Parse(r, gitlab.MergeRequestEvents, gitlab.PipelineEvents)
		if err != nil {
			if err == gitlab.ErrEventNotFound {
				// handle error
			}
		}
		switch payload.(type) {

		// case gitlab.MergeRequestEventPayload:
		// 	mergeRequest := payload.(gitlab.MergeRequestEventPayload)
		// 	// In case merge pull request, do whatever you want from here...
		// 	fmt.Printf("%+v", mergeRequest)

		case gitlab.PipelineEventPayload:
			pipeline := payload.(gitlab.PipelineEventPayload)
			CIStatus := pipeline.ObjectAttributes.Status
			authorEmail := pipeline.Commit.Author.Email
			_, chatwork := findChatworkOfMember(memberInfo, authorEmail)
			mergeUrl := pipeline.MergeRequest.URL
			pullRequestTitle := pipeline.MergeRequest.Title
			message := makeMessage(chatwork, pullRequestTitle, mergeUrl, CIStatus)

			if CIStatus == "success" || CIStatus == "failed" {
				sendMessageToChatwork(message)
			}
		}
	})
	http.ListenAndServe(config.ListenPort, nil)
}
