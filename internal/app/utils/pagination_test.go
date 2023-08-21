package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewPagination(t *testing.T) {
	tests := []struct {
		inputLimit     int
		inputOffset    int
		expectedLimit  int
		expectedOffset int
	}{
		{15, 5, 15, 5},
		{0, 10, 10, 10},
		{10, -5, 10, 0},
		{-5, -5, 10, 0},
	}

	for _, tt := range tests {
		p := NewPagination(tt.inputLimit, tt.inputOffset)
		assert.Equal(t, tt.expectedLimit, p.Limit, "They should be equal")
		assert.Equal(t, tt.expectedOffset, p.Offset, "They should be equal")
	}
}

func TestDefaultPagination(t *testing.T) {
	p := DefaultPagination()
	assert.Equal(t, 10, p.Limit, "Default Limit should be 10")
	assert.Equal(t, 0, p.Offset, "Default Offset should be 0")
}
