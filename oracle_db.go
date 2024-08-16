package core

import "database/sql"

type oracleDBSession struct {
	*sql.DB
}
