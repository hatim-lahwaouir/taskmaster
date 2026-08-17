package supervisor



type Msg struct {
	clientId int
	msg string
}


func NewMsg(clientID int, msg string) *Msg{
	return &Msg{clientId: clientID, msg: msg}
}


func (m *Msg) GetClientID() int{ 
	return m.clientId
}

func (m *Msg) GetMesg() string {
	return m.msg
}