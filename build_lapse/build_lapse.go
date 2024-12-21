package build_lapse

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"voidlapse/timestamp"

	"time"
)

func createTimelapse(outputPath string, inputPath string) string {

	err := os.MkdirAll(outputPath, 0755)
	if err != nil {
		log.Printf("Error creating output directory: %v\n", err)
	}

	layout := "02-01-2006-15:04"
	videoName := "timelapse_" + time.Now().Format(layout) + ".mp4"

	fmt.Println("*********************************************************************************")
	fmt.Println(`
	▗▖  ▗▖ ▗▄▖ ▗▄▄▄▖▗▄▄▄     ▗▖    ▗▄▖ ▗▄▄▖  ▗▄▄▖▗▄▄▄▖    
	▐▌  ▐▌▐▌ ▐▌  █  ▐▌  █    ▐▌   ▐▌ ▐▌▐▌ ▐▌▐▌   ▐▌       
	▐▌  ▐▌▐▌ ▐▌  █  ▐▌  █    ▐▌   ▐▛▀▜▌▐▛▀▘  ▝▀▚▖▐▛▀▀▘    
	 ▝▚▞▘ ▝▚▄▞▘▗▄█▄▖▐▙▄▄▀    ▐▙▄▄▖▐▌ ▐▌▐▌   ▗▄▄▞▘▐▙▄▄▖    
																											 
	`)

	fmt.Printf("Starting to process timelapse creation:%v\n", time.Now())
	// Construct the gstreamer command

	command := exec.Command(
		"ffmpeg",
		"-framerate", "30",
		"-i", filepath.Join(inputPath, "image_%05d.jpg"),
		"-c:v", "libx264",
		"-pix_fmt", "yuv420p",
		"-preset", "ultrafast",
		filepath.Join(outputPath, videoName),
	)
	err = command.Start()
	if err != nil {
		log.Printf("Error starting command: %v\n", err)

	}

	startTime := time.Now()
	command.Wait()
	endTime := time.Since(startTime)
	fmt.Printf("Timelapse created successfully with videoname: %s. elapsed Time: %s\n", videoName, endTime)
	fmt.Println("*********************************************************************************")

	return videoName
}

func BuildLapse(inputPath, outputPath, textColor, timestampFormat, wantTimestamp string) {

	// Check if required flags are provided
	if inputPath == "" {
		fmt.Println("Input path is required:")
		flag.PrintDefaults()
		os.Exit(1)
	}
	if outputPath == "" {
		outputPath = "../timelapse_photos"
	}
	// Convert output path to absolute path
	absOutputPath, err := filepath.Abs(outputPath)
	if err != nil {
		fmt.Printf("Error getting absolute output path: %v\n", err)
		os.Exit(1)
	}

	absInputPath, err := filepath.Abs(inputPath)
	if err != nil {
		fmt.Printf("Error getting absolute input path: %v\n", err)
		os.Exit(1)
	}
	_, err = os.Stat(absInputPath)
	if err != nil {
		// Check if the error is because the file does not exist
		if errors.Is(err, os.ErrNotExist) {
			fmt.Printf("file not found: %s\n", absInputPath)
			os.Exit(1)
		}
		// Return other possible errors
		fmt.Println("error checking file: %w", err)
		os.Exit(1)
	}

	videoName := createTimelapse(absOutputPath, absInputPath)
	if wantTimestamp == "y" || wantTimestamp == "Y" {
		fmt.Println()
		fmt.Println("*********************************************************************************")
		st := time.Now()

		fmt.Println("Adding timestamp to your video startTime: %s", st)
		timestamp.AddTimeStamp(absInputPath, filepath.Join(absOutputPath, "ts_"+videoName), textColor, timestampFormat)
		fmt.Println("Video processing timestamp completed successfully. Time Elapsed:%s", time.Since(st))
		fmt.Println("*********************************************************************************")
	}
}
