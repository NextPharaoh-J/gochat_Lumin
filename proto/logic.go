package proto

type ConnectRequest struct {
	AuthToken string `json:"auth_token"`
	RoomId    int    `json:"room_id"`
	ServerId  string `json:"server_id"`
}
type DisConnectRequest struct {
	RoomId int
	UserId int
}

type ConnectReply struct {
	UserId int
}
type DisConnectReply struct {
	Has bool
}
