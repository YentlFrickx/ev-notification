package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

type MbpResult struct {
	Status string
}
type MbpResponse struct {
	Count   int
	Results []MbpResult
}

func getCurrentStatus() (string, []string) {
	mobilityPlusLocation := os.Getenv("MBP_LOCATION")
	r, err := http.Get(fmt.Sprintf("https://my.mobilityplus.be/sp/api/20/user/charging/locations/%s/evses/", mobilityPlusLocation))
	if err != nil {
		log.WithError(err)
		return "unknown", []string{}
	}
	defer r.Body.Close()

	var mbpResonse MbpResponse

	err = json.NewDecoder(r.Body).Decode(&mbpResonse)
	if err != nil {
		log.WithError(err)
		return "unknown", []string{}
	}

	availableCount := 0
	statuses := []string{}
	for _, res := range mbpResonse.Results {
		if res.Status == "available" {
			availableCount += 1
		}
		statuses = append(statuses, res.Status)
	}

	return fmt.Sprintf("Chargers available: %d", availableCount), statuses
}

type PushBulletData struct {
	Iden  string `json:"iden"`
	Body  string `json:"body"`
	Title string `json:"title"`
	Type  string `json:"type"`
}

func sendPBAlert(status string, statuses []string, pushBulletApiKey string, deviceId string) bool {

	bodyString := "Status of chargers:\n"

	for _, status := range statuses {
		bodyString += fmt.Sprintf(" • %s\n", status)
	}

	data := PushBulletData{
		Iden:  deviceId,
		Body:  bodyString,
		Title: status,
		Type:  "note",
	}
	body, _ := json.Marshal(data)
	request, err := http.NewRequest("POST", "https://api.pushbullet.com/v2/pushes", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.Header.Set("Access-token", pushBulletApiKey)

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		log.WithError(err)
		return false
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("Statuscode: %d", resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		log.Printf("response Body: %s", string(body))
		return false
	}

	return true
}

func main() {
	pushBulletApiKey := os.Getenv("PB_KEY")
	deviceId := os.Getenv("DEVICE_ID")
	lastStatus := ""

	for {
		status, statuses := getCurrentStatus()
		log.Infoln(status)
		if status != lastStatus {
			log.Infoln("Sending alert")
			success := sendPBAlert(status, statuses, pushBulletApiKey, deviceId)
			if success {
				lastStatus = status
			}
		}
		time.Sleep(1 * time.Minute)
	}

}
