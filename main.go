// main.go
package main

import (
	"flag"
	"log"

	"voidlapse/add_clock"
	"voidlapse/build_lapse"
	"voidlapse/shoot_lapse"
)

func main() {
	var (
		inputDir        string
		outputDir       string
		inputFile       string
		outputFile      string
		textColor       string
		timestampFormat string
		mode            string
		durationHours   float64
		intervalSeconds int
		shootoutputPath string
		wantTimestamp   string
	)

	flag.Float64Var(&durationHours, "d", 0.0, "SL:Duration in hours to capture screenshots(required)")
	flag.IntVar(&intervalSeconds, "i", 0, "SL:Interval in seconds between screenshots(required)")
	flag.StringVar(&shootoutputPath, "o", "", "SL:Output directory path for screenshots to be saved(required)")
	flag.StringVar(&inputDir, "ip", "", "BL:Location of folder where this image files are located(requried)")
	flag.StringVar(&outputDir, "op", "", "BL:Location of folder where this image files will be stored (optional)")
	flag.StringVar(&wantTimestamp, "t", "n", "BL:Do you want timestamp(optional) type y")

	flag.StringVar(&inputFile, "it", "", "AC:input shot image files")
	flag.StringVar(&outputFile, "ot", "", "AC:output video file's folder")
	flag.StringVar(&textColor, "c", "white", "AC:text color (white or black)")
	flag.StringVar(&timestampFormat, "f", "datetime", "AC:timestamp format (datetime, date, time)")
	flag.StringVar(&mode, "m", "build", "mode (build or shoot or just add clock)")
	flag.Parse()

	switch mode {
	case "build":
		build_lapse.BuildLapse(inputDir, outputDir, textColor, timestampFormat, wantTimestamp)
	case "shoot":
		shoot_lapse.ShootLapse(durationHours, intervalSeconds, shootoutputPath)
	case "timestamp":
		add_clock.AddClock(inputFile, outputFile, textColor, timestampFormat)
	default:
		log.Fatalf("Invalid mode specified use -m build or shoot or timestamp")
	}
}
