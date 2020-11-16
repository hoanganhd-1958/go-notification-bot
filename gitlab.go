package main

import (
	"fmt"
	"net/http"

	"gopkg.in/go-playground/webhooks.v5/gitlab"
)

const (
	path = "/"
)

func main() {
	hook, _ := gitlab.New(gitlab.Options.Secret("hoanganhd"))
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
	http.ListenAndServe(":3000", nil)
}
