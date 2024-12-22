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
			"-vf", "format=yuvj422p", // Fix deprecated pixel format warning
			"-f", "mjpeg", // Output format
			"-q:v", "2", // High quality (1-31, lower is better)
			"-r", "24", // 24 fps
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
				log.Printf("Error writing frame footer: %v", err)
				return
			}

			framesSent++
			if framesSent%24 == 0 {
				elapsed := time.Since(startTime).Seconds()
				fps := float64(framesSent) / elapsed
				log.Printf("Streaming frames at %.2f fps", fps)
			}
		}
	}
}

func main() {
	stream := newStream()
	go stream.startCapture()

	http.Handle("/", stream)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
