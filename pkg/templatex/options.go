package templatex

type Option func(*options)

type options struct {
	metrics Metrics
}

func defaultOptions() options {
	return options{
		metrics: NoopMetrics{},
	}
}

func WithMetrics(metrics Metrics) Option {
	return func(o *options) {
		if metrics != nil {
			o.metrics = metrics
		}
	}
}
