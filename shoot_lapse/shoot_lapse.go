package shoot_lapse

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func captureScreenshots(durationHours float64, intervalSeconds int, outputPath string) {
	// Convert hours to seconds
	durationSeconds := int(durationHours * 3600)

	// Calculate expected number of images
	expectedImages := durationSeconds / intervalSeconds

	// Create output directory if it doesn't exist
	err := os.MkdirAll(outputPath, 0755)
	if err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		return
	}

	fmt.Printf("Starting capture for %f hours\n", durationHours)
	fmt.Printf("Taking screenshots every %d seconds\n", intervalSeconds)
	fmt.Printf("Expected number of images: %d\n", expectedImages)

	// Construct the gstreamer command
	command := exec.Command("gst-launch-1.0",
		"v4l2src",
		"!",
		"videorate",
		"!",
		fmt.Sprintf("video/x-raw,framerate=1/%d", intervalSeconds),
		"!",
		"jpegenc",
		"!",
		"multifilesink",
		fmt.Sprintf("location=%s/image_%%05d.jpg", outputPath))

	// Start the gstreamer process
	// command.Start() runs the external process in a separate thread of execution.
	// The method does not block; it immediately returns control to the program
	// while the command continues to execute in the background.
	err = command.Start()
	if err != nil {
		fmt.Printf("Error starting command: %v\n", err)
		return
	}

	// Calculate end time
	startTime := time.Now()
	endTime := startTime.Add(time.Duration(durationSeconds) * time.Second)

	// Wait and monitor progress, loops through forever until the time is reached
	for time.Now().Before(endTime) {
		// Count current images
		files, err := os.ReadDir(outputPath)
		if err != nil {
			fmt.Printf("Error reading directory: %v\n", err)
			continue
		}

		currentImages := 0
		for _, file := range files {
			if strings.HasPrefix(file.Name(), "image_") && strings.HasSuffix(file.Name(), ".jpg") {
				currentImages++
			}
		}

		fmt.Printf("\rProgress: %d/%d images captured", currentImages, expectedImages)

		if currentImages >= expectedImages {
			break
		}

		// Sleep for a shorter interval to be more responsive
		sleepDuration := time.Second
		if time.Duration(intervalSeconds)*time.Second < sleepDuration {
			sleepDuration = time.Duration(intervalSeconds) * time.Second
		}
		time.Sleep(sleepDuration)
	}

	// Terminate the process if we press ctrl c
	if err := command.Process.Signal(os.Interrupt); err != nil {
		fmt.Printf("\nError interrupting process: %v\n", err)
	}

	// wait for the process to finish. if we are not waiting then the program will end from the main thread thus
	// it will exit
	command.Wait()

	// Count final images
	files, err := os.ReadDir(outputPath)
	if err != nil {
		fmt.Printf("\nError reading directory: %v\n", err)
		return
	}

	finalImages := 0
	for _, file := range files {
		if strings.HasPrefix(file.Name(), "image_") && strings.HasSuffix(file.Name(), ".jpg") {
			finalImages++
		}
	}

	fmt.Printf("\nCapture completed. %d images saved to: %s\n", finalImages, outputPath)
}

func ShootLapse(durationHours float64, intervalSeconds int, outputPath string) {

	// Check if required flags are provided
	if durationHours == 0.0 || intervalSeconds == 0 || outputPath == "" {
		fmt.Println("All flags are required:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Convert output path to absolute path
	absPath, err := filepath.Abs(outputPath)
	if err != nil {
		fmt.Printf("Error getting absolute path: %v\n", err)
		os.Exit(1)
	}

	captureScreenshots(durationHours, intervalSeconds, absPath)
}
