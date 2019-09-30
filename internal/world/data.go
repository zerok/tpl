package world

// System acts as a container for various info points about the system tpl
// is executed on.
type Data struct {
	world *World
}

// System allows access to various pieces of information regarding the system
// tpl is executed on.
func (w *World) Data() *Data {
	return &Data{
		world: w,
	}
}
