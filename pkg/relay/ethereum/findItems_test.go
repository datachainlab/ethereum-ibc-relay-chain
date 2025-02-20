package ethereum

import (
	"fmt"
	"slices"
	"testing"
)

func TestFindItems(t *testing.T) {
	cases := []struct {
		size      int
		expect    int
		expectLog []int
	}{
		{
			size:      0,
			expect:    0,
			expectLog: []int{},
		},
		{
			size:      10,
			expect:    10,
			expectLog: []int{10},
		},
		{
			size:      10,
			expect:    9,
			expectLog: []int{10, 5, 8, 9},
		},
		{
			size:      10,
			expect:    0,
			expectLog: []int{10, 5, 3, 2, 1},
		},
		{
			size:      10,
			expect:    2,
			expectLog: []int{10, 5, 3, 2},
		},
	}

	for _, c := range cases {
		type Data struct {
			expect int
			log    []int
		}
		data := Data{expect: c.expect, log: make([]int, 0, c.size)}
		result, err := findItems(c.size, func(count int) error {
			data.log = append(data.log, count)
			if count <= data.expect {
				return nil
			} else {
				return fmt.Errorf("fail at count=%d", count)
			}
		})
		if c.expect == 0 {
			if err == nil {
				t.Errorf("findItems(%d,%d) unexpectedly returned %v", c.size, c.expect, result)
			}
		} else if err != nil {
			t.Errorf("findItems(%d,%d) unexpectedly failed: %v", c.size, c.expect, err)
		} else if result != c.expect {
			t.Errorf("findItems(%d,%d) has been mistakenly resulted in %v", c.size, c.expect, result)
		}
		if slices.Compare(data.log, c.expectLog) != 0 {
			t.Errorf("findItems(%d,%d) log unpexpected %v, expected=%v", c.size, c.expect, data.log, c.expectLog)
		}
	}
}
