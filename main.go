package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

func getSleepDuration(retryAfter string) (time.Duration, error) {
	sleepDuration, err := strconv.Atoi(retryAfter)
	if err == nil {
		// wait for the amount indicated by Retry-After and hit the server again
		return time.Duration(sleepDuration) * time.Second, nil
	}

	// if parsing as a string to int failed then it may be date
	retryDate, err := http.ParseTime(retryAfter)
	if err == nil {
		now := time.Now()
		// subtract the current time from the retry date
		sleepDuration := retryDate.Sub(now)
		return sleepDuration, nil
	}

	return 0, fmt.Errorf("Unable to parse Retry-After header: %q", retryAfter)
}

func main() {
	// make a request to the server
	serverURL := "http://localhost:8080"

	for true {
		// keep making a GET request unless we error out
		res, err := http.Get(serverURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "There was an error making a request to the server: %s", err)
			os.Exit(1)
		}

		// check status codes
		if res.StatusCode == 429 {
			sleepDuration, err := getSleepDuration(res.Header.Get("retry-after"))
			if err == nil {
				fmt.Println("Retrying again in: ", sleepDuration.Seconds())
			}
			time.Sleep(sleepDuration)
		} else if res.StatusCode == 200 {
			body, err := io.ReadAll(res.Body)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not read the response body: %q", err)
			}

			fmt.Printf("The weather forecast: %s\n", body)
			break
		}

	}
}
