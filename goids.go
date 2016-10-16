package goids

import (
	"math"
	"math/rand"
	"time"

	"github.com/go-gl/mathgl/mgl32"
)

const (
	maxSpeed = 0.001
	maxForce = 0.0001
)

type World struct {
	Goids []*Goid
}

func (w *World) Step(dt time.Duration) {
	for _, g := range w.Goids {
		g.Step(w, dt)
	}
	// Separation
	// Velocity matching
	// Cohesion
}

type Goid struct {
	X        float32
	Y        float32
	Velocity mgl32.Vec2
	Color    mgl32.Vec4
}

func (g *Goid) Heading() float32 {
	x, y := g.Velocity.Elem()
	h := float32(math.Atan2(float64(y), float64(x)))
	return mgl32.RadToDeg(h)
}

func (g *Goid) Speed() float32 {
	return g.Velocity.Len()
}

func ToVel(degrees float32, speed float32) mgl32.Vec2 {
	result := mgl32.Vec2{}
	theta := mgl32.DegToRad(degrees)
	result[0] = float32(math.Cos(float64(theta)))
	result[1] = float32(math.Sin(float64(theta)))
	result = result.Mul(speed)
	return result
}

func NewGoid(speed float32) *Goid {
	color := ColorGreen
	if rand.NormFloat64() > 0.5 {
		color = ColorBlue
	}
	return &Goid{
		X:        0,
		Y:        0,
		Velocity: ToVel(float32(rand.Intn(360)), speed),
		Color:    color,
	}
}

func (g *Goid) Step(w *World, t time.Duration) {
	//sep := separation(w.Goids)
	accel := mgl32.Vec2{}
	c := g.cohesion(w.Goids)
	accel = accel.Add(c)

	speed := g.Velocity.Len()
	v := g.Velocity
	v = v.Add(accel)
	v = v.Normalize().Mul(speed)
	g.Velocity = v

	g.X += v.X()
	g.Y += v.Y()
}

func (g *Goid) cohesion(goids []*Goid) mgl32.Vec2 {
	// compute average pos
	result := mgl32.Vec2{}
	for _, goid := range goids {
		result = result.Add(mgl32.Vec2{goid.X, goid.Y})
	}
	target := result.Mul(1.0 / float32(len(goids)))
	desired := target.Sub(mgl32.Vec2{g.X, g.Y})
	return desired.Sub(g.Velocity).Normalize().Mul(maxForce)
}

func separation(goids []*Goid) mgl32.Vec2 {
	return mgl32.Vec2{}
}
