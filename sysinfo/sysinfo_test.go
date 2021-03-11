package sysinfo

import (
	"reflect"
	"testing"
)

func TestSysInfo(t *testing.T) {
	type void struct{}
	var member void

	sysInfo := SysInfo()

	t.Logf(`SysInfo() = %v`, sysInfo)

	expectedKeys := make(map[string]void)
	for _, key := range []string{"os", "arch", "ncpu", "totalMemory", "freeMemory"} {
		expectedKeys[key] = member
	}

	receivedKeys := make(map[string]void)
	for key := range sysInfo {
		receivedKeys[key] = member
	}

	if !reflect.DeepEqual(expectedKeys, receivedKeys) {
		t.Fatalf(`SysInfo() keys = %v, want %v, error`, receivedKeys, expectedKeys)
	}
}
