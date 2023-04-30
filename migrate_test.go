package migrate

import (
	"fmt"
	"golang.org/x/exp/slices"
	"testing"
)

func TestSort(t *testing.T) {
	var strs = []string{
		"AB.x",
		"AB-2.x",
	}
	slices.SortFunc(strs, func(a, b string) bool {
		return a < b
	})
	fmt.Printf("%#v\n", strs)
}
