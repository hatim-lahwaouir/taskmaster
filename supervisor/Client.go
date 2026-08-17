package supervisor

import (
	"net"
	"sync"
	"sync/atomic"
)



type Client struct {
	conn net.Conn
	id int
	mu sync.Mutex
	msg chan *Msg
	bad atomic.Bool
}

var id int = 0

func NewClient(conn net.Conn, msg chan *Msg) *Client{
	id++

	return &Client{conn: conn, id: id, msg: msg}
}

func (c *Client) ID() int {
	return c.id
}



func (c *Client) IsBad() bool {
	return c.bad.Load()
}






func (c *Client) IsGood()  bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn != nil
}



func (c *Client) ReadGorotine(wg *sync.WaitGroup) {
	defer wg.Done()

	for ;; {
		buf := make([]byte, 800)
		n, err := c.conn.Read(buf)
		if err != nil {
			c.bad.Store(true)
			c.conn.Close()
			return 
		}
		c.msg <- NewMsg(c.ID(), string(buf[:n]))
	}
}



func (c *Client) WriteGorotine(buf []byte) {
	_, err := c.conn.Write(buf)
	if err != nil {
		c.bad.Store(true)
		c.conn.Close()
	}
}