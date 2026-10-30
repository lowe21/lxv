package model

import (
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/text/gstr"
)

func PolygonRaw(coords [][]float64) gdb.Raw {
	points := make([]string, 0, len(coords)+1)
	for _, coord := range coords {
		if len(coord) != 2 {
			continue
		}
		points = append(points, gstr.JoinAny(coord, " "))
	}

	if len(points) > 0 && points[0] != points[len(points)-1] {
		points = append(points, points[0])
	}

	return gdb.Raw("ST_GeomFromText('POLYGON((" + gstr.Join(points, ", ") + "))', 0)")
}
