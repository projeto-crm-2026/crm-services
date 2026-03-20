package utils

import (
	"database/sql"
	"strconv"
	"strings"
	"time"
)

func QueryAny(v string) any {
	if v == "" {
		return nil
	}
	return v
}

func QueryPtr(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

func QueryTags(v string) []string {
	if v == "" {
		return nil
	}
	return strings.Split(v, ",")
}

func QueryUint(v string) *uint {
	if v == "" {
		return nil
	}
	id, err := strconv.ParseUint(v, 10, 32)
	if err != nil {
		return nil
	}
	uid := uint(id)
	return &uid
}

func QueryDate(v string) *time.Time {
	if v == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", v)
	if err != nil {
		return nil
	}
	return &t
}

func NullString(ns sql.NullString) any {
	if ns.Valid {
		return ns.String
	}
	return nil
}
