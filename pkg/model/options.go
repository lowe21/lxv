package model

import (
	"github.com/gogf/gf/v2/crypto/gsha256"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
)

type Sort string

const (
	Asc  Sort = "ASC"
	Desc Sort = "DESC"
)

type Options struct {
	uk         gdb.Map
	cacheKey   string
	column     string
	conditions []gdb.Map
	order      string
	page       int
	limit      int
}

type Option func(*Options)

func WithUK(column string, value any) Option {
	return func(options *Options) {
		if column != "" {
			options.uk[column] = value
		}
	}
}

func WithColumn(column string) Option {
	return func(options *Options) {
		if column != "" {
			options.column = column
		}
	}
}

func WithCondition(condition gdb.Map) Option {
	return func(options *Options) {
		if condition != nil {
			options.conditions = append(options.conditions, condition)
		}
	}
}

func WithOrder(column string, sort Sort) Option {
	return func(options *Options) {
		if column != "" {
			options.order = column
			if sort == Asc || sort == Desc {
				options.order = gstr.Join([]string{options.order, string(sort)}, " ")
			}
		}
	}
}

func WithPage(page int) Option {
	return func(options *Options) {
		if page > 0 {
			options.page = page
		}
	}
}

func WithLimit(limit int) Option {
	return func(options *Options) {
		if limit > 0 {
			options.limit = limit
		}
	}
}

func parseOptions(opts ...Option) *Options {
	options := &Options{
		uk: make(gdb.Map),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(options)
		}
	}
	if len(options.uk) > 0 {
		options.cacheKey = gsha256.Encrypt(gconv.String(options.uk))
	}

	return options
}
