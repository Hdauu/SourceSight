package core

import (
	"math"
	"sort"
)

type Vec3 struct {
	X float64
	Y float64
	Z float64
}

func ToRad(deg float64) float64 {
	return deg * (math.Pi / 180.0)
}

func ToDeg(rad float64) float64 {
	return rad * (180.0 / math.Pi)
}

func Dist3D(a, b Vec3) float64 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	dz := a.Z - b.Z
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}

func Dist2D(a, b Vec3) float64 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	return math.Sqrt(dx*dx + dy*dy)
}

func Length(v Vec3) float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
}

func Normalize(v Vec3) Vec3 {
	len := math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
	if len < 1e-9 {
		return Vec3{}
	}
	return Vec3{X: v.X / len, Y: v.Y / len, Z: v.Z / len}
}

func Dot(a, b Vec3) float64 {
	return a.X*b.X + a.Y*b.Y + a.Z*b.Z
}

func AngleBetween(a, b Vec3) float64 {
	na := Normalize(a)
	nb := Normalize(b)
	d := Dot(na, nb)
	if d > 1.0 {
		d = 1.0
	} else if d < -1.0 {
		d = -1.0
	}
	return ToDeg(math.Acos(d))
}

// EyeAnglesForward converts Source Engine Euler angles (pitch, yaw) to unit forward direction vector.
// Convention: positive pitch = down (-Z)
func EyeAnglesForward(pitch, yaw float32) Vec3 {
	p := ToRad(float64(pitch))
	y := ToRad(float64(yaw))
	return Vec3{
		X: math.Cos(p) * math.Cos(y),
		Y: math.Cos(p) * math.Sin(y),
		Z: -math.Sin(p),
	}
}

// RayIntersectsAABB performs Kay & Kajiya slab method for ray-box intersection.
func RayIntersectsAABB(origin, dir, boxMin, boxMax Vec3) bool {
	const eps = 1e-9
	tMin := math.Inf(-1)
	tMax := math.Inf(1)

	axes := []struct {
		o, d, lo, hi float64
	}{
		{origin.X, dir.X, boxMin.X, boxMax.X},
		{origin.Y, dir.Y, boxMin.Y, boxMax.Y},
		{origin.Z, dir.Z, boxMin.Z, boxMax.Z},
	}

	for _, a := range axes {
		if math.Abs(a.d) < eps {
			if a.o < a.lo || a.o > a.hi {
				return false
			}
		} else {
			t1 := (a.lo - a.o) / a.d
			t2 := (a.hi - a.o) / a.d
			if t1 > t2 {
				t1, t2 = t2, t1
			}
			if t1 > tMin {
				tMin = t1
			}
			if t2 < tMax {
				tMax = t2
			}
			if tMin > tMax {
				return false
			}
		}
	}

	return tMax >= 0
}

// SegmentIntersectsSphere reports whether the straight line from a to b passes through
// the sphere. Only the part between the two points counts: a sphere beyond b (behind
// the target) or behind a doesn't block anything between them.
func SegmentIntersectsSphere(a, b, center Vec3, radius float64) bool {
	ab := Vec3{X: b.X - a.X, Y: b.Y - a.Y, Z: b.Z - a.Z}
	ac := Vec3{X: center.X - a.X, Y: center.Y - a.Y, Z: center.Z - a.Z}
	len2 := Dot(ab, ab)
	t := 0.0
	if len2 > 0 {
		t = math.Max(0, math.Min(1, Dot(ac, ab)/len2))
	}
	// Closest point of the segment to the centre
	p := Vec3{X: a.X + t*ab.X - center.X, Y: a.Y + t*ab.Y - center.Y, Z: a.Z + t*ab.Z - center.Z}
	return Dot(p, p) <= radius*radius
}

// NormalizeAngle normalizes angle delta to (-180, 180] degrees to prevent wrap-around false snaps.
func NormalizeAngle(angle float64) float64 {
	return math.Mod(math.Mod(angle+180.0, 360.0)+360.0, 360.0) - 180.0
}

func Mean(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	var sum float64
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}

func StdDev(vals []float64) float64 {
	if len(vals) < 2 {
		return 0
	}
	m := Mean(vals)
	var sumSq float64
	for _, v := range vals {
		diff := v - m
		sumSq += diff * diff
	}
	return math.Sqrt(sumSq / float64(len(vals)-1))
}

// Median returns the middle value (upper median for even counts) without
// modifying vals.
func Median(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	c := append([]float64(nil), vals...)
	sort.Float64s(c)
	return c[len(c)/2]
}
