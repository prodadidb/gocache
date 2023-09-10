package store_test

import (
	"testing"

	"github.com/prodadidb/gocache/store"
	"github.com/stretchr/testify/assert"
)

func TestInvalidateOptionsTagsValue(t *testing.T) {
	// Given
	options := store.InvalidateOptions{
		Tags: []string{"tag1", "tag2", "tag3"},
	}

	// When - Then
	assert.Equal(t, []string{"tag1", "tag2", "tag3"}, options.Tags)
}
