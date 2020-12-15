package controller

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	deliveriesCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "argocd_notifications_deliveries_total",
			Help: "Number of delivered notifications.",
		},
		[]string{"template", "service", "succeeded"},
	)

	triggerEvaluationsCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "argocd_notifications_trigger_eval_total",
			Help: "Number of trigger evaluations.",
		},
		[]string{"name", "triggered"},
	)
)

func NewMetricsRegistry() *controllerRegistry {
	registry := &controllerRegistry{
		Registry:                  prometheus.NewRegistry(),
		deliveriesCounter:         deliveriesCounter,
		triggerEvaluationsCounter: triggerEvaluationsCounter,
	}
	registry.MustRegister(deliveriesCounter)
	registry.MustRegister(triggerEvaluationsCounter)
	return registry
}

type controllerRegistry struct {
	*prometheus.Registry
	deliveriesCounter         *prometheus.CounterVec
	triggerEvaluationsCounter *prometheus.CounterVec
}

func (r *controllerRegistry) IncDeliveriesCounter(template string, service string, succeeded bool) {
	r.deliveriesCounter.WithLabelValues(template, service, strconv.FormatBool(succeeded)).Inc()
}

func (r *controllerRegistry) IncTriggerEvaluationsCounter(name string, triggered bool) {
	r.triggerEvaluationsCounter.WithLabelValues(name, strconv.FormatBool(triggered)).Inc()
}
