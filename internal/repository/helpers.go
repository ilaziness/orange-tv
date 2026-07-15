package repository

import "github.com/uptrace/bun"

func bunIn(ids []int64) bun.InValues {
	return bun.In(ids)
}
