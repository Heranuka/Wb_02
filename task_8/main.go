package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/beevik/ntp"
)

func main() {
	currentTime, err := getNTPTime("pool.ntp.org")
	if err != nil {
		log.Printf("Error getting NTP time: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(currentTime.Format(time.RFC3339))
}

func getNTPTime(server string) (time.Time, error) {
	ntpTime, err := ntp.Time(server)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to query NTP server %q: %w", server, err)
	}

	return ntpTime, nil
}
