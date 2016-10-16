package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/tmc/goids"
)

func init() {
	// This is needed to arrange that main() runs on main thread.
	// See documentation for functions that are only allowed to be called from the main thread.
	runtime.LockOSThread()
}

var (
	colorGreen = mgl32.Vec4{0.1, 0.8, 0.1, 1}
	colorBlue  = mgl32.Vec4{0.1, 0.1, 0.8, 1}
)

func main() {
	err := glfw.Init()
	if err != nil {
		panic(err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	windowHeight, windowWidth := 640, 640
	window, err := glfw.CreateWindow(windowHeight, windowWidth, "Goids!", nil, nil)
	if err != nil {
		panic(err)
	}
	window.SetPos(0, 0)
	window.MakeContextCurrent()
	window.SetKeyCallback(keyCallback)

	// Initialize Glow
	if err := gl.Init(); err != nil {
		panic(err)
	}

	version := gl.GoStr(gl.GetString(gl.VERSION))
	fmt.Println("OpenGL version", version)

	vertexShader, err := goids.LoadShader("main.vert")
	if err != nil {
		panic(err)
	}
	fragmentShader, err := goids.LoadShader("main.frag")
	if err != nil {
		panic(err)
	}
	program, err := goids.NewProgram(vertexShader, fragmentShader)
	if err != nil {
		panic(err)
	}
	gl.UseProgram(program)

	model := mgl32.Ident4()
	modelUniform := gl.GetUniformLocation(program, gl.Str("model\x00"))
	// scale down
	scaleDown := model.Mul4(mgl32.Scale3D(0.06, 0.06, 1))

	projection := mgl32.Perspective(mgl32.DegToRad(45.0), float32(windowWidth)/float32(windowHeight), 0.1, 10.0)
	projectionUniform := gl.GetUniformLocation(program, gl.Str("projection\x00"))
	gl.UniformMatrix4fv(projectionUniform, 1, false, &projection[0])

	camera := mgl32.LookAtV(mgl32.Vec3{0, 0, 5.0}, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})
	cameraUniform := gl.GetUniformLocation(program, gl.Str("camera\x00"))
	gl.UniformMatrix4fv(cameraUniform, 1, false, &camera[0])

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(goidOutline)*4, gl.Ptr(goidOutline), gl.STATIC_DRAW)

	vertAttrib := uint32(gl.GetAttribLocation(program, gl.Str("vert\x00")))
	gl.EnableVertexAttribArray(vertAttrib)
	gl.VertexAttribPointer(vertAttrib, 3, gl.FLOAT, false, 3*4, gl.PtrOffset(0))
	gl.ClearColor(0.1, 0.1, 0.1, 1.0)

	world := &World{
		Goids: []*Goid{
			{0, 0, 0, 0.01, colorBlue},
			{0, 0.5, 90, 0.01, colorBlue},
			{0, 0, 180, 0.01, colorBlue},
		},
	}
	for i := 0; i < 100; i++ {
		world.Goids = append(world.Goids, newGoid(0.01+0.0001*(float32(i))))
	}
	for i := 0; i < 100; i++ {
		world.Goids = append(world.Goids, newGoid(0.02))
	}
	mu := sync.Mutex{}
	go func() {
		for {
			dt := time.Millisecond * 10
			time.Sleep(dt)
			mu.Lock()
			headingInc := float32(1)
			for _, goid := range world.Goids {
				goid.Heading += headingInc
				if goid.Heading >= 360 {
					goid.Heading = 0
				}
			}
			mu.Unlock()
		}
	}()
	//gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	previousTime := glfw.GetTime()
	for !window.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		now := glfw.GetTime()
		elapsed := now - previousTime
		previousTime = now

		mu.Lock()
		world.Step(time.Duration(elapsed) * time.Millisecond)
		mu.Unlock()

		gl.BindVertexArray(vao)
		for _, goid := range world.Goids {
			goidModel := model
			goidModel = goidModel.Mul4(mgl32.Translate3D(goid.X, goid.Y, 0))
			goidModel = goidModel.Mul4(mgl32.HomogRotate3DZ(mgl32.DegToRad(goid.Heading)))
			goidModel = goidModel.Mul4(scaleDown)

			gl.UniformMatrix4fv(modelUniform, 1, false, &goidModel[0])

			colorUniform := gl.GetUniformLocation(program, gl.Str("color\x00"))
			gl.Uniform4fv(colorUniform, 1, &goid.Color[0])
			gl.DrawArrays(gl.TRIANGLES, 0, 3)
		}
		window.SwapBuffers()
		glfw.PollEvents()
	}

}

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

func newGoid(speed float32) *Goid {
	color := colorGreen
	if rand.NormFloat64() > 0.5 {
		color = colorBlue
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

var goidOutline = []float32{
	-1, -1, 0,
	0, 2, 0,
	1, -1, 0,
}

func keyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press {
		switch key {
		case glfw.KeyEscape:
			glfw.Terminate()
			os.Exit(0)
		}
	}
}
