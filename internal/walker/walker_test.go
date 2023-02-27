package walker

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNeed2Skip(t *testing.T) {
	t.Parallel()

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

func TestGetSizeAndSum(t *testing.T) {
	t.Parallel()

	// prepare test file
	tF, err := os.CreateTemp("", "tmpfile-")
	assert.NoError(t, err)
	defer func() {
		_ = os.Remove(tF.Name())
	}()
	for i := 0; i < 100*1024; i++ {
		// nolint
		_, err = tF.Write([]byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Urna condimentum mattis pellentesque id. Et odio pellentesque diam volutpat commodo sed egestas. In dictum non consectetur a erat nam at lectus urna. Sapien et ligula ullamcorper malesuada proin. Eget mi proin sed libero enim sed. Nunc lobortis mattis aliquam faucibus purus in massa tempor. Nisl vel pretium lectus quam id leo. Amet mauris commodo quis imperdiet massa tincidunt nunc pulvinar. Amet consectetur adipiscing elit ut aliquam. Id semper risus in hendrerit gravida rutrum quisque. Nibh cras pulvinar mattis nunc sed blandit. Justo donec enim diam vulputate ut. Malesuada bibendum arcu vitae elementum curabitur vitae nunc. Nec dui nunc mattis enim ut tellus elementum sagittis vitae. Vitae tortor condimentum lacinia quis vel. Posuere sollicitudin aliquam ultrices sagittis orci a scelerisque. Tellus orci ac auctor augue. Mattis rhoncus urna neque viverra justo nec ultrices dui sapien."))
		assert.NoError(t, err)
	}
	tF.Close()
	sum, size, err := getSizeAndSum(tF.Name())
	assert.NoError(t, err)
	assert.Equal(t, sum, "6c62adc96b28bb8a141ca009f74ad345226c265806fe0eeecadcb524769f88c5")
	assert.Equal(t, size, uint64(0x63e7000))
}
