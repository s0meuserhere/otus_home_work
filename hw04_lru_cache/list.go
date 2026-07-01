package hw04lrucache

type List interface {
	Len() int
	Front() *ListItem
	Back() *ListItem
	PushFront(v interface{}) *ListItem
	PushBack(v interface{}) *ListItem
	Remove(i *ListItem)
	MoveToFront(i *ListItem)
}

type ListItem struct {
	Value interface{}
	Next  *ListItem
	Prev  *ListItem
}

type list struct {
	first *ListItem
	last  *ListItem
	len   int
}

func NewList() List {
	return &list{}
}

func (l *list) Len() int {
	return l.len
}

func (l *list) Front() *ListItem {
	return l.first
}

func (l *list) Back() *ListItem {
	return l.last
}

func (l *list) PushFront(v interface{}) *ListItem {
	l.first = &ListItem{
		Value: v,
		Next:  l.Front(),
		Prev:  nil,
	}

	if l.Front().Next != nil {
		l.Front().Next.Prev = l.Front()
	}

	if l.Len() == 0 {
		l.last = l.Front()
	}

	l.len++

	return l.Front()
}

func (l *list) PushBack(v interface{}) *ListItem {
	l.last = &ListItem{
		Value: v,
		Next:  nil,
		Prev:  l.Back(),
	}

	if l.Back().Prev != nil {
		l.Back().Prev.Next = l.Back()
	}

	if l.Len() == 0 {
		l.first = l.Back()
	}

	l.len++

	return l.Back()
}

func (l *list) Remove(i *ListItem) {
	if i == nil {
		return
	}
	l.removeWithoutDec(i)
	l.len--
}

func (l *list) MoveToFront(i *ListItem) {
	if i == nil || i == l.first {
		return
	}

	l.removeWithoutDec(i)

	i.Prev = nil
	i.Next = l.first

	if l.first != nil {
		l.first.Prev = i
	}
	l.first = i
}

func (l *list) removeWithoutDec(i *ListItem) {
	if i == nil {
		return
	}

	if i.Prev != nil {
		i.Prev.Next = i.Next
	} else {
		l.first = i.Next
	}

	if i.Next != nil {
		i.Next.Prev = i.Prev
	} else {
		l.last = i.Prev
	}
}
