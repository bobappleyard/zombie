package data

type Queue[T any] struct {
	items []T
	put   int
	take  int
}

func (q *Queue[T]) Enqueue(x T) {
	if len(q.items) == 0 || q.next(q.put) == q.take {
		q.grow()
	}
	q.items[q.put] = x
	q.put = q.next(q.put)
}

func (q *Queue[T]) Dequeue() T {
	if !q.Ready() {
		panic("empty queue")
	}
	res := q.items[q.take]
	q.take = q.next(q.take)
	return res
}

func (q *Queue[T]) Ready() bool {
	return q.put != q.take
}

func (q *Queue[T]) next(x int) int {
	return (x + 1) % len(q.items)
}

func (q *Queue[T]) grow() {
	if len(q.items) == 0 {
		q.items = make([]T, 4)
		return
	}
	next := make([]T, len(q.items)*2)
	for i, j := 0, q.take; j != q.put; i, j = i+1, q.next(j) {
		next[i] = q.items[j]
	}
	q.put = len(q.items) - 1
	q.take = 0
	q.items = next
}
