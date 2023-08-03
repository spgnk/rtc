package peer

import "github.com/spgnk/rtc/utils"

// func (p *Peers) Error(v string) {
// 	p.logger.ERROR(v, map[string]any{
// 		"signal_id":     p.getSignalID(),
// 		"signal_status": "error",
// 	})
// }

// func (p *Peers) Info(v string) {
// 	p.logger.INFO(v, map[string]any{
// 		"signal_id":     p.getSignalID(),
// 		"signal_status": "info",
// 	})
// }

// func (p *Peers) Stack(v ...string) {
// 	p.logger.STACK(v...)
// }

// func (p *Peers) Warn(v string) {
// 	p.logger.WARN(v, map[string]any{
// 		"signal_id":     p.getSignalID(),
// 		"signal_status": "warn",
// 	})
// }

type peersLog struct {
	id     string
	logger utils.Log
}

func (w *peersLog) ERROR(v string, tags map[string]any) {
	if tags == nil {
		tags = make(map[string]any)
	}
	tags["signal_id"] = w.id
	tags["status"] = "error"
	w.logger.ERROR(v, tags)
}

func (w *peersLog) INFO(v string, tags map[string]any) {
	if tags == nil {
		tags = make(map[string]any)
	}
	tags["signal_id"] = w.id
	tags["status"] = "info"
	w.logger.INFO(v, tags)
}

func (w *peersLog) STACK(v ...string) {
	w.logger.STACK(v...)
}

func (w *peersLog) WARN(v string, tags map[string]any) {
	if tags == nil {
		tags = make(map[string]any)
	}
	tags["signal_id"] = w.id
	tags["status"] = "warn"
	w.logger.WARN(v, tags)
}

func (w *peersLog) DEBUG(v string, tags map[string]any) {
	if tags == nil {
		tags = make(map[string]any)
	}
	tags["signal_id"] = w.id
	tags["status"] = "debug"
	w.logger.DEBUG(v, tags)
}
