package graylog

import (
	"crypto/rand"
	"fmt"
	"log"
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
					if err := conn.Send(chunk); err != nil {
						_ = conn.Close()
						log.Printf("send error, %v", err)
						break loop
					}
				}
			}
		}

		time.Sleep(g.options.ReconnectInterval)
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
	if dataSize := len(data); dataSize > g.options.MaxChunkSize {
		headerSize := 12
		if g.options.MaxChunkSize <= headerSize {
			err = errcode.New(fmt.Sprintf("max chunk size must be greater than header size %d", headerSize))
			return
		}

		chunkSize := g.options.MaxChunkSize - headerSize
		chunkNumber := (dataSize + chunkSize - 1) / chunkSize
		if chunkNumber > 128 {
			err = errcode.New("chunks too large")
			return
		}

		id := make([]byte, 8)
		if _, err = rand.Read(id); err != nil {
			return
		}

		chunks = make([][]byte, 0, chunkNumber)
		currentSize := 0
		currentNumber := 0
		for currentSize < dataSize {
			nextSize := min(currentSize+chunkSize, dataSize)
			chunk := make([]byte, 0, headerSize+nextSize-currentSize)
			chunk = append(chunk, 0x1e, 0x0f)
			chunk = append(chunk, id...)
			chunk = append(chunk, byte(currentNumber), byte(chunkNumber))
			chunk = append(chunk, data[currentSize:nextSize]...)
			chunks = append(chunks, chunk)
			currentSize = nextSize
			currentNumber++
		}
	} else {
		chunks = [][]byte{data}
	}

	return
}
