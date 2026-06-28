package postgres

import (
	"errors"
	"fmt"
	"strings"
)

var ErrUnsupportedDialect = errors.New("unsupported dialect")

type IndexMethod string

const (
	BTree  IndexMethod = "btree"
	Hash   IndexMethod = "hash"
	GIN    IndexMethod = "gin"
	GiST   IndexMethod = "gist"
	SPGiST IndexMethod = "spgist"
	BRIN   IndexMethod = "brin"
	Bloom  IndexMethod = "bloom"
)

type Index struct {
	Name    string
	Table   string
	Columns []string
	Method  IndexMethod
}

func (i Index) SQL(dialect string) (string, error) {
	if dialect != "postgres" && dialect != "postgresql" {
		return "", ErrUnsupportedDialect
	}
	method := i.Method
	if method == "" {
		method = BTree
	}
	columns := make([]string, len(i.Columns))
	for index, column := range i.Columns {
		columns[index] = quote(column)
	}
	return fmt.Sprintf("CREATE INDEX %s ON %s USING %s (%s)", quote(i.Name), quote(i.Table), method, strings.Join(columns, ", ")), nil
}
