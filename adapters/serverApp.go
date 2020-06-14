package adapters

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
)

func SaveVideoLink(userID, adminID string, filename string) error {
	client := http.Client{}
	user, _ := strconv.Atoi(userID)
	admin, _ := strconv.Atoi(adminID)
	reqBody, err := json.Marshal(struct {
		UserID   int    `json:"user_id"`
		AdminID  int    `json:"admin_id"`
		Filename string `json:"filename"`
	}{
		UserID:   user,
		AdminID:  admin,
		Filename: filename,
	})
	if err != nil {
		return err
	}
	req, _ := http.NewRequest(http.MethodPost,
		fmt.Sprintf("%s/video/save", os.Getenv("SERVERAPP_HOST")), bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil || res.StatusCode != http.StatusCreated {
		return errors.New("bad request: serverApp")
	}

	return nil
}
