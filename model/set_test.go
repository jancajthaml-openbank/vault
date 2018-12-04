package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransactionSet_New(t *testing.T) {
	s := NewTransactionSet()
	assert.Equal(t, s.Size(), 0)
}

func TestTransactionSet_Add(t *testing.T) {
	s := NewTransactionSet()
	s.Add("A", "B", "C", "D", "X", "Y", "E", "F")
	s.Add("G")
	actualOutput := s.Values()
	expectedOutput := []string{"A", "B", "C", "D", "X", "Y", "E", "F", "G"}
	assert.Equal(t, expectedOutput, actualOutput)
}

func TestTransactionSet_Remove(t *testing.T) {
	s := NewTransactionSet()
	s.Add("A", "B", "C", "D", "X", "Y", "E", "F")
	s.Add("G")

	s.Remove("X", "Y")

	actualOutput := s.Values()
	expectedOutput := []string{"A", "B", "C", "D", "E", "F", "G"}
	assert.Equal(t, expectedOutput, actualOutput)

	s.Remove("X", "Y")
	actualOutput = s.Values()
	expectedOutput = []string{"A", "B", "C", "D", "E", "F", "G"}

	assert.Equal(t, expectedOutput, actualOutput)
}

func TestTransactionSet_Contains(t *testing.T) {
	s := NewTransactionSet()
	s.Add("A", "B", "C", "D", "X", "Y", "E", "F")
	s.Add("G")

	table := []struct {
		input          []string
		expectedOutput bool
	}{
		{[]string{"X", "Y"}, true},
		{[]string{"H"}, false},
	}

	for _, test := range table {
		actualOutput := s.Contains(test.input...)
		assert.Equal(t, test.expectedOutput, actualOutput)
	}
}

func TestTransactionSet_Values(t *testing.T) {
	s := NewTransactionSet()
	s.Add("A", "B", "C", "D", "X", "Y", "E", "F")
	s.Add("G")

	actualOutput := s.Values()
	expectedOutput := []string{"A", "B", "C", "D", "X", "Y", "E", "F", "G"}

	assert.Equal(t, expectedOutput, actualOutput)
}

func TestTransactionSet_Size(t *testing.T) {
	s := NewTransactionSet()
	require.Equal(t, s.Size(), 0)

	s.Add("A", "B", "C", "D", "X", "Y", "E", "F")
	assert.Equal(t, s.Size(), 8)

	s.Add("A", "B", "C", "D", "X", "Y", "E", "F", "G")
	assert.Equal(t, s.Size(), 9)

	s.Remove("A", "B", "C", "D", "X", "Y", "E", "F", "G")
	assert.Equal(t, s.Size(), 0)
}

func BenchmarkTransactionSet_Add(b *testing.B) {
	s := NewTransactionSet()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Add(string(i))
	}
}

func BenchmarkTransactionSet_Remove(b *testing.B) {
	s := NewTransactionSet()
	for i := 0; i < 10000; i++ {
		s.Add(string(i))
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Remove(string(i))
	}
}

func BenchmarkTransactionSet_Contains(b *testing.B) {
	s := NewTransactionSet()
	for i := 0; i < 10000; i++ {
		s.Add(string(i))
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Contains(string(i))
	}
}

func BenchmarkTransactionSet_Size(b *testing.B) {
	s := NewTransactionSet()
	for i := 0; i < 10000; i++ {
		s.Add(string(i))
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Size()
	}
}

func BenchmarkTransactionSet_Values(b *testing.B) {
	s := NewTransactionSet()
	for i := 0; i < 10000; i++ {
		s.Add(string(i))
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Values()
	}
}

func BenchmarkTransactionSet_String(b *testing.B) {
	s := NewTransactionSet()
	for i := 0; i < 10000; i++ {
		s.Add(string(i))
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.String()
	}
}
