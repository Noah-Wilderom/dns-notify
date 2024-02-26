package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

var timeout = flag.Int("timeout", 3, "Ex: 3")

var host = flag.String("host", "", "Ex: example.com")

func main() {
	flag.Parse()

	if *host == "" {
		fmt.Println("Usage: dns-notify --host example.com")
		os.Exit(1)
	}

	fmt.Println("Target:", *host)
	fmt.Println("Timeout:", *timeout)

	for {

		if ping(*host) {
			fmt.Printf("Host %s found\n", *host)
			cmd := exec.Command("notify-send", "DNS Notify", fmt.Sprintf("Host %s is found!", *host))

			err := cmd.Run()
			if err != nil {
				fmt.Println("Error executing command:", err)
				return
			}

			os.Exit(0)
		}

		fmt.Println("Pinging host failed...")

		time.Sleep(time.Second * time.Duration(*timeout))
	}
}

func ping(host string) bool {
	if !strings.HasPrefix(host, "https://") && !strings.HasPrefix(host, "http://") {
		host = fmt.Sprintf("https://%s", host)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	client := &http.Client{}

	req, err := http.NewRequestWithContext(ctx, "GET", host, nil)
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
