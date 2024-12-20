# Void Lapse  
Void Lapse is an easy-to-use, Linux-based timelapse creation tool written in Go. It simplifies the process of creating hours-long timelapses by providing an intuitive command-line interface.

![Terminal View](/img/1.png)  
![Sample Timelapse Image](/img/2.jpg)

---

## Features  
- **Three distinct modes:**
  1. **Shoot Mode (SL):** Captures images at regular intervals.
  2. **Build Lapse Mode (BL):** Converts captured images into a timelapse video.
  3. **Add Clock Mode (AC):** Adds timestamps to existing videos based on frame metadata.  
- Flexible timestamp customization options, including format and color.  
- Lightweight and designed for Linux environments.

---

## Installation  

### Prerequisites  
Ensure the following dependencies are installed:  

1. **Go Libraries:**  
   ```bash
   go get github.com/fogleman/gg
   go get github.com/u2takey/ffmpeg-go
   go get golang.org/x/image/font/basicfont
   ```
2. **System Dependencies:**  
   ```bash
   sudo apt-get install ffmpeg libavcodec-dev libavformat-dev libswscale-dev libv4l-dev
   sudo apt-get install -y v4l-utils \
       libgstreamer1.0-0 gstreamer1.0-plugins-base gstreamer1.0-plugins-good \
       gstreamer1.0-plugins-bad gstreamer1.0-plugins-ugly gstreamer1.0-libav \
       gstreamer1.0-tools gstreamer1.0-x gstreamer1.0-alsa \
       gstreamer1.0-gl gstreamer1.0-gtk3 gstreamer1.0-qt5 gstreamer1.0-pulseaudio
   ```

### Verify Camera Detection  
Plug in your webcam and verify detection:  

```bash
ls -l /dev/video*
```
You should see something like `video0`.  

```bash
v4l2-ctl --list-devices
```

---

## How to Use  

### Setup  
```bash
go mod init voidlapse
go mod tidy
go build voidlapse
```
To see available flags:
```bash
./voidlapse -h
```
![Flags Overview](./img/3.png)

### Modes  

#### 1. **Shoot Mode (SL)**  
Captures images at regular intervals. Use the following flags:  
- `-d`: Duration in hours.  
- `-i`: Interval between each image in seconds.  
- `-o`: Output directory where the images will be stored.  

Example:  
```bash
./voidlapse -m shoot -d 0.014 -i 5 -o ./timelapse_photos
```

#### 2. **Build Lapse Mode (BL)**  
Converts captured images into a timelapse video. Use the following flags:  
- `-ip`: Input folder containing images from Shoot Mode.  
- `-op`: Output location for the timelapse video.  
- `-t`: (y or n) Add a timestamp to the video.  
- `-c`: Timestamp color (`white` or `black`).  
- `-f`: Timestamp format (`date`, `time`, or `datetime`).  

Example:  
```bash
./voidlapse -m build -ip ./timelapse_photos -op ./timelapse.mp4 -c white -f date -t y
```

#### 3. **Add Clock Mode (AC)**  
Adds timestamps to an existing video based on metadata. Use the following flags:  
- `-it`: Input video file.  
- `-ot`: Output video file with timestamps.  
- `-c`: Timestamp color (`white` or `black`).  
- `-f`: Timestamp format (`date`, `time`, or `datetime`).  

Example:  
```bash
./voidlapse -m timestamp -it ./timelapse_photos/timelapse.mp4 -ot ./timelapse_photos/ts_timelapse.mp4 -c white -f date
```

---

## License  
Void Lapse is open-source and available under the MIT License.

---
