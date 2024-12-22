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
		frames: make(chan []byte, 30),
	}
}

func (s *Stream) startCapture() {
	log.Println("Starting camera capture...")

	// Simplified ffmpeg command without assuming input format
	ffmpeg := exec.Command("ffmpeg",
		"-f", "v4l2",
		"-i", "/dev/video0",
		"-f", "mjpeg",
		"-frames:v", "0", // Unlimited frames
		"-r", "24", // Output framerate
		"-q:v", "8", // Higher quality value (1-31, lower is better)
		"-",
	)

	log.Printf("Executing command: %v", ffmpeg.String())

	ffmpegOutput, err := ffmpeg.StdoutPipe()
	if err != nil {
		log.Fatalf("Failed to create stdout pipe: %v", err)
	}

	ffmpeg.Stderr = log.Writer()

	err = ffmpeg.Start()
	if err != nil {
		log.Fatalf("Failed to start ffmpeg: %v", err)
	}

	log.Println("FFmpeg started successfully")

	buffer := make([]byte, 1024*1024)
	frameCount := 0
	startTime := time.Now()

	for {
		n, err := ffmpegOutput.Read(buffer)
		if err != nil {
			log.Printf("Error reading frame: %v", err)
			continue
		}

		frameCount++
		if frameCount%100 == 0 {
			elapsed := time.Since(startTime).Seconds()
			fps := float64(frameCount) / elapsed
			log.Printf("Capturing frames at %.2f fps", fps)
		}

		frame := make([]byte, n)
		copy(frame, buffer[:n])

		select {
		case s.frames <- frame:
		default:
			log.Println("Frame dropped - buffer full")
		}
	}
}

func (s *Stream) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request for: %s", r.URL.Path)

	if r.URL.Path == "/" {
		log.Println("Serving HTML page")
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `
<!DOCTYPE html>
<html>
<head>
    <title>Webcam Stream</title>
</head>
<body style="margin:0;padding:0;display:flex;justify-content:center;align-items:center;min-height:100vh;background:#000;">
    <img src="/stream" style="max-width:100%;max-height:100vh;object-fit:contain;">
</body>
</html>`)
		return
	}

	if r.URL.Path == "/stream" {
		log.Println("Starting stream response")
		w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=frame")

		framesSent := 0
		startTime := time.Now()

		for frame := range s.frames {
			_, err := w.Write([]byte("--frame\r\nContent-Type: image/jpeg\r\n\r\n"))
			if err != nil {
				log.Printf("Error writing frame boundary: %v", err)
				return
			}

			_, err = w.Write(frame)
			if err != nil {
				log.Printf("Error writing frame data: %v", err)
				return
			}

			_, err = w.Write([]byte("\r\n"))
			if err != nil {
				log.Printf("Error writing frame ending: %v", err)
				return
			}

			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}

			framesSent++
			if framesSent%100 == 0 {
				elapsed := time.Since(startTime).Seconds()
				fps := float64(framesSent) / elapsed
				log.Printf("Streaming to client at %.2f fps", fps)
			}
		}
		log.Println("Stream ended")
		return
	}

	http.NotFound(w, r)
}

func main() {
	log.Println("Initializing stream...")
	stream := newStream()

	log.Println("Starting capture routine...")
	go stream.startCapture()

	time.Sleep(2 * time.Second)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", stream))
}
