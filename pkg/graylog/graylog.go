package graylog

import (
	"crypto/rand"
	"log"
	"math"
	"time"

	"github.com/gogf/gf/v2/encoding/gcompress"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/net/gudp"

	"github.com/lowe21/lxv/pkg/errcode"
)

type Graylog struct {
	options *Options
	gelf    chan *Gelf
}

func (g *Graylog) Send(gelf *Gelf) {
	if gelf != nil {
		gelf.Version = g.options.Version
	}

	g.gelf <- gelf
}

func (g *Graylog) worker() {
	for {
		conn, err := gudp.NewClientConn(g.options.Address)
		if err != nil {
			log.Printf("worker error, %v", err)
		} else {
		loop:
			for gelf := range g.gelf {
				chunks, err := g.compress(gelf)
				if err != nil {
					log.Printf("compress error, %v", err)
					continue
				}
				for _, chunk := range chunks {
					if conn.Send(chunk) != nil {
						_ = conn.Close()
						break loop
					}
				}
			}
		}

		<-time.After(g.options.ReconnectInterval)
	}
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

		chunkNumber := int(math.Ceil(float64(dataSize) / float64(g.options.ChunkSize)))
		if chunkNumber > 128 {
			err = errcode.New("chunks too large")
			return
		}

		currentSize := 0
		currentNumber := 0

		for currentSize < dataSize && currentNumber < chunkNumber {
			nextSize := currentSize + g.options.ChunkSize

			chunk := []byte{0x1e, 0x0f}
			chunk = append(chunk, id...)
			chunk = append(chunk, byte(currentNumber))
			chunk = append(chunk, byte(chunkNumber))
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
