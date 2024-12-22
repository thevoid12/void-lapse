package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"time"
)

type Stream struct {
	frames chan []byte
}

func newStream() *Stream {
	return &Stream{
		frames: make(chan []byte, 30), // Buffer up to 30 frames
	}
}

func (s *Stream) startCapture() {
	// Using ffmpeg with v4l2 input for USB webcam
	// Typically your webcam will be at /dev/video0
	ffmpeg := exec.Command("ffmpeg",
		"-f", "v4l2", // Use V4L2 input
		"-framerate", "24", // Input framerate
		"-video_size", "640x480", // Input size
		"-i", "/dev/video0", // Input device
		"-f", "mjpeg", // Output format
		"-q:v", "5", // Quality (1-31, 1 is highest)
		"-", // Output to stdout
	)

	ffmpegOutput, err := ffmpeg.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	err = ffmpeg.Start()
	if err != nil {
		log.Fatal(err)
	}

	// Read MJPEG frames
	buffer := make([]byte, 1024*1024) // 1MB buffer
	for {
		n, err := ffmpegOutput.Read(buffer)
		if err != nil {
			log.Printf("Error reading frame: %v", err)
			continue
		}

		// Copy frame to prevent buffer reuse issues
		frame := make([]byte, n)
		copy(frame, buffer[:n])

		// Try to send frame, drop if channel is full
		select {
		case s.frames <- frame:
		default:
			// Drop frame if channel is full
		}
	}
}

func (s *Stream) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		// Serve the HTML page
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `
<!DOCTYPE html>
<html>
<body style="margin:0;padding:0;display:flex;justify-content:center;align-items:center;min-height:100vh;background:#000;">
    <img src="/stream" style="max-width:100%;max-height:100vh;object-fit:contain;">
</body>
</html>`)
		return
	}

	if r.URL.Path == "/stream" {
		// Set up MJPEG stream
		w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=frame")
		for frame := range s.frames {
			_, err := w.Write([]byte("--frame\r\nContent-Type: image/jpeg\r\n\r\n"))
			if err != nil {
				return
			}
			_, err = w.Write(frame)
			if err != nil {
				return
			}
			_, err = w.Write([]byte("\r\n"))
			if err != nil {
				return
			}
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
		return
	}

	http.NotFound(w, r)
}

func main() {
	stream := newStream()
	go stream.startCapture()

	// Give some time for the camera to initialize
	time.Sleep(2 * time.Second)

	fmt.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", stream))
}
