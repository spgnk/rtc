package peer

func (p *Peer) Error(v string) {
	p.logger.ERROR(v, map[string]any{"id": p.GetPeerConnectionID()})
}

func (p *Peer) Info(v string) {
	p.logger.INFO(v, map[string]any{"id": p.GetPeerConnectionID()})
}

func (p *Peer) Stack(v ...string) {
	p.logger.STACK(v...)
}

func (p *Peer) Warn(v string) {
	p.logger.WARN(v, map[string]any{"id": p.GetPeerConnectionID()})
}
