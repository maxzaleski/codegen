package slice

import (
	"reflect"
	"strconv"
	"testing"
)

func TestMap(t *testing.T) {
	input := []int{1, 2, 3, 4, 5}
	expected := []string{"1", "2", "3", "4", "5"}

	result := Map(input, func(i int) string {
		return strconv.Itoa(i)
	})
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, but got %v", expected, result)
	}
}
