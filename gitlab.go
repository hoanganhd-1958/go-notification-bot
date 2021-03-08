package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os/exec"
	"strings"

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
}

func findChatworkOfMember(slice []MemberInfo, val string) (int, string) {
	for i, item := range slice {
		if before(item.Email, "@") == val {
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
	message := "[info][title]CI tool report[/title]" + pullRequestTitle + "\nAuthor: " + chatwork + "\nCommit URL: " + mergeUrl + "\nStatus: " + status + "[/info]"
	return message
}

func before(value string, a string) string {
	// Get substring before a string.
	pos := strings.Index(value, a)
	if pos == -1 {
		return ""
	}
	return value[0:pos]
}

func main() {
	config = loadConfig()

	hook, _ := gitlab.New(gitlab.Options.Secret(config.SecretToken))
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		payload, err := hook.Parse(r, gitlab.MergeRequestEvents, gitlab.PipelineEvents)
		// memberInfo := fetchMemberInfoFromGGSheet()
		if err != nil {
			if err == gitlab.ErrEventNotFound {
				// handle error
			}
		}
		switch payload.(type) {

		case gitlab.MergeRequestEventPayload:
			mergeRequest := payload.(gitlab.MergeRequestEventPayload)
			// In case merge pull request, do whatever you want from here...
			if mergeRequest.ObjectAttributes.State == "merged" {
				go func() {
					cmd := exec.Command("bash", "-c", "doo deploy -c ~/devops_tool/deploy/stg_dwh_sapphire.yml")
					cmd.Run()
				}()
				
				fmt.Println("DONE!")
			}
			
			// case gitlab.PipelineEventPayload:
			// 	pipeline := payload.(gitlab.PipelineEventPayload)
			// 	fmt.Printf("%+v", pipeline)
			// 	CIStatus := pipeline.ObjectAttributes.Status
			// 	authorEmail := pipeline.Commit.Author.Email
			// 	fmt.Printf("%+v", authorEmail)
			// 	_, chatwork := findChatworkOfMember(memberInfo, before(authorEmail, "@"))
			// 	mergeUrl := pipeline.Commit.URL
			// 	pullRequestTitle := pipeline.Commit.Title
			// 	message := makeMessage(chatwork, pullRequestTitle, mergeUrl, CIStatus)

			// 	if CIStatus == "success" || CIStatus == "failed" {
			// 		sendMessageToChatwork(message)
			// 	}
		}
	})
	http.ListenAndServe(config.ListenPort, nil)
}
