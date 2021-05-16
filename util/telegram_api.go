package util

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/raydwaipayan/cowin_alerts/db"
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

type Response struct {
	ChatId int64  `json:"chat_id"`
	Text   string `json:"text"`
}

var (
	strPost []byte
)

func init() {
	strPost = []byte("POST")
}

func doRequest(url string) (*fasthttp.Response, error) {
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(url)

	resp := fasthttp.AcquireResponse()
	client := &fasthttp.Client{}
	err := client.Do(req, resp)
	fasthttp.ReleaseRequest(req)

	return resp, err
}

func sendMessage(firstname string, chatid int64, text string) error {
	log.Printf("Sending message to user %s with chatid %d: %s\n", firstname, chatid, text)

	url := "https://api.telegram.org/bot" + os.Getenv("BOT_TOKEN") + "/sendMessage"
	data := Response{
		ChatId: chatid,
		Text:   text,
	}
	jsonData, _ := json.Marshal(data)

	req := fasthttp.AcquireRequest()
	req.SetRequestURI(url)
	req.SetBody(jsonData)
	req.Header.SetMethodBytes(strPost)
	req.Header.SetContentType("application/json")
	res := fasthttp.AcquireResponse()
	if err := fasthttp.Do(req, res); err != nil {
		return err
	}
	fasthttp.ReleaseRequest(req)
	fasthttp.ReleaseResponse(res)

	return nil
}

func registerUser(firstname string, pincode int, chatid int64) error {
	pindata := []Pincode{}
	resp, err := doRequest(fmt.Sprintf("https://api.postalpincode.in/pincode/%d", pincode))
	if err != nil {
		return err
	}

	err = json.Unmarshal(resp.Body(), &pindata)
	if err != nil {
		return err
	}
	fasthttp.ReleaseResponse(resp)

	location := pindata[0].Postoffice[0]
	err = db.AddUserEntry(firstname, pincode, chatid)
	if err != nil {
		return err
	}

	msg := fmt.Sprintf("Successfully registered for pin: %d (%s, %s)", pincode, location.City, location.State)

	sendMessage(firstname, chatid, msg)
	return nil
}

func listEntries(firstname string, chatid int64) error {
	entries, err := db.GetUserEntries(chatid)
	if err != nil {
		return err
	}

	msg := "You have registered alerts for the following pincodes:"
	for _, entry := range entries {
		msg += "\n" + fmt.Sprint(entry.Pincode)
	}

	sendMessage(firstname, chatid, msg)
	return nil
}

func removeEntries(firstname string, chatid int64) error {
	err := db.RemoveUserEntries(chatid)
	if err != nil {
		return err
	}

	msg := "All alerts are disabled"

	sendMessage(firstname, chatid, msg)
	return nil
}

func getMsgForPin(pincode int, dateStr string) string {
	centers, _ := getCenters(pincode, dateStr)
	msg := ""

	for _, center := range centers {
		msg += fmt.Sprintf("CENTER: %s\nAddress: %s\n", center.Name, center.Address)
		for _, session := range center.Sessions {
			msg += fmt.Sprintf("\n\tDate: %s", session.Date)
			msg += fmt.Sprintf("\n\tSlots available: %d", session.Available)
			msg += fmt.Sprintf("\n\tAge Limit: %d", session.AgeLimit)
			msg += fmt.Sprintf("\n\tVaccine type: %s", session.Vaccine)
			msg += "\n"
		}
		msg += "-----------------------------------\n\n"
	}
	return msg
}

func SendUpdates() error {
	entries, err := db.GetAllEntries()
	if err != nil {
		log.Print("Fatal: Couldn't load db entries")
		return err
	}

	currentTime := time.Now()
	dateStr := currentTime.Format("02-01-2006")

	for _, entry := range entries {
		msg := getMsgForPin(entry.Pincode, dateStr)
		if msg != "" {
			msg = fmt.Sprintf("SLOTS AVAILABLE FOR PIN %d\n\n", entry.Pincode) + msg
			err = sendMessage(entry.FirstName, entry.Chatid, msg)
			if err != nil {
				log.Printf("Fatal: Couldn't send updates to user: %s\n", entry.FirstName)
			}
		} else {
			log.Printf("No entries found for pin: %d\n", entry.Pincode)
		}
	}
	return nil
}

func showHelp(firstname string, chatid int64) error {
	msg := "Options:\n/register PINCODE - Register alerts for the given pin"
	msg += "\n/list - List all pincodes registered"
	msg += "\n/disable - Disable all alerts"
	msg += "\n/status PINCODE- Get vaccine status immediately for a pin (please note that the bot automatically checks the status every 5 minutes"
	sendMessage(firstname, chatid, msg)
	return nil
}

func ReceiveWebhook(ctx *fasthttp.RequestCtx) error {
	var update Update
	err := json.Unmarshal(ctx.PostBody(), &update)
	if err != nil {
		return err
	}

	firstname := update.Message.From.FirstName
	message := update.Message.Text
	chatid := update.Message.Chat.Id

	params := strings.Fields(message)
	switch params[0] {
	case "/register":
		if len(params) >= 2 {
			i, err := strconv.Atoi(params[1])
			if err != nil {
				msg := "Invalid pincode.\nUsage: /register PINCODE\n"
				sendMessage(firstname, chatid, msg)
				break
			}
			err = registerUser(firstname, i, chatid)
			if err != nil {
				msg := "Failed to register alerts. Please try again later"
				sendMessage(firstname, chatid, msg)
			}
		} else {
			msg := "Please enter the pincode to register.\nExample: /register 742101"
			sendMessage(firstname, chatid, msg)
		}
	case "/list":
		err := listEntries(firstname, chatid)
		if err != nil {
			msg := "An unexpected error has occured. Please try again later"
			sendMessage(firstname, chatid, msg)
		}
	case "/disable":
		err := removeEntries(firstname, chatid)
		if err != nil {
			msg := "An unexpected error has occured. Please try again later"
			sendMessage(firstname, chatid, msg)
		}
	case "/status":
		if len(params) >= 2 {
			i, err := strconv.Atoi(params[1])
			if err != nil {
				msg := "Invalid pincode"
				sendMessage(firstname, chatid, msg)
				break
			}
			currentTime := time.Now()
			dateStr := currentTime.Format("02-01-2006")
			msg := getMsgForPin(i, dateStr)

			if msg == "" {
				msg = "No vaccine slots found for the given pincode"
			} else {
				msg = "Found vaccine slots:\n\n" + msg
			}
			err = sendMessage(firstname, chatid, msg)
			if err != nil {
				msg := "An unexpected error has occured. Please try again later"
				sendMessage(firstname, chatid, msg)
			}
		} else {
			msg := "Invalid pincode.\nUsage: /status PINCODE\n"
			sendMessage(firstname, chatid, msg)
		}
	case "/help":
		showHelp(firstname, chatid)
	default:
		msg := "Invalid command."
		sendMessage(firstname, chatid, msg)
	}

	return nil
}
