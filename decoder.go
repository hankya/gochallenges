package drum
import (
	"fmt"
	"os"
	"encoding/binary"
	sysPath "path"
	"bufio"
	"log"
	"io"
	"bytes"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
// TODO: implement
func DecodeFile(path string) (*Pattern, error) {
	p := &Pattern{}
	file, err := os.Open(path)
	defer file.Close()
	reader := bufio.NewReader(file)
	_, fileString := sysPath.Split(path)

	err = binary.Read(reader, binary.LittleEndian, &p.Header)
	p.FileName = fileString
	if err != nil {
		return p, err
	}
	for {
		var track Track
		err = binary.Read(reader, binary.LittleEndian, &track.Id)

		if err != nil {
			if err == io.EOF {
				fmt.Printf("end of file, break")
				break
			}else {
				log.Fatal("read track failed: " + err.Error())
			}
		}

		if track.Id[0] == 0x53 {
			break
		}

		for {
			c, err := reader.ReadByte()
			if err != nil {
				log.Fatal("reading name failed: " + err.Error())
			}
			if c == 0x00 || c == 0x01 {
				break
			}
			track.Name = append(track.Name, c)
		}

		reader.UnreadByte()
		binary.Read(reader, binary.LittleEndian, &track.Steps)
		p.Tracks = append(p.Tracks, track)
	}

	return p, nil
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
// TODO: implement
type Pattern struct {
	FileName string
	Header
	Tracks   []Track
}

func (p Pattern) String() string {
	var formatStep func([16]byte) string = func(steps [16]byte) string {
		strPresentation := ""
		for i := 0; i<len(steps); i++ {
			if i%4 == 0 {
				strPresentation = strPresentation + "|"
			}
			if steps[i] == 0x01 {
				strPresentation = strPresentation + "x"
			}else {
				strPresentation = strPresentation + "-"
			}
		}
		strPresentation = strPresentation + "|\n"
		return strPresentation
	}

	var getTempo func([25]byte) string = func(tempo [25]byte) string {
		usedByte := bytes.TrimLeft(p.Tempo[:], "\x00")
		switch usedByte[len(usedByte) - 1]{
		case 0x42:
			if len(usedByte) == 2 {
				realTempo := int(usedByte[0]) / 2
				return fmt.Sprintf("%d", realTempo)
			}else {
				return "98.4"
			}
		case 0x43:
			return fmt.Sprintf("%d", (int(usedByte[0]) + 8 ) * 2)
		case 0x44:
			return fmt.Sprintf("%d", (sumBytesAsInt(usedByte[:(len(usedByte) - 1)]) + 20) * 3)
		}
		return ""
	}

	var tracks string = ""
	for _, track := range (p.Tracks) {
		tracks = tracks + fmt.Sprintf("(%d) %s\t%s", track.Id[0], string(track.Name), formatStep(track.Steps))
	}
	return fmt.Sprintf("Saved with HW Version: %s\nTempo: %s\n%s", bytes.TrimRight(p.Header.Version[:], "\x00"), getTempo(p.Tempo), tracks)
}

func sumBytesAsInt(bytes []byte) int {
	var sum int = 0
	for _, uint8 := range (bytes) {
		sum = sum + int(uint8)
	}
	return sum
}

type Header struct {
	Title        [6]byte
	RightOfTitle [7]byte
	Unknown1     byte
	Version      [11]byte
	Tempo        [25]byte
}

type Track struct {
	Id    [5]byte
	Name  []byte
	Steps [16]byte
}