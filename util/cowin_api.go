package util

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Session struct {
	Date       string `json:"date"`
	Available  int    `json:"available_capacity"`
	Available1 int    `json:"available_capacity_dose1"`
	Available2 int    `json:"available_capacity_dose2"`
	AgeLimit   int    `json:"min_age_limit"`
	Vaccine    string `json:"vaccine"`
}

type Center struct {
	Name     string    `json:"name"`
	Address  string    `json:"address"`
	Sessions []Session `json:"sessions"`
}

type CenterData struct {
	Centers []Center `json:"centers"`
}

var apifmtstring string = "https://cdn-api.co-vin.in/api/v2/appointment/sessions/public/calendarByPin?pincode=%d&date=%s"
var ageLimit int = 25

func init() {
	godotenv.Load()
	ageString := os.Getenv("AGE_LIMIT")

	i, err := strconv.Atoi(ageString)
	if err == nil {
		ageLimit = i
	}
}

func getCenters(pincode int, date string, dose int) ([]Center, error) {
	url := fmt.Sprintf(apifmtstring, pincode, date)
	resp, err := doRequest(url)
	if err != nil {
		log.Printf("Query to query cowin api")
		return []Center{}, nil
	}

	var data CenterData
	err = json.Unmarshal(resp.Body(), &data)
	if err != nil {
		log.Printf("Failed to parse cowin api response data")
		return []Center{}, nil
	}

	availableCenters := []Center{}
	for _, center := range data.Centers {
		sessions := []Session{}
		for _, session := range center.Sessions {
			if session.Available > 0 && session.AgeLimit <= ageLimit {
				if dose == 0 ||
					(dose == 1 && session.Available1 > 0) ||
					(dose == 2 && session.Available2 > 0) {
					sessions = append(sessions, session)
				}
			}
		}
		if len(sessions) > 0 {
			availableCenters = append(availableCenters, Center{
				Name:     center.Name,
				Address:  center.Address,
				Sessions: sessions,
			})
		}
	}
	return availableCenters, nil
}
