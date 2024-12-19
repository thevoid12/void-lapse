#!/usr/bin/env python3
import argparse
import subprocess
import time
import os

def capture_screenshots(duration_hours, interval_seconds, output_path):
    # Convert hours to seconds
    duration_seconds = duration_hours * 3600
    
    # Calculate expected number of images
    expected_images = int(duration_seconds // interval_seconds)
    
    # Create output directory if it doesn't exist
    os.makedirs(output_path, exist_ok=True)
    
    print(f"Starting capture for {duration_hours} hours")
    print(f"Taking screenshots every {interval_seconds} seconds")
    print(f"Expected number of images: {expected_images}")
    
    # Construct the gstreamer command
    command = [
        'gst-launch-1.0',
        'v4l2src',
        '!',
        'videorate',
        '!',
        f'video/x-raw,framerate=1/{interval_seconds}',
        '!',
        'jpegenc',
        '!',
        'multifilesink',
        f'location={output_path}/image_%05d.jpg'
    ]
    
    try:
        # Start the gstreamer process
        process = subprocess.Popen(command)
        
        # Calculate end time
        start_time = time.time()
        end_time = start_time + duration_seconds
        
        # Wait and monitor progress
        while time.time() < end_time:
            # Count current images
            current_images = len([f for f in os.listdir(output_path) 
                                if f.startswith('image_') and f.endswith('.jpg')])
            
            print(f"\rProgress: {current_images}/{expected_images} images captured", 
                  end='', flush=True)
            
            if current_images >= expected_images:
                break
                
            # Sleep for a shorter interval to be more responsive
            time.sleep(min(1, interval_seconds))
        
        # Terminate the process
        process.terminate()
        process.wait()
        
        final_images = len([f for f in os.listdir(output_path) 
                           if f.startswith('image_') and f.endswith('.jpg')])
        
        print(f"\nCapture completed. {final_images} images saved to: {output_path}")
        
    except subprocess.CalledProcessError as e:
        print(f"Error running gstreamer command: {e}")
    except KeyboardInterrupt:
        print("\nCapture interrupted by user")
        process.terminate()
        process.wait()
    except Exception as e:
        print(f"An error occurred: {e}")

def main():
    # Parse command line arguments
    parser = argparse.ArgumentParser(description='Capture screenshots using gstreamer')
    parser.add_argument('-d', '--duration', type=float, required=True,
                        help='Duration in hours to capture screenshots')
    parser.add_argument('-i', '--interval', type=int, required=True,
                        help='Interval in seconds between screenshots')
    parser.add_argument('-o', '--output', type=str, required=True,
                        help='Output directory path for screenshots')
    
    args = parser.parse_args()
    
    # Call the capture function with provided arguments
    capture_screenshots(args.duration, args.interval, args.output)

if __name__ == "__main__":
    main()
