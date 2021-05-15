package util

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/valyala/fasthttp"
)

type User struct {
	Id        int64  `json:"id"`
	FirstName string `json:"first_name"`
}

type Chat struct {
	Id int64 `json:"id"`
}

type Message struct {
	Id   int64  `json:"message_id"`
	From User   `json:"from"`
	Date int64  `json:"date"`
	Chat Chat   `json:"chat"`
	Text string `json:"text"`
}

type Update struct {
	Id      int64   `json:"update_id"`
	Message Message `json:"message"`
}

type Postoffice struct {
	City     string `json:"Block"`
	District string `json:"District"`
	State    string `json:"State"`
}

type Pincode struct {
	Status     string       `json:"Status"`
	Postoffice []Postoffice `json:"PostOffice"`
}

func doRequest(url string) []byte {
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(url)

	resp := fasthttp.AcquireResponse()
	client := &fasthttp.Client{}
	client.Do(req, resp)

	bodyBytes := resp.Body()
	return bodyBytes
}

func registerUser(username string, pincode string, chatid int64) error {
	pindata := []Pincode{}
	output := doRequest(fmt.Sprintf("https://api.postalpincode.in/pincode/%s", pincode))
	err := json.Unmarshal(output, &pindata)
	if err != nil {
		return err
	}
	location := pindata[0].Postoffice[0]

	log.Print(username, pincode, chatid, location)
	return nil
}

func ReceiveWebhook(ctx *fasthttp.RequestCtx) error {
	var update Update
	err := json.Unmarshal(ctx.PostBody(), &update)
	if err != nil {
		return err
	}

	username := update.Message.From.FirstName
	message := update.Message.Text
	chat := update.Message.Chat.Id

	params := strings.Fields(message)
	switch params[0] {
	case "/register":
		if len(params) >= 2 {
			registerUser(username, params[1], chat)
		}
	}

	return nil
}
