package peer

func (p *Peer) Error(v string, tags map[string]any) {
	if tags == nil {
		tags = make(map[string]any)
	}
	tags["peer_id"] = p.GetPeerConnectionID()
	tags["status"] = "error"
	p.logger.ERROR(v, tags)
}

func (p *Peer) Info(v string, tags map[string]any) {
	if tags == nil {
		tags = make(map[string]any)
	}
	tags["peer_id"] = p.GetPeerConnectionID()
	tags["status"] = "info"
	p.logger.INFO(v, tags)
}

func (p *Peer) Stack(v ...string) {
	p.logger.STACK(v...)
}

func (p *Peer) Warn(v string, tags map[string]any) {
	if tags == nil {
		tags = make(map[string]any)
	}
	tags["peer_id"] = p.GetPeerConnectionID()
	tags["status"] = "warn"
	p.logger.WARN(v, tags)
}
