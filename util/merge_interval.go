package util

import "sort"

type Interval struct {
	Start int
	End   int
}

type IntervalSlice []Interval

func (s IntervalSlice) Len() int           { return len(s) }
func (s IntervalSlice) Less(i, j int) bool { return s[i].Start < s[j].Start }
func (s IntervalSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// 合并上传记录
func Merge(s IntervalSlice) []Interval {
	if len(s) <= 1 {
		return s
	}
	result := []Interval{}
	tmp := s[0]
	sort.Sort(s)
	for _, i := range s[1:] {
		if tmp.End >= i.Start-1 {
			tmp.End = i.End
		} else {
			result = append(result, tmp)
			tmp = i
		}
	}
	result = append(result, tmp)
	return result
}
