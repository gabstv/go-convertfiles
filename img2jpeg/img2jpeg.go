package img2jpeg

import (
	"bytes"
	"errors"
	"log"
	"os"
	"os/exec"
	"strconv"
)

type ConvertOptions struct {
	Verbose        bool
	Density        int
	Resize         int
	Trim           bool
	Quality        int
	Sharpen        bool
	WhiteBG        bool
	Flatten        bool
	SRGBColorspace bool
}

func defaultOptions() ConvertOptions {
	return ConvertOptions{
		true,
		0,
		0,
		true,
		60,
		true,
		true,
		true,
		true,
	}
}

// Checks if the image is a png,jpeg,bmp,psd
func IsValid(file *os.File) (bool, error) {
	buff := make([]byte, 8)
	n, err := file.Read(buff)
	if err != nil {
		return false, err
	}
	if n < 8 {
		return false, errors.New("The file is too small to be valid (unable to read file signature).")
	}
	// PNG 89 50 4E 47 0D 0A 1A 0A
	if bytes.Compare(buff[:8], []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}) == 0 {
		return true, nil
	}
	// JPG FF D8 FF
	if bytes.Compare(buff[:3], []byte{0xFF, 0xD8, 0xFF}) == 0 {
		return true, nil
	}
	// BMP 42 4D
	if bytes.Compare(buff[:2], []byte{0x42, 0x4D}) == 0 {
		return true, nil
	}
	// PSD 38 42 50 53
	if bytes.Compare(buff[:4], []byte{0x38, 0x42, 0x50, 0x53}) == 0 {
		return true, nil
	}
	return false, nil
}

// Checks if the image is a png,jpeg,bmp,psd
func IsValidP(inputPath string) (bool, error) {
	file, err := os.Open(inputPath)
	if err != nil {
		return false, err
	}
	defer file.Close()
	return IsValid(file)
}

func ConvertToJpeg(inputPath, outputPath string, opts ...ConvertOptions) error {
	opt := defaultOptions()
	if len(opts) > 0 {
		opt = opts[0]
	}
	args := make([]string, 0, 100)
	if opt.Verbose {
		args = append(args, "-verbose")
	}
	if opt.SRGBColorspace {
		args = append(args, "-colorspace", "sRGB")
	}
	if opt.Density > 0 {
		args = append(args, "-density", strconv.Itoa(opt.Density))
	}
	// append input file now
	args = append(args, inputPath)
	// other options come after
	if opt.Quality > 0 {
		args = append(args, "-quality", strconv.Itoa(opt.Quality))
	}
	if opt.Sharpen {
		args = append(args, "-sharpen", "0x1.0")
	}
	if opt.Resize > 0 {
		args = append(args, "-resize", strconv.Itoa(opt.Resize)+"%")
	}
	if opt.WhiteBG {
		args = append(args, "-background", "white")
	}
	if opt.Flatten {
		args = append(args, "-flatten")
	}
	// append output file now
	args = append(args, outputPath)
	cmd := exec.Command("convert", args...)
	buff := new(bytes.Buffer)
	cmd.Stderr = buff
	cmd.Stdout = buff
	err := cmd.Start()
	if err != nil {
		return err
	}
	v, err := cmd.Process.Wait()
	if err != nil {
		return err
	}
	if v.Success() {
		if opt.Verbose {
			log.Println("[[img verbose]]", buff.String())
		}
		return nil
	}
	return errors.New(buff.String())
}
