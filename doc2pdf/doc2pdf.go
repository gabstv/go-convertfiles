package doc2pdf

// Convert .doc .docx files to PDF

import (
	"bytes"
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"runtime"
)

// Checks file signature
// http://en.wikipedia.org/wiki/List_of_file_signatures
func IsValid(file *os.File) (bool, error) {
	buff := make([]byte, 8)
	n, err := file.Read(buff)
	if err != nil {
		return false, err
	}
	if n < 8 {
		return false, errors.New("The file is too small to be valid (unable to read file signature).")
	}
	// DOC D0 CF 11 E0
	if bytes.Compare(buff[:4], []byte{0xD0, 0xCF, 0x11, 0xE0}) == 0 {
		return true, nil
	}
	// DOCX WIKIPEDIA FILE SIGNATURE
	//
	//                        INVESTIGATE!
	//                       /
	//                  ___
	//          D>=G==='   '.
	//                |======|
	//                |======|
	//            )--/]IIIIII]
	//               |_______|
	//              C O  O O D
	//             C  O  O  O D
	//           C O  O  O  O  D
	//          C__O___O__O__O__D
	//         [________________]
	//
	// DOCX 50 4B 03 04, 50 4B 05 06 (empty archive) [WIKIPEDIA ~ INVESTIGATE THIS: http://en.wikipedia.org/wiki/List_of_file_signatures]
	if bytes.Compare(buff, []byte{0x50, 0x4B, 0x03, 0x04, 0x50, 0x4B, 0x05, 0x06}) == 0 {
		return true, nil
	}
	// DOCX 50 4B 07 08 (spanned archive) [WIKIPEDIA ~ INVESTIGATE THIS]
	if bytes.Compare(buff[:4], []byte{0x50, 0x4B, 0x07, 0x08}) == 0 {
		return true, nil
	}
	// DOCX 50 4b 03 04 14 00 06 00 [OFFICE 2007] [http://www.filesignatures.net/index.php?page=search&search=504B030414000600&mode=SIG]
	if bytes.Compare(buff, []byte{0x50, 0x4B, 0x03, 0x04, 0x14, 0x00, 0x06, 0x00}) == 0 {
		return true, nil
	}
	// DOCX PROMISCUOUS MODE
	if bytes.Compare(buff[:4], []byte{0x50, 0x4B, 0x03, 0x04}) == 0 {
		log.Printf("PROMISCUOUS DOC %x\n", buff)
		return true, nil
	}
	//log.Printf("%x\n", buff)
	return false, nil
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

func ConvertToPDF(inputPath, outputPath string) error {
	if ok, er := IsValidP(inputPath); !ok {
		if er != nil {
			return errors.New("IsValid: " + er.Error())
		}
		return errors.New("Invalid file signature.")
	}
	var cmd *exec.Cmd
	log.Println("doc2pdf GOOS", runtime.GOOS)
	if runtime.GOOS == "darwin" {
		cmd = exec.Command("/Applications/LibreOffice.app/Contents/MacOS/python", "/usr/local/bin/unoconv", "-f", "pdf", "-o", outputPath, inputPath)
	} else if runtime.GOOS == "linux" {
		cmd = exec.Command("libreoffice", "--headless", "--convert-to", "pdf", "--outdir", "/tmp", inputPath)
		log.Println("[[doc verbose]]", "libreoffice --headless --convert-to pdf --outdir /tmp "+inputPath)
	} else {
		return errors.New("Only OSX and Ubuntu supported for now!")
	}
	buff := new(bytes.Buffer)
	wtr := io.MultiWriter(buff, os.Stdout)
	cmd.Stderr = wtr
	cmd.Stdout = wtr
	err := cmd.Start()
	if err != nil {
		return err
	}
	log.Println("[[doc verbose]]", "conversion started")
	v, err := cmd.Process.Wait()
	if err != nil {
		return err
	}
	log.Println("[[doc verbose]]", v)
	if v.Success() {
		log.Println("[[doc verbose]]", buff.String())
		// move file
		if runtime.GOOS == "linux" {
			enxt := path.Ext(inputPath)
			_, fn0 := path.Split(inputPath)
			er9 := os.Rename(path.Join("/tmp", fn0[:len(fn0)-len(enxt)]+".pdf"), outputPath)
			if er9 != nil {
				log.Println("[[doc verbose]]", "linux rename error", er9)
			}
		}
		return nil
	}
	return errors.New(buff.String())
}
