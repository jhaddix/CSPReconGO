# CspReconGo

CspReconGo is a command-line tool designed for cybersecurity analysts, web developers, and IT professionals. 

It automates the extraction and analysis of domains from Content Security Policy (CSP) headers and JavaScript files on websites. This tool is essential for conducting detailed web security audits, understanding external resource interactions, and monitoring changes in CSP and JavaScript-based domain references.

## Key Features

- **CSP Header Analysis:** Parses CSP headers to identify domains, helping users understand the website's security policies and external dependencies.
- **JavaScript File Analysis:** Automatically fetches and analyzes JavaScript files linked by the website, extracting domain references to reveal third-party integrations and external scripts.
- TODO **Domain Tracking Across Runs:** Compares results between runs, highlighting newly detected domains, which is invaluable for monitoring changes over time.
- **Structured Output:** Neatly organizes and documents the detected domains with a count of unique entries, outputting the results in a user-friendly format for further analysis.

## Getting Started

### Prerequisites

- **Go:** Ensure Go is installed on your system. You can download it from [Go's official download page](https://golang.org/dl/).

### Installation

1. **Download Source Code**: Clone the repository or download the source code to your local machine.

   ```shell
   git clone https://github.com/jhaddix/CspReconGo.git && cd CspReconGo
   ```
2. **Initialize a Go module in the directory**:
   ```
   go mod init CspReconGo.go
   ```
3. **Download Dependancies**:
   ```
   go get github.com/chromedp/chromedp
   go get github.com/chromedp/cdproto/network
   ```

### Use

1. **Use as**:

```
go run CspReconGo.go https://www.examplewebsite.com
```
