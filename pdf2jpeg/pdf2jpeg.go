package pdf2jpeg

// by Gabriel Ochsenhofer
// Uses imagemagick to convert a pdf file into a jpeg
// NEEDS ImageMagick convert command
// Ubuntu: sudo apt-get install imagemagick
// OSX: brew install imagemagick OR sudo port install imagemagick

import (
	"bytes"
	"errors"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
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

type PDFInfo struct {
	Width  int
	Height int
	Pages  int
}

func defaultOptions() ConvertOptions {
	return ConvertOptions{
		true,
		225,
		60,
		true,
		60,
		true,
		true,
		true,
		true,
	}
}

// Checks file signature
// http://en.wikipedia.org/wiki/List_of_file_signatures
func IsValidP(inputPath string) (bool, error) {
	file, err := os.Open(inputPath)
	if err != nil {
		return false, err
	}
	defer file.Close()
	return IsValid(file)
}

// Checks file signature
// http://en.wikipedia.org/wiki/List_of_file_signatures
func IsValid(file *os.File) (bool, error) {
	buff := make([]byte, 4)
	n, err := file.Read(buff)
	if err != nil {
		return false, err
	}
	if n < 4 {
		return false, errors.New("The file is too small to be valid (unable to read file signature).")
	}
	// 25 50 44 46 (%PDF)
	if bytes.Compare(buff, []byte{0x25, 0x50, 0x44, 0x46}) == 0 {
		return true, nil
	}
	return false, nil
}

func GetNumPages(inputPath string) (int, error) {
	if ok, er := IsValidP(inputPath); !ok {
		if er != nil {
			return 0, errors.New("IsValid: " + er.Error())
		}
		return 0, errors.New("Invalid file signature.")
	}
	cmd := exec.Command("identify", "-format", "%n", inputPath)
	buff := new(bytes.Buffer)
	cmd.Stdout = buff
	//cmd.Stderr = buff
	err := cmd.Run()
	if err != nil {
		log.Println("ERROR RUNNING IDENTIFY COMMAND", err.Error())
		return 0, err
	}
	trimmed := strings.TrimSpace(buff.String())
	return strconv.Atoi(trimmed)
}

func GetInfo(inputPath string) (*PDFInfo, error) {
	if ok, er := IsValidP(inputPath); !ok {
		if er != nil {
			return nil, errors.New("IsValid: " + er.Error())
		}
		return nil, errors.New("Invalid file signature.")
	}
	cmd := exec.Command("identify", "-format", "\"%w,%h,%n\"", inputPath)
	buff := new(bytes.Buffer)
	cmd.Stdout = buff
	//cmd.Stderr = buff
	err := cmd.Run()
	if err != nil {
		log.Println("ERROR RUNNING IDENTIFY COMMAND", err.Error())
		return nil, err
	}
	trimmed := strings.TrimSpace(buff.String())
	slices := strings.Split(trimmed, ",")
	pdfi := &PDFInfo{}
	if len(slices) < 3 {
		return nil, errors.New("Could not get data! " + buff.String())
	}
	pdfi.Width, err = strconv.Atoi(slices[0])
	if err != nil {
		return pdfi, err
	}
	pdfi.Height, err = strconv.Atoi(slices[1])
	if err != nil {
		return pdfi, err
	}
	pdfi.Pages, err = strconv.Atoi(slices[2])
	return pdfi, err
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
			log.Println("[[pdf verbose]]", buff.String())
		}
		return nil
	}
	return errors.New(buff.String())
}
