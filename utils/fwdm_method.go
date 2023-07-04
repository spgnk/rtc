package utils

func (f *ForwarderMannager) checkClose() bool {
	return f.isClosed
}

func (f *ForwarderMannager) setClose(state bool) {
	f.isClosed = state
}

func (f *ForwarderMannager) getForwarder(id *string) *Forwarder {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.forwadrders[*id]
}

func (f *ForwarderMannager) setForwarder(id *string, fwd *Forwarder) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.forwadrders[*id] = fwd
}

func (f *ForwarderMannager) deleteForwarder(id *string) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	delete(f.forwadrders, *id)
}

func (f *ForwarderMannager) closeForwarder(id *string) {
	fw := f.getForwarder(id)
	f.deleteDataTime(*id)
	if fw != nil {
		f.deleteForwarder(id)
		fw.Close()
		fw = nil
	}
}

func (f *ForwarderMannager) closeForwaders() {
	// f.mutex.RLock()
	// defer f.mutex.RUnlock()
	// for _, fwd := range f.forwadrders {
	// 	fwd.Close()
	// }
	// fwds := f.getAllFwd()
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	for _, fwd := range f.forwadrders {
		fwd.Close()
	}
}

func (f *ForwarderMannager) setDatatime(c *ClientDataTime) {
	f.mutex.Lock()
	f.dataTime[c.id] = c.t
	f.mutex.Unlock()
}

func (f *ForwarderMannager) getDatatime() map[string]int64 {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.dataTime
}

func (f *ForwarderMannager) deleteDataTime(fwdID string) {
	f.mutex.Lock()
	delete(f.dataTime, fwdID)
	f.mutex.Unlock()
}
