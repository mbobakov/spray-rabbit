package elasticspray

import (
	"net/url"

	"github.com/pkg/errors"
)

func normalizeURL(uri string) (string, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", errors.Wrap(err, "Coudn't parse url for elastic")
	}
	if u.Scheme == "" {
		u.Scheme = defaultElasticSheme
	}
	u.Path = elasticAPIBulkPath
	return u.String(), nil
}
