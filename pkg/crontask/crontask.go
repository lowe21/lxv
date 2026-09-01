package crontask

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"

	"github.com/gogf/gf/v2/container/garray"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcron"

	"github.com/lowe21/lxv/pkg/errcode"
)

type CronTask struct {
	options []*Option
	cron    *gcron.Cron
	ctx     context.Context
	cancel  context.CancelFunc
	once    sync.Once
}

func (c *CronTask) Start() {
	c.once.Do(func() {
		c.ctx, c.cancel = context.WithCancel(context.Background())

		rows := []string{"#", "NAME", "PATTERN", "TASKER", "STATUS"}
		table := tablewriter.NewTable(os.Stdout, tablewriter.WithConfig(tablewriter.Config{
			Header: tw.CellConfig{
				Merging: tw.CellMerging{Mode: tw.MergeBoth},
			},
			Row: tw.CellConfig{
				Alignment: tw.CellAlignment{PerColumn: []tw.Align{tw.AlignCenter}},
			},
		}))
		table.Header(garray.New().Pad(len(rows), "CRONTASK").Slice())
		if err := table.Append(rows); err != nil {
			panic(err)
		}

		index := 0
		for _, option := range c.options {
			tasker, err := GetTasker(option.Name)
			if err != nil {
				g.Log().Error(c.ctx, err)
				continue
			}
			if err = c.AddTask(c.ctx, option.Name, option.Pattern, tasker); err != nil {
				g.Log().Error(c.ctx, err)
				continue
			}
			index += 1
			if err = table.Append(index, option.Name, option.Pattern, fmt.Sprintf("%T", tasker), "OK"); err != nil {
				panic(err)
			}
		}

		if index > 0 {
			_ = table.Render()
		} else {
			_ = table.Close()
		}
	})
}

func (c *CronTask) AddTask(ctx context.Context, name, pattern string, tasker Tasker) (err error) {
	if c.cron.Search(name) != nil {
		err = errcode.New(fmt.Sprintf("task already exists, name: %s", name))
		return
	}

	if _, err = c.cron.AddSingleton(ctx, pattern, func(ctx context.Context) {
		defer func() {
			if exception := recover(); exception != nil {
				g.Log().Errorf(ctx, "task run panic, %+v", exception)
			}
		}()

		if err := tasker.Run(ctx); err != nil {
			g.Log().Error(ctx, gerror.Wrap(err, fmt.Sprintf("%s %s", name, pattern)))
		}
	}, name); err != nil {
		err = errcode.New(fmt.Sprintf("task add error, %v", err))
	}

	return
}

func (c *CronTask) RemoveTask(name string) {
	c.cron.Remove(name)
}

func (c *CronTask) Stop() {
	if c.cancel != nil {
		c.cancel()
	}
	c.cron.Stop()
}
