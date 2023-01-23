package main

import (
	"log"
	"testing"
	"time"
)

func TestWebsiteRunner(t *testing.T) {
	websites := []string{"google.com:80", "amazon.com:80", "facebook.com:80", "cnn.com:80"}
	var websiteMetricList []*websiteMetrics
	for _, website := range websites {
		go func(website string) {
			log.Printf("Checking website %s", website)
			websiteMetric, err := websiteConnectionTest(website)
			if err != nil {
				log.Printf("Website %s is down", website)
			} else {
				websiteMetricList = append(websiteMetricList, websiteMetric)
				log.Printf("Website %s is up", website)
			}
		}(website)
		time.Sleep(10 * time.Millisecond)
	}
	for _, websiteMetric := range websiteMetricList {
		log.Printf("Website %s is up", websiteMetric.Website)
		log.Printf("Latency: %s", websiteMetric.Latency)
	}
}
