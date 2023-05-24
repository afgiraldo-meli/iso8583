package encoding

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestText(t *testing.T) {
	enc := &textEncoder{}

	t.Run("Decode", func(t *testing.T) {
		res, read, err := enc.Decode([]byte("hello"), 5)

		require.NoError(t, err)
		require.Equal(t, []byte("hello"), res)
		require.Equal(t, 5, read)

		res, read, err = enc.Decode([]byte("hello, 世界!"), 14)

		require.NoError(t, err)
		require.Equal(t, []byte("hello, 世界!"), res)
		require.Equal(t, 14, read)

		_, _, err = enc.Decode([]byte("hello"), 6)
		require.Error(t, err)
		require.EqualError(t, err, "not enough data to decode. expected len 6, got 5")

		_, _, err = enc.Decode(nil, 6)
		require.Error(t, err)
		require.EqualError(t, err, "not enough data to decode. expected len 6, got 0")

		_, _, err = enc.Decode(nil, -1)
		require.Error(t, err)
		require.EqualError(t, err, "invalid length: -1")
	})

	t.Run("Encode", func(t *testing.T) {
		res, err := enc.Encode([]byte("hello"))

		require.NoError(t, err)
		require.Equal(t, []byte("hello"), res)

		res, err = enc.Encode([]byte("hello, 世界!"))

		require.NoError(t, err)
		require.Equal(t, []byte("hello, 世界!"), res)
	})
}
