package channels

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"server/types"

	"github.com/go-ini/ini"
)

var VIBER_AUTH_TOKEN string

type Viber struct{}

func (v Viber) Init() {
	cfg, err := ini.Load("conf/channels_conf.ini")
	if err != nil {
		log.Println("Failed to read conf file: ", err)
	}

	VIBER_AUTH_TOKEN = cfg.Section("viber").Key("token").String()
}

func (v Viber) GetSenderInfo(data map[string]interface{}) (string, string) {

	var senderUniqueID string
	var senderName string

	if data["sender"] != nil {
		senderInfo := data["sender"].(map[string]interface{})
		senderUniqueID = senderInfo["id"].(string)
		senderName = senderInfo["name"].(string)
	} else {
		senderUniqueID = data["user_id"].(string)
	}

	return senderUniqueID, senderName
}

func (v Viber) ParseReceivedData(body []byte) (string, map[string]interface{}) {

	var data map[string]interface{}
	var event string

	json.Unmarshal(body, &data)

	event = data["event"].(string)

	if event == "message" {
		event = types.EVENT_NEW_MESSAGE
	} else if event == "delivered" || event == "seen" {
		event = types.EVENT_MESSAGE_STATUS_UPDATED
	}

	return event, data
}

func (v Viber) GetMessageInfo(data map[string]interface{}) (string, uint) {

	var (
		messageText string
		timestamp   uint
	)

	msgInfo := data["message"].(map[string]interface{})
	messageText = msgInfo["text"].(string)
	timestamp = uint(data["timestamp"].(float64))

	return messageText, timestamp
}

func (v Viber) GetMessageStatus(data map[string]interface{}) types.MessageStatus {

	var status types.MessageStatus

	event := data["event"].(string)

	if event == "delivered" {
		status = types.Delivered
	} else if event == "seen" {
		status = types.Seen
	}

	return status
}

func (v Viber) SendMessage(customerID string, messageText string, autoreply bool) {

	postData := make(map[string]interface{})

	postData["receiver"] = customerID
	postData["type"] = "text"
	postData["text"] = messageText
	postData["min_api_version"] = 1

	var senderName string
	if autoreply {
		senderName = " "
	} else {
		senderName = "Agent"
	}

	postData["sender"] = map[string]string{
		"name":   senderName,
		"avatar": "",
	}

	postBody, _ := json.Marshal(postData)
	req, _ := http.NewRequest("POST", "https://chatapi.viber.com/pa/send_message", bytes.NewReader(postBody))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Viber-Auth-Token", VIBER_AUTH_TOKEN)

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Failed to send message to Viber", err)
	}

	defer response.Body.Close()
}
