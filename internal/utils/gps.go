package utils

import (
	"math"
)

const earthRadius = 6371000 // Earth radius in meters

// CalculateDistance calculates distance between two GPS coordinates using Haversine formula
// Returns distance in meters
func CalculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	dLat := toRadians(lat2 - lat1)
	dLon := toRadians(lon2 - lon1)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRadians(lat1))*math.Cos(toRadians(lat2))*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	distance := earthRadius * c
	return distance
}

// ValidateLocation checks if user is within the allowed radius
func ValidateLocation(userLat, userLon, locationLat, locationLon, radius float64) (bool, float64) {
	distance := CalculateDistance(userLat, userLon, locationLat, locationLon)
	return distance <= radius, distance
}

// toRadians converts degrees to radians
func toRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

// GetNearbyLocations returns locations within specified radius (in kilometers)
func IsWithinRadius(userLat, userLon, locationLat, locationLon, radiusKm float64) bool {
	distance := CalculateDistance(userLat, userLon, locationLat, locationLon)
	return distance <= (radiusKm * 1000) // Convert km to meters
}
