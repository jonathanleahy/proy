package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/jonathanleahy/prroxy/reporter/internal/config"
	"github.com/jonathanleahy/prroxy/reporter/internal/reporter"
)

func main() {
	// Parse command-line flags
	configFile := flag.String("config", "config.json", "Path to configuration file")
	outputFormat := flag.String("format", "markdown", "Output format: json, markdown")
	outputFile := flag.String("output", "", "Output file (default: stdout)")
	maxFailures := flag.Int("max-failures", 10, "Stop after N failures/mismatches (default: 10, 0 for no limit)")
	outputDir := flag.String("output-dir", "", "Directory for individual test results")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Create and run reporter
	r := reporter.NewReporter(cfg, *outputDir, *maxFailures)
	fmt.Fprintf(os.Stderr, "Running tests (will stop after %d failures)...\n", *maxFailures)

	report := r.Run()

	// Format output
	var output string
	switch *outputFormat {
	case "json":
		output, err = reporter.FormatJSON(report)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error formatting JSON: %v\n", err)
			os.Exit(1)
		}
	case "markdown":
		output = reporter.FormatMarkdown(report)
	default:
		fmt.Fprintf(os.Stderr, "Unknown format: %s\n", *outputFormat)
		os.Exit(1)
	}

	// Write output
	if *outputFile != "" {
		err = os.WriteFile(*outputFile, []byte(output), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Report written to %s\n", *outputFile)
	} else {
		fmt.Println(output)
	}

	// Exit with error code if there were failures
	if report.FailedEndpoints > 0 {
		fmt.Fprintf(os.Stderr, "\n❌ %d/%d endpoints failed\n", report.FailedEndpoints, report.TotalEndpoints)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "\n✅ All %d endpoints matched\n", report.MatchedEndpoints)
}
