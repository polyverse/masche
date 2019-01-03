package process

// #cgo CFLAGS: -std=c99
// #cgo CFLAGS: -DPSAPI_VERSION=1
// #cgo LDFLAGS: -lpsapi
// #include "process.h"
// #include "process_windows.h"
import "C"

import (
	"bytes"
	"fmt"
	"os/exec"
	"reflect"
	"strconv"
	"unsafe"

	"github.com/polyverse/masche/cresponse"
)

func (p process) Name() (name string, harderror error, softerrors []error) {
	var cname uintptr
	r := C.GetProcessName(p.hndl, (**C.char)(unsafe.Pointer(&cname)))

	harderror, softerrors = cresponse.GetResponsesErrors(unsafe.Pointer(r))
	C.response_free(r)
	if harderror == nil {
		name = C.GoString((*C.char)(unsafe.Pointer(cname)))
		C.free(unsafe.Pointer(cname))
	}
	return
}

func getAllPids() (pids []int, harderror error, softerrors []error) {
	r := C.getAllPids()
	defer C.EnumProcessesResponse_Free(r)
	if r.error != 0 {
		return nil, fmt.Errorf("getAllPids failed with error %d", r.error), nil
	}

	pids = make([]int, 0, r.length)
	// We use this to access C arrays without doing manual pointer arithmetic.
	cpids := *(*[]C.DWORD)(unsafe.Pointer(
		&reflect.SliceHeader{
			Data: uintptr(unsafe.Pointer(r.pids)),
			Len:  int(r.length),
			Cap:  int(r.length)}))
	for i, _ := range cpids {
		pid := int(cpids[i])
		// pids 0 and 4 are reserved in windows.
		if pid == 0 || pid == 4 {
			continue
		}
		pids = append(pids, pid)
	}

	return pids, nil, nil
}

type LinuxProcess int

func (p LinuxProcess) Pid() int {
	return int(p)
}

func (p LinuxProcess) Name() (name string, harderror error, softerrors []error) {
	name, err := ProcessExe(p.Pid())
	return name, err, nil
}

func (p LinuxProcess) Close() (harderror error, softerrors []error) {
	return nil, nil
}

func (p LinuxProcess) Handle() uintptr {
	wmicCommand := exec.Command("wmic", "path", "win32_process", "where", "processid=" +
		strconv.FormatUint(uint64(p), 10), "get", "handle")
	wmicOutput, err := wmicCommand.Output()
	if err != nil {
		return 0    // Can't return a negative since it's unsigned. Can't return an error. Not sure what to do here
					// besides panic or pretend nothing is wrong.
	}
	wmicLines := bytes.Split(wmicOutput, []byte("\n"))
	processLine := wmicLines[1]
	handleString := string(processLine)
	handle, _ := strconv.ParseUint(handleString, 10, 64)
	return uintptr(handle)
}