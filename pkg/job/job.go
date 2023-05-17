// Copyright (c) Facebook, Inc. and its affiliates.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package job

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/linuxboot/contest/pkg/test"
	"github.com/linuxboot/contest/pkg/types"

	"github.com/insomniacslk/xjson"
)

// JobDescriptorMajorVersion, JobDescriptorMinorVersion are the current
// version of the job descriptor that the client must speaks to descripe jobs.
// It has two numbers to denote breaking and non-breaking changes
const JobDescriptorMajorVersion uint = 1
const JobDescriptorMinorVersion uint = 0

// Descriptor models the deserialized version of the JSON text given as
// input to the job creation request.
type Descriptor struct {
	JobName                     string
	Version                     string
	Tags                        []string
	Runs                        uint
	RunInterval                 xjson.Duration
	TestDescriptors             []*test.TestDescriptor
	Reporting                   Reporting
	TargetManagerAcquireTimeout *xjson.Duration // optional
	TargetManagerReleaseTimeout *xjson.Duration // optional
}

// Validate performs sanity checks on the job descriptor
func (d *Descriptor) Validate() error {

	if len(d.TestDescriptors) == 0 {
		return errors.New("need at least one TestDescriptor in the JobDescriptor")
	}
	if d.JobName == "" {
		return errors.New("job name cannot be empty")
	}
	if d.RunInterval < 0 {
		return errors.New("run interval must be non-negative")
	}

	if len(d.Reporting.RunReporters) == 0 && len(d.Reporting.FinalReporters) == 0 {
		return errors.New("at least one run reporter or one final reporter must be specified in a job")
	}
	for _, reporter := range d.Reporting.RunReporters {
		if strings.TrimSpace(reporter.Name) == "" {
			return errors.New("run reporters cannot have empty or all-whitespace names")
		}
	}
	return nil
}

// CheckVersion checks the compatibility of the received descriptor
// version against the supported one
func (d *Descriptor) CheckVersion() error {
	if d.Version == "" {
		return fmt.Errorf("version Error: Empty Job Descriptor Version Field")
	}
	// Convert the version string into 2 numbers
	versionNums := strings.Split(d.Version, ".")
	if len(versionNums) != 2 {
		return fmt.Errorf("version Error: Incorrect Job Descriptor Version %v", d.Version)
	}
	majorVersion, err := strconv.Atoi(versionNums[0])
	if err != nil {
		return fmt.Errorf("version Error: %w", err)
	}
	minorVersion, err := strconv.Atoi(versionNums[1])
	if err != nil {
		return fmt.Errorf("version Error: %w", err)
	}

	// checks the major, minor numbers against the supported version
	// If the major don't match of the minor is ahead of the currently supported
	// , return an error msg
	if majorVersion != int(JobDescriptorMajorVersion) || minorVersion > int(JobDescriptorMinorVersion) {
		return fmt.Errorf(
			"version Error: The Job Descriptor Version %s is not compatible with the server: %d.%d",
			d.Version,
			JobDescriptorMajorVersion,
			JobDescriptorMinorVersion,
		)
	}

	return nil
}

// CurrentDescriptorVersion returns current JobDescriptor version as a string
// e.g "1.0"
func CurrentDescriptorVersion() string {
	return fmt.Sprintf("%d.%d", JobDescriptorMajorVersion, JobDescriptorMinorVersion)
}

// ExtendedDescriptor is a job descriptor which has been extended with the full
// description of the test obtained from the test fetcher (which might not be a
// literal embedded in the job descriptor itself). A TestStepDescriptors object
// represents the test steps of a single test, associated to the test name.
// As one job might consist in multiple tests, the extended descriptor needs to
// capture a list of TestStepDescriptors, one for each test.
type ExtendedDescriptor struct {
	Descriptor
	TestStepsDescriptors []test.TestStepsDescriptors
}

// Job is used to run a type of test job on a given set of targets.
type Job struct {
	ID   types.JobID
	Name string
	// a freeform list of strings that the user can provide to tag a job, and
	// subsequently use to search and aggregate.
	Tags []string

	// How many times a job has to run. 0 means infinite.
	// A "run" is the execution of a sequence of tests. For example, setting
	// Runs to 2 will execute all the tests defined in `Tests` once, and then
	// will execute them again.
	Runs uint

	// RunInterval is the interval between multiple runs, if more than one, or
	// unlimited, are specified.
	RunInterval time.Duration

	// TargetManagerAcquireTimeout represents the maximum time that JobManager should wait for the execution of the Acquire function from the chosen TargetManager.
	TargetManagerAcquireTimeout time.Duration

	// TargetManagerReleaseTimeout represents the maximum time that JobManager should wait for the execution of the Release function from the chosen TargetManager.
	TargetManagerReleaseTimeout time.Duration

	// ExtendedDescriptor represents the descriptor submitted by the client that
	// resulted in the creation of this ConTest job.
	ExtendedDescriptor *ExtendedDescriptor

	// Tests represents the instantiated description of the tests
	Tests []*test.Test

	// RunReporterBundles and FinalReporterBundles wrap the reporter instances
	// chosen for the Job and its associated parameters, which have already
	// gone through validation
	RunReporterBundles   []*ReporterBundle
	FinalReporterBundles []*ReporterBundle
}

type State int

const (
	JobStateUnknown            State = iota
	JobStateStarted                  // 1
	JobStateCompleted                // 2
	JobStateFailed                   // 3
	JobStatePaused                   // 4
	JobStatePauseFailed              // 5
	JobStateCancelling               // 6
	JobStateCancelled                // 7
	JobStateCancellationFailed       // 8
)

func (js State) String() string {
	if js > 8 {
		return fmt.Sprintf("JobState%d", js)
	}
	return []string{
		"JobStateUnknown",
		string(EventJobStarted),
		string(EventJobCompleted),
		string(EventJobFailed),
		string(EventJobPaused),
		string(EventJobPauseFailed),
		string(EventJobCancelling),
		string(EventJobCancelled),
		string(EventJobCancellationFailed),
	}[js]
}

// InfoFetcher defines how to fetch job information
type InfoFetcher interface {
	FetchJob(types.JobID) (*Job, error)
	FetchJobs([]types.JobID) ([]*Job, error)
	FetchJobIDsByServerID(serverID string) ([]types.JobID, error)
}
