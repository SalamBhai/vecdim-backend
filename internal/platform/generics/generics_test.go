package generics_test

import (
	"strings"
	"testing"

	"github.com/salambhai/vecdim-backend/internal/platform/generics"
	"github.com/stretchr/testify/assert"
)

func TestGenericMap(t *testing.T) {
	words := []string{"lagos", "kano", "osun"}
	upper := generics.Map(words, strings.ToUpper)
	assert.Equal(t, []string{"LAGOS", "KANO", "OSUN"}, upper)
}

func TestGenericFilter(t *testing.T) {
	numbers := []int{10, 25, 50, 75, 100}
	greaterThan50 := generics.Filter(numbers, func(n int) bool {
		return n > 50
	})
	assert.Equal(t, []int{75, 100}, greaterThan50)
}

func TestGenericChunk(t *testing.T) {
	items := []string{"A", "B", "C", "D", "E"}
	chunks := generics.Chunk(items, 2)
	assert.Len(t, chunks, 3)
	assert.Equal(t, []string{"A", "B"}, chunks[0])
	assert.Equal(t, []string{"C", "D"}, chunks[1])
	assert.Equal(t, []string{"E"}, chunks[2])
}

func TestGenericContains(t *testing.T) {
	states := []string{"Lagos", "Kano", "Osun"}
	assert.True(t, generics.Contains(states, "Osun"))
	assert.False(t, generics.Contains(states, "Kaduna"))
}
