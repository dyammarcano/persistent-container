package cache2you

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewCacheData(t *testing.T) {
	c := NewCacheData(time.Minute, time.Minute)
	assert.NotNil(t, c)
}

func TestCacheData_SetGet(t *testing.T) {
	c := NewCacheData(time.Minute, time.Minute)
	c.Set("key", "value", time.Minute)

	value, found := c.Get("key")
	assert.True(t, found)

	strVal, ok := value.(string)
	assert.True(t, ok)
	assert.Equal(t, "value", strVal)
}
