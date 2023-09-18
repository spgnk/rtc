package utils

import (
	"fmt"
)

func (f *Forwarder) getID() string {
	// f.mutex.RLock()
	// defer f.mutex.RUnlock()
	return f.id
}

func (f *Forwarder) checkClose() bool {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.isClosed
}

func (f *Forwarder) setClose(state bool) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.isClosed = state
}

// info to export log info
func (f *Forwarder) info(v ...interface{}) {
	f.logger.INFO(fmt.Sprintf("[%s] %v", f.id, v), nil)
}

// error to export error info
func (f *Forwarder) error(v ...interface{}) {
	f.logger.ERROR(fmt.Sprintf("[%s] %v", f.id, v), nil)
}

func (f *Forwarder) getClient(id string) *Client {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.clients[id]
}

func (f *Forwarder) addClient(id string, c *Client) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.clients[id] = c
}

func (f *Forwarder) removeClient(id string) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	delete(f.clients, id)
}

func (f *Forwarder) closeClient(id string) {
	c := f.getClient(id)
	if c == nil {
		return
	}
	f.removeClient(id)
	c.cancelFunc()
}

// Handlepanic prevent panic
func handlepanic(data ...interface{}) {
	if a := recover(); a != nil {
		fmt.Println("===========This data make fwd panic==============")
		fmt.Println(data...)
		fmt.Println("RECOVER", a)
	}
}
