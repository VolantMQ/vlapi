package vlhealth

import (
	"github.com/heptiolabs/healthcheck"
)

type Handler interface {
	AddLivenessCheck(string, healthcheck.Check)
	AddReadinessCheck(string, healthcheck.Check)
}
