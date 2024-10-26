package database

type File struct {
	Id     int
	Name   string
	Path   string
	Hash   string
	HashId int64
	Size   int64
}
