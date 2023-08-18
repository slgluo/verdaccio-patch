package test

import (
	"reflect"
	"testing"
	"verdaccio-patch/dependency"
)

func TestGetSortedVersions(t *testing.T) {
	t.Run("sort by version", func(t *testing.T) {
		var versionMap = map[string]string{
			"1.9.4": "2023-04-01T22:27:14.649Z",
			"1.9.5": "2023-04-01T22:27:14.659Z",
		}
		var expected = []string{"1.9.5", "1.9.4"}
		var actual = dependency.GetSortedVersions(versionMap)
		if !reflect.DeepEqual(expected, actual) {
			t.Error("error")
		}
	})

	t.Run("sort by time", func(t *testing.T) {
		var versionMap = map[string]string{
			"a":     "2023-04-01T22:27:14.649Z",
			"1.9.5": "2023-04-01T22:27:14.659Z",
		}
		var expected = []string{"1.9.5", "a"}
		var actual = dependency.GetSortedVersions(versionMap)
		if !reflect.DeepEqual(expected, actual) {
			t.Error("error")
		}
	})
}
