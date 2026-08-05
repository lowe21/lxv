package graylog

import (
	"context"
	"crypto/rand"
	"log"
	"math"

	"github.com/gogf/gf/v2/encoding/gcompress"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/net/gudp"
	"github.com/gogf/gf/v2/os/gtimer"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gogf/gf/v2/util/gutil"
)

type Graylog struct {
	options *Options
	gelf    chan *Gelf
}

func (g *Graylog) Send(ctx context.Context, gelf *Gelf) {
	if err := gutil.Try(ctx, func(_ context.Context) {
		if gelf != nil {
			gelf.Version = g.options.Version
		}

		g.gelf <- gelf
	}); err != nil {
		log.Print(err)
	}
}

func (g *Graylog) worker() {
	conn, err := gudp.NewClientConn(g.options.Address)
	if err != nil {
		goto ERROR
	}

	for gelf := range g.gelf {
		chunks, err := g.compress(gelf)
		if err != nil {
			continue
		}
		for _, chunk := range chunks {
			if conn.Send(chunk) != nil {
				goto ERROR
			}
		}
	}
ERROR:
	if conn != nil {
		_ = conn.Close()
	}

	log.Printf("graylog connect failed, try to reconnect after %.0f seconds", g.options.ReconnectInterval.Seconds())
	gtimer.SetTimeout(context.Background(), g.options.ReconnectInterval, func(_ context.Context) {
		g.worker()
	})
}

func (g *Graylog) compress(gelf *Gelf) (chunks [][]byte, err error) {
	json, err := gjson.Encode(gelf)
	if err != nil {
		return
	}

	data, err := gcompress.Gzip(json)
	if err != nil {
		return
	}
	if dataSize := len(data); dataSize > g.options.ChunkSize {
		id := make([]byte, 8)
		if _, err = rand.Read(id); err != nil {
			return
		}

		currentSize := 0
		currentNumber := 0
		chunkNumber := gconv.Int(math.Ceil(gconv.Float64(dataSize) / gconv.Float64(g.options.ChunkSize)))
		for currentSize < dataSize && currentNumber < chunkNumber {
			nextSize := currentSize + g.options.ChunkSize

			chunk := []byte{0x1e, 0x0f}
			chunk = append(chunk, id...)
			chunk = append(chunk, gconv.Byte(currentNumber%128))
			chunk = append(chunk, gconv.Byte(chunkNumber%128))
			if nextSize < dataSize {
				chunk = append(chunk, data[currentSize:nextSize]...)
			} else {
				chunk = append(chunk, data[currentSize:]...)
			}
			chunks = append(chunks, chunk)

			currentSize = nextSize
			currentNumber += 1
		}
	} else {
		chunks = [][]byte{data}
	}

	return
}
