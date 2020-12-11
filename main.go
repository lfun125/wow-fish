package main

import (
	"fish/circle"
	"time"

	"github.com/go-vgo/robotgo"
)

func main() {
	x, y := robotgo.GetScreenSize()
	radius := 100
	c := circle.NewCircle(radius, x, y)
	for _, r := range c.ListCoordinates() {
		robotgo.MoveMouse(r.X, r.Y)
		time.Sleep(5 * time.Millisecond)
	}
}
