package core

type TailQuery struct {
	queries    []string
	operator   []string
	isHasWhere bool
}

const (
	OPERATOR_AND  = "AND"
	OPERATOR_OR   = "OR"
	OPERATOR_NONE = " "
)

func NewTailQuery() *TailQuery {
	return &TailQuery{
		queries:    []string{},
		operator:   []string{},
		isHasWhere: true,
	}
}

func (dbWhere *TailQuery) Add(query string, operator string) {
	dbWhere.queries = append(dbWhere.queries, query)
	if len(dbWhere.queries) > 1 {
		dbWhere.operator = append(dbWhere.operator, operator)
	}
}

func (dbWhere *TailQuery) PopLastQuery() {
	if len(dbWhere.queries) != 0 {
		dbWhere.queries = dbWhere.queries[:len(dbWhere.queries)-1]
	}

	if len(dbWhere.operator) != 0 {
		dbWhere.operator = dbWhere.operator[:len(dbWhere.operator)-1]
	}
}

func (dbWhere *TailQuery) GetQuery() string {
	whereQuery := " "
	if len(dbWhere.queries) == 0 {
		return whereQuery
	}

	if dbWhere.isHasWhere {
		whereQuery += "WHERE "
	}

	whereQuery += dbWhere.queries[0]
	for i := 1; i < len(dbWhere.queries); i++ {
		whereQuery += " " + dbWhere.operator[i-1] + " " + dbWhere.queries[i]
	}

	return whereQuery
}

func (dbWhere *TailQuery) HasWhere(val bool) {
	dbWhere.isHasWhere = val
}
