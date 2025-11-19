package adapters

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	leverBaseURL = "https://jobs.lever.co"
	leverSource  = "lever-html"
)

// FetchLever scrapes Lever-hosted job boards for the provided company slug.
func FetchLever(ctx context.Context, client *http.Client, ua, company string) ([]Job, error) {
	if client == nil {
		return nil, fmt.Errorf("http client is required")
	}

	company = strings.TrimSpace(company)
	if company == "" {
		return nil, fmt.Errorf("company cannot be empty")
	}

	url := fmt.Sprintf("%s/%s", leverBaseURL, company)
	doc, err := fetchDocument(ctx, client, ua, url)
	if err != nil {
		return nil, err
	}

	jobs := make([]Job, 0)
	doc.Find("div.posting").Each(func(_ int, node *goquery.Selection) {
		link := node.Find(".posting-title > a").First()
		title := strings.TrimSpace(link.Text())
		href, _ := link.Attr("href")
		href = ensureAbsoluteURL(leverBaseURL, href)

		if title == "" || href == "" {
			return
		}

		location := strings.TrimSpace(node.Find(".sort-by-location").Text())
		jobs = append(jobs, Job{
			Title:    title,
			Company:  company,
			URL:      href,
			Location: location,
			Source:   leverSource,
			Level:    inferLevel(title),
		})
	})

	return jobs, nil
}
