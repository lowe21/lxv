package model

import (
	"github.com/gogf/gf/v2/text/gstr"
)

func WhereEqual(column string) string {
	return parseColumn(column) + "=?"
}

func WhereNotEqual(column string) string {
	return parseColumn(column) + "!=?"
}

func WhereLT(column string) string {
	return parseColumn(column) + "<?"
}

func WhereLTE(column string) string {
	return parseColumn(column) + "<=?"
}

func WhereGT(column string) string {
	return parseColumn(column) + ">?"
}

func WhereGTE(column string) string {
	return parseColumn(column) + ">=?"
}

func WhereBetween(column string) string {
	return parseColumn(column) + " BETWEEN ? AND ?"
}

func WhereNotBetween(column string) string {
	return parseColumn(column) + " NOT BETWEEN ? AND ?"
}

func WhereIn(column string) string {
	return parseColumn(column) + " IN (?)"
}

func WhereNotIn(column string) string {
	return parseColumn(column) + " NOT IN (?)"
}

func WhereLike(column string) string {
	return parseColumn(column) + " LIKE ?"
}

func WhereNotLike(column string) string {
	return parseColumn(column) + " NOT LIKE ?"
}

func WhereIsNull(column string) string {
	return parseColumn(column) + " IS NULL"
}

func WhereIsNotNull(column string) string {
	return parseColumn(column) + " IS NOT NULL"
}

func parseColumn(column string) string {
	return "`" + gstr.ReplaceByArray(column, []string{"`", "``", ".", "`.`"}) + "`"
}
