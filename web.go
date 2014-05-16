package main

// Slack outgoing and incoming webhooks are handled here. Requests come in and
// are examined to see if we need to respond. If we do, we set a timer to check
// if a response was already posted by calling the history api. Responses are
// sent back using an incoming webhook.
//
// Create an outgoing webhook in your Slack here:
// https://my.slack.com/services/new/outgoing-webhook
//
// Create an incoming webhook in your Slack here:
// https://my.slack.com/services/new/incoming-webhook

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type WebhookResponse struct {
	Username string `json:"username"`
	Text     string `json:"text"`
	Channel  string `json:"channel"`
}

type ChannelMessage struct {
	Type     string
	Subtype  string
	Username string
	TS       string
	User     string
	Text     string
}

type HistoryResponse struct {
	Ok       bool
	Error    string
	Messages []ChannelMessage
	Latest   string
	Oldest   string
	HasMore  bool
}

func init() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		incomingText := r.PostFormValue("text")
		log.Printf("Handling incoming request: %s", incomingText)

		if incomingText == ":point_right:" {
			// Wait 1 minute and then see if anyone replied
			log.Print("Waiting to bump")
			go func() {
				time.Sleep(30 * time.Second)

				log.Print("Checking history")
				messages, err := MakeHistoryCall(r.PostFormValue("channel_id"), r.PostFormValue("timestamp"))
				if err == nil {
					log.Printf("History returned %d new messages", len(messages))
					needsResponse := true
					for _, m := range messages {
						if strings.Contains(m.Text, ":point_left") {
							needsResponse = false
							break
						}
					}

					if needsResponse {
						log.Print("Completing bump")
						err := MakeIncomingWebhookCall(r.PostFormValue("team_domain"), r.PostFormValue("channel_id"), ":point_left:")
						if err != nil {
							log.Fatal(err)
						}
					} else {
						log.Print("Bump has already been completed")
					}
				} else {
					log.Fatal(err)
				}
			}()
		}
	})
}

func StartServer(port int) {
	log.Printf("Starting HTTP server on %d", port)
	err := http.ListenAndServe(":"+strconv.Itoa(port), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func MakeHistoryCall(channel_id string, ts string) ([]ChannelMessage, error) {
	var method string
	if strings.HasPrefix(channel_id, "C") {
		method = "channels.history"
	} else if strings.HasPrefix(channel_id, "G") {
		method = "groups.history"
	} else if strings.HasPrefix(channel_id, "D") {
		method = "im.history"
	} else {
		return nil, errors.New("Unknown channel type")
	}

	// Build our query
	apiUrl := url.URL{
		Scheme: "https",
		Host:   "slack.com",
		Path:   "/api/" + method,
	}

	apiParams := url.Values{}
	apiParams.Set("token", apiKey)
	apiParams.Set("channel", channel_id)
	apiParams.Set("oldest", ts)

	apiUrl.RawQuery = apiParams.Encode()

	// Execute call
	log.Print(apiUrl.String())
	res, err := http.Get(apiUrl.String())
	if err != nil {
		return nil, err
	}

	// read response
	body, _ := ioutil.ReadAll(res.Body)

	// parse into json
	var h HistoryResponse
	err = json.Unmarshal(body, &h)
	if err != nil {
		return nil, err
	}

	if h.Ok != true {
		return nil, errors.New("API Error: " + h.Error)
	}

	return h.Messages, nil
}

func MakeIncomingWebhookCall(domain string, channel_id string, text string) error {
	// Build our query
	webhookUrl := url.URL{
		Scheme: "https",
		Host:   domain + ".slack.com",
		Path:   "/services/hooks/incoming-webhook",
	}

	// Construct response payload
	var response WebhookResponse
	response.Username = botUsername
	response.Text = text
	response.Channel = channel_id

	payload, err := json.Marshal(response)
	if err != nil {
		return err
	}

	webhookParams := url.Values{}
	webhookParams.Set("token", webhookToken)
	webhookParams.Set("payload", string(payload))

	webhookUrl.RawQuery = webhookParams.Encode()

	// Execute call
	_, err = http.PostForm(webhookUrl.String(), webhookParams)
	if err != nil {
		return err
	}

	return nil
}
