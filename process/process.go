// this package provides functions to interact with the os processes
// You can list all the processes running on the os, filter them via a regexp
// and then use them from in other masche modules, because they are already open.
package process

import (
	"github.com/pkg/errors"
	"regexp"
)

// Process type represents a running processes that can be used by other modules.
// In order to get a Process on of the Open* functions must be called, and once it's not needed it must be closed.
type Process interface {
	// Pid returns the process' pid.
	Pid() int

	// Name returns the process' binary full path.
	Name() (string, error, []error)

	// Info
	Info() (ProcessInfo, error)
}

type ProcessInfo interface {
}

// OpenFromPid opens a process by its pid.
func ProcessFromPid(pid int) (Process, error, []error) {
	// This function is implemented by the OS-specific openFromPid function.
	return processFromPid(pid)
}

// GetAllPids returns a slice with al the running processes' pids.
func GetAllProcesses() ([]Process, error, []error) {
	softerrors := []error{}

	// This function is implemented by the OS-specific getAllPids function.
	allProcs, harderror1, softerrors1 := getAllProcesses()
	softerrors = append(softerrors, softerrors1...)
	if harderror1 != nil {
		return nil, harderror1, softerrors
	} // if
	return allProcs, nil, softerrors
}

// OpenByName recieves a Regexp an returns a slice with all the Processes whose name matches it.
func ProcessesByName(r *regexp.Regexp) (ps []Process, harderror error, softerrors []error) {
	procs, harderror, softerrors := GetAllProcesses()
	if harderror != nil {
		return nil, errors.Wrapf(harderror, "Unable to get Processes on this host"), nil
	}

	matchs := make([]Process, 0)

	for _, p := range procs {
		name, err, softs := p.Name()
		if err != nil {
			softerrors = append(softerrors, err)
		}
		if softs != nil {
			softerrors = append(softerrors, softs...)
		}

		if r.MatchString(name) {
			matchs = append(matchs, p)
		}
	}

	return matchs, nil, softerrors
}
