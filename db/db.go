package db

import (
	"log"
	"os"
	"path"

	"github.com/asdine/storm/q"
	"github.com/asdine/storm/v3"
)

type UserEntry struct {
	ID        int   `storm:"id,increment"`
	Chatid    int64 `storm:"index"`
	FirstName string
	Pincode   int
}

var (
	db *storm.DB
)

func init() {
	db_dir := path.Join(os.Getenv("DATA_DIR"), "data.db")
	var err error
	db, err = storm.Open(db_dir)
	if err == nil {
		db.Init(&UserEntry{})
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
	err := query.Find(&entry)

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
