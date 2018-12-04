package memaccess

import (
	"errors"
	"github.com/polyverse/masche/process"
)

func nextMemoryRegion(p process.Process, address uintptr) (region MemoryRegion, harderror error, softerrors []error) {
	return MemoryRegion{}, errors.New("Not implemented yet"), nil
}

func copyMemory(p process.Process, address uintptr, buffer []byte) (harderror error, softerrors []error) {
	return errors.New("Not implemented yet"), nil
}
