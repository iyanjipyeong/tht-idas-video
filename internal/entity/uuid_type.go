package entity

type UUID string

func (id UUID) String() string {
	return string(id)
}
