package rule

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"
	"zscaler/core/probe"

	log "github.com/Sirupsen/logrus"
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
	ServiceName string
	Scale       Scaler
	Probe       probe.Probe
	RefreshRate time.Duration
}

// Check the probe, UP and DOWN at top and low quater
func (r Default) Check() error {
	probe := r.Probe.Value()
	log.Debug(fmt.Sprintf("["+r.ServiceName+"] "+r.Probe.Name()+"at %.2f\n", probe))
	if probe > 0.75 {
		r.Scale.Up()
	}
	if probe < 0.25 {
		r.Scale.Down()
	}
	return nil
}

// CheckInterval return the time to wait between each check
func (r Default) CheckInterval() time.Duration {
	return r.RefreshRate
}

// FloatValue handler
type FloatValue struct {
	ServiceName string
	Scale       Scaler
	Probe       probe.Probe
	RefreshRate time.Duration
	Up          func(v float64) bool
	Down        func(v float64) bool
}

// Check the probe, UP and DOWN
func (r FloatValue) Check() error {
	probe := r.Probe.Value()
	log.Debug(fmt.Sprintf("["+r.ServiceName+"] "+r.Probe.Name()+"at %.2f\n", probe))
	if r.Up(probe) && r.Down(probe) {
		log.Warning("[" + r.ServiceName + "] try to scale up and down at the same time! (nothing done)")
		return nil
	}
	if r.Up(probe) {
		r.Scale.Up()
	}
	if r.Down(probe) {
		r.Scale.Down()
	}
	return nil
}

// CheckInterval return the time to wait between each check
func (r FloatValue) CheckInterval() time.Duration {
	return r.RefreshRate
}

// Decode a logical rule (ex. ">0.75")
func Decode(order string) (func(float64) bool, error) {
	// check syntax
	regex := regexp.MustCompile("^[[:space:]]*([>|<|==|!=])[[:space:]]*([[:digit:]]*(?:\\.[[:digit:]]*)?)$")
	matches := regex.FindStringSubmatch(order)
	if len(matches) == 3 {
		value, err := strconv.ParseFloat(matches[2], 64)
		if err != nil {
			return nil, nil
		}
		switch matches[1] {
		case ">":
			return func(p float64) bool { return p > value }, nil
		case "<":
			return func(p float64) bool { return p < value }, nil
		case "==":
			return func(p float64) bool { return p == value }, nil
		case "!=":
			return func(p float64) bool { return p != value }, nil
		}
	}
	return func(p float64) bool { return false }, errors.New("Error decoding rule [" + order + "]")
}
