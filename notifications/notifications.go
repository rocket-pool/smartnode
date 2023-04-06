package notifications

import{
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
}

func sendDiscord(){
	if !cfg.IsNativeMode  {

		url := "http://192.168.1.52:8001/notify"

		// Read the JSON data from file
		data, err := ioutil.ReadFile("./example.json")
		if err != nil {
		fmt.Println("Error reading file:", err)
		}

		// Create the HTTP POST request
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
		if err != nil {
			fmt.Println("Error creating request:", err)
		}

		// Set headers if needed
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error making request:", err)
		}
		defer resp.Body.Close()

		// Handle the response
		fmt.Println("Response Status:", resp.Status)
		fmt.Println("Response Body:", resp.Body)
	}
}