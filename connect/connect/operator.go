package connect

import "gochat_my/proto"

type Operator interface {
	Connect(conn *proto.ConnectRequest) (int, error)
	DisConnect(conn *proto.DisConnectRequest) error
}
type DefaultOperator struct{}

func (this *DefaultOperator) Connect(conn *proto.ConnectRequest) (uid int, err error) {
	rpcConnect := new(RpcConnect)
	uid, err = rpcConnect.Connect(conn)
	return
}
func (this *DefaultOperator) DisConnect(conn *proto.DisConnectRequest) (err error) {
	rpcConnect := new(RpcConnect)
	err = rpcConnect.DisConnect(conn)
	return
}
