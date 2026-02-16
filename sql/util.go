package sql

import "database/sql"

func isAffected(res sql.Result) bool {
	rows, err := res.RowsAffected()
	if err != nil {
		return true
	}
	return rows != 0
}

func getAffected(res sql.Result) int64 {
	ret, err := res.RowsAffected()
	if err != nil {
		return -1
	}
	return ret
}
