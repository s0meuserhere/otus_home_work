package hw04lrucache

import "slices"

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
	items []*ListItem
}

func NewList() List {
	data := make([]*ListItem, 0)
	return &list{data}
}

func (l *list) Len() int {
	if l.items == nil {
		return 0
	}

	return len(l.items)
}

func (l *list) Front() *ListItem {
	if l.Len() == 0 {
		return nil
	}
	return l.items[0]
}

func (l *list) Back() *ListItem {
	if l.Len() == 0 {
		return nil
	}
	return l.items[l.Len()-1]
}

func (l *list) PushFront(v interface{}) *ListItem {
	if l.items == nil {
		return nil
	}

	li := &ListItem{v, nil, nil}

	if l.Len() == 0 {
		l.items = append(l.items, li)
		return l.Front()
	}

	li.Next = l.Front()
	l.Front().Prev = li

	l.items = slices.Insert(l.items, 0, li)

	return l.Front()
}

func (l *list) PushBack(v interface{}) *ListItem {
	if l.items == nil {
		return nil
	}

	li := &ListItem{v, nil, nil}

	if l.Len() == 0 {
		l.items = append(l.items, li)
		return l.Back()
	}

	li.Prev = l.Back()
	l.Back().Next = li

	l.items = slices.Insert(l.items, l.Len(), li)

	return l.Back()
}

func (l *list) Remove(i *ListItem) {
	if i == nil || l.items == nil {
		return
	}

	index := slices.IndexFunc(l.items, func(item *ListItem) bool {
		return item.Value == i.Value
	})

	if index > -1 {
		deleting := l.items[index]

		if deleting.Prev != nil {
			deleting.Prev.Next = deleting.Next
		}
		if deleting.Next != nil {
			deleting.Next.Prev = deleting.Prev
		}

		l.items = slices.Delete(l.items, index, index+1)
	}
}

func (l *list) MoveToFront(i *ListItem) {
	if i == nil || l.items == nil {
		return
	}

	index := slices.IndexFunc(l.items, func(item *ListItem) bool {
		return item.Value == i.Value
	})

	if index > -1 {
		l.Remove(i)
		l.PushFront(i.Value)
	}
}
