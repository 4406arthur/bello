package stream

import (
	"sync"

	"github.com/nats-io/nats.go"
)

// Pool is a simple connection pool for nats.io connections. It will create a small
// pool of initial connections, and if more connections are needed they will be
// created on demand. If a connection is Put back and the pool is full it will
// be closed.
type Pool struct {
	pool chan *nats.Conn
	df   DialFunc

	stopOnce sync.Once

	// The network/address that the pool is connecting to. These are going to be
	// whatever was passed into the New function. These should not be
	// changed after the pool is initialized
	Network, Addr string
}

// DialFunc is a function which can be passed into NewCustom
type DialFunc func(url string, options ...nats.Option) (*nats.Conn, error)

// NewPoolCustom is like New except you can specify a DialFunc which will be
// used when creating new connections for the pool. The common use-case is to do
// authentication for new connections.
func NewPoolCustom(addr string, size int, df DialFunc) (*Pool, error) {
	var client *nats.Conn
	var err error
	pool := make([]*nats.Conn, 0, size)
	for i := 0; i < size; i++ {
		client, err = df(addr)
		if err != nil {
			for _, client = range pool {
				client.Close()
			}
			pool = pool[:0]
			break
		}
		pool = append(pool, client)
	}
	p := Pool{
		Addr: addr,
		pool: make(chan *nats.Conn, len(pool)),
		df:   df,
	}
	for i := range pool {
		p.pool <- pool[i]
	}

	if size < 1 {
		return &p, err
	}

	return &p, err
}

// New creates a new NatsPool whose connections are all created using
// nats.Connect. The size indicates the maximum number of idle
// connections to have waiting to be used at any given moment. If an error is
// encountered an empty (but still usable) pool is returned alongside that error
func NewPool(addr string, size int) (*Pool, error) {
	return NewPoolCustom(addr, size, nats.Connect)
}

// Get retrieves an available nats connections. If there are none available it will
// create a new one on the fly
func (p *Pool) Get() (*nats.Conn, error) {
	select {
	case conn := <-p.pool:
		return conn, nil
	default:
		return p.df(p.Addr)
	}
}

// Put returns a client back to the pool. If the pool is full the client is
// closed instead. If the client is already closed (due to connection failure or
// what-have-you) it will not be put back in the pool
func (p *Pool) Put(conn *nats.Conn) {
	select {
	case p.pool <- conn:
	default:
		conn.Close()
	}
}

// Empty removes and calls Close() on all the connections currently in the pool.
// Assuming there are no other connections waiting to be Put back this method
// effectively closes and cleans up the pool.
func (p *Pool) Empty() {
	var conn *nats.Conn
	for {
		select {
		case conn = <-p.pool:
			conn.Close()
		default:
			return
		}
	}
}

// Avail returns the number of connections currently available to be gotten from
// the NatsPool using Get. If the number is zero then subsequent calls to Get will
// be creating new connections on the fly
func (p *Pool) Avail() int {
	return len(p.pool)
}
