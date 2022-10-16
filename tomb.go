package gone

func NewTomb(goner Goner) Tomb {
	return &tomb{goner: goner}
}

type tomb struct {
	id    GonerId
	goner Goner
}

func (t *tomb) SetId(id GonerId) Tomb {
	t.id = id
	return t
}

func (t *tomb) GetId() GonerId {
	return t.id
}

func (t *tomb) GetGoner() Goner {
	return t.goner
}
