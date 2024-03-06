package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

var timeout = flag.Int("timeout", 3, "Ex: 3")
var webhook = flag.String("webhook", "", "Ex: example.com/webhook")
var host = flag.String("host", "", "Ex: example.com")

type Notifier func(host string)

type WebhookPayload struct {
	Host string `json:"host"`
	Time string `json:"time"`
}

var start = time.Now()

func main() {
	flag.Parse()

	var notifier Notifier = desktopNotification

	if *host == "" {
		fmt.Println("Usage: dns-notify --host example.com")
		os.Exit(1)
	}

	if *webhook != "" {
		notifier = webhookNotification
	}

	fmt.Println("Target:", *host)
	fmt.Println("Timeout:", *timeout)

	for {

		if ping(*host) {
			fmt.Printf("Host %s found in %s\n", *host, time.Since(start))

			notifier(*host)

			os.Exit(0)
		}

		fmt.Println("Pinging host failed...")

		time.Sleep(time.Second * time.Duration(*timeout))
	}
}

func ping(url string) bool {
	if !strings.HasPrefix(url, "https://") && !strings.HasPrefix(url, "http://") {
		url = fmt.Sprintf("https://%s", url)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	client := &http.Client{}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}

	defer resp.Body.Close()

	return true
}

func desktopNotification(url string) {
	cmd := exec.Command("notify-send", "DNS Notify", fmt.Sprintf("Host %s is found!", url))

	err := cmd.Run()
	if err != nil {
		fmt.Println("Error executing command:", err)
		return
	}
}

func webhookNotification(url string) {
	webhookUrl := getWebhookUrl(*webhook)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	data, err := json.Marshal(WebhookPayload{
		Host: url,
		Time: time.Since(start).String(),
	})

	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "POST", webhookUrl, bytes.NewBuffer(data))

	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	fmt.Printf("Webhook %s has been notified\n", webhookUrl)
}

func getWebhookUrl(url string) string {
	if strings.HasPrefix(*webhook, "http://") || strings.HasPrefix(*webhook, "https://") {
		return *webhook
	}

	return fmt.Sprintf("https://%s", *webhook)
}
