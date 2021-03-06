package process

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/polyverse/masche/test"
)

func TestOpenFromPid(t *testing.T) {
	cmd, err := test.LaunchTestCase()
	if err != nil {
		t.Fatal(err)
	}
	defer cmd.Process.Kill()

	pid := int(cmd.Process.Pid)
	proc, err, softerrors := OpenFromPid(pid)
	defer proc.Close()
	test.PrintSoftErrors(softerrors)
	if err != nil {
		t.Fatal(err)
	}
}

func TestProcessName(t *testing.T) {
	cmd, err := test.LaunchTestCase()
	if err != nil {
		t.Fatal(err)
	}
	defer cmd.Process.Kill()

	pid := int(cmd.Process.Pid)
	proc, err, softerrors := OpenFromPid(pid)
	defer proc.Close()
	test.PrintSoftErrors(softerrors)
	if err != nil {
		t.Fatal(err)
	}

	name, err, softerrors := proc.Name()
	test.PrintSoftErrors(softerrors)
	if err != nil {
		t.Fatal(err)
	}

	if name != test.GetTestCasePath() {
		t.Error("Expected name", test.GetTestCasePath(), "and got", name)
	}
}

func TestOpenByName(t *testing.T) {
	cmd, err := test.LaunchTestCase()
	if err != nil {
		t.Fatal(err)
	}
	defer cmd.Process.Kill()

	r := regexp.MustCompile("test[/\\\\]tools[/\\\\]test")
	procs, err, softerrors := OpenByName(r)
	defer CloseAll(procs)
	test.PrintSoftErrors(softerrors)
	if len(procs) == 0 {
		t.Error("The test case was launched and not opened.")
	}

	for _, proc := range procs {
		name, err, softerrors := proc.Name()
		test.PrintSoftErrors(softerrors)
		if err != nil {
			t.Fatal(err)
		}

		if name != test.GetTestCasePath() {
			t.Error("Expected name", test.GetTestCasePath(), "and got", name)
		}
	}
}

func TestProcessInfo(t *testing.T) {
	cmd, err := test.LaunchTestCase()
	if err != nil {
		t.Fatal(err)
	}
	defer cmd.Process.Kill()

	pid := int(cmd.Process.Pid)
	procInfo, err := ProcessInfo(pid)
	if err != nil {
		t.Fatalf("Error when calling ProcInfo: %v", err)
	}

	fmt.Printf("ProcessInfo: %+v\n", procInfo)
}
