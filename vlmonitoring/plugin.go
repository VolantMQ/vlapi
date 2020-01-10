package vlmonitoring

type IFace interface {
	Push(stats Stats)
	Shutdown() error
}
