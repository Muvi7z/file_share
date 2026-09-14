package storage

type Storage struct {
	PosterPath string
}

func New(PosterPath string) *Storage {
	return &Storage{
		PosterPath: PosterPath,
	}
}
