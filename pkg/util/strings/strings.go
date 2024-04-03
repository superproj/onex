// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package strings

import (
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/asaskevich/govalidator"
)

type frequencyInfo struct {
	s         string
	frequency int
}

type frequencyInfoSlice []frequencyInfo

func (fi frequencyInfoSlice) Len() int {
	return len(fi)
}

func (fi frequencyInfoSlice) Swap(i, j int) {
	fi[i], fi[j] = fi[j], fi[i]
}

func (fi frequencyInfoSlice) Less(i, j int) bool {
	return fi[j].frequency > fi[i].frequency
}

// Creates an slice of slice values not included in the other given slice.
func Diff(base, exclude []string) (result []string) {
	excludeMap := make(map[string]bool)
	for _, s := range exclude {
		excludeMap[s] = true
	}
	for _, s := range base {
		if !excludeMap[s] {
			result = append(result, s)
		}
	}

	return result
}

// Creates an slice of slice values included in the other given slice.
func Include(base, include []string) (result []string) {
	baseMap := make(map[string]bool)
	for _, s := range base {
		baseMap[s] = true
	}
	for _, s := range include {
		if baseMap[s] {
			result = append(result, s)
		}
	}

	return result
}

func Unique(ss []string) (result []string) {
	smap := make(map[string]bool)
	for _, s := range ss {
		smap[s] = true
	}
	for s := range smap {
		result = append(result, s)
	}

	return result
}

func CamelCaseToUnderscore(str string) string {
	return govalidator.CamelCaseToUnderscore(str)
}

func UnderscoreToCamelCase(str string) string {
	return govalidator.UnderscoreToCamelCase(str)
}

func FindString(array []string, str string) int {
	for index, s := range array {
		if str == s {
			return index
		}
	}

	return -1
}

func StringIn(str string, array []string) bool {
	return FindString(array, str) > -1
}

func Reverse(s string) string {
	size := len(s)
	buf := make([]byte, size)
	for start := 0; start < size; {
		r, n := utf8.DecodeRuneInString(s[start:])
		start += n
		utf8.EncodeRune(buf[size-start:], r)
	}

	return string(buf)
}

// Filter filters a list for a string.
func Filter(list []string, strToFilter string) (newList []string) {
	for _, item := range list {
		if item != strToFilter {
			newList = append(newList, item)
		}
	}

	return
}

// Contains returns true if a list contains a string.
func Contains(list []string, strToSearch string) bool {
	for _, item := range list {
		if item == strToSearch {
			return true
		}
	}

	return false
}

func FrequencySort(list []string) []string {
	cnt := map[string]int{}

	for _, s := range list {
		cnt[s]++
	}

	infos := make([]frequencyInfo, 0, len(cnt))
	for s, c := range cnt {
		infos = append(infos, frequencyInfo{
			s:         s,
			frequency: c,
		})
	}

	sort.Sort(frequencyInfoSlice(infos))

	ret := make([]string, 0, len(infos))
	for _, info := range infos {
		ret = append(ret, info.s)
	}

	return ret
}

// ContainsEqualFold returns true if a given slice 'slice' contains string 's' under unicode case-folding.
func ContainsEqualFold(slice []string, s string) bool {
	for _, item := range slice {
		if strings.EqualFold(item, s) {
			return true
		}
	}

	return false
}
