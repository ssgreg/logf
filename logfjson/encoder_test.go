package logfjson

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/ssgreg/logf"
	"github.com/stretchr/testify/require"
)

type encoderTestCase struct {
	Name   string
	Entry  []logf.Entry
	Golden string
}

func TestEncoder(t *testing.T) {

	testCases := []encoderTestCase{
		{
			"Simple",
			[]logf.Entry{
				{
					LoggerID: int32(rand.Int()),
					Level:    logf.LevelInfo,
					Text:     "message",
				},
			},
			`{"level":"info","ts":"0001-01-01T00:00:00Z","msg":"message"}` + "\n",
		},
		{
			"SimpleCheck",
			[]logf.Entry{
				{
					LoggerID: int32(rand.Int()),
					Level:    logf.LevelInfo,
					Text:     "message",
					Fields: []logf.Field{
						logf.Any("s", "str"),
					},
				},
			},
			`{"level":"info","ts":"0001-01-01T00:00:00Z","msg":"message","s":"str"}` + "\n",
		},
	}

	enc := newTestEncoder()

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			b := logf.NewBuffer()
			for _, e := range tc.Entry {
				enc.Encode(b, e)
			}

			require.EqualValues(t, tc.Golden, b.String())

			var a map[string]interface{}
			require.NoError(t, json.NewDecoder(bytes.NewBuffer(b.Bytes())).Decode(&a))
		})
	}

}

func newTestEncoder() logf.Encoder {
	return NewEncoder(EncoderConfig{})
}
