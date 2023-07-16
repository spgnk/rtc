package worker

func (w *PeerWorker) Error(v string) {
	w.logger.ERROR(v, map[string]any{"id": w.id})
}

func (w *PeerWorker) Info(v string) {
	w.logger.INFO(v, map[string]any{"id": w.id})
}

func (w *PeerWorker) Stack(v ...string) {
	w.logger.STACK(v...)
}

func (w *PeerWorker) Warn(v string) {
	w.logger.WARN(v, map[string]any{"id": w.id})
}
