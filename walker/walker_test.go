package walker

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNeed2Skip(t *testing.T) {
	cases := []struct {
		Excludes *[]string
		Path     string
		Skip     bool
	}{
		{
			Excludes: &[]string{
				".*.go",
			},
			Path: "/home/test/Some/Test/test.go",
			Skip: true,
		},
		{
			Excludes: &[]string{
				".*/Some/.*",
			},
			Path: "/home/test/Some/Test/test.go",
			Skip: true,
		},
		{
			Excludes: &[]string{
				"NotFound",
			},
			Path: "/home/test/Some/Test/test.go",
			Skip: false,
		},
	}

	for i, c := range cases {
		fmt.Printf("Matching '%s' against '%#v'\n", c.Path, *c.Excludes)
		skip := need2skip(c.Path, c.Excludes)
		assert.Equal(t, skip, c.Skip,
			fmt.Sprintf("they should be equal in iteration %d", i))
	}
}
