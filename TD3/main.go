package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"math/rand"
	"os"
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

	width := 4096
	height := 4096
	//Créer un objet Scène
	scene := Scene{}

	//Initialiser la scène
	populateScene(&scene)
	//Créer une caméra
	camera := Camera{Vec3f{0, 0, -5}, Vec3f{0, 1, 0}, Vec3f{0, 0, 5}}

	image := Image{make([]rgbRepresentation, width*height), width, height}
	//fonction de rendu
	renderFrame(image, camera, scene)
	//Sauvegarde de l'image
	image.save("./result.png")

}
