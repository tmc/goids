package main

import (
	"fmt"
	"math"
	"os"
	"runtime"
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

	world := &goids.World{
		Goids: []*goids.Goid{
			{0, 0, goids.ToVel(0, 0.001), goids.ColorBlue},
			{0, -1, goids.ToVel(0, 0.001), goids.ColorGreen},
			{-0.1, -1, goids.ToVel(0, 0.001), goids.ColorRed},
		},
	}
	/*
		for i := 0; i < 100; i++ {
			world.Goids = append(world.Goids, goids.NewGoid(0.001+0.0001*(float32(i))))
		}
		for i := 0; i < 100; i++ {
			world.Goids = append(world.Goids, goids.NewGoid(0.002))
		}
	*/
	/*
		go func() {
			return
			for {
				dt := time.Millisecond * 10
				time.Sleep(dt)
				headingInc := float32(1)
				for _, goid := range world.Goids {
					goid.Heading += headingInc
					if goid.Heading >= 360 {
						goid.Heading = 0
					}
				}
			}
		}()
	*/
	//gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	previousTime := glfw.GetTime()
	for !window.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		now := glfw.GetTime()
		elapsed := now - previousTime
		previousTime = now

		world.Step(time.Duration(elapsed) * time.Millisecond)

		gl.BindVertexArray(vao)
		for _, goid := range world.Goids {
			goidModel := model
			goidModel = goidModel.Mul4(mgl32.Translate3D(goid.X, goid.Y, 0))
			goidModel = goidModel.Mul4(mgl32.HomogRotate3DZ(mgl32.DegToRad(goid.Heading())))
			goidModel = goidModel.Mul4(scaleDown)

			// first rotate model to point in the X+
			goidModel = goidModel.Mul4(mgl32.HomogRotate3DZ(-math.Pi / 2))
			gl.UniformMatrix4fv(modelUniform, 1, false, &goidModel[0])

			colorUniform := gl.GetUniformLocation(program, gl.Str("color\x00"))
			gl.Uniform4fv(colorUniform, 1, &goid.Color[0])
			gl.DrawArrays(gl.TRIANGLES, 0, 3)
		}
		window.SwapBuffers()
		glfw.PollEvents()
	}

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
