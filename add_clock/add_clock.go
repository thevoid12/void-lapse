package add_clock

import (
	"flag"
	"fmt"
	"os"
	"voidlapse/timestamp"
)

func AddClock(inputFile, outputFile, textColor, timestampType string) {
	if inputFile == "" || outputFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	timestamp.AddTimeStamp(inputFile, outputFile, textColor, timestampType)
	fmt.Println("Video processing timestamp completed successfully!")
}
