package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	// Directory containing your media files
	dir := "."

	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	// Check if ffmpeg is available in PATH
	if err := haveFFmpeg(); err != nil {
		fmt.Println("ffmpeg not found in PATH:", err)
		return
	}

	// Read directory entries
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Println("Failed to open directory:", err)
		return
	}

	// Audio filter: compressor + loudness normalization
	filter := `acompressor,loudnorm=I=-14:TP=-1.8:LRA=11`

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		in := filepath.Join(dir, e.Name())
		ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(in), "."))
		if !isAudio(ext) && !isVideo(ext) {
			continue
		}

		// Temporary output file
		tmp := tmpName(in)

		// Build ffmpeg command
		args := []string{"-y", "-i", in, "-filter:a", filter}
		if isVideo(ext) {
			args = append(args, "-c:v", "copy") // keep video stream as-is
		} else {
			args = append(args, "-vn") // disable video if it's audio only
		}
		args = append(args, tmp)

		// Run ffmpeg
		fmt.Println(">>>", "ffmpeg", strings.Join(args, " "))
		if err := run("ffmpeg", args...); err != nil {
			fmt.Println("Failed:", in)
			_ = os.Remove(tmp)
			continue
		}

		// Backup original file
		bak := in + ".bak"
		_ = os.Remove(bak)
		if err := os.Rename(in, bak); err != nil {
			fmt.Println("Backup failed:", err)
			_ = os.Remove(tmp)
			continue
		}

		// Replace original with processed file
		if err := os.Rename(tmp, in); err != nil {
			_ = os.Rename(bak, in)
			fmt.Println("Replace failed:", err)
			continue
		}
		_ = os.Remove(bak)
		fmt.Println("✔ Success:", in)
	}
	fmt.Println("done.")
}

// Check ffmpeg availability
func haveFFmpeg() error {
	return exec.Command("ffmpeg", "-version").Run()
}

// Supported audio extensions
func isAudio(ext string) bool {
	switch ext {
	case "mp3", "m4a", "aac", "wav", "flac", "ogg":
		return true
	}
	return false
}

// Supported video extensions
func isVideo(ext string) bool {
	switch ext {
	case "mp4", "mov", "mkv", "webm":
		return true
	}
	return false
}

// Run a command and capture output
func run(bin string, args ...string) error {
	cmd := exec.Command(bin, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v\n%s", err, string(out))
	}
	return nil
}

// Generate temporary filename in the same directory
func tmpName(in string) string {
	dir := filepath.Dir(in)
	base := strings.TrimSuffix(filepath.Base(in), filepath.Ext(in))
	ext := filepath.Ext(in)
	return filepath.Join(dir, base+".__tmp_norm__"+ext)
}
