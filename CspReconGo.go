package main

import (
    "context"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "regexp"
    "strings"
    "sync"
    "time"

    "github.com/chromedp/chromedp"
    "github.com/chromedp/cdproto/network"
)

func main() {
    // Check for correct usage by ensuring a URL is provided as an argument.
    if len(os.Args) < 2 {
        log.Fatalf("Usage: %s <URL>", os.Args[0])
    }
    targetURL := os.Args[1]

    // Initialize the ChromeDP headless browser context.
    ctx, cancel := chromedp.NewExecAllocator(context.Background(), chromedp.Flag("headless", true))
    defer cancel()

    ctx, cancel = chromedp.NewContext(ctx)
    defer cancel()

    // Set a timeout for the entire operation to avoid hanging indefinitely.
    ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
    defer cancel()

    var scriptURLs []string // To hold URLs of scripts found in the page.
    cspHeaders := make(map[string]bool) // To store unique CSP headers encountered.

    // Set up a listener to capture network requests and responses.
    chromedp.ListenTarget(ctx, func(ev interface{}) {
        switch evt := ev.(type) {
        case *network.EventRequestWillBeSent:
            // If the request is for a script, add its URL to scriptURLs.
            if evt.Type == network.ResourceTypeScript {
                scriptURLs = append(scriptURLs, evt.Request.URL)
            }
        case *network.EventResponseReceived:
            // Examine response headers for CSP directives, case-insensitively.
            for name, value := range evt.Response.Headers {
                if strings.Contains(strings.ToLower(name), "csp") {
                    if v, ok := value.(string); ok {
                        cspHeaders[v] = true
                    }
                }
            }
        }
    })

    // Navigate to the target URL and wait for the page to load.
    if err := chromedp.Run(ctx, network.Enable(), chromedp.Navigate(targetURL), chromedp.Sleep(2*time.Second)); err != nil {
        log.Fatalf("Failed to navigate: %v", err)
    }

    domains := make(map[string]struct{}) // A set to hold unique domains found.
    var wg sync.WaitGroup

    // Process CSP headers to extract domains and potential JS URLs.
    for cspHeader := range cspHeaders {
        jsURLs := extractJSURLsFromCSP(cspHeader)
        // Add any JS URLs found in CSP headers to scriptURLs for processing.
        scriptURLs = append(scriptURLs, jsURLs...)
        extractDomainsFromCSP(cspHeader, domains)
    }

    // Concurrently fetch and parse each JS file for domains.
    for _, url := range scriptURLs {
        wg.Add(1)
        go func(url string) {
            defer wg.Done()
            fmt.Printf("Fetching JS: %s\n", url)
            if ds, err := fetchAndParseJS(url); err == nil {
                for _, d := range ds {
                    domains[d] = struct{}{} // Add found domains to the set.
                }
            }
        }(url)
    }

    wg.Wait()

    // Print all unique domains detected.
    fmt.Println("\nDetected the following domains from CSP and Referenced JS:")
    for domain := range domains {
        fmt.Println(domain)
    }
}

// fetchAndParseJS downloads a JavaScript file and parses it for domains.
func fetchAndParseJS(url string) ([]string, error) {
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    return parseDomains(string(body)), nil
}

// parseDomains uses a regular expression to find domains in a given text (JavaScript code).
func parseDomains(js string) []string {
    domainRegex := regexp.MustCompile(`https?://[^\s/"'<>]+`)
    matches := domainRegex.FindAllString(js, -1)

    seen := make(map[string]struct{}) // To de-duplicate domains.
    var domains []string
    for _, match := range matches {
        if _, exists := seen[match]; !exists {
            seen[match] = struct{}{}
            domains = append(domains, match)
        }
    }
    return domains
}

// extractDomainsFromCSP parses a CSP header for domains.
func extractDomainsFromCSP(csp string, domains map[string]struct{}) {
    domainRegex := regexp.MustCompile(`https?://[^\s/"'<>]+`)
    matches := domainRegex.FindAllString(csp, -1)
    for _, match := range matches {
        domains[match] = struct{}{} // Add found domains to the set.
    }
}

// extractJSURLsFromCSP attempts to find URLs to JS files specifically in a CSP header string.
func extractJSURLsFromCSP(cspHeader string) []string {
    jsURLRegex := regexp.MustCompile(`https?://[^\s/"'<>]+\.js`)
    return jsURLRegex.FindAllString(cspHeader, -1)
}
