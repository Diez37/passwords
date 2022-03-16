package config

const (
	HashSaltFieldName = "hash.salt"

	HashSaltDefault = "1zJT7As5HyRs9rCzbRXE"
)

type Hash struct {
	Salt string
}

func NewHash() *Hash {
	return &Hash{}
}
