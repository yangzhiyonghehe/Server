package LinkList

import (
	"sync"

	"github.com/astaxie/beego"
)

//双向链表

type Node struct {
	Value    interface{}
	Next     *Node
	Previous *Node
}

func NewNode(v interface{}) *Node {
	return &Node{Value: v}
}

type LinkeList struct {
	length int
	head   *Node
	lock   *sync.Mutex
}

func NewLinkList() *LinkeList {
	l := new(LinkeList)
	l.length = 0
	l.lock = new(sync.Mutex)
	return l
}

func (l *LinkeList) HeadInsert(value interface{}) {
	l.lock.Lock()
	defer l.lock.Unlock()

	node := NewNode(value)
	node.Next = l.head
	if l.head != nil {
		l.head.Previous = node
	}

	l.head = node
	l.length++
}

func (l *LinkeList) TailInsert(value interface{}) {
	l.lock.Lock()
	defer l.lock.Unlock()

	tmp := l.head

	node := NewNode(value)

	if tmp == nil {
		l.head = node
		l.length++
		return
	}

	for tmp.Next != nil {
		tmp = tmp.Next
	}

	tmp.Next = node
	node.Previous = tmp
	l.length++
}

func (l *LinkeList) Remove(node *Node) *Node {
	l.lock.Lock()
	defer l.lock.Unlock()
	if node.Previous == nil {
		if node.Next != nil {
			l.head = node.Next
			l.length--
			return node.Next
		} else {
			node = nil
			l.head = nil
			l.length--
			return nil
		}
	} else {
		if node.Next != nil {
			node.Previous.Next = node.Next
			node.Next.Previous = node.Previous
			l.length--
			return node.Next
		} else {
			node.Previous.Next = nil
			l.length--
			return nil
		}
	}
}

func (l *LinkeList) GetHead() *Node {
	return l.head
}

func (l *LinkeList) Print() {
	tmp := l.head
	beego.Debug("linklist:")
	for tmp != nil {
		beego.Debug(tmp.Value)
		tmp = tmp.Next
	}
}
