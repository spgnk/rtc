package peer

func (p *Peers) Error(v string) {
	p.logger.ERROR(v, map[string]any{
		"signal_id":     p.getSignalID(),
		"signal_status": "error",
	})
}

func (p *Peers) Info(v string) {
	p.logger.INFO(v, map[string]any{
		"signal_id":     p.getSignalID(),
		"signal_status": "info",
	})
}

func (p *Peers) Stack(v ...string) {
	p.logger.STACK(v...)
}

func (p *Peers) Warn(v string) {
	p.logger.WARN(v, map[string]any{
		"signal_id":     p.getSignalID(),
		"signal_status": "warn",
	})
}
