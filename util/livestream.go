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
		frames: make(chan []byte, 24), // Buffer for 1 second at 24fps
	}
}

func (s *Stream) startCapture() {
	log.Println("Starting camera capture from /dev/video0...")

	for {
		ffmpeg := exec.Command("ffmpeg",
			"-f", "v4l2",
			"-video_size", "1280x720", // HD resolution
			"-i", "/dev/video0", // Explicitly use video0
			"-f", "mjpeg",
			"-frames:v", "0", // Continuous stream
			"-r", "24", // 24 fps
			"-q:v", "2", // High quality (1-31, lower is better)
			"-update", "1", // Enable update mode
			"-", // Output to pipe
		)

		log.Printf("Executing command: %v", ffmpeg.String())

		ffmpegOutput, err := ffmpeg.StdoutPipe()
		if err != nil {
			log.Printf("Failed to create stdout pipe: %v", err)
			time.Sleep(time.Second)
			continue
		}

		ffmpeg.Stderr = log.Writer()

		err = ffmpeg.Start()
		if err != nil {
			log.Printf("Failed to start ffmpeg: %v", err)
			time.Sleep(time.Second)
			continue
		}

		log.Println("FFmpeg started successfully")

		buffer := make([]byte, 2*1024*1024) // 2MB buffer for HD frames
		frameCount := 0
		startTime := time.Now()

		for {
			n, err := ffmpegOutput.Read(buffer)
			if err != nil {
				log.Printf("Error reading frame: %v", err)
				break
			}

			frameCount++
			if frameCount%24 == 0 { // Log every second (24 frames)
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

		ffmpeg.Process.Kill()
		ffmpeg.Wait()
		log.Println("FFmpeg process ended, restarting in 1 second...")
		time.Sleep(time.Second)
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
    <title>24FPS Webcam Stream</title>
    <style>
        body {
            margin: 0;
            padding: 0;
            background: #000;
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
            overflow: hidden;
        }
        .stream-container {
            max-width: 1280px;
            max-height: 720px;
            width: 100%;
            height: 100vh;
            display: flex;
            justify-content: center;
            align-items: center;
        }
        .stream-container img {
            max-width: 100%;
            max-height: 100%;
            object-fit: contain;
        }
    </style>
</head>
<body>
    <div class="stream-container">
        <img src="/stream" alt="Live Stream">
    </div>
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
			if framesSent%24 == 0 { // Log every second (24 frames)
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
	// Check if video0 exists
	if _, err := exec.Command("ls", "/dev/video0").Output(); err != nil {
		log.Fatal("No video device found at /dev/video0")
	}

	log.Println("Initializing stream...")
	stream := newStream()

	log.Println("Starting capture routine...")
	go stream.startCapture()

	time.Sleep(2 * time.Second)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", stream))
}
