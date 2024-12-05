package data

import (
	"testing"

	"github.com/bobappleyard/zombie/util/assert"
)

func TestQueue(t *testing.T) {
	var q Queue[int]

	q.Enqueue(1)
	q.Enqueue(2)
	q.Enqueue(3)
	q.Enqueue(4)
	q.Enqueue(5)
	q.Enqueue(6)

	assert.Equal(t, q.Dequeue(), 1)

	q.Enqueue(7)
	q.Enqueue(8)
	q.Enqueue(9)
	q.Enqueue(10)
	q.Enqueue(11)

	expect := 2
	for q.Ready() {
		assert.Equal(t, q.Dequeue(), expect)
		expect++
	}
	assert.Equal(t, expect, 12)
}
