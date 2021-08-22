package opensensemap

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

const SERVER = "ingress.opensensemap.org"

func PostFloatValue(apiToken string, measurement float32, digits int, boxId string, sensorId string) {
	url := "https://" + SERVER + "/boxes/" + boxId + "/" + sensorId

	content := fmt.Sprintf("{\"value\":\"%."+strconv.Itoa(digits)+"f\"}", measurement)
	req, _ := http.NewRequest("POST", url, bytes.NewBufferString(content))
	req.Header.Add("Authorization", apiToken)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Panic(err)
		return
	}
	defer resp.Body.Close()

	log.Println("response Status:", resp.Status)
	body, _ := ioutil.ReadAll(resp.Body)
	log.Println("response Body:", string(body))
}
