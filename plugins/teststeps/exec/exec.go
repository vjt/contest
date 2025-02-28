// Copyright (c) Facebook, Inc. and its affiliates.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package exec

import (
	"encoding/json"
	"fmt"

	"github.com/insomniacslk/xjson"
	"github.com/linuxboot/contest/pkg/event"
	"github.com/linuxboot/contest/pkg/event/testevent"
	"github.com/linuxboot/contest/pkg/test"
	"github.com/linuxboot/contest/pkg/xcontext"
	"github.com/linuxboot/contest/plugins/teststeps"
)

type stepParams struct {
	Bin struct {
		Path string   `json:"path"`
		Args []string `json:"args"`
		// TODO: add max execution timer
	} `json:"bin"`

	Transport struct {
		Proto   string          `json:"proto"`
		Options json.RawMessage `json:"options,omitempty"`
	} `json:"transport,omitempty"`

	OCPOutput bool `json:"ocp_output"`

	Constraints struct {
		TimeQuota xjson.Duration `json:"time_quota,omitempty"`
	} `json:"constraints,omitempty"`

	ExitCodeMap map[int]string `json:"exitcode_map,omitempty"`
}

// Name is the name used to look this plugin up.
var Name = "Exec"

// TestStep implementation for the exec plugin
type TestStep struct {
	stepParams
}

// Name returns the name of the Step
func (ts TestStep) Name() string {
	return Name
}

// Run executes the step.
func (ts *TestStep) Run(
	ctx xcontext.Context,
	ch test.TestStepChannels,
	ev testevent.Emitter,
	stepsVars test.StepsVariables,
	params test.TestStepParameters,
	resumeState json.RawMessage,
) (json.RawMessage, error) {
	if err := ts.populateParams(params); err != nil {
		return nil, err
	}

	tr := NewTargetRunner(ts, ev, stepsVars)
	return teststeps.ForEachTarget(Name, ctx, ch, tr.Run)
}

func (ts *TestStep) populateParams(stepParams test.TestStepParameters) error {
	bag := stepParams.GetOne("bag").JSON()

	if err := json.Unmarshal(bag, &ts.stepParams); err != nil {
		return fmt.Errorf("failed to deserialize parameters")
	}

	return nil
}

// ValidateParameters validates the parameters associated to the step
func (ts *TestStep) ValidateParameters(_ xcontext.Context, stepParams test.TestStepParameters) error {
	return ts.populateParams(stepParams)
}

// New initializes and returns a new exec step.
func New() test.TestStep {
	return &TestStep{}
}

// Load returns the name, factory and events which are needed to register the step.
func Load() (string, test.TestStepFactory, []event.Name) {
	return Name, New, Events
}
