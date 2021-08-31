package website

import "context"

type Website struct {
	URL    string
	Format string
	Path   string
}

func (w *Website) GetIP(ctx context.Context) (string, error) {
	return roundtrip(ctx, -1, w.URL, w.Format, w.Path)
}
