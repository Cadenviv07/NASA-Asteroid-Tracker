package physics

import (
	"math"
	"time"
)

type Vector3 struct {
	X, Y, Z float64
}

type OrbitalElements struct {
	SemiMajorAxis float64
	Eccentricity  float64
	Inclination   float64
	AscendingNode float64
	Perihelion    float64
	MeanAnomaly   float64
}

func TimetoJulian(t time.Time) float64 {
	t = t.UTC()

	year := float64(t.Year())
	month := float64(t.Month())
	day := float64(t.Day())

	//Astronomy treats january and febuary as the 13 and 14 month of the year
	if month <= 2 {
		year -= 1
		month += 12
	}

	A := math.Floor(year / 100)
	B := 2 - A + math.Floor(A/4)

	jd := math.Floor(365.25*(year+4716)) + math.Floor(30.6001*(month+1)) + day + B - 1524.5

	fraction := float64(t.Hour())/24.0 + float64(t.Minute())/1440.0 + float64(t.Second())/86400.0

	return jd + fraction
}

// Caculates mean anomoly which tells you what precentage of the orbit you have passed through
func PositionInOrbit(currentT, meanMotion float64, meanAnomoly float64, epochDate float64) float64 {
	currentTime := currentT
	deltaT := currentTime - epochDate
	M := meanMotion*deltaT + meanAnomoly

	M = math.Mod(M, 360)

	if M < 0 {
		M += 360
	}

	return M
}

func CalculateEccentricAnomaly(M_deg float64, e float64) float64 {
	rad := math.Pi / 180.0
	M := M_deg * rad
	E := M

	for {
		f := E - e*math.Sin(E) - M

		slope := 1.0 - e*math.Cos(E)

		delta := f / slope

		E = E - delta

		if math.Abs(delta) < 0.000001 {
			break
		}
	}

	return E
}

func GetPlaneCoordinates(E float64, e float64, a float64) (float64, float64) {
	//Calculate the coordinate of the eclipse shifted so the sun is at the center of the eclipse
	x := a * (math.Cos(E) - e)

	factor := math.Sqrt(1 - (e * e))
	y := a * factor * math.Sin(E)

	return x, y
}

func RotatePlane(x float64, y float64, i float64, omega float64, w float64) Vector3 {

	rad := math.Pi / 180.0
	inclination := i * rad // Inclinatino of plane
	Node := omega * rad    // longitudnal argument of plane
	PA := w * rad          // Perihelion argument of plane

	//Multiply current plane by three different rotaiton matrixes

	cosO := math.Cos(Node)
	sinO := math.Sin(Node)
	cosw := math.Cos(PA)
	sinw := math.Sin(PA)
	cosi := math.Cos(inclination)
	sini := math.Sin(inclination)

	xf := (cosO*cosw-sinO*sinw*cosi)*x + (-cosO*sinw-sinO*cosw*cosi)*y

	yf := (sinO*cosw+cosO*sinw*cosi)*x + (-sinO*sinw+cosO*cosw*cosi)*y

	z := (sinw*sini)*x + (cosw*sini)*y

	return Vector3{X: xf, Y: yf, Z: z}
}

func GetEarthsPosition(currentJD float64) Vector3 {

	earth := OrbitalElements{
		SemiMajorAxis: 1.00000011,
		Eccentricity:  0.01671022,
		Inclination:   0.00005,
		AscendingNode: -11.26064,
		Perihelion:    102.94719,
		MeanAnomaly:   100.46435,
	}

	const EarthEpoch = 2451545.0
	const EarthMeanMotion = 0.98560766

	timePassed := currentJD - EarthEpoch

	M := earth.MeanAnomaly + (EarthMeanMotion * timePassed)
	M = math.Mod(M, 360.0)
	M = M * math.Pi / 180.0

	E := CalculateEccentricAnomaly(M, earth.Eccentricity)

	xPlane, yPlane := GetPlaneCoordinates(E, earth.Eccentricity, earth.SemiMajorAxis)

	return RotatePlane(xPlane, yPlane, earth.Inclination, earth.AscendingNode, earth.Perihelion)

}
