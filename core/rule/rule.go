package rule

import (
	"fmt"
	"time"
	"zscaler/core/probe"
)

// A Rule must be able to perform a check
type Rule interface {
	Check() error                 // performe a check on the target and act if needed
	CheckInterval() time.Duration // time to wait between each check
}

// Watcher check periodically the rule and report back errors
// TODO channel back to kill if needed
func Watcher(c chan error, r Rule) {
	for {
		err := r.Check()
		if err != nil {
			c <- err
			return
		}
		time.Sleep(r.CheckInterval())
	}
}

// Default provide a basic rule implementation
type Default struct {
	Target Service
	Probe  probe.Probe
}

// Service describes the object to scale
type Service struct {
	Name  string
	Scale Scaler
}

// Check the probe, UP and DOWN at top and low quater
func (r Default) Check() error {
	fmt.Printf("Checking probe... ")
	probe := r.Probe.Value()
	fmt.Printf("at %.2f\n", probe)
	if probe > 0.75 {
		r.Target.Scale.Up()
	}
	if probe < 0.25 {
		r.Target.Scale.Down()
	}
	return nil
}

// CheckInterval return the time to wait between each check
func (r Default) CheckInterval() time.Duration {
	return 3 * time.Second
}
