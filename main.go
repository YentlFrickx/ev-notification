package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"gopkg.in/yaml.v3"

	log "github.com/sirupsen/logrus"
)

type MbpResult struct {
	Status string
}
type MbpResponse struct {
	Count   int
	Results []MbpResult
}

type MbpConfig struct {
	LocationGroups []struct {
		GroupName string   `yaml:"groupName"`
		Locations []string `yaml:"locations"`
	} `yaml:"locationGroups"`
}

func getLocationConfig() (*MbpConfig, error) {
	filePath, set := os.LookupEnv("MBP_CONFIG_FILE")
	if set {
		buf, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}

		config := &MbpConfig{}
		err = yaml.Unmarshal(buf, config)
		if err != nil {
			return nil, fmt.Errorf("in file %q: %w", filePath, err)
		}
		return config, nil
	} else {
		location, set := os.LookupEnv("MBP_LOCATION")
		if !set {
			log.Fatalf("No location configured")
		}
		return &MbpConfig{
			LocationGroups: []struct {
				GroupName string   `yaml:"groupName"`
				Locations []string `yaml:"locations"`
			}([]struct {
				GroupName string
				Locations []string
			}{
				{
					GroupName: "",
					Locations: []string{location},
				},
			}),
		}, nil
	}
}

func getCurrentStatus(locationId string) (int, []string) {
	r, err := http.Get(fmt.Sprintf("https://my.mobilityplus.be/sp/api/20/user/charging/locations/%s/evses/", locationId))
	if err != nil {
		log.WithError(err)
		return 0, []string{}
	}
	defer r.Body.Close()

	var mbpResonse MbpResponse

	err = json.NewDecoder(r.Body).Decode(&mbpResonse)
	if err != nil {
		log.WithError(err)
		return 0, []string{}
	}

	availableCount := 0
	statuses := []string{}
	for _, res := range mbpResonse.Results {
		if res.Status == "available" {
			availableCount += 1
		}
		statuses = append(statuses, res.Status)
	}

	return availableCount, statuses
}

type PushBulletData struct {
	Iden  string `json:"iden"`
	Body  string `json:"body"`
	Title string `json:"title"`
	Type  string `json:"type"`
}

func sendPBAlert(title string, bodyString string, pushBulletApiKey string, deviceId string) bool {
	data := PushBulletData{
		Iden:  deviceId,
		Body:  bodyString,
		Title: title,
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

func createNotification(groupName string, availableAmount int, statuses []string) (string, string) {
	if groupName != "" {
		groupName = " in " + groupName
	}
	title := fmt.Sprintf("%d chargers available%s", availableAmount, groupName)
	bodyString := fmt.Sprintf("Status of chargers%s:\n", groupName)
	for _, status := range statuses {
		bodyString += fmt.Sprintf(" • %s\n", status)
	}
	return title, bodyString
}

func main() {
	pushBulletApiKey, set := os.LookupEnv("PB_KEY")
	if !set {
		log.Fatalf("Missing PushBullet API key")
	}

	deviceId, set := os.LookupEnv("DEVICE_ID")
	if !set {
		log.Fatalf("Missing PushBullet device id")
	}

	conf, err := getLocationConfig()
	if err != nil {
		log.WithError(err)
		return
	}

	groups := conf.LocationGroups
	for _, group := range groups {
		log.Printf("Setup group: %s", group.GroupName)
		group := group
		go func() {
			lastTotalAmount := 0
			for {
				var currentStatuses []string
				totalAmount := 0
				for _, location := range group.Locations {
					status, statuses := getCurrentStatus(location)
					currentStatuses = append(currentStatuses, statuses...)
					totalAmount += status
				}
				title, notification := createNotification(group.GroupName, totalAmount, currentStatuses)
				if lastTotalAmount != totalAmount {
					log.Infof("Sending notification for %s", group.GroupName)
					sendPBAlert(title, notification, pushBulletApiKey, deviceId)
					lastTotalAmount = totalAmount
				}
				time.Sleep(1 * time.Minute)
			}
		}()
	}

	select {}

}
