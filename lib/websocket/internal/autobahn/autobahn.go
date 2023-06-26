package autobahn

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

const (
	testResultInformational = `INFORMATIONAL`
	testResultNonStrict     = `NON-STRICT`
	testResultOK            = `OK`
)

type autobahnTestReport struct {
	Behaviour       string `json:"behavior"`
	BehaviourClose  string `json:"behaviorClose"`
	ReportFile      string `json:"reportfile"`
	Duration        int    `json:"duration"`
	RemoteCloseCode int    `json:"remoteCloseCode"`
}

func (tr *autobahnTestReport) isBehaviourSuccess() bool {
	if tr.Behaviour == testResultOK {
		return true
	}
	if tr.Behaviour == testResultInformational {
		return true
	}
	if tr.Behaviour == testResultNonStrict {
		return true
	}
	return false
}

func (tr *autobahnTestReport) isBehaviourCloseSuccess() bool {
	if tr.Behaviour == testResultOK {
		return true
	}
	if tr.Behaviour == testResultInformational {
		return true
	}
	if tr.Behaviour == testResultNonStrict {
		return true
	}
	return false

}

type autobahnCaseReport map[string]autobahnTestReport

type autobahnReport map[string]autobahnCaseReport

// PrintReports read the JSON reports from fileReportsJson and print the
// total test cases and the failed one.
func PrintReports(fileReportsJson string) {
	var (
		logp = `PrintReports`

		report autobahnReport
		raw    []byte
		err    error
	)

	raw, err = os.ReadFile(fileReportsJson)
	if err != nil {
		log.Fatalf(`%s: %s`, logp, err)
	}

	err = json.Unmarshal(raw, &report)
	if err != nil {
		log.Fatalf(`%s: %s`, logp, err)
	}

	var (
		name       string
		caseName   string
		listCase   autobahnCaseReport
		testReport autobahnTestReport
	)
	for name, listCase = range report {
		fmt.Printf("Test: %s\n", name)
		fmt.Printf("Total test cases: %d\n", len(listCase))

		var totalSuccess int

		for caseName, testReport = range listCase {
			if !testReport.isBehaviourSuccess() {
				fmt.Printf("  Test case %s: %s\n", caseName, testReport.Behaviour)
				continue
			}
			if !testReport.isBehaviourCloseSuccess() {
				fmt.Printf("  Test case %s: close: %s\n", caseName, testReport.BehaviourClose)
				continue
			}
			totalSuccess++
		}

		fmt.Printf("Total success: %d\n", totalSuccess)
		fmt.Printf("Total failed : %d\n", len(listCase)-totalSuccess)
	}
}
