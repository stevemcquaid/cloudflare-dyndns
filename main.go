package main

import (
	"fmt"
	"log"
	"os"

	"errors"
	"flag"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/joho/godotenv"
)

var OLD_IP string
var DOMAIN string
var CF_API_KEY string
var CF_API_EMAIL string
var SUBDOMAIN string

func argParse() error {
	configfile := flag.String("config", "", "Absolute path to the config env file")
	flag.Parse()

	if *configfile != "" {
		// Load dotenv file into environment, overriding existing vars
		err := godotenv.Load(*configfile)
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	// Get vars from environment
	DOMAIN = os.Getenv("DOMAIN")
	if DOMAIN == "" {
		msg := fmt.Sprintf("Need to define DOMAIN var")
		return errors.New(msg)
	}
	CF_API_KEY = os.Getenv("CF_API_KEY")
	if CF_API_KEY == "" {
		msg := fmt.Sprintf("Need to define CF_API_KEY var")
		return errors.New(msg)
	}
	CF_API_EMAIL = os.Getenv("CF_API_EMAIL")
	if CF_API_EMAIL == "" {
		msg := fmt.Sprintf("Need to define CF_API_EMAIL var")
		return errors.New(msg)
	}
	SUBDOMAIN = os.Getenv("SUBDOMAIN")
	if SUBDOMAIN == "" {
		msg := fmt.Sprintf("Need to define SUBDOMAIN var")
		return errors.New(msg)
	}

	return nil
}

func main() {
	err := argParse()
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(os.Stdout)

	//OLD_IP = getMyIP(4)
	//dynDNS(OLD_IP)
	checkIP()

	log.Println("Entering Control Loop... ")
	for {
		time.Sleep(60 * time.Second)
		go checkIP()
	}
}

func checkIP() {
	log.Printf("Checking IP...\n")
	new_ip := getMyIP(4)
	if OLD_IP == "" {
		// First Run
		dynDNS(new_ip)
	} else if OLD_IP != new_ip {
		log.Printf("IP Address Changed: %s -> %s", OLD_IP, new_ip)
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
	log.Println("Set DNSRecord:", newRecord.Name, newRecord.Content, "\n")

	// Print records
	//showCurrentRecords(zoneID, api)
}

func updateRecord(zoneID string, api *cloudflare.API, newRecord *cloudflare.DNSRecord) {
	// Get current records
	//log.Println("Getting old dns records... ")
	dns := cloudflare.DNSRecord{Type: newRecord.Type, Name: newRecord.Name}
	old_records, err := api.DNSRecords(zoneID, dns)
	if err != nil {
		log.Fatal(err)
		return
	}

	if len(old_records) == 1 {
		// Update
		err := api.UpdateDNSRecord(zoneID, old_records[0].ID, *newRecord)
		if err != nil {
			log.Fatal(err)
			return
		}
		return
	}

	if len(old_records) > 1 {
		// Delete every record
		for _, record := range old_records {
			err := api.DeleteDNSRecord(zoneID, record.ID)
			if err != nil {
				log.Fatal(err)
				return
			}
			msg := fmt.Sprintf("Deleted DNSRecord: %s - %s: %s", record.Type, record.Name, record.Content)
			log.Println(msg)
		}
	}

	// Create if < 1 or > 1
	_, err = api.CreateDNSRecord(zoneID, *newRecord)
	if err != nil {
		log.Fatal(err)
		return
	}
	//log.Println("Done")
}

func showCurrentRecords(zoneID string, api *cloudflare.API) {
	// Fetch all DNS records for example.org
	records, err := api.DNSRecords(zoneID, cloudflare.DNSRecord{})
	if err != nil {
		log.Println(err)
		return
	}

	for _, r := range records {
		msg := fmt.Sprintf("%s: %s", r.Name, r.Content)
		log.Println(msg)
	}
}

func getMyIP(protocol int) string {
	var target string
	if protocol == 4 {
		target = "http://myexternalip.com/raw"
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
