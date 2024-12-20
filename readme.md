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
