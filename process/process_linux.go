package process

import (
	"bufio"
	"fmt"
	"github.com/polyverse/masche/common"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type linuxProcess struct {
	pid int
}

func (p *linuxProcess) Pid() int {
	return p.pid
}

func (p *linuxProcess) Name() (name string, harderror error, softerrors []error) {
	name, err := processExe(p.Pid())

	if err != nil {
		// If the exe link doesn't take us to the real path of the binary of the process maybe it's not present anymore
		// or the process didn't started from a file. We mimic this ps(1) trick and take the name form
		// /proc/<pid>/status in that case.

		statusPath := filepath.Join("/proc", fmt.Sprintf("%d", p.Pid()), "status")
		statusFile, err := os.Open(statusPath)
		if err != nil {
			return name, err, nil
		}
		defer statusFile.Close()

		r := bufio.NewReader(statusFile)
		for line, _, err := r.ReadLine(); err != io.EOF; line, _, err = r.ReadLine() {
			if err != nil {
				return name, err, nil
			}

			namePrefix := "Name:"
			if strings.HasPrefix(string(line), namePrefix) {
				name := strings.Trim(string(line[len(namePrefix):]), " \t")

				// We add the square brackets to be consistent with ps(1) output.
				return "[" + name + "]", nil, nil
			}
		}

		return name, fmt.Errorf("No name found for pid %v", p.Pid()), nil
	}

	return name, err, nil
}

func (p *linuxProcess) Handle() uintptr {
	return uintptr(p.pid)
}

func (p *linuxProcess) Info() (ProcessInfo, error) {
	statusPath := filepath.Join("/proc", fmt.Sprintf("%d", p), "status")
	statusFile, err := os.Open(statusPath)
	if err != nil {
		return nil, fmt.Errorf("Unable to open proc %d's status file at %s (%v)", p.Pid(), statusPath, err)
	}
	defer statusFile.Close()

	data, err := ioutil.ReadAll(statusFile)
	if err != nil {
		return nil, fmt.Errorf("Unable to read data from proc %d's status file at %s (%v)", p.Pid(), statusPath, err)
	}

	lpi := &linuxProcessInfo{}
	err = parseStatusToStruct(data, lpi)
	if err != nil {
		return nil, fmt.Errorf("Unable to process data from %s into LinuxProcessInfo struct (%v)", statusPath, err)
	}

	//we ignore this error
	lpi.Executable, err = processExe(p.Pid())
	if err != nil {
		fmt.Fprintf(os.Stderr, "[Warning] Error when expanding symlink to executable: %v\n", err)
	}

	return lpi, err
}

func getAllProcesses() ([]Process, error, []error) {
	files, err := ioutil.ReadDir("/proc/")
	if err != nil {
		return nil, err, nil
	}

	procs := []Process{}

	for _, f := range files {
		pid, err := strconv.Atoi(f.Name())
		if err != nil {
			continue
		}
		procs = append(procs, &linuxProcess{pid: pid})
	}

	return procs, nil, nil
}

func processFromPid(pid int) (Process, error, []error) {
	// Check if we have premissions to read the process memory
	procPath := common.ProcPathFromPid(uint(pid))
	_, err := os.Stat(procPath)
	if err != nil {
		return nil, fmt.Errorf("Error when testing existence of process with pid %v", pid), nil
	}

	return &linuxProcess{pid: pid}, nil, nil
}
