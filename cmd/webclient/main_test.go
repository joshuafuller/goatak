package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeUid(t *testing.T) {
	c := makeUid("s1")
	fmt.Println(c)
	assert.Len(t, c, 24)
	assert.Equal(t, makeUid("string1"), makeUid("string1"))
	assert.NotEqual(t, makeUid("s1"), makeUid("s2"))
}
