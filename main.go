package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"time"
)

type MbpResult struct {
	Status string
}
type MbpResponse struct {
	Count   int
	Results []MbpResult
}

func getCurrentStatus() string {
	r, err := http.Get("https://my.mobilityplus.be/sp/api/20/user/charging/locations?window=50.88267564845102,4.681929475410969,50.88045876964255,4.687765962227375&limit=100&offset=0&has_point=true")
	if err != nil {
		log.WithError(err)
		return "unknown"
	}
	defer r.Body.Close()

	var mbpResonse MbpResponse

	err = json.NewDecoder(r.Body).Decode(&mbpResonse)
	if err != nil {
		log.WithError(err)
		return "unknown"
	}
	return mbpResonse.Results[0].Status
}

type PushBulletData struct {
	Iden  string `json:"iden"`
	Body  string `json:"body"`
	Title string `json:"title"`
	Type  string `json:"type"`
}

func sendPBAlert(status string, pushBulletApiKey string, deviceId string) bool {
	data := PushBulletData{
		Iden:  deviceId,
		Body:  fmt.Sprintf("Status of the nearby charger changed to %s", status),
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
		status := getCurrentStatus()
		log.Infoln("Current status: ", status)
		if status != lastStatus {
			log.Infoln("Sending alert")
			success := sendPBAlert(status, pushBulletApiKey, deviceId)
			if success {
				lastStatus = status
			}
		}
		time.Sleep(10 * time.Second)
	}

}
