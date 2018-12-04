package process

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

type linuxProcessInfo struct {
	Id              int    `json:"id" statusFileKey:"Pid"`
	Command         string `json:"command" statusFileKey:"Name"`
	UserId          int    `json:"userId" statusFileKey:"Uid"`
	UserName        string `json:"userName" statusFileKey:""`
	GroupId         int    `json:"groupId" statusFileKey:"Gid"`
	GroupName       string `json:"groupName" statusFileKey:""`
	ParentProcessId int    `json:"parentProcessId" statusFileKey:"PPid"`
	Executable      string `json:"executable"`
}

var (
	tmpLpi         = linuxProcessInfo{}
	keyToFieldName = map[string]string{}
	mtx            = &sync.RWMutex{}
)

func processExe(pid int) (string, error) {
	exePath := filepath.Join("/proc", fmt.Sprintf("%d", pid), "exe")
	name, err := filepath.EvalSymlinks(exePath)
	if err != nil {
		return "", fmt.Errorf("Unable to expand process executable symlink %s (%v)", exePath, err)
	}
	return name, nil
}

func parseStatusToStruct(data []byte, lpi *linuxProcessInfo) error {
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
