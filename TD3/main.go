package main

import (
	"encoding/gob"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"
)

// -------------------------------
type Vec2f struct {
	x, y float32
}

type Vec3f struct {
	x, y, z float32
}

func (v Vec3f) inverte() Vec3f {
	return Vec3f{-v.x, -v.y, -v.z}
}

func Add(v1, v2 Vec3f) Vec3f {
	return Vec3f{v1.x + v2.x, v1.y + v2.y, v1.z + v2.z}
}

func (v Vec3f) mul(f float32) Vec3f {
	return Vec3f{v.x * f, v.y * f, v.z * f}
}

func Mul(v1, v2 Vec3f) Vec3f {
	return Vec3f{v1.x * v2.x, v1.y * v2.y, v1.z * v2.z}
}
func Dot(v1, v2 Vec3f) float32 {
	return v1.x*v2.x + v1.y*v2.y + v1.z*v2.z
}

func cross(v1, v2 Vec3f) Vec3f {
	return Vec3f{v1.y*v2.z - v2.y*v1.z, v1.z*v2.x - v2.z*v1.x, v1.x*v2.y - v2.x*v1.y}
}

func (v Vec3f) norme() float32 {
	return float32(math.Sqrt(float64(v.x*v.x + v.y*v.y + v.z*v.z)))
}
func (v *Vec3f) normalize() {
	norme := v.norme()
	v.x /= norme
	v.y /= norme
	v.z /= norme
}
func (v Vec3f) normalized() Vec3f {
	norme := v.norme()
	return Vec3f{v.x / norme, v.y / norme, v.z / norme}
}

// --------------------------------
type rgbRepresentation struct {
	r, g, b uint8
}

// --------------------------------
type Image struct {
	frameBuffer   []rgbRepresentation
	width, height int
}

func (i Image) save(path string) error {
	// Création de l'image
	img := image.NewRGBA(image.Rect(0, 0, i.width, i.height))
	for y := 0; y < i.height; y++ {
		for x := 0; x < i.width; x++ {
			idx := (y*i.width + x)
			r, g, b := i.frameBuffer[idx].r, i.frameBuffer[idx].b, i.frameBuffer[idx].g
			img.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}

	pngFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer pngFile.Close()
	if err := png.Encode(pngFile, img); err != nil {
		return err
	}
	return nil
}

// ------------------
type Light struct {
	color    Vec3f
	position Vec3f
}

// --------------------------------
type Scene struct {
	objects []GeometricObject
	lights  []Light
}

func (s *Scene) addLight(l Light) {
	s.lights = append(s.lights, l)
}
func (s *Scene) addElement(g GeometricObject) {
	s.objects = append(s.objects, g)
}

type Phong struct {
	ka Vec3f
	kd Vec3f
	ks Vec3f
	n  float32
}

func (p Phong) render(rio, rdi, normal Vec3f, t float32, scene Scene) rgbRepresentation {
	hitPoint := Add(rio, rdi.mul(t))
	var finalColor Vec3f = Vec3f{0, 0, 0}

	for _, light := range scene.lights {
		lightDir := Add(light.position, hitPoint.inverte())
		lightDir.normalize()
		viewDir := rio.inverte().normalized()
		ambient := Mul(p.ka, light.color)
		diffuseFactor := Dot(normal, lightDir)
		if diffuseFactor < 0 {
			diffuseFactor = 0
		}
		diffuse := Mul(p.kd, light.color.mul(diffuseFactor))
		reflectDir := Add(lightDir.inverte(), normal.mul(2*Dot(normal, lightDir)))
		reflectDir.normalize()
		specularFactor := Dot(reflectDir, viewDir)
		if specularFactor < 0 {
			specularFactor = 0
		}
		specularFactor = float32(math.Pow(float64(specularFactor), float64(p.n)))
		specular := Mul(p.ks, light.color.mul(specularFactor))
		lightContribution := Add(ambient, Add(diffuse, specular))
		finalColor = Add(finalColor, lightContribution)
	}

	finalColor.x = float32(math.Min(float64(finalColor.x), 1.0))
	finalColor.y = float32(math.Min(float64(finalColor.y), 1.0))
	finalColor.z = float32(math.Min(float64(finalColor.z), 1.0))

	return rgbRepresentation{
		r: uint8(finalColor.x * 255),
		g: uint8(finalColor.y * 255),
		b: uint8(finalColor.z * 255),
	}
}

type RenderJob struct {
	StartX, EndX  int
	StartY, EndY  int
	Width, Height int
	Camera        Camera
	Scene         Scene
}

type RenderResult struct {
	StartX, StartY int
	Width, Height  int
	Pixels         []rgbRepresentation
}

type TCPServer struct {
	address          string
	scene            Scene
	camera           Camera
	imageWidth       int
	imageHeight      int
	clients          []net.Conn
	clientsMutex     sync.Mutex
	frameBuffer      []rgbRepresentation
	completedJobs    int
	totalJobs        int
	completedJobsMux sync.Mutex
}

func NewTCPServer(address string, scene Scene, camera Camera, width, height int) *TCPServer {
	return &TCPServer{
		address:     address,
		scene:       scene,
		camera:      camera,
		imageWidth:  width,
		imageHeight: height,
		frameBuffer: make([]rgbRepresentation, width*height),
	}
}

func (s *TCPServer) Start() error {
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		return fmt.Errorf("failed to start TCP server: %v", err)
	}
	defer listener.Close()

	fmt.Printf("Server listening on %s\n", s.address)

	var wg sync.WaitGroup
	wg.Add(1)
	go s.acceptConnections(listener, &wg)

	fmt.Println("Press Enter to start distributed rendering...")
	fmt.Scanln()

	s.distributeJobs()
	s.waitForCompletion()

	img := Image{s.frameBuffer, s.imageWidth, s.imageHeight}
	err = img.save("distributed_result.png")
	if err != nil {
		return fmt.Errorf("failed to save image: %v", err)
	}

	fmt.Println("Rendering complete! Image saved as distributed_result.png")

	s.clientsMutex.Lock()
	for _, client := range s.clients {
		client.Close()
	}
	s.clientsMutex.Unlock()

	wg.Wait()
	return nil
}

func (s *TCPServer) acceptConnections(listener net.Listener, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}

		fmt.Printf("New client connected: %s\n", conn.RemoteAddr())

		s.clientsMutex.Lock()
		s.clients = append(s.clients, conn)
		s.clientsMutex.Unlock()

		go s.handleClient(conn)
	}
}

func (s *TCPServer) handleClient(conn net.Conn) {
	decoder := gob.NewDecoder(conn)

	for {
		var result RenderResult
		err := decoder.Decode(&result)
		if err != nil {
			fmt.Printf("Client disconnected or error: %v\n", err)

			s.clientsMutex.Lock()
			// Remove client from our list
			for i, client := range s.clients {
				if client == conn {
					s.clients = append(s.clients[:i], s.clients[i+1:]...)
					break
				}
			}
			s.clientsMutex.Unlock()

			conn.Close()
			return
		}

		s.processResult(result)
	}
}

func (s *TCPServer) distributeJobs() {
	s.clientsMutex.Lock()
	numClients := len(s.clients)
	s.clientsMutex.Unlock()

	if numClients == 0 {
		fmt.Println("No clients connected. Rendering locally...")
		renderFrame(Image{s.frameBuffer, s.imageWidth, s.imageHeight}, s.camera, s.scene)
		return
	}

	rowsPerClient := s.imageHeight / numClients

	var jobs []RenderJob

	for i := 0; i < numClients; i++ {
		startY := i * rowsPerClient
		endY := startY + rowsPerClient

		if i == numClients-1 {
			endY = s.imageHeight
		}

		job := RenderJob{
			StartX: 0,
			EndX:   s.imageWidth,
			StartY: startY,
			EndY:   endY,
			Width:  s.imageWidth,
			Height: s.imageHeight,
			Camera: s.camera,
			Scene:  s.scene,
		}

		jobs = append(jobs, job)
	}

	s.totalJobs = len(jobs)
	fmt.Printf("Distributing %d jobs to %d clients\n", s.totalJobs, numClients)

	s.clientsMutex.Lock()
	for i, client := range s.clients {
		if i < len(jobs) {
			encoder := gob.NewEncoder(client)
			err := encoder.Encode(jobs[i])
			if err != nil {
				fmt.Printf("Error sending job to client: %v\n", err)
			}
		}
	}
	s.clientsMutex.Unlock()
}

func (s *TCPServer) processResult(result RenderResult) {
	for y := 0; y < result.Height; y++ {
		for x := 0; x < result.Width; x++ {
			globalX := result.StartX + x
			globalY := result.StartY + y

			if globalX >= 0 && globalX < s.imageWidth && globalY >= 0 && globalY < s.imageHeight {
				index := globalY*s.imageWidth + globalX
				resultIndex := y*result.Width + x

				if resultIndex < len(result.Pixels) && index < len(s.frameBuffer) {
					s.frameBuffer[index] = result.Pixels[resultIndex]
				}
			}
		}
	}

	s.completedJobsMux.Lock()
	s.completedJobs++
	completed := s.completedJobs
	total := s.totalJobs
	s.completedJobsMux.Unlock()

	fmt.Printf("Received results: %d/%d jobs completed\n", completed, total)
}

func (s *TCPServer) waitForCompletion() {
	for {
		s.completedJobsMux.Lock()
		if s.completedJobs >= s.totalJobs {
			s.completedJobsMux.Unlock()
			break
		}
		s.completedJobsMux.Unlock()

		time.Sleep(100 * time.Millisecond)
	}
}

type TCPClient struct {
	serverAddr string
	conn       net.Conn
	encoder    *gob.Encoder
	decoder    *gob.Decoder
}

func NewTCPClient(serverAddr string) (*TCPClient, error) {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %v", err)
	}

	return &TCPClient{
		serverAddr: serverAddr,
		conn:       conn,
		encoder:    gob.NewEncoder(conn),
		decoder:    gob.NewDecoder(conn),
	}, nil
}

func (c *TCPClient) Start(numWorkers int) error {
	fmt.Printf("Connected to server at %s\n", c.serverAddr)

	jobChan := make(chan RenderJob)
	resultChan := make(chan RenderResult)

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go c.renderWorker(i, jobChan, resultChan, &wg)
	}

	go func() {
		for result := range resultChan {
			err := c.encoder.Encode(result)
			if err != nil {
				fmt.Printf("Error sending result to server: %v\n", err)
				return
			}
			fmt.Println("Sent render result to server")
		}
	}()

	for {
		var job RenderJob
		err := c.decoder.Decode(&job)
		if err != nil {
			fmt.Printf("Server disconnected or error: %v\n", err)
			close(jobChan)
			break
		}

		fmt.Printf("Received job: Render region (%d,%d) to (%d,%d)\n",
			job.StartX, job.StartY, job.EndX, job.EndY)

		jobChan <- job
	}

	wg.Wait()
	close(resultChan)

	c.conn.Close()
	return nil
}

func (c *TCPClient) renderWorker(id int, jobs <-chan RenderJob, results chan<- RenderResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {
		fmt.Printf("Worker %d processing job...\n", id)

		width := job.EndX - job.StartX
		height := job.EndY - job.StartY

		pixels := make([]rgbRepresentation, width*height)

		ro := job.Camera.position
		cosFovy := float32(0.66)

		aspect := float32(job.Width) / float32(job.Height)
		horizontal := (cross(job.Camera.direction(), job.Camera.up)).normalized().mul(cosFovy * aspect)
		vertical := (cross(horizontal, job.Camera.direction())).normalized().mul(cosFovy)

		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				globalX := job.StartX + x
				globalY := job.StartY + y

				uvx := (float32(globalX) + float32(0.5)) / float32(job.Width)
				uvy := (float32(globalY) + float32(0.5)) / float32(job.Height)

				rd := Add(Add(job.Camera.direction(), horizontal.mul(uvx-float32(0.5))), vertical.mul(uvy-float32(0.5))).normalized()

				pixels[y*width+x] = renderPixel(job.Scene, ro, rd)
			}
		}

		result := RenderResult{
			StartX: job.StartX,
			StartY: job.StartY,
			Width:  width,
			Height: height,
			Pixels: pixels,
		}

		results <- result
		fmt.Printf("Worker %d completed job\n", id)
	}
}

func init() {
	gob.Register(Vec3f{})
	gob.Register(Camera{})
	gob.Register(Scene{})
	gob.Register(Light{})
	gob.Register(Sphere{})
	gob.Register(Lambert{})
	gob.Register(Phong{})
}

func serverMain() {
	scene := Scene{}
	populateSceneWithPhong(&scene)

	camera := Camera{Vec3f{0, 0, -5}, Vec3f{0, 1, 0}, Vec3f{0, 0, 5}}

	server := NewTCPServer(":8081", scene, camera, 2048, 2048)

	err := server.Start()
	if err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func clientMain() {
	client, err := NewTCPClient("localhost:8081")
	if err != nil {
		fmt.Printf("Client error: %v\n", err)
		return
	}

	numWorkers := 4
	err = client.Start(numWorkers)
	if err != nil {
		fmt.Printf("Client error: %v\n", err)
	}
}

func NewPhongMaterial(diffuseColor Vec3f, specularStrength float32, shininess float32) Phong {
	ambientCoef := 0.1

	return Phong{
		ka: diffuseColor.mul(float32(ambientCoef)),
		kd: diffuseColor,
		ks: Vec3f{specularStrength, specularStrength, specularStrength},
		n:  shininess,
	}
}

func populateSceneWithPhong(scene *Scene) {
	scene.addElement(Sphere{1, Vec3f{0, 0, 8}, NewPhongMaterial(Vec3f{1.0, 0, 0}, 0.8, 32)})
	scene.addElement(Sphere{0.3, Vec3f{2, 1.5, 4}, NewPhongMaterial(Vec3f{0.0, 1.0, 0}, 0.5, 16)})
	scene.addElement(Sphere{0.9, Vec3f{0, -1, 5}, Lambert{Vec3f{0.0, 0, 1.0}}})
	scene.addElement(Sphere{0.5, Vec3f{-2, -2, 5}, NewPhongMaterial(Vec3f{1.0, 1.0, 1.0}, 0.9, 64)})

	randomSpheres := generateRandomSpheresWithMixedMaterials(15, 0.2, 0.7, Vec3f{5, 5, 10})
	for _, sphere := range randomSpheres {
		scene.addElement(sphere)
	}

	scene.addLight(Light{Vec3f{1.0, 1.0, 1.0}, Vec3f{0, 10, 0}})
	scene.addLight(Light{Vec3f{0.5, 0.5, 0.8}, Vec3f{-10, 5, -5}})
}

func generateRandomSpheresWithMixedMaterials(count int, minRadius, maxRadius float32, boundingBox Vec3f) []Sphere {
	rand.Seed(time.Now().UnixNano())

	spheres := make([]Sphere, count)

	for i := 0; i < count; i++ {
		position := Vec3f{
			x: (rand.Float32()*2 - 1) * boundingBox.x,
			y: (rand.Float32()*2 - 1) * boundingBox.y,
			z: (rand.Float32()*2-1)*boundingBox.z + 5,
		}

		radius := minRadius + rand.Float32()*(maxRadius-minRadius)

		color := Vec3f{
			x: rand.Float32(),
			y: rand.Float32(),
			z: rand.Float32(),
		}

		var material Materials
		if rand.Float32() < 0.5 {
			material = Lambert{kd: color}
		} else {
			specularStrength := rand.Float32() * 0.9
			shininess := 8 + rand.Float32()*56
			material = NewPhongMaterial(color, specularStrength, shininess)
		}

		spheres[i] = Sphere{radius, position, material}
	}

	return spheres
}

// ----------------------------------
type Materials interface {
	render(rio, rdi, n Vec3f, t float32, scene Scene) rgbRepresentation
}

type Lambert struct {
	kd Vec3f
}

func (l Lambert) render(rio, rdi, n Vec3f, t float32, scene Scene) rgbRepresentation {
	// res := Mul(l.kd, scene.lights[0].color) // res := l.kd
	// return rgbRepresentation{uint8(res.x), uint8(res.y), uint8(res.z)}
	omega := Add(rio, rdi.mul(t))
	omega = Add(scene.lights[0].position, Vec3f{-omega.x, -omega.y, -omega.z})
	Li := Mul(l.kd, scene.lights[0].color.mul(Dot(n, omega))).mul(1 / 3.14)
	return rgbRepresentation{uint8(Li.x * 255), uint8(Li.y * 255), uint8(Li.z * 255)}
}

type GeometricObject interface {
	isIntersectedByRay(ro, rd Vec3f) (bool, float32)
	render(rio, rdi Vec3f, t float32, scene Scene) rgbRepresentation
}

// -------------------------------
type Sphere struct {
	radius   float32
	position Vec3f
	Material Materials
}

func (s Sphere) render(rio, rdi Vec3f, t float32, scene Scene) rgbRepresentation {
	/*
	* Le calcul de la normal sur une sphère est l'inverse du rayon incident.
	* C'est pourquoi n = rd1.inverte()
	 */
	return s.Material.render(rio, rdi, rdi.inverte(), t, scene)
}
func (s Sphere) isIntersectedByRay(ro, rd Vec3f) (bool, float32) {
	L := Add(ro, Vec3f{-s.position.x, -s.position.y, -s.position.z})

	a := Dot(rd, rd)
	b := 2.0 * Dot(rd, L)
	c := Dot(L, L) - s.radius*s.radius
	delta := b*b - 4.0*a*c

	t0 := (-b - float32(math.Sqrt(float64(delta)))) / 2 * a
	t1 := (-b + float32(math.Sqrt(float64(delta)))) / 2 * a
	t := t0
	t = min(t, t1)

	if delta > 0 {
		return true, t
	}
	return false, 0.0
}

// ------------------------------
type Camera struct {
	position, up, at Vec3f
}

func (c Camera) direction() Vec3f {
	dir := Add(c.at, c.position.inverte())
	return dir.mul(float32(1) / dir.norme())
}

func generateRandomSpheres(count int, minRadius, maxRadius float32, boundingBox Vec3f) []Sphere {
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	spheres := make([]Sphere, count)

	for i := 0; i < count; i++ {
		position := Vec3f{
			x: (rand.Float32()*2 - 1) * boundingBox.x,
			y: (rand.Float32()*2 - 1) * boundingBox.y,
			z: (rand.Float32()*2-1)*boundingBox.z + 5,
		}

		radius := minRadius + rand.Float32()*(maxRadius-minRadius)

		material := Lambert{
			kd: Vec3f{
				x: rand.Float32(),
				y: rand.Float32(),
				z: rand.Float32(),
			},
		}

		spheres[i] = Sphere{radius, position, material}
	}

	return spheres
}

// ------------------------------

func renderPixel(scene Scene, ro, rd Vec3f) rgbRepresentation {
	var tmin float32
	tmin = 9999999999.0
	res := rgbRepresentation{}
	for _, object := range scene.objects {
		isIntersected, t := object.isIntersectedByRay(ro, rd)
		if isIntersected && t < tmin {
			tmin = t
			res = object.render(ro, rd, t, scene)
		}
	}
	return res
}

func renderFrame(image Image, camera Camera, scene Scene) {
	ro := camera.position
	cosFovy := float32(0.66)

	aspect := float32(image.width) / float32(image.height)
	horizontal := (cross(camera.direction(), camera.up)).normalized().mul(cosFovy * aspect)
	vertical := (cross(horizontal, camera.direction())).normalized().mul(cosFovy)

	for x := 0; x < image.width; x++ {
		for y := 0; y < image.height; y++ {

			uvx := (float32(x) + float32(0.5)) / float32(image.width)
			uvy := (float32(y) + float32(0.5)) / float32(image.height)

			rd := Add(Add(camera.direction(), horizontal.mul(uvx-float32(0.5))), vertical.mul(uvy-float32(0.5))).normalized()

			image.frameBuffer[y*image.width+x] = renderPixel(scene, ro, rd)
		}
	}

}

func populateScene(scene *Scene) {
	scene.addElement(Sphere{1, Vec3f{0, 0, 8}, Lambert{Vec3f{1.0, 0, 0}}})

	randomSpheres := generateRandomSpheresWithMixedMaterials(40, 0.2, 0.7, Vec3f{5, 5, 10})
	for _, sphere := range randomSpheres {
		scene.addElement(sphere)
	}

	scene.addLight(Light{Vec3f{1.0, 1.0, 1.0}, Vec3f{0, 10, 0}})
}

func main() {

	// width := 4096
	// height := 4096
	// //Créer un objet Scène
	// scene := Scene{}

	// //Initialiser la scène
	// populateScene(&scene)
	// //Créer une caméra
	// camera := Camera{Vec3f{0, 0, -5}, Vec3f{0, 1, 0}, Vec3f{0, 0, 5}}

	// image := Image{make([]rgbRepresentation, width*height), width, height}
	// //fonction de rendu
	// renderFrame(image, camera, scene)
	// //Sauvegarde de l'image
	// image.save("./result.png")

	serverMain()
	clientMain()

}
