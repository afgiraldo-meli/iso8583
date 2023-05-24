package prefix

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_BackslashPrefix_EncodeLength(t *testing.T) {
	testCases := []struct {
		name    string
		maxLen  int
		dataLen int
		wantErr string
	}{
		{
			name:    "Success",
			maxLen:  2,
			dataLen: 2,
		},
		{
			name:    "Error_When_MaxLenAchieved",
			maxLen:  2,
			dataLen: 3,
			wantErr: "field length: 3 is larger than maximum: 2",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b := NewDelimiter('\x5C', "ASCIIBackslash")

			data, err := b.EncodeLength(tc.maxLen, tc.dataLen)
			if err != nil || tc.wantErr != "" {
				assert.EqualError(t, err, tc.wantErr)
				return
			}

			assert.Equal(t, []byte{}, data)
		})
	}
}

func Test_BackslashPrefix_DecodeLength(t *testing.T) {
	testCases := []struct {
		name    string
		maxLen  int
		data    []byte
		wantLen int
		wantErr string
	}{
		{
			name:    "Success_When_CharInLastByte",
			maxLen:  5,
			data:    []byte("Data\\"),
			wantLen: 5,
		},
		{
			name:    "Success_When_CharInTheMiddleOfData",
			maxLen:  10,
			data:    []byte("Data\\remaining"),
			wantLen: 5,
		},
		{
			name:    "NoCharFound_When_MaxLenAchieved",
			maxLen:  5,
			data:    []byte("More data\\"),
			wantErr: "delimiter not found in first 5 bytes",
		},
		{
			name:    "NoCharFound_When_TotalDataIterated",
			maxLen:  10,
			data:    []byte("Total data"),
			wantErr: "delimiter not found",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b := NewDelimiter('\x5C', "ASCIIBackslash")

			length, _, err := b.DecodeLength(tc.maxLen, tc.data)
			if err != nil || tc.wantErr != "" {
				assert.EqualError(t, err, tc.wantErr)
				return
			}

			assert.Equal(t, tc.wantLen, length)
		})
	}
}

func Test_BackslashPrefix_Inspect(t *testing.T) {
	b := NewDelimiter('\x5C', "ASCIIBackslash")

	inspection := b.Inspect()

	assert.Equal(t, "ASCIIBackslashDelimiter", inspection)
}
