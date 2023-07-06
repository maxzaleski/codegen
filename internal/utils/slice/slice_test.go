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

func TestFilter(t *testing.T) {
	input := []int{1, 2, 3, 4, 5}
	expected := []int{2, 4}

	result := Filter(input, func(i int) bool {
		return i%2 == 0
	})
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, but got %v", expected, result)
	}
}

func TestMapKeys(t *testing.T) {
	input := map[string]int{
		"one":   1,
		"two":   2,
		"three": 3,
	}
	expected := []string{"one", "two", "three"}

	result := MapKeys(input)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, but got %v", expected, result)
	}
}
