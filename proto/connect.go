package proto

type Msg struct {
	Ver       int    `json:"ver"`  // proto version
	Operation int    `json:"op"`   // operation for request
	SeqId     int    `json:"seq"`  // sequence number chosen by client
	Body      []byte `json:"body"` // binary body bytes
}

type PushMsgRequest struct {
	UserId int
	Msg    Msg
}

type PushRoomMsgRequest struct {
	RoomId int
	Msg    Msg
}

type PushRoomCountRequest struct {
	RoomId int
	Count  int
}
