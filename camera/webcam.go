package camera

import (
	"gocv.io/x/gocv"
)

func setupWebcam(width, height, framerate int) *gocv.VideoCapture {
	webcam, _ := gocv.OpenVideoCapture(0)
	//window := gocv.NewWindow("Hello")

	webcam.Set(gocv.VideoCaptureFrameWidth, float64(width))
	webcam.Set(gocv.VideoCaptureFrameHeight, float64(height))
	webcam.Set(gocv.VideoCaptureFPS, float64(framerate))
	return webcam
}
