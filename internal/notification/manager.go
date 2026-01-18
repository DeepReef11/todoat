package notification

// manager implements NotificationManager
type manager struct {
	channels        []NotificationChannel
	enabled         bool
	commandExecutor CommandExecutor
}

// NewManager creates a new NotificationManager based on configuration
func NewManager(cfg *Config, opts ...Option) (NotificationManager, error) {
	m := &manager{
		channels: []NotificationChannel{},
		enabled:  cfg.Enabled,
	}

	// Apply options first to get command executor
	for _, opt := range opts {
		opt(m)
	}

	if !cfg.Enabled {
		return m, nil
	}

	if cfg.OSNotification.Enabled {
		var osOpts []Option
		if m.commandExecutor != nil {
			osOpts = append(osOpts, WithCommandExecutor(m.commandExecutor))
		}
		osChannel := NewOSNotificationChannel(&cfg.OSNotification, osOpts...)
		m.channels = append(m.channels, osChannel)
	}

	if cfg.LogNotification.Enabled {
		logChannel := NewLogNotificationChannel(&cfg.LogNotification)
		m.channels = append(m.channels, logChannel)
	}

	return m, nil
}

// Send dispatches notification to all enabled channels
func (m *manager) Send(n Notification) error {
	if !m.enabled {
		return nil
	}

	var lastErr error
	for _, ch := range m.channels {
		if err := ch.Send(n); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// SendAsync dispatches notification without blocking
func (m *manager) SendAsync(n Notification) {
	go func() {
		_ = m.Send(n)
	}()
}

// Close cleans up resources
func (m *manager) Close() error {
	var lastErr error
	for _, ch := range m.channels {
		if err := ch.Close(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// ChannelCount returns the number of active channels
func (m *manager) ChannelCount() int {
	return len(m.channels)
}
