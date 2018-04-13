package main

import (
	"fmt"
	"log"
	"os"

	"github.com/cloudflare/cloudflare-go"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

var OLD_IP string
var DOMAIN string
var CF_API_KEY string
var CF_API_EMAIL string
var SUBDOMAIN string

func main() {
	DOMAIN = os.Getenv("DOMAIN")
	CF_API_KEY = os.Getenv("CF_API_KEY")
	CF_API_EMAIL = os.Getenv("CF_API_EMAIL")
	SUBDOMAIN = os.Getenv("SUBDOMAIN")

	OLD_IP = getMyIP(4)
	dynDNS(OLD_IP)

	fmt.Println("Entering Control Loop... ")
	for {
		time.Sleep(60 * time.Second)
		go checkIP()
	}
}

func checkIP() {
	fmt.Println("Checking IP...")
	new_ip := getMyIP(4)
	if OLD_IP != new_ip {
		fmt.Sprintln("IP Address Changed: %s -> %s", OLD_IP, new_ip)
		dynDNS(new_ip)
	}
	OLD_IP = new_ip
}

func dynDNS(ip string) {
	// Construct a new API object
	api, err := cloudflare.New(CF_API_KEY, CF_API_EMAIL)
	if err != nil {
		log.Fatal(err)

	}

	// Fetch the zone ID
	zoneID, err := api.ZoneIDByName(DOMAIN) // Assuming example.com exists in your Cloudflare account already
	if err != nil {
		log.Fatal(err)
		return
	}

	// Record to create
	newRecord := cloudflare.DNSRecord{
		Type:    "A",
		Name:    SUBDOMAIN + "." + DOMAIN,
		Content: getMyIP(4),
	}

	updateRecord(zoneID, api, &newRecord)
	fmt.Println("Set DNSRecord:", newRecord.Name, newRecord.Content, "\n")

	// Print records
	//showCurrentRecords(zoneID, api)
}

func updateRecord(zoneID string, api *cloudflare.API, newRecord *cloudflare.DNSRecord) {
	// Get current records
	//fmt.Println("Getting old dns records... ")

	dns := cloudflare.DNSRecord{Type: newRecord.Type, Name: newRecord.Name}
	old_records, err := api.DNSRecords(zoneID, dns)
	if err != nil {
		fmt.Println(err)
		return
	}

	if len(old_records) == 1 {
		// Update
		err := api.UpdateDNSRecord(zoneID, old_records[0].ID, *newRecord)
		if err != nil {
			fmt.Println(err)
			return
		}
		return
	}

	if len(old_records) > 1 {
		// Delete every record
		for _, record := range old_records {
			err := api.DeleteDNSRecord(zoneID, record.ID)
			if err != nil {
				fmt.Println(err)
				return
			}
			msg := fmt.Sprintf("Deleted DNSRecord: %s - %s: %s", record.Type, record.Name, record.Content)
			fmt.Println(msg)
		}
	}

	// Create if < 1 or > 1
	_, err = api.CreateDNSRecord(zoneID, *newRecord)
	if err != nil {
		fmt.Println(err)
		return
	}
	//fmt.Println("Done")
}

func showCurrentRecords(zoneID string, api *cloudflare.API) {
	// Fetch all DNS records for example.org
	records, err := api.DNSRecords(zoneID, cloudflare.DNSRecord{})
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, r := range records {
		msg := fmt.Sprintf("%s: %s", r.Name, r.Content)
		fmt.Println(msg)
	}
}

func getMyIP(protocol int) string {
	var target string
	if protocol == 4 {
		target = "http://ipv4.myexternalip.com/raw"

	} else if protocol == 6 {
		target = "http://ipv6.myexternalip.com/raw"

	} else {
		return ""

	}
	resp, err := http.Get(target)
	if err == nil {
		contents, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			defer resp.Body.Close()
			return strings.TrimSpace(string(contents))

		}

	}
	return ""
}
