package builders

import sq "github.com/Masterminds/squirrel"

type OrderByField string

type OrderByDirection string

const (
	AscDirection OrderByDirection = "ASC"
)

type OrderBy struct {
	Field     OrderByField
	Direction OrderByDirection
}

func newQueryBuilder() sq.StatementBuilderType {
	return sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
}
func newDeleteQueryBuilder() sq.StatementBuilderType {
	return sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
}

type WhereBuilder[T any] interface {
	Where(pred interface{}, args ...interface{}) T
}

type Specification interface {
	Predicates() []sq.Sqlizer
}

func ApplyFilter[B WhereBuilder[B]](builder B, s Specification) B {
	for _, p := range s.Predicates() {
		builder = builder.Where(p)
	}

	return builder
}
