package cnb

import (
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestHello(t *testing.T) {
	key := okay()

	require.Equal(t, key, okay())
	require.Equal(t, key, okay())
	require.Equal(t, key, okay())
	require.Equal(t, key, okay())
	require.Equal(t, key, okay())
	require.Equal(t, key, okay())
	require.Equal(t, key, okay())
	require.Equal(t, key, okay())
	require.Equal(t, key, okay())
	require.Equal(t, key, okay())
	require.Equal(t, key, okay())
	require.Equal(t, key, okay())
	require.Equal(t, key, okay())
	require.Equal(t, key, okay())
	require.Equal(t, key, okay())
	require.Equal(t, key, okay())
	require.Equal(t, key, okay())
	require.Equal(t, key, okay())
	require.Equal(t, key, okay())
	require.Equal(t, key, okay())
	require.Equal(t, key, okay())
	require.Equal(t, key, okay())
	require.Equal(t, key, okay())
	require.Equal(t, key, okay())
	require.Equal(t, key, okay())
	require.Equal(t, key, okay())
	require.Equal(t, key, okay())
	require.Equal(t, key, okay())
	require.Equal(t, key, okay())
	require.Equal(t, key, okay())
	require.Equal(t, key, okay())
	require.Equal(t, key, okay())
	require.Equal(t, key, okay())
	require.Equal(t, key, okay())
	require.Equal(t, key, okay())
	require.Equal(t, key, okay())
	require.Equal(t, key, okay())
	require.Equal(t, key, okay())
	require.Equal(t, key, okay())
	require.Equal(t, key, okay())
}

func okay() string {
	myMap := map[string]string{
		"hello":   "okay",
		"hello1":  "okay",
		"hello2":  "okay",
		"hello3":  "okay",
		"hello4":  "okay",
		"hello5":  "okay",
		"hello6":  "okay",
		"hello7":  "okay",
		"hello8":  "okay",
		"hello9":  "okay",
		"hello10": "okay",
		"hello11": "okay",
		"hello12": "okay",
		"hello13": "okay",
		"hello14": "okay",
		"hello15": "okay",
	}

	builder := strings.Builder{}
	for key := range myMap {
		builder.WriteString(key)
	}

	return builder.String()

}
