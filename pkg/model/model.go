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
	Ctx(ctx context.Context) *gdb.Model
	cacheOption(key string, ttl ...time.Duration) gdb.CacheOption
}

func Transaction(model Model, ctx context.Context, fn func(context.Context, gdb.TX) error) (err error) {
	if fn == nil {
		return
	}

	isTx := gdb.TXFromCtx(ctx, model.DB().GetGroup()) != nil

	invalidator := cacheInvalidatorFromCtx(ctx)
	if isTx && invalidator == nil {
		return errcode.New(gcode.CodeDbOperationError, "transaction context missing cache invalidator")
	}

	if invalidator == nil {
		invalidator = &cacheInvalidator{}
		ctx = context.WithValue(ctx, invalidatorCtxKey, invalidator)
	}

	if err = model.Ctx(ctx).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		return fn(context.WithValue(ctx, invalidatorCtxKey, invalidator), tx)
	}); err != nil {
		return
	}

	if !isTx {
		invalidator.flush(ctx, model.DB())
	}

	return
}

func Exist(model Model, ctx context.Context, opts ...Option) (exist bool, err error) {
	options := parseOptions(opts...)

	m := model.Ctx(ctx)
	for _, condition := range options.conditions {
		m = m.Where(condition)
	}

	return m.Exist()
}

func Count(model Model, ctx context.Context, opts ...Option) (count int, err error) {
	options := parseOptions(opts...)

	m := model.Ctx(ctx)
	for _, condition := range options.conditions {
		m = m.Where(condition)
	}

	return m.Count()
}

func Sum(model Model, ctx context.Context, opts ...Option) (sum float64, err error) {
	options := parseOptions(opts...)

	m := model.Ctx(ctx)
	for _, condition := range options.conditions {
		m = m.Where(condition)
	}

	return m.Sum(options.column)
}

func Avg(model Model, ctx context.Context, opts ...Option) (avg float64, err error) {
	options := parseOptions(opts...)

	m := model.Ctx(ctx)
	for _, condition := range options.conditions {
		m = m.Where(condition)
	}

	return m.Avg(options.column)
}

func FindList(model Model, ctx context.Context, opts ...Option) (result gdb.Result, totalCount int, err error) {
	options := parseOptions(opts...)
	if options.page <= 0 {
		options.page = 1
	}
	if options.limit <= 0 {
		options.limit = 10
	}

	m := model.Ctx(ctx)
	for _, condition := range options.conditions {
		m = m.Where(condition)
	}
	if options.order != "" {
		m = m.Order(options.order)
	}

	return m.Page(options.page, options.limit).AllAndCount(true)
}

func FindAll(model Model, ctx context.Context, opts ...Option) (result gdb.Result, err error) {
	options := parseOptions(opts...)

	m := model.Ctx(ctx)
	for _, condition := range options.conditions {
		m = m.Where(condition)
	}
	if options.order != "" {
		m = m.Order(options.order)
	}
	if options.limit > 0 {
		m = m.Limit(options.limit)
	}

	return m.All()
}

func FindOne(model Model, ctx context.Context, opts ...Option) (record gdb.Record, err error) {
	options := parseOptions(opts...)
	if len(options.uk) == 0 {
		err = errcode.New(gcode.CodeDbOperationError, "unique key is empty")
		return
	}

	m := model.Ctx(ctx)
	if len(options.conditions) > 0 {
		for _, condition := range options.conditions {
			m = m.Where(condition)
		}
	} else {
		m = m.Cache(
			model.cacheOption(options.cacheKey, time.Hour),
		)
	}

	return m.One()
}

func InsertOne(model Model, ctx context.Context, do any, opts ...Option) (result sql.Result, err error) {
	options := parseOptions(opts...)

	m := model.Ctx(ctx)
	if options.cacheKey != "" {
		m = m.Hook(
			cacheHandler(model.DB(), model.cacheOption(options.cacheKey).Name),
		)
	}

	return m.Insert(do)
}

func UpdateOne(model Model, ctx context.Context, do any, opts ...Option) (result sql.Result, err error) {
	options := parseOptions(opts...)
	if len(options.uk) == 0 {
		err = errcode.New(gcode.CodeDbOperationError, "unique key is empty")
		return
	}

	m := model.Ctx(ctx).Where(options.uk)
	for _, condition := range options.conditions {
		m = m.Where(condition)
	}
	if options.cacheKey != "" {
		m = m.Hook(
			cacheHandler(model.DB(), model.cacheOption(options.cacheKey).Name),
		)
	}

	return m.Update(do)
}

func DeleteOne(model Model, ctx context.Context, opts ...Option) (result sql.Result, err error) {
	options := parseOptions(opts...)
	if len(options.uk) == 0 {
		err = errcode.New(gcode.CodeDbOperationError, "unique key is empty")
		return
	}

	m := model.Ctx(ctx).Where(options.uk)
	for _, condition := range options.conditions {
		m = m.Where(condition)
	}
	if options.cacheKey != "" {
		m = m.Hook(
			cacheHandler(model.DB(), model.cacheOption(options.cacheKey).Name),
		)
	}

	return m.Delete()
}

func DeleteCache(model Model, ctx context.Context, opts ...Option) (err error) {
	options := parseOptions(opts...)
	if len(options.uk) == 0 {
		err = errcode.New(gcode.CodeDbOperationError, "unique key is empty")
		return
	}

	if options.cacheKey != "" {
		cacheInvalidate(ctx, model.DB(), model.cacheOption(options.cacheKey).Name)
	}

	return
}
