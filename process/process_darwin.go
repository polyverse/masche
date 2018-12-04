package process

import "C"
import (
	"bytes"
	"encoding/binary"
	"github.com/pkg/errors"
	"syscall"
	"unsafe"
)

type darwinProcess struct {
	pid    int
	ppid   int
	binary string
}

func (p *darwinProcess) Pid() int {
	return p.pid
}

func (p *darwinProcess) Name() (string, error, []error) {
	return p.binary, nil, nil
}

func (p *darwinProcess) Info() (ProcessInfo, error) {
	return &darwinProcessInfo{}, nil
}

func processFromPid(pid int) (Process, error, []error) {
	ps, err, softerrors := getAllProcesses()
	if err != nil {
		return nil, err, softerrors
	}

	for _, p := range ps {
		if p.Pid() == pid {
			return p, nil, softerrors
		}
	}

	return nil, nil, softerrors
}

func getAllProcesses() ([]Process, error, []error) {
	buf, err := darwinSyscall()
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to get all processes on darwin"), nil
	}

	procs := make([]*kinfoProc, 0, 50)
	k := 0
	for i := _KINFO_STRUCT_SIZE; i < buf.Len(); i += _KINFO_STRUCT_SIZE {
		proc := &kinfoProc{}
		err = binary.Read(bytes.NewBuffer(buf.Bytes()[k:i]), binary.LittleEndian, proc)
		if err != nil {
			return nil, errors.Wrapf(err, "Unable to read process info from darwin syscall"), nil
		}

		k = i
		procs = append(procs, proc)
	}

	darwinProcs := make([]Process, len(procs))
	for i, p := range procs {
		darwinProcs[i] = &darwinProcess{
			pid:    int(p.Pid),
			ppid:   int(p.PPid),
			binary: darwinCstring(p.Comm),
		}
	}

	return darwinProcs, nil, nil
}

func darwinCstring(s [16]byte) string {
	i := 0
	for _, b := range s {
		if b != 0 {
			i++
		} else {
			break
		}
	}

	return string(s[:i])
}

func darwinSyscall() (*bytes.Buffer, error) {
	mib := [4]int32{_CTRL_KERN, _KERN_PROC, _KERN_PROC_ALL, 0}
	size := uintptr(0)

	_, _, errno := syscall.Syscall6(
		syscall.SYS___SYSCTL,
		uintptr(unsafe.Pointer(&mib[0])),
		4,
		0,
		uintptr(unsafe.Pointer(&size)),
		0,
		0)

	if errno != 0 {
		return nil, errno
	}

	bs := make([]byte, size)
	_, _, errno = syscall.Syscall6(
		syscall.SYS___SYSCTL,
		uintptr(unsafe.Pointer(&mib[0])),
		4,
		uintptr(unsafe.Pointer(&bs[0])),
		uintptr(unsafe.Pointer(&size)),
		0,
		0)

	if errno != 0 {
		return nil, errno
	}

	return bytes.NewBuffer(bs[0:size]), nil
}

const (
	_CTRL_KERN         = 1
	_KERN_PROC         = 14
	_KERN_PROC_ALL     = 0
	_KINFO_STRUCT_SIZE = 648
)

type kinfoProc struct {
	_    [40]byte
	Pid  int32
	_    [199]byte
	Comm [16]byte
	_    [301]byte
	PPid int32
	_    [84]byte
}
