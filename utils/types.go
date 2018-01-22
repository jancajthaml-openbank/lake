package utils

import "context"

const bufferSize = 100

// ZMQClient is a fascade for ZMQ queue
type ZMQClient struct {
	push   chan string
	sub    chan string
	stop   context.CancelFunc
	host   string
	region string
}

func newClient(region, host string, cancel context.CancelFunc) *ZMQClient {

	return &ZMQClient{
		push:   make(chan string, bufferSize),
		sub:    make(chan string, bufferSize),
		stop:   cancel,
		host:   host,
		region: region,
	}
}
