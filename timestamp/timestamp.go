package timestamp

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fogleman/gg"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"golang.org/x/image/font/basicfont"
)

// AddTimestamp module adds timestamp to the video provided picking each frame's metadata and creating file
// inputFile: the origial video file
// Output file: the file name(include path) for the output time stamp
// textColor : white or black color timestamp
// timestampType: I support 3 types of timestamp type datetime,date (date only),time (time only)
func AddTimeStamp(inputFile, outputFile, textColor, timestampType string) {
	// Get video creation time
	videoStartTime, err := getVideoCreationTime(inputFile)
	if err != nil {
		log.Printf("Warning: Could not get video creation time: %v", err)
		videoStartTime = time.Now() // Fallback to current time
	}

	// Create temporary directory for frames
	tmpDir, err := os.MkdirTemp("", "video-frames-")
	if err != nil {
		log.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Extract frames with PTS metadata
	err = ffmpeg.Input(inputFile).
		Output(filepath.Join(tmpDir, "frame-%d.png"),
			ffmpeg.KwArgs{
				"vf":           "fps=30,showinfo", // showinfo filter will print PTS info
				"frame_pts":    "1",
				"start_number": "0",
			}).
		OverWriteOutput().
		Run()
	if err != nil {
		log.Fatalf("Failed to extract frames: %v", err)
	}

	// Process frames
	files, err := filepath.Glob(filepath.Join(tmpDir, "frame-*.png"))
	if err != nil {
		log.Fatalf("Failed to list frames: %v", err)
	}

	textCol := getColor(textColor)

	// Get frame rate from video
	probe, err := ffmpeg.Probe(inputFile)
	if err != nil {
		log.Fatalf("Failed to probe video: %v", err)
	}

	// Extract frame rate from probe data
	var frameRate float64 = 30 // default fallback
	if strings.Contains(probe, "r_frame_rate") {
		lines := strings.Split(probe, "\n")
		for _, line := range lines {
			if strings.Contains(line, "r_frame_rate") {
				rateStr := strings.Split(line, ": ")[1]
				rateStr = strings.Trim(rateStr, "\"")
				nums := strings.Split(rateStr, "/")
				if len(nums) == 2 {
					num, _ := strconv.ParseFloat(nums[0], 64)
					den, _ := strconv.ParseFloat(nums[1], 64)
					if den != 0 {
						frameRate = num / den
					}
				}
			}
		}
	}

	for i, file := range files {
		// Read frame
		dc, err := gg.LoadImage(file)
		if err != nil {
			log.Fatalf("Failed to load frame %s: %v", file, err)
		}

		// Calculate frame time based on frame number and frame rate
		frameTime := videoStartTime.Add(time.Duration(float64(i) * float64(time.Second) / frameRate))
		timestamp := getTimestampFormat(frameTime, timestampType)

		// Process frame
		processedFrame, err := processFrame(dc, textCol, timestamp)
		if err != nil {
			log.Fatalf("Failed to process frame %s: %v", file, err)
		}

		// Save processed frame
		outFile := filepath.Join(tmpDir, fmt.Sprintf("processed-%d.png", i))
		err = gg.SavePNG(outFile, processedFrame)
		if err != nil {
			log.Fatalf("Failed to save processed frame: %v", err)
		}

		// Show progress
		if i%30 == 0 {
			fmt.Printf("Processed %d frames...\n", i)
		}
	}

	// Combine frames into video
	err = ffmpeg.Input(filepath.Join(tmpDir, "processed-%d.png"),
		ffmpeg.KwArgs{
			"framerate":    strconv.FormatFloat(frameRate, 'f', -1, 64),
			"start_number": "0",
		}).
		Output(outputFile, ffmpeg.KwArgs{
			"c:v":     "libx264",
			"pix_fmt": "yuv420p",
			"preset":  "medium",
			"crf":     "23",
		}).
		OverWriteOutput().
		Run()
	if err != nil {
		log.Fatalf("Failed to create output video: %v", err)
	}
}

func getColor(colorName string) color.Color {
	switch colorName {
	case "white":
		return color.White
	case "black":
		return color.Black
	default:
		return color.White
	}
}

func getTimestampFormat(t time.Time, format string) string {
	switch format {
	case "datetime":
		return t.Format("2006-01-02 15:04:05")
	case "date":
		return t.Format("2006-01-02")
	case "time":
		return t.Format("15:04:05")
	default:
		return t.Format("2006-01-02 15:04:05")
	}
}

func getVideoCreationTime(filename string) (time.Time, error) {
	probe, err := ffmpeg.Probe(filename)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to probe video: %v", err)
	}

	// Try to get creation_time from metadata
	if strings.Contains(probe, "creation_time") {
		lines := strings.Split(probe, "\n")
		for _, line := range lines {
			if strings.Contains(line, "creation_time") {
				timeStr := strings.Split(line, ": ")[1]
				timeStr = strings.Trim(timeStr, "\"")
				// Parse the time string
				t, err := time.Parse("2006-01-02T15:04:05.000000Z", timeStr)
				if err == nil {
					return t, nil
				}
			}
		}
	}

	// If creation_time not found, use file modification time
	fileInfo, err := os.Stat(filename)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get file info: %v", err)
	}
	return fileInfo.ModTime(), nil
}

func processFrame(frame image.Image, textColor color.Color, timestamp string) (image.Image, error) {
	width := frame.Bounds().Dx()
	height := frame.Bounds().Dy()

	dc := gg.NewContext(width, height)
	dc.DrawImage(frame, 0, 0)

	// Configure text drawing
	dc.SetColor(textColor)
	fontSize := float64(height) / 30

	// Try to load custom font, fall back to basic font if it fails
	err := dc.LoadFontFace("/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf", fontSize)
	if err != nil {
		dc.SetFontFace(basicfont.Face7x13)
		fontSize = float64(height) / 40
	}

	textY := float64(height) - fontSize
	textX := fontSize

	// Draw outline
	outlineColor := color.Black
	if textColor == color.Black {
		outlineColor = color.White
	}

	for dx := -2.0; dx <= 2.0; dx++ {
		for dy := -2.0; dy <= 2.0; dy++ {
			if dx == 0 && dy == 0 {
				continue
			}
			dc.SetColor(outlineColor)
			dc.DrawString(timestamp, textX+dx, textY+dy)
		}
	}

	dc.SetColor(textColor)
	dc.DrawString(timestamp, textX, textY)

	return dc.Image(), nil
}
