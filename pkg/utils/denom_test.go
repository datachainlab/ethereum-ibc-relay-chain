package utils

import (
	"testing"
)

func TestParseEtherAmount(t *testing.T) {
	cases := []struct {
		data     string
		expected string
		isErr    bool
	}{
		{
			data:     "123wei",
			expected: "123",
		},
		{
			data:     "123gwei",
			expected: "123000000000",
		},
		{
			data:     "123ether",
			expected: "123000000000000000000",
		},
		{
			data:  "123eth",
			isErr: true,
		},
		{
			data:  "a123wei",
			isErr: true,
		},
		{
			data:  "gwei",
			isErr: true,
		},
		{
			data:  "eth",
			isErr: true,
		},
	}

	for _, c := range cases {
		a, err := ParseEtherAmount(c.data)
		if c.isErr {
			if err == nil {
				t.Errorf("ParseEtherAmount(%s) unexpectedly returned %v", c.data, a)
			}
		} else if err != nil {
			t.Errorf("ParseEtherAmount(%s) unexpectedly failed: %v", c.data, err)
		} else if a.String() != c.expected {
			t.Errorf("%s has been mistakenly parsed to %v, expected=%s", c.data, a, c.expected)
		}
	}
}
