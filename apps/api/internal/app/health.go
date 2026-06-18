package app

type ServiceName string

const (
	ServiceNameStuffStash ServiceName = "stuff-stash"
)

type HealthStatusValue string

const (
	HealthStatusHealthy HealthStatusValue = "healthy"
)

type HealthStatus struct {
	Service ServiceName
	Status  HealthStatusValue
}
