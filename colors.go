package goids

import (
	"math"
	"math/rand"
	"time"

	"github.com/go-gl/mathgl/mgl32"
)

type World struct {
	Goids []*Goid
}

func (w *World) Step(dt time.Duration) {
	for _, g := range w.Goids {
		g.Step(dt)
	}
	// Separation
	// Velocity matching
	// Cohesion
}

type Goid struct {
	X       float32
	Y       float32
	Heading float32 // in degrees
	Speed   float32
	Color   mgl32.Vec4
}

func NewGoid(speed float32) *Goid {
	color := ColorGreen
	if rand.NormFloat64() > 0.5 {
		color = ColorBlue
	}
	return &Goid{
		X:       0,
		Y:       0,
		Heading: float32(rand.Intn(360)),
		Speed:   speed,
		Color:   color,
	}
}

func (g *Goid) Velocity() mgl32.Vec2 {
	velocity := mgl32.Vec2{}
	headingRad := float64(mgl32.DegToRad(g.Heading))
	// correct for '0' meaning Y+
	//headingRad += math.Pi
	velocity[0] = -1 * float32(math.Sin(headingRad)) * g.Speed
	velocity[1] = float32(math.Cos(headingRad)) * g.Speed
	return velocity
}

func (g *Goid) Step(t time.Duration) {
	v := g.Velocity()
	g.X += v.X()
	g.Y += v.Y()
}
