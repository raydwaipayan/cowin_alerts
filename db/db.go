package db

import (
	"log"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/asdine/storm/q"
	"github.com/asdine/storm/v3"
	"github.com/joho/godotenv"
)

type UserEntry struct {
	ID        int   `storm:"id,increment"`
	Chatid    int64 `storm:"index"`
	FirstName string
	Pincode   int
}

type Alerted struct {
	ID        int   `storm:"id,increment"`
	Chatid    int64 `storm:"index"`
	Pincode   int
	Timestamp time.Time
}

var (
	db            *storm.DB
	alertDuration int
)

func init() {
	godotenv.Load()
	db_dir := path.Join(os.Getenv("DATA_DIR"), "data.db")

	var err error
	db, err = storm.Open(db_dir)
	if err == nil {
		db.Init(&UserEntry{})
		db.Init(&Alerted{})
	}

	alertDuration = 24
	durationString := os.Getenv("ALERT_DURATION")

	i, err := strconv.Atoi(durationString)
	if err == nil {
		alertDuration = i
	}
}

func GetAllEntries() ([]UserEntry, error) {
	var entries []UserEntry
	err := db.All(&entries)
	if err == storm.ErrNotFound {
		return entries, nil
	}
	return entries, err
}

func GetUserEntries(chatid int64) ([]UserEntry, error) {
	var entries []UserEntry
	err := db.Find("Chatid", chatid, &entries)
	if err == storm.ErrNotFound {
		return entries, nil
	}

	return entries, err
}

func RemoveUserEntries(chatid int64) error {
	query := db.Select(q.Eq("Chatid", chatid))
	return query.Delete(new(UserEntry))
}

func AddUserEntry(firstname string, pincode int, chatid int64) error {
	query := db.Select(q.Eq("Chatid", chatid), q.Eq("Pincode", pincode))
	var entry UserEntry
	err := query.First(&entry)

	if err != nil {
		entry = UserEntry{
			Chatid:    chatid,
			Pincode:   pincode,
			FirstName: firstname,
		}
		err = db.Save(&entry)
		if err != nil {
			log.Printf("Could not add pincode entry for user %s\n", firstname)
			log.Print(err)
			return err
		}
	}

	return nil
}

func UpdateAlerted(chatid int64, pincode int) error {
	query := db.Select(q.Eq("Chatid", chatid), q.Eq("Pincode", pincode))
	var data Alerted
	err := query.First(&data)

	if err == nil {
		query.Delete(new(Alerted))
	}

	data = Alerted{
		Chatid:    chatid,
		Pincode:   pincode,
		Timestamp: time.Now(),
	}
	err = db.Save(&data)
	if err != nil {
		log.Printf("Failed to write alert timestamp")
		return err
	}
	return nil
}

func ShouldAlert(chatid int64, pincode int) bool {
	query := db.Select(q.Eq("Chatid", chatid), q.Eq("Pincode", pincode))
	var data Alerted
	err := query.First(&data)

	if err != nil {
		return true
	}

	duration := time.Since(data.Timestamp)
	return duration.Hours() >= float64(alertDuration)
}
