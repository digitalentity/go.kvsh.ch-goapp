package module

type Bundle []Module

func (b *Bundle) Add(m Module) {
	*b = append(*b, m)
}
