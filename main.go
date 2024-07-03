package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

var visited_urls = make(map[string]bool)
var emails = make(map[string]bool)
var domain_urls = make(map[string]bool)

func scrape_page(page_url string) {
	log.Println("Scraping", page_url)

	visited_urls[page_url] = true

	resp, err := http.Get(page_url)
	if err != nil {
		fmt.Printf("Failed to scrape %s: %v\n", page_url, err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to read response body: %v\n", err)
		return
	}

	bodyStr := string(body)

	// Find all emails
	emailRegex := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	newEmails := emailRegex.FindAllString(bodyStr, -1)
	for _, email := range newEmails {
		log.Println("Found email", email)
		emails[email] = true
	}

	// Find all URLs of the same domain
	parsedUrl, err := url.Parse(page_url)
	if err != nil {
		fmt.Printf("Failed to parse URL %s: %v\n", page_url, err)
		return
	}
	baseUrl := fmt.Sprintf("%s://%s", parsedUrl.Scheme, parsedUrl.Host)

	linkRegex := regexp.MustCompile(`href="([^"]+)"`)
	matches := linkRegex.FindAllStringSubmatch(bodyStr, -1)
	for _, match := range matches {
		href := match[1]
		fullUrl := href
		if !strings.HasPrefix(href, "http") {
			fullUrl = baseUrl + href
		}
		if strings.Contains(fullUrl, baseUrl) && !visited_urls[fullUrl] {
			domain_urls[fullUrl] = true
		}
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run main.go <URL>")
		os.Exit(1)
	}

	startUrl := os.Args[1]
	domain_urls[startUrl] = true

	for len(domain_urls) > 0 {
		for currentUrl := range domain_urls {
			delete(domain_urls, currentUrl)
			if !visited_urls[currentUrl] {
				scrape_page(currentUrl)
			}
		}
	}

	// Save emails to a CSV file
	file, err := os.Create("emails.csv")
	if err != nil {
		fmt.Printf("Failed to create CSV file: %v\n", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"Email"})
	for email := range emails {
		writer.Write([]string{email})
	}
}
