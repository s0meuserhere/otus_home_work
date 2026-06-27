package hw04lrucache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		l := NewList()

		// пустой список: длина 0, Front и Back — nil
		require.Equal(t, 0, l.Len())
		require.Nil(t, l.Front())
		require.Nil(t, l.Back())
	})

	t.Run("complex", func(t *testing.T) {
		l := NewList()

		l.PushFront(10) // [10]
		l.PushBack(20)  // [10, 20]
		l.PushBack(30)  // [10, 20, 30]
		require.Equal(t, 3, l.Len())

		middle := l.Front().Next // 20
		l.Remove(middle)         // [10, 30]
		require.Equal(t, 2, l.Len())

		for i, v := range [...]int{40, 50, 60, 70, 80} {
			if i%2 == 0 {
				l.PushFront(v)
			} else {
				l.PushBack(v)
			}
		} // [80, 60, 40, 10, 30, 50, 70]

		require.Equal(t, 7, l.Len())
		require.Equal(t, 80, l.Front().Value)
		require.Equal(t, 70, l.Back().Value)

		l.MoveToFront(l.Front()) // [80, 60, 40, 10, 30, 50, 70]
		l.MoveToFront(l.Back())  // [70, 80, 60, 40, 10, 30, 50]

		elems := make([]int, 0, l.Len())
		for i := l.Front(); i != nil; i = i.Next {
			elems = append(elems, i.Value.(int))
		}
		require.Equal(t, []int{70, 80, 60, 40, 10, 30, 50}, elems)
	})

	t.Run("single element", func(t *testing.T) {
		l := NewList()

		l.PushFront(79) // [79]

		// один элемент: Front == Back, ссылки Prev/Next — nil
		require.Equal(t, 1, l.Len())
		require.Equal(t, 79, l.Front().Value)
		require.Equal(t, l.Front(), l.Back())
		require.Nil(t, l.Front().Prev)
		require.Nil(t, l.Back().Next)
	})

	t.Run("remove front and back", func(t *testing.T) {
		l := NewList()

		l.PushBack(1) // [1]
		l.PushBack(2) // [1, 2]
		l.PushBack(3) // [1, 2, 3]

		l.Remove(l.Front()) // [2, 3] — удалён первый элемент
		require.Equal(t, []int{2, 3}, listValues(l))

		l.Remove(l.Back()) // [2] — удалён последний элемент
		require.Equal(t, []int{2}, listValues(l))
	})

	t.Run("move to front middle", func(t *testing.T) {
		l := NewList()

		l.PushBack(1) // [1]
		l.PushBack(2) // [1, 2]
		l.PushBack(3) // [1, 2, 3]

		middle := l.Front().Next // 2
		l.MoveToFront(middle)    // [2, 1, 3]

		require.Equal(t, []int{2, 1, 3}, listValues(l))
	})
}

func listValues(l List) []int {
	elems := make([]int, 0, l.Len())
	for i := l.Front(); i != nil; i = i.Next {
		elems = append(elems, i.Value.(int))
	}
	return elems
}
