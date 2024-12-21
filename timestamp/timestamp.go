package timestamp

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/fogleman/gg"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"golang.org/x/image/font/basicfont"
)

type frameInfo struct {
	path      string
	timestamp time.Time
}

// AddTimestamp module adds timestamp to the video provided picking each frame's metadata and creating file
// inputFile: the origial video file
// Output file: the file name(include path) for the output time stamp
// textColor : white or black color timestamp
// timestampType: I support 3 types of timestamp type datetime,date (date only),time (time only)
func AddTimeStamp(inputFile, outputFile, textColor, timestampType string) {
	// Get video creation time

	// Create temporary directory for frames
	tmpDir, err := os.MkdirTemp("", "video-frames-")
	if err != nil {
		log.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Get original video info
	probe, err := ffmpeg.Probe(inputFile)
	if err != nil {
		log.Fatalf("Failed to probe video: %v", err)
	}

	// Extract original video parameters
	frameRate, _, _ := extractVideoParams(probe)

	// Extract frames preserving original frame rate and format
	err = ffmpeg.Input(inputFile).
		Output(filepath.Join(tmpDir, "frame-%08d.png"),
			ffmpeg.KwArgs{
				"vsync":        "0", // Maintain exact frame timing
				"frame_pts":    "1", // Preserve presentation timestamps
				"start_number": "0",
				"vf":           "drawtext=text='%{pts\\:hms}':x=0:y=0:fontsize=1:fontcolor=white@0.0",
			}).
		OverWriteOutput().
		Run()
	if err != nil {
		log.Fatalf("Failed to extract frames: %v", err)
	}

	// Get frame timestamps using ffprobe
	frames, err := getFrameTimestamps(inputFile)
	if err != nil {
		log.Fatalf("Failed to get frame timestamps: %v", err)
	}

	// Process frames
	files, err := filepath.Glob(filepath.Join(tmpDir, "frame-*.png"))
	if err != nil {
		log.Fatalf("Failed to list frames: %v", err)
	}

	// Sort frames numerically
	sort.Slice(files, func(i, j int) bool {
		numI := extractFrameNumber(files[i])
		numJ := extractFrameNumber(files[j])
		return numI < numJ
	})

	textCol := getColor(textColor)

	// Process each frame
	for i, file := range files {
		dc, err := gg.LoadImage(file)
		if err != nil {
			log.Fatalf("Failed to load frame %s: %v", file, err)
		}

		timestamp := getTimestampFormat(frames[i].timestamp, timestampType)

		processedFrame, err := processFrame(dc, textCol, timestamp)
		if err != nil {
			log.Fatalf("Failed to process frame %s: %v", file, err)
		}

		outFile := filepath.Join(tmpDir, fmt.Sprintf("processed-%08d.png", i))
		err = gg.SavePNG(outFile, processedFrame)
		if err != nil {
			log.Fatalf("Failed to save processed frame: %v", err)
		}

		if i%30 == 0 {
			fmt.Printf("Processed %d frames...\n", i)
		}
	}

	// Recreate video with original parameters
	// err = ffmpeg.Input(filepath.Join(tmpDir, "processed-%08d.png"),
	// 	ffmpeg.KwArgs{
	// 		"framerate": strconv.FormatFloat(frameRate, 'f', -1, 64),
	// 		"vsync":     "0",
	// 	}).
	// 	Output(outputFile, ffmpeg.KwArgs{
	// 		"c:v":     codec,  // Use original codec
	// 		"pix_fmt": pixFmt, // Use original pixel format
	// 		"vsync":   "0",    // Maintain exact frame timing
	// 		"preset":  "medium",
	// 		"crf":     "23",
	// 	}).
	// 	OverWriteOutput().
	// 	Run()
	// if err != nil {
	// 	log.Fatalf("Failed to create output video: %v", err)
	// }
	// Combine frames into video
	// err = ffmpeg.Input(filepath.Join(tmpDir, "processed-%08d.png"),
	// 	ffmpeg.KwArgs{
	// 		"framerate":    strconv.FormatFloat(frameRate, 'f', -1, 64),
	// 		"start_number": "0",
	// 	}).
	// 	Output(outputFile, ffmpeg.KwArgs{
	// 		"c:v":     codec,
	// 		"pix_fmt": pixFmt,
	// 		"preset":  "medium",
	// 		"crf":     "23",
	// 	}).
	// 	OverWriteOutput().
	// 	Run()
	// if err != nil {
	// 	log.Fatalf("Failed to create output video: %v", err)
	// }

	// Verify frames exist before running ffmpeg
	processedFiles, err := filepath.Glob(filepath.Join(tmpDir, "processed-*.png"))
	if err != nil || len(processedFiles) == 0 {
		log.Fatalf("No processed frames found in %s", tmpDir)
	}
	log.Printf("Found %d processed frames", len(processedFiles))

	command := exec.Command(
		"ffmpeg",
		"-framerate", strconv.FormatFloat(frameRate, 'f', -1, 64),
		"-i", filepath.Join(tmpDir, "processed-%08d.png"),
		"-c:v", "libx264",
		"-pix_fmt", "yuv420p",
		"-preset", "ultrafast",
		"-y", // Force overwrite
		outputFile,
	)

	// Capture command output
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	// Log the command being executed
	log.Printf("Executing command: %s", command.String())

	// Run the command and wait for completion
	err = command.Run()
	if err != nil {
		log.Printf("FFmpeg stdout: %s", stdout.String())
		log.Printf("FFmpeg stderr: %s", stderr.String())
		log.Fatalf("Failed to create video: %v", err)
	}

	command.Wait()
	// Verify the output file was created
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		log.Fatalf("Output file was not created at: %s", outputFile)
	}

}

func extractVideoParams(probe string) (float64, string, string) {
	var frameRate float64 = 30 // default fallback
	codec := "libx264"         // default fallback
	pixFmt := "yuv420p"        // default fallback

	lines := strings.Split(probe, "\n")
	for _, line := range lines {
		// Extract frame rate
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
		// Extract codec
		if strings.Contains(line, "codec_name") {
			codec = strings.Split(line, ": ")[1]
			codec = strings.Trim(codec, "\"")
		}
		// Extract pixel format
		if strings.Contains(line, "pix_fmt") {
			pixFmt = strings.Split(line, ": ")[1]
			pixFmt = strings.Trim(pixFmt, "\"")
		}
	}
	return frameRate, codec, pixFmt
}

// [Rest of the helper functions remain the same]

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

func extractFrameNumber(filename string) int {
	// Extract the number between "frame-" and ".png"
	base := filepath.Base(filename)
	numStr := strings.TrimPrefix(base, "frame-")
	numStr = strings.TrimSuffix(numStr, ".png")
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return -1
	}
	return num
}

func getFrameTimestamps(inputFile string) ([]frameInfo, error) {
	// Get video start time
	videoStartTime, err := getVideoCreationTime(inputFile)
	if err != nil {
		return nil, err
	}

	// Get frame timestamps using ffprobe
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-select_streams", "v:0",
		"-show_entries", "frame=pts_time",
		"-of", "csv=p=0",
		inputFile)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to get frame timestamps: %v", err)
	}

	// Parse timestamps
	var frames []frameInfo
	for _, line := range strings.Split(stdout.String(), "\n") {
		if line == "" {
			continue
		}
		ptsTime, err := strconv.ParseFloat(line, 64)
		if err != nil {
			continue
		}

		// Calculate actual timestamp for this frame
		frameTime := videoStartTime.Add(time.Duration(ptsTime * float64(time.Second)))
		frames = append(frames, frameInfo{
			timestamp: frameTime,
		})
	}

	return frames, nil
}
