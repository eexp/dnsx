package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/logrusorgru/aurora"
	"github.com/pkg/errors"

	"github.com/eexp/dnsx/internal/testutils"
)

var (
	debug   = os.Getenv("DEBUG") == "true"
	success = aurora.Green("[✓]").String()
	failed  = aurora.Red("[✘]").String()
	errored = false

	mainDnsxBinary = flag.String("main", "", "Main Branch Dnsx Binary")
	devDnsxBinary  = flag.String("dev", "", "Dev Branch Dnsx Binary")
	testcases      = flag.String("testcases", "", "Test cases file for dnsx functional tests")
)

func main() {
	flag.Parse()

	if err := runFunctionalTests(); err != nil {
		log.Fatalf("Could not run functional tests: %s\n", err)
	}
	if errored {
		os.Exit(1)
	}
}

func runFunctionalTests() error {
	file, err := os.Open(*testcases)
	if err != nil {
		return errors.Wrap(err, "could not open test cases")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}
		if err := runIndividualTestCase(text); err != nil {
			errored = true
			fmt.Fprintf(os.Stderr, "%s Test \"%s\" failed: %s\n", failed, text, err)
		} else {
			fmt.Printf("%s Test \"%s\" passed!\n", success, text)
		}
	}
	return nil
}

func runIndividualTestCase(testcase string) error {
	parts := strings.Fields(testcase)

	var finalArgs []string
	var target string
	if len(parts) > 1 {
		finalArgs = parts[2:]
		target = parts[0]
	}
	mainOutput, err := testutils.RunDnsxBinaryAndGetResults(target, *mainDnsxBinary, debug, finalArgs)
	if err != nil {
		return errors.Wrap(err, "could not run dnsx main test")
	}
	devOutput, err := testutils.RunDnsxBinaryAndGetResults(target, *devDnsxBinary, debug, finalArgs)
	if err != nil {
		return errors.Wrap(err, "could not run dnsx dev test")
	}
	if len(mainOutput) == len(devOutput) {
		return nil
	}
	return fmt.Errorf("%s main is not equal to %s dev", mainOutput, devOutput)
}
