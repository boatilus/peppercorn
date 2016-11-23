package session

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateSessionID(t *testing.T) {
	got, err := GenerateSessionID()

	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, len(got), 172) // We should have retrieved a 172-character string
}
