// this package provides functions to interact with the os processes
// You can list all the processes running on the os, filter them via a regexp
// and then use them from in other masche modules, because they are already open.
package process

import (
	"fmt"
	"regexp"
	"sort"
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

// OpenFromPid opens a process by its pid.
func ProcessFromPid(pid int) (Process, error, []error) {
	// This function is implemented by the OS-specific openFromPid function.
	return processFromPid(pid)
}

// GetAllPids returns a slice with al the running processes' pids.
func GetAllProcesses() ([]Process, error, []error) {
	softerrors := []error{}

	// This function is implemented by the OS-specific getAllPids function.
	allPids, harderror1, softerrors1 := getAllPids()
	softerrors = append(softerrors, softerrors1...)
	if harderror1 != nil {
		return nil, harderror1, softerrors
	} // if
	sort.Ints(allPids)

	procs := []Process{}
	for _, pid := range allPids {
		proc, harderr2, softerrors2 := ProcessFromPid(pid)
		softerrors = append(softerrors, softerrors2...)
		softerrors = append(softerrors,
			fmt.Errorf("Error occurred when getting Process from Pid %d: %v", pid, harderr2))

		procs = append(procs, proc)
	}

	return procs, nil, softerrors
}

// OpenByName recieves a Regexp an returns a slice with all the Processes whose name matches it.
func ProcessesByName(r *regexp.Regexp) (ps []Process, harderror error, softerrors []error) {
	procs, harderror, softerrors := GetAllProcesses()
	if harderror != nil {
		return nil, fmt.Errorf("", harderror), nil
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
