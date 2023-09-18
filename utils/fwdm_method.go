package utils

import "time"

func (f *ForwarderMannager) checkClose() bool {
	return f.isClosed
}

func (f *ForwarderMannager) setClose(state bool) {
	f.isClosed = state
}

func (f *ForwarderMannager) getForwarder(trackID string) *Forwarder {
	fwd, has := f.forwadrders.Get(trackID)
	if !has {
		f.logger.WARN("Cannor find fwd with input trackID", map[string]any{
			"track_id": trackID,
			"time":     time.Now(),
		})
	}
	return fwd
}

func (f *ForwarderMannager) setForwarder(trackID string, fwd *Forwarder) {
	f.forwadrders.Set(trackID, fwd)
}

func (f *ForwarderMannager) deleteForwarder(trackID string) {
	f.forwadrders.Remove(trackID)
}

func (f *ForwarderMannager) closeForwarder(trackID string) {
	fwd := f.getForwarder(trackID)
	if fwd != nil {
		f.deleteForwarder(trackID)
		fwd.Close()
	}
}

func (f *ForwarderMannager) closeForwaders() {
	for _, fwd := range f.forwadrders.Items() {
		fwd.Close()
	}
}
