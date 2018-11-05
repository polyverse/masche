package process

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

type LinuxProcessInfo struct {
	Id              int    `json:"id" statusFileKey:"Pid"`
	Command         string `json:"command" statusFileKey:"Name"`
	UserId          int    `json:"userId" statusFileKey:"Uid"`
	UserName        string `json:"userName" statusFileKey:""`
	GroupId         int    `json:"groupId" statusFileKey:"Gid"`
	GroupName       string `json:"groupName" statusFileKey:""`
	ParentProcessId int    `json:"parentProcessId" statusFileKey:"PPid"`
	ThreadId        int    `json:"threadId" statusFileKey:""`
	SessionId       int    `json:"sessionId" statusFileKey:""`
}

var (
	tmpLpi         = LinuxProcessInfo{}
	keyToFieldName = map[string]string{}
	mtx            = &sync.RWMutex{}
)

func ProcessInfo(pid int) (*LinuxProcessInfo, error) {
	statusPath := filepath.Join("/proc", fmt.Sprintf("%d", pid), "status")
	statusFile, err := os.Open(statusPath)
	if err != nil {
		return nil, fmt.Errorf("Unable to open proc %d's status file at %s (%v)", pid, statusPath, err)
	}
	defer statusFile.Close()

	data, err := ioutil.ReadAll(statusFile)
	if err != nil {
		return nil, fmt.Errorf("Unable to read data from proc %d's status file at %s (%v)", pid, statusPath, err)
	}

	lpi := &LinuxProcessInfo{}
	err = parseStatusToStruct(data, lpi)
	return lpi, err
}

func parseStatusToStruct(data []byte, lpi *LinuxProcessInfo) error {
	if lpi == nil {
		return fmt.Errorf("Cannot parse Process Status into a nil LinuxProcessInfo")
	}

	r := bufio.NewReader(bytes.NewReader(data))
	for line, err := r.ReadString('\n'); err != io.EOF; line, err = r.ReadString('\n') {
		if err != nil {
			return fmt.Errorf("Error when parsing Status line from Proc Status data (%v)", err)
		}

		statusComponents := strings.Split(line, ":")
		if len(statusComponents) != 2 {
			continue
		}

		key := strings.TrimSpace(statusComponents[0])
		value := strings.TrimSpace(statusComponents[1])

		vals := strings.Fields(value)
		if len(vals) > 0 {
			value = vals[0]
		}

		fieldName := getFieldNameForKey(key)
		vfield := reflect.ValueOf(lpi).Elem().FieldByName(fieldName)
		if !vfield.IsValid() {
			continue //Nobody wants this value
		}

		val, err := stringToReflectValue(value, vfield.Type())
		if err != nil {
			return err
		}

		vfield.Set(val)
	}
	return nil
}

func stringToReflectValue(value string, t reflect.Type) (reflect.Value, error) {
	switch t.Name() {
	case "string":
		return reflect.ValueOf(value), nil
	case "int":
		intVal, err := strconv.Atoi(value)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("Error converting string %s into an integer. (%v)", value, err)
		}
		return reflect.ValueOf(intVal), nil
	}
	return reflect.Value{}, fmt.Errorf("Unsupported Converstion: string %s to value of type %v", value, t)
}

func getFieldNameForKey(key string) string {
	mtx.RLock()
	fieldName, ok := keyToFieldName[key]
	mtx.RUnlock()
	if ok {
		return fieldName
	}

	t := reflect.TypeOf(tmpLpi)
	fieldForKey, found := t.FieldByNameFunc(func(name string) bool {
		fieldCandidate, found := t.FieldByName(name)
		if found && fieldCandidate.Tag.Get("statusFileKey") == key {
			return true
		}
		return false
	})

	if !found {
		return ""
	}

	mtx.Lock()
	defer mtx.Unlock()
	keyToFieldName[key] = fieldForKey.Name
	return keyToFieldName[key]
}

// DEPRECATED
func deprecatedCommandLineProcessInfo(pid int) (*LinuxProcessInfo, []error) {
	errs := []error{}

	command, err := attributeByPid(pid, "comm")
	errs = appendError(errs, err, "Unable to get attribute comm for Pid %d", pid)

	userid, err := intAttributeByPid(pid, "uid")
	errs = appendError(errs, err, "Unable to get attribute uid for Pid %d", pid)

	username, err := attributeByPid(pid, "user")
	errs = appendError(errs, err, "Unable to get attribute user for Pid %d", pid)

	groupId, err := intAttributeByPid(pid, "gid")
	errs = appendError(errs, err, "Unable to get attribute gid for Pid %d", pid)

	groupName, err := attributeByPid(pid, "group")
	errs = appendError(errs, err, "Unable to get attribute group for Pid %d", pid)

	parentProcessId, err := intAttributeByPid(pid, "ppid")
	errs = appendError(errs, err, "Unable to get attribute ppid for Pid %d", pid)

	threadId, err := intAttributeByPid(pid, "tid")
	errs = appendError(errs, err, "Unable to get attribute tid for Pid %d", pid)

	sessionId, err := intAttributeByPid(pid, "sid")
	errs = appendError(errs, err, "Unable to get attribute sid for Pid %d", pid)

	return &LinuxProcessInfo{
		Id:              pid,
		Command:         command,
		UserId:          userid,
		UserName:        username,
		GroupId:         groupId,
		GroupName:       groupName,
		ParentProcessId: parentProcessId,
		ThreadId:        threadId,
		SessionId:       sessionId,
	}, errs
}

// TODO: Replace with non-command
func attributeByPid(pid int, attribute string) (string, error) {
	ret, err := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", attribute+"=").Output()
	if err != nil {
		return "", fmt.Errorf("Unable to get process attribute %s for pid %d, by calling the 'ps' command. (%v)", attribute, pid, err)
	}
	return strings.Replace(string(ret), "\n", "", 1), nil
}

func intAttributeByPid(pid int, attribute string) (int, error) {
	strVal, err := attributeByPid(pid, attribute)
	if err != nil {
		return 0, fmt.Errorf("Unable to get integer value of attribute %s for pid %d due to underlying error. (%v)", attribute, pid, err)
	}

	ret, err := strconv.Atoi(strings.TrimSpace(strVal))
	if err != nil {
		return 0, fmt.Errorf("Unable to parse integer from value %s of attribute %s for pid %d due to underlying error. (%v)", strVal, attribute, pid, err)
	}

	return ret, nil
}

func appendError(errs []error, err error, format string, params ...interface{}) []error {
	if err == nil {
		return errs
	}

	params = append(params, err)
	wrappedErr := fmt.Errorf(format+" (%v)", params...)
	errs = append(errs, wrappedErr)
	return errs
}
