python shoot-lapse.py -d 0.014 -i 5 -o ./timelapse_photos
ffmpeg -framerate 30 -i image_%05d.jpg -c:v libx264 -pix_fmt yuv420p -preset ultrafast timelapse1.mp4
ffmpeg -framerate 30 -i /path/to/images/image_%05d.jpg -c:v libx264 -pix_fmt yuv420p -preset ultrafast timelapse1.mp4

```go
go mod init shoot_lapse
go mod tidy
go build shoot_lapse.go
./shoot_lapse -d 0.014 -i 5 -o ./timelapse_photos
```
```bash
sudo apt update
sudo apt install ffmpeg
```
ffmpeg -f v4l2 -video_size 1280x720 -i /dev/video0 -s 30 -frames 1 out.jpg
fswebcam -r 1280x720 -p YUYV --set -D 2 -S 60 -F 10  test_image7.jpg



# Update package lists
sudo apt-get update

# Install build tools and dependencies
sudo apt-get install -y build-essential cmake pkg-config
sudo apt-get install -y libjpeg-dev libpng-dev libtiff-dev
sudo apt-get install -y libavcodec-dev libavformat-dev libswscale-dev libv4l-dev
sudo apt-get install -y libxvidcore-dev libx264-dev
sudo apt-get install -y libgtk-3-dev
sudo apt-get install -y libatlas-base-dev gfortran
sudo apt-get install -y python3-dev

# Install OpenCV 4
sudo apt-get install -y python3-opencv
sudo apt-get install -y libopencv-dev


sudo apt-get update
sudo apt-get install build-essential pkg-config
sudo apt-get install libopencv-dev
sudo apt-get install libavcodec-dev libavformat-dev libswscale-dev libv4l-dev
go get -u gocv.io/x/gocv


go run main.go -c black -i date -input input.mp4 -output output.mp4



# Install required dependencies
go get github.com/fogleman/gg
go get github.com/u2takey/ffmpeg-go

# Make sure ffmpeg is installed
sudo apt-get install ffmpeg

# Run the program
go run main.go -input input.mp4 -output output.mp4 -c white -t datetime
go run add_clock.go -input "../timelapse_photos/timelapse.mp4" -output "../timelapse_photos/output.mp4" -c white -t datetime
go get golang.org/x/image/font/basicfont
