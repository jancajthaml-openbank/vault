package model

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPromises(t *testing.T) {

	t.Log("initially empty")
	{
		a := NewPromises()
		assert.Equal(t, a.Size(), 0)
	}

	t.Log("does not panic on nil")
	{
		var s *Promises
		s.Add("X")
		s.Remove("X")
	}
}

func TestPromises_Add(t *testing.T) {
	s := NewPromises()
	values := []string{"A", "B", "C", "D", "X", "Y", "E", "F", "G"}
	for _, value := range values {
		s.Add(value)
	}
	for _, value := range values {
		assert.True(t, s.Contains(value))
	}
	assert.Equal(t, len(values), s.Size())
}

func TestPromises_Remove(t *testing.T) {
	s := NewPromises()
	s.Add("X")
	s.Add("Y")
	values := []string{"A", "B", "C", "D", "E", "F", "G"}
	for _, value := range values {
		s.Add(value)
	}
	for _, value := range values {
		assert.True(t, s.Contains(value))
	}
	s.Remove("X")
	s.Remove("Y")
	assert.Equal(t, len(values), s.Size())
}

func TestPromises_Contains(t *testing.T) {
	s := NewPromises()
	s.Add("A")
	s.Add("B")
	s.Add("C")
	s.Add("D")
	s.Add("X")
	s.Add("Y")
	s.Add("E")
	s.Add("F")
	s.Add("G")

	table := []struct {
		input          []string
		expectedOutput bool
	}{
		{[]string{"X", "Y"}, true},
		{[]string{"H"}, false},
	}

	for _, test := range table {
		for _, value := range test.input {
			actualOutput := s.Contains(value)
			assert.Equal(t, test.expectedOutput, actualOutput)
		}
	}
}

func TestPromises_Size(t *testing.T) {
	s := NewPromises()
	require.Equal(t, s.Size(), 0)

	s.Add("A")
	s.Add("B")
	s.Add("C")
	s.Add("D")
	s.Add("X")
	s.Add("Y")
	s.Add("E")
	s.Add("F")
	assert.Equal(t, s.Size(), 8)

	s.Add("A")
	s.Add("B")
	s.Add("C")
	s.Add("D")
	s.Add("X")
	s.Add("Y")
	s.Add("E")
	s.Add("F")
	s.Add("G")
	assert.Equal(t, s.Size(), 9)

	s.Remove("A")
	s.Remove("B")
	s.Remove("C")
	s.Remove("D")
	s.Remove("X")
	s.Remove("Y")
	s.Remove("E")
	s.Remove("F")
	s.Remove("G")
	assert.Equal(t, s.Size(), 0)
}

func BenchmarkPromises_Add(b *testing.B) {
	s := NewPromises()

	fixture := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		fixture[i] = fmt.Sprintf("%d", i)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.Add(fixture[i%1000])
	}
}

func BenchmarkPromises_Remove(b *testing.B) {
	s := NewPromises()

	fixture := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		fixture[i] = fmt.Sprintf("%d", i)
		s.Add(fixture[i])
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.Remove(fixture[i%1000])
	}
}

func BenchmarkPromises_Contains(b *testing.B) {
	s := NewPromises()

	fixture := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		fixture[i] = fmt.Sprintf("%d", i)
		s.Add(fixture[i])
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.Contains(fixture[i%1000])
	}
}

func BenchmarkPromises_Size(b *testing.B) {
	s := NewPromises()

	for i := 0; i < 1000; i++ {
		s.Add(fmt.Sprintf("%d", i))
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.Size()
	}
}
