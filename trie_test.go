package cidlookup

import (
	"testing"

	"golang.org/x/exp/rand"

	"github.com/stretchr/testify/require"
)

func TestInserts(t *testing.T) {
	tr := cidTrie{}
	require.NoError(t, tr.Add([]byte{1, 2, 3, 4}))
	require.NoError(t, tr.Add([]byte{4, 3, 2, 1}))
	require.NoError(t, tr.Add([]byte{6, 3, 2, 1}))
	require.NoError(t, tr.Add([]byte{1, 2, 3, 6}))
	require.NoError(t, tr.Add([]byte{1, 2, 9, 10}))
}

func TestCollisionDetection(t *testing.T) {
	tr := cidTrie{}
	require.NoError(t, tr.Add([]byte{1, 2, 3, 4}))
	require.ErrorIs(t, tr.Add([]byte{1, 2, 3, 4, 5, 6}), errCollision)
	require.ErrorIs(t, tr.Add([]byte{1, 2, 3}), errCollision)
}

func TestSingleLookup(t *testing.T) {
	tr := cidTrie{}
	require.NoError(t, tr.Add([]byte{1, 2, 3, 4}))

	b, err := tr.Lookup([]byte{1, 2, 3, 4, 5, 6, 7})
	require.NoError(t, err)
	require.Equal(t, []byte{1, 2, 3, 4}, b)

	b, err = tr.Lookup([]byte{1, 2, 3, 4})
	require.NoError(t, err)
	require.Equal(t, []byte{1, 2, 3, 4}, b)

	_, err = tr.Lookup([]byte{1, 2, 3})
	require.ErrorIs(t, err, errNotFound)

	_, err = tr.Lookup([]byte{4, 3, 2, 1})
	require.ErrorIs(t, err, errNotFound)
}

func TestMultipleLookups(t *testing.T) {
	tr := cidTrie{}
	require.NoError(t, tr.Add([]byte{1, 2, 3, 4}))
	require.NoError(t, tr.Add([]byte{1, 2, 3, 6}))
	require.NoError(t, tr.Add([]byte{1, 2, 3, 5, 7}))

	b, err := tr.Lookup([]byte{1, 2, 3, 4, 5, 6, 7})
	require.NoError(t, err)
	require.Equal(t, []byte{1, 2, 3, 4}, b)

	b, err = tr.Lookup([]byte{1, 2, 3, 6, 8})
	require.NoError(t, err)
	require.Equal(t, []byte{1, 2, 3, 6}, b)

	b, err = tr.Lookup([]byte{1, 2, 3, 5, 7, 0})
	require.NoError(t, err)
	require.Equal(t, []byte{1, 2, 3, 5, 7}, b)
}

const numCIDs = 100000

func BenchmarkTrieLookup(b *testing.B) {
	b.Run("4 byte CIDs", func(b *testing.B) { benchmarkTrieLookup(b, 4) })
	b.Run("8 byte CIDs", func(b *testing.B) { benchmarkTrieLookup(b, 8) })
	b.Run("15 byte CIDs", func(b *testing.B) { benchmarkTrieLookup(b, 15) })
	b.Run("20 byte CIDs", func(b *testing.B) { benchmarkTrieLookup(b, 20) })
}

func BenchmarkMapLookup(b *testing.B) {
	b.Run("4 byte CIDs", func(b *testing.B) { benchmarkMapLookup(b, 4) })
	b.Run("8 byte CIDs", func(b *testing.B) { benchmarkMapLookup(b, 8) })
	b.Run("15 byte CIDs", func(b *testing.B) { benchmarkMapLookup(b, 15) })
	b.Run("20 byte CIDs", func(b *testing.B) { benchmarkMapLookup(b, 20) })
}

func benchmarkTrieLookup(b *testing.B, cidLen int) {
	var cids = make([][]byte, 0, numCIDs)
	tr := cidTrie{}
	var numInserted int
	for numInserted < numCIDs {
		buf := make([]byte, cidLen)
		rand.Read(buf)
		if err := tr.Add(buf); err == nil {
			numInserted++
			cids = append(cids, buf)
		}
	}

	packet := make([]byte, 20)
	buf := make([]byte, 20)
	rand.Read(packet)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy(buf, packet)
		copy(buf[:cidLen], cids[i%len(cids)])
		cid, err := tr.Lookup(buf)
		if err != nil {
			b.Fatal(err)
		}
		_ = cid
	}
}

var Sink bool

func benchmarkMapLookup(b *testing.B, cidLen int) {
	m := make(map[string]struct{})

	var cids = make([][]byte, 0, numCIDs)
	for i := 0; i < numCIDs; i++ {
		buf := make([]byte, cidLen)
		rand.Read(buf)
		m[string(buf)] = struct{}{}
		cids = append(cids, buf)
	}

	packet := make([]byte, 20)
	buf := make([]byte, 20)
	rand.Read(packet)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy(buf, packet)
		copy(buf[:cidLen], cids[i%len(cids)])
		_, ok := m[string(buf)]
		Sink = ok
	}
}
