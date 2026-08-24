package crontab

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"

	"github.com/gogf/gf/v2/container/garray"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcron"
)

type (
	Handler interface {
		Name() string
		Run(ctx context.Context) error
	}

	Config struct {
		Name    string
		Pattern string
	}
)

var (
	handlers map[string]Handler
	started  bool
	mutex    sync.RWMutex
	once     sync.Once
)

func SetHandler(handler Handler) {
	mutex.Lock()
	defer mutex.Unlock()

	if started {
		panic("cannot set handler after crontab started")
	}

	name := handler.Name()
	if name == "" {
		panic(fmt.Sprintf("%T missing handler name", handler))
	}

	if handlers == nil {
		handlers = make(map[string]Handler)
	}
	if _, ok := handlers[name]; ok {
		panic(fmt.Sprintf(`%T handler name "%s" already exists`, handler, name))
	}
	handlers[name] = handler
}

func Start() {
	once.Do(func() {
		mutex.Lock()
		defer mutex.Unlock()

		started = true

		configs := make([]*Config, 0)
		if err := g.Config().MustGet(nil, "crontab").Scan(&configs); err != nil {
			panic(err)
		}

		rows := []string{"#", "NAME", "PATTERN", "HANDLER", "STATUS"}
		table := tablewriter.NewTable(os.Stdout, tablewriter.WithConfig(tablewriter.Config{
			Header: tw.CellConfig{
				Merging: tw.CellMerging{Mode: tw.MergeBoth},
			},
			Row: tw.CellConfig{
				Alignment: tw.CellAlignment{PerColumn: []tw.Align{tw.AlignCenter}},
			},
		}))
		table.Header(garray.New().Pad(len(rows), "CRONTAB").Slice())
		if err := table.Append(rows); err != nil {
			panic(err)
		}
		defer func() {
			if exception := recover(); exception != nil {
				panic(exception)
			}
			if len(configs) > 0 {
				_ = table.Render()
			} else {
				_ = table.Close()
			}
		}()

		for index, config := range configs {
			name := config.Name
			pattern := config.Pattern
			handler, ok := handlers[name]
			if !ok {
				panic(fmt.Sprintf(`crontab "%s" not found`, name))
			}

			if err := Runner(context.Background(), name, pattern, handler); err != nil {
				panic(err)
			}

			if err := table.Append(index+1, name, pattern, fmt.Sprintf("%T", handler), "OK"); err != nil {
				panic(err)
			}
		}
	})
}

func Runner(ctx context.Context, name, pattern string, handler Handler) (err error) {
	if _, err = gcron.AddSingleton(ctx, pattern, func(ctx context.Context) {
		defer func() {
			if exception := recover(); exception != nil {
				g.Log().Error(ctx, exception)
			}
		}()

		if err := handler.Run(ctx); err != nil {
			g.Log().Errorf(ctx, `crontab "%s" error: %v`, name, err)
		} else {
			g.Log().Infof(ctx, `crontab "%s" executed`, name)
		}
	}, name); err != nil {
		return
	}

	return
}
