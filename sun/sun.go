package sun

import (
	"log"
	"time"

	"github.com/nathan-osman/go-sunrise"
)

const INVALIDE_VALUE = -200.0

var latitude = INVALIDE_VALUE
var longitude = INVALIDE_VALUE

func InitLocation(newLatitude, newLongitude float64) {
	latitude = newLatitude
	longitude = newLongitude
}

func GetDayInfo() (sunriseTime, sunsetTime time.Time) {
	if latitude == INVALIDE_VALUE || longitude == INVALIDE_VALUE {
		log.Println("InitLocation must be called before.")
		return time.Now(), time.Now()
	}
	now := time.Now()
	sunriseTime, sunsetTime = sunrise.SunriseSunset(latitude, longitude, now.Year(), now.Month(), now.Day())
	// From UTC to local time
	sunriseTime = sunriseTime.In(now.Location())
	sunsetTime = sunsetTime.In(now.Location())
	log.Println("Sunrise:", sunriseTime.Format("15:04:05"))
	log.Println("Sunset:", sunsetTime.Format("15:04:05"))
	return sunriseTime, sunsetTime
}
