package main

import (
	"fmt"
	"time"
)

var (
	roundshotHiedanranta = &roundshotUrlBuilder{"582431e6e60a55.13915440", "Europe/Helsinki"}
)

type roundshotSizeVariant string

const (
	roundshotImageSizeVariantFull roundshotSizeVariant = "full"
	// commented out b/c lint complains..
	// roundshotImageSizeVariantHalf roundshotSizeVariant = "half"
)

type roundshotUrlBuilder struct {
	cameraId       string
	cameraTimezone string
}

func (r *roundshotUrlBuilder) Url(ts time.Time, size roundshotSizeVariant) (string, error) {
	// these pranksters save them to storage in local time..
	cameraTimezone, err := time.LoadLocation(r.cameraTimezone)
	if err != nil {
		return "", err
	}

	tsLocally := ts.In(cameraTimezone)

	dayComponent := tsLocally.Format("2006-01-02")
	timeComponent := tsLocally.Format("15-04-05")

	// looks like 2020-06-07-13-30-00_full.jpg
	filename := fmt.Sprintf("%s-%s_%s.jpg", dayComponent, timeComponent, size)

	// looks like https://storage.roundshot.com/582431e6e60a55.13915440/2020-06-07/13-30-00/2020-06-07-13-30-00_full.jpg
	return fmt.Sprintf(
		"https://storage.roundshot.com/%s/%s/%s/%s",
		r.cameraId,
		dayComponent,
		timeComponent,
		filename), nil
}
