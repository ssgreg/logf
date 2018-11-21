package logf

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBufferInitial(t *testing.T) {
	capacity := 10

	t.Run("Initial", func(t *testing.T) {
		buf := NewBufferWithCapacity(capacity)

		require.EqualValues(t, 0, buf.Len())
		require.EqualValues(t, capacity, buf.Cap())
	})
}

func TestBufferAppend(t *testing.T) {
	t.Run("AppendString", func(t *testing.T) {
		capacity := 10
		data := "12345678"
		dataLen := len(data)
		buf := NewBufferWithCapacity(capacity)

		buf.AppendString(data)
		require.EqualValues(t, dataLen, buf.Len())
		require.EqualValues(t, capacity, buf.Cap())
		require.EqualValues(t, data, string(buf.Data))

		buf.AppendBytes([]byte(data))
		require.EqualValues(t, dataLen*2, buf.Len())
		require.True(t, buf.Cap() >= dataLen*2)
		require.EqualValues(t, data+data, buf.String())

		n, err := buf.Write([]byte(data))
		assert.Equal(t, len(data), n)
		assert.NoError(t, err)
		require.EqualValues(t, dataLen*3, buf.Len())
		require.True(t, buf.Cap() >= dataLen*3)
		require.EqualValues(t, data+data+data, buf.String())

		buf.AppendByte('\n')
		require.EqualValues(t, dataLen*3+1, buf.Len())
		require.True(t, buf.Cap() >= dataLen*3+1)
		require.EqualValues(t, data+data+data+"\n", buf.String())
	})
}

func TestBufferEnsureSize(t *testing.T) {
	capacity := 10
	data := "12345678"
	dataLen := len(data)

	t.Run("DoNothingWithEnoughCapacity", func(t *testing.T) {
		buf := NewBufferWithCapacity(capacity)
		buf.AppendString(data)

		buf.EnsureSize(capacity - dataLen)
		require.EqualValues(t, dataLen, buf.Len())
		require.EqualValues(t, capacity, buf.Cap())
	})

	t.Run("ReallocIfCapacityIsNotEnough", func(t *testing.T) {
		buf := NewBufferWithCapacity(capacity)
		buf.AppendString(data)

		buf.EnsureSize((capacity - dataLen) * 2)
		require.EqualValues(t, dataLen, buf.Len())
		require.True(t, buf.Cap() > capacity+(capacity-dataLen)*2)
		require.EqualValues(t, data, buf.String())
	})
}

func TestBufferExtendBytes(t *testing.T) {
	t.Run("WithoutRealloc", func(t *testing.T) {
		capacity := 10
		buf := NewBufferWithCapacity(capacity)

		buf.ExtendBytes(capacity)
		require.EqualValues(t, capacity, buf.Len())
		require.EqualValues(t, capacity, buf.Cap())
	})
	t.Run("WithRealloc", func(t *testing.T) {
		capacity := 10
		buf := NewBufferWithCapacity(capacity)

		buf.ExtendBytes(capacity * 2)
		require.EqualValues(t, capacity*2, buf.Len())
		require.True(t, buf.Cap() >= capacity*2)
	})
}

func TestBufferBack(t *testing.T) {
	capacity := 10
	buf := NewBufferWithCapacity(capacity)
	buf.AppendByte('\n')
	assert.EqualValues(t, '\n', buf.Back())
}

func TestBufferAppendFunctions(t *testing.T) {
	capacity := 10
	buf := NewBufferWithCapacity(capacity)

	AppendUint(buf, 123456789012345678)
	assert.Equal(t, []byte("123456789012345678"), buf.Bytes())
	buf.Reset()

	AppendInt(buf, -123456789012345678)
	assert.Equal(t, []byte("-123456789012345678"), buf.Bytes())
	buf.Reset()

	AppendFloat32(buf, 123456.16)
	assert.Equal(t, []byte("123456.16"), buf.Bytes())
	buf.Reset()

	AppendFloat64(buf, 123456.16)
	assert.Equal(t, []byte("123456.16"), buf.Bytes())
	buf.Reset()

	AppendBool(buf, true)
	assert.Equal(t, []byte("true"), buf.Bytes())
	buf.Reset()
}
