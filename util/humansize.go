package util

import (
	"fmt"
	"math"
)

const (
	Byte uint64 = 1

	// base 1024
	Kb = Byte * 1024
	Mb = Kb * 1024
	Gb = Mb * 1024
	Tb = Gb * 1024
	Pb = Tb * 1024
	Eb = Pb * 1024
)

var (
	SuffixList = []string{
		"b", "Kb", "Mb", "Gb", "Tb", "Pb", // "Eb",
	}
	NumSuffix = len(SuffixList)
)

func FormatSize(size uint64) string {
	var n int
	var unit uint64
	var unit_name string

	base := uint64(1024)
	unit_name = SuffixList[0]

	for n = NumSuffix; n >= 0; n-- {
		unit = uint64(math.Pow(float64(base), float64(n)))
		if size > unit {
			unit_name = SuffixList[n]
			break
		}
	}
	// fmt.Printf("size %d,  unit %d (%s)\n", size, unit, unit_name)

	v := float64(size) / float64(unit)

	return fmt.Sprintf("%.2f%s", v, unit_name)
}

func HumanSize(size uint64) string {
	return fmt.Sprintf("%s (%d bytes)", FormatSize(size), size)
}
