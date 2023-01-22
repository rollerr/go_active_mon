package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

type websiteMetrics struct {
	Website     string
	Latency     time.Duration
	CheckTime   time.Time
	ErrorReason string
}

func websiteConnectionTest(website string) (*websiteMetrics, error) {
	timeout := time.Second
	start := time.Now()
	var metrics *websiteMetrics
	log.Printf("Checking connection to %s", website)
	conn, err := net.DialTimeout("tcp", website, timeout)
	if err != nil {
		log.Printf("Connection to %s failed: %s)", website, err)
		metrics = &websiteMetrics{Website: website, Latency: time.Since(start), CheckTime: time.Now(), ErrorReason: err.Error()}
	} else {
		defer conn.Close()
		log.Printf("Connection to %s succeeded (%s)", website, time.Since(start))
		metrics = &websiteMetrics{Website: website, Latency: time.Since(start), CheckTime: time.Now()}
		cloudWatchPublisher([]*websiteMetrics{metrics})
	}
	return metrics, err
}

func websiteRunner() {
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

func cloudWatchPublisher(websiteMetrics []*websiteMetrics) {
	// Publish to CloudWatch
	log.Printf("Publishing to CloudWatch %v", websiteMetrics[0].Latency.Seconds()*1000)
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc := cloudwatch.New(sess)

	metric := &cloudwatch.MetricDatum{
		MetricName: aws.String("WebsiteLatency"),
		Unit:       aws.String("Milliseconds"),
		Value:      aws.Float64(websiteMetrics[0].Latency.Seconds() * 1000),
	}
	input := &cloudwatch.PutMetricDataInput{
		MetricData: []*cloudwatch.MetricDatum{metric},
		Namespace:  aws.String("WebsiteLatencyAggregate"),
	}
	_, err := svc.PutMetricData(input)
	if err != nil {
		log.Printf("Error publishing to CloudWatch: %s", err)
	} else {
		log.Printf("Successfully published to CloudWatch")
	}
}

func main() {
	ticker := time.Tick(5 * time.Second)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	for {
		select {
		case <-ticker:
			websiteRunner()
		case <-c:
			log.Print("Stopping website runner")
			os.Exit(0)
		}
	}
}
