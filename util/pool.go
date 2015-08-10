package util

// Pool holds Clients.
type Pool struct {
	pool chan string
}

// NewPool creates a new pool of Clients.
func NewPool(max int) *Pool {
	p := &Pool{
		pool: make(chan string, max),
	}
	for i := 0; i < max; i++ {
		s := "client"
		p.Return(s)
	}
	return p
}

// Borrow a Client from the pool.
func (p *Pool) Borrow() string {
	return <-p.pool
}

// Return returns a Client to the pool.
func (p *Pool) Return(c string) {
	p.pool <- c
}
