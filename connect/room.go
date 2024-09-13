package connect

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gochat_my/proto"
	"sync"
)

const NoRoom = -1

type Room struct {
	Id          int
	OnlineCount int
	rLock       sync.RWMutex
	drop        bool // room is dropped
	next        *Channel
}

func NewRoom(id int) *Room {
	room := new(Room)
	room.Id = id
	room.drop = false
	room.next = nil
	room.OnlineCount = 0
	return room
}

func (r *Room) Push(msg *proto.Msg) { // send a message in room
	r.rLock.RLock()
	for ch := r.next; ch != nil; ch = ch.Next {
		if err := ch.Push(msg); err != nil {
			logrus.Infof("push msg err : %s", err.Error())
		}
	}
	r.rLock.Unlock()
	return
}

func (r *Room) Put(ch *Channel) (err error) {
	// a room include many channel (user session)
	// channel is self-linked by doubly linked list
	r.rLock.Lock()
	defer r.rLock.Unlock()
	if !r.drop { // ch.Next 代表上一条信息，room.next表示最后加入的ch
		if r.next != nil {
			r.next.Prev = ch
		}
		ch.Next = r.next
		ch.Prev = nil
		r.next = ch
		r.OnlineCount++
	} else {
		err = errors.New("room drop")
	}
	return
}

func (r *Room) RemoveChannel(ch *Channel) bool {
	r.rLock.Lock()
	defer r.rLock.Unlock()
	if ch.Next != nil {
		ch.Next.Prev = ch.Prev
	}
	if ch.Prev != nil {
		ch.Prev.Next = ch.Next
	} else {
		r.next = ch.Next
	}

	r.OnlineCount--
	r.drop = false
	if r.OnlineCount <= 0 {
		r.drop = true
	}
	return r.drop
}
