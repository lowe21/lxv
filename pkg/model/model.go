package model

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gcode"

	"github.com/lowe21/lxv/pkg/errcode"
)

type DAO[T any] interface {
	DB() gdb.DB
	Table() string
	Ctx(ctx context.Context) *gdb.Model
	Entity() *T
}

func Transaction[T any](ctx context.Context, dao DAO[T], fn func(context.Context, gdb.TX) error) (err error) {
	if fn != nil {
		inTransaction := gdb.TXFromCtx(ctx, dao.DB().GetGroup()) != nil

		invalidator := cacheInvalidatorFromCtx(ctx)
		if invalidator == nil {
			if inTransaction {
				return errcode.New(gcode.CodeDbOperationError, "transaction context missing cache invalidator")
			}
			invalidator = &cacheInvalidator{}
		}

		if err = dao.Ctx(ctx).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
			return fn(context.WithValue(ctx, invalidatorCtxKey, invalidator), tx)
		}); err != nil {
			return
		}

		if !inTransaction {
			invalidator.Flush(ctx, dao.DB())
		}
	}

	return
}

func Exist[T any](ctx context.Context, dao DAO[T], opts ...Option) (exist bool, err error) {
	options := parseOptions(opts...)

	model := dao.Ctx(ctx)
	for _, condition := range options.conditions {
		model = model.Where(condition)
	}

	return model.Exist()
}

func Count[T any](ctx context.Context, dao DAO[T], opts ...Option) (count int, err error) {
	options := parseOptions(opts...)

	model := dao.Ctx(ctx)
	for _, condition := range options.conditions {
		model = model.Where(condition)
	}

	return model.Count()
}

func Sum[T any](ctx context.Context, dao DAO[T], opts ...Option) (sum float64, err error) {
	options := parseOptions(opts...)

	model := dao.Ctx(ctx)
	for _, condition := range options.conditions {
		model = model.Where(condition)
	}

	return model.Sum(options.column)
}

func Avg[T any](ctx context.Context, dao DAO[T], opts ...Option) (avg float64, err error) {
	options := parseOptions(opts...)

	model := dao.Ctx(ctx)
	for _, condition := range options.conditions {
		model = model.Where(condition)
	}

	return model.Avg(options.column)
}

func FindList[T any](ctx context.Context, dao DAO[T], opts ...Option) (entities []*T, totalCount int, err error) {
	options := parseOptions(opts...)

	model := dao.Ctx(ctx)
	for _, condition := range options.conditions {
		model = model.Where(condition)
	}
	if options.order != "" {
		model = model.Order(options.order)
	}
	if options.page <= 0 {
		options.page = 1
	}
	if options.limit <= 0 {
		options.limit = 10
	}

	result, totalCount, err := model.Page(options.page, options.limit).AllAndCount(true)
	if err != nil {
		return
	}

	err = result.Structs(&entities)

	return
}

func FindAll[T any](ctx context.Context, dao DAO[T], opts ...Option) (entities []*T, err error) {
	options := parseOptions(opts...)

	model := dao.Ctx(ctx)
	for _, condition := range options.conditions {
		model = model.Where(condition)
	}
	if options.order != "" {
		model = model.Order(options.order)
	}
	if options.limit > 0 {
		model = model.Limit(options.limit)
	}

	err = model.Scan(&entities)

	return
}

func FindOne[T any](ctx context.Context, dao DAO[T], opts ...Option) (entity *T, err error) {
	options := parseOptions(opts...)
	if len(options.uk) == 0 {
		err = errcode.New(gcode.CodeDbOperationError, "unique key is empty")
		return
	}

	model := dao.Ctx(ctx).Where(options.uk)
	if len(options.conditions) > 0 {
		for _, condition := range options.conditions {
			model = model.Where(condition)
		}
	} else {
		model = model.Cache(
			cacheOption(dao.DB(), dao.Table(), options.cacheKey, time.Hour),
		)
	}

	err = model.Scan(&entity)

	return
}

func InsertOne[T any](ctx context.Context, dao DAO[T], do any, opts ...Option) (lastInsertID int64, err error) {
	options := parseOptions(opts...)

	model := dao.Ctx(ctx)
	if len(options.cacheKeys) > 0 {
		model = model.Hook(
			cacheHandler(dao.DB(), dao.Table(), options.cacheKeys...),
		)
	}

	return model.InsertAndGetId(do)
}

func UpdateOne[T any](ctx context.Context, dao DAO[T], do any, opts ...Option) (affected int64, err error) {
	options := parseOptions(opts...)
	if len(options.uk) == 0 {
		err = errcode.New(gcode.CodeDbOperationError, "unique key is empty")
		return
	}

	model := dao.Ctx(ctx).Where(options.uk)
	for _, condition := range options.conditions {
		model = model.Where(condition)
	}
	if len(options.cacheKeys) > 0 {
		model = model.Hook(
			cacheHandler(dao.DB(), dao.Table(), options.cacheKeys...),
		)
	}

	return model.Limit(1).UpdateAndGetAffected(do)
}

func DeleteOne[T any](ctx context.Context, dao DAO[T], opts ...Option) (affected int64, err error) {
	options := parseOptions(opts...)
	if len(options.uk) == 0 {
		err = errcode.New(gcode.CodeDbOperationError, "unique key is empty")
		return
	}

	model := dao.Ctx(ctx).Where(options.uk)
	for _, condition := range options.conditions {
		model = model.Where(condition)
	}
	if len(options.cacheKeys) > 0 {
		model = model.Hook(
			cacheHandler(dao.DB(), dao.Table(), options.cacheKeys...),
		)
	}

	result, err := model.Limit(1).Delete()
	if err != nil {
		return
	}

	return result.RowsAffected()
}

func DeleteCache[T any](ctx context.Context, dao DAO[T], opts ...Option) (err error) {
	options := parseOptions(opts...)

	if len(options.cacheKeys) > 0 {
		cacheInvalidate(ctx, dao.DB(), dao.Table(), options.cacheKeys...)
	}

	return
}
