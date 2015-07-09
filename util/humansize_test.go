package util

import (
	"fmt"
	"testing"
)

func TestHumanSize(t *testing.T) {

	values := map[uint64]string{
		1:                                      "1.00b",
		1080:                                   "1.05Kb",
		1024 * 1024 * 5:                        "5.00Mb",
		(1024 * 1024 * 1024 * 11) + 1048576*50: "11.05Gb",
	}

	for bytes, expected := range values {
		res := HumanSize(bytes)

		if expected != res {
			t.Errorf("bytse=%d: got %s, expected %s", bytes, res, expected)
		}
		fmt.Printf("%v -> %v\n", bytes, res)
	}

}
