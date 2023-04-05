package domain

type Visit struct {
	Page  Link
	Links []Link
}

type Link string

func (l Link) String() string {
	return string(l)
}
