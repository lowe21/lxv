package model

import (
	"context"
	"database/sql"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gcode"

	"github.com/lowe21/lxv/pkg/errcode"
)

type Model interface {
	DB() gdb.DB
	Table() string
	Ctx(ctx context.Context) *gdb.Model
}

func Transaction(mod Model, ctx context.Context, fn func(context.Context, gdb.TX) error) (err error) {
	if fn != nil {
		inTransaction := gdb.TXFromCtx(ctx, mod.DB().GetGroup()) != nil

		invalidator := cacheInvalidatorFromCtx(ctx)
		if invalidator == nil {
			if inTransaction {
				return errcode.New(gcode.CodeDbOperationError, "transaction context missing cache invalidator")
			}
			invalidator = &cacheInvalidator{}
		}

		if err = mod.Ctx(ctx).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
			return fn(context.WithValue(ctx, invalidatorCtxKey, invalidator), tx)
		}); err != nil {
			return
		}

		if !inTransaction {
			invalidator.Flush(ctx, mod.DB())
		}
	}

	return
}

func Exist(mod Model, ctx context.Context, opts ...Option) (exist bool, err error) {
	options := parseOptions(opts...)

	model := mod.Ctx(ctx)
	for _, condition := range options.conditions {
		model = model.Where(condition)
	}

	return model.Exist()
}

func Count(mod Model, ctx context.Context, opts ...Option) (count int, err error) {
	options := parseOptions(opts...)

	model := mod.Ctx(ctx)
	for _, condition := range options.conditions {
		model = model.Where(condition)
	}

	return model.Count()
}

func Sum(mod Model, ctx context.Context, opts ...Option) (sum float64, err error) {
	options := parseOptions(opts...)

	model := mod.Ctx(ctx)
	for _, condition := range options.conditions {
		model = model.Where(condition)
	}

	return model.Sum(options.column)
}

func Avg(mod Model, ctx context.Context, opts ...Option) (avg float64, err error) {
	options := parseOptions(opts...)

	model := mod.Ctx(ctx)
	for _, condition := range options.conditions {
		model = model.Where(condition)
	}

	return model.Avg(options.column)
}

func FindList(mod Model, ctx context.Context, opts ...Option) (result gdb.Result, totalCount int, err error) {
	options := parseOptions(opts...)

	model := mod.Ctx(ctx)
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

	return model.Page(options.page, options.limit).AllAndCount(true)
}

func FindAll(mod Model, ctx context.Context, opts ...Option) (result gdb.Result, err error) {
	options := parseOptions(opts...)

	model := mod.Ctx(ctx)
	for _, condition := range options.conditions {
		model = model.Where(condition)
	}
	if options.order != "" {
		model = model.Order(options.order)
	}
	if options.limit > 0 {
		model = model.Limit(options.limit)
	}

	return model.All()
}

func FindOne(mod Model, ctx context.Context, opts ...Option) (record gdb.Record, err error) {
	options := parseOptions(opts...)
	if len(options.uk) == 0 {
		err = errcode.New(gcode.CodeDbOperationError, "unique key is empty")
		return
	}

	model := mod.Ctx(ctx).Where(options.uk)
	if len(options.conditions) > 0 {
		for _, condition := range options.conditions {
			model = model.Where(condition)
		}
	} else {
		model = model.Cache(
			cacheOption(mod.DB(), mod.Table(), options.cacheKey, time.Hour),
		)
	}

	return model.One()
}

func InsertOne(mod Model, ctx context.Context, do any, opts ...Option) (result sql.Result, err error) {
	options := parseOptions(opts...)

	model := mod.Ctx(ctx)
	if len(options.cacheKeys) > 0 {
		model = model.Hook(
			cacheHandler(mod.DB(), mod.Table(), options.cacheKeys...),
		)
	}

	return model.Insert(do)
}

func UpdateOne(mod Model, ctx context.Context, do any, opts ...Option) (result sql.Result, err error) {
	options := parseOptions(opts...)
	if len(options.uk) == 0 {
		err = errcode.New(gcode.CodeDbOperationError, "unique key is empty")
		return
	}

	model := mod.Ctx(ctx).Where(options.uk)
	for _, condition := range options.conditions {
		model = model.Where(condition)
	}
	if len(options.cacheKeys) > 0 {
		model = model.Hook(
			cacheHandler(mod.DB(), mod.Table(), options.cacheKeys...),
		)
	}

	return model.Update(do)
}

func DeleteOne(mod Model, ctx context.Context, opts ...Option) (result sql.Result, err error) {
	options := parseOptions(opts...)
	if len(options.uk) == 0 {
		err = errcode.New(gcode.CodeDbOperationError, "unique key is empty")
		return
	}

	model := mod.Ctx(ctx).Where(options.uk)
	for _, condition := range options.conditions {
		model = model.Where(condition)
	}
	if len(options.cacheKeys) > 0 {
		model = model.Hook(
			cacheHandler(mod.DB(), mod.Table(), options.cacheKeys...),
		)
	}

	return model.Delete()
}

func DeleteCache(mod Model, ctx context.Context, opts ...Option) (err error) {
	options := parseOptions(opts...)

	if len(options.cacheKeys) > 0 {
		cacheInvalidate(ctx, mod.DB(), mod.Table(), options.cacheKeys...)
	}

	return
}
