package model

type Fields interface {
	// Iterator of string
	Terms(field string) Terms
}
