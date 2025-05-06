package extractor

import (
	"encoding/binary"
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"

	"golang.org/x/text/encoding/japanese"
)

type MenuEntry struct {
	JumpID      uint32
	Unk0C       uint32
	AreaID      uint16
	AreaID2     uint16
	AreaID3     uint16
	Unk18       uint16
	PosX        float32
	PosY        float32
	PosZ        float32
	Unk28       uint8
	Unk29       uint8
	Unk30       uint16
	PosX1       float32
	PosY1       float32
	PosZ1       float32
	Unk38       uint8
	Unk39       uint8
	Unk40       uint16
	Title       string
	Description string
}

func ExtractData() {
	inputPath := filepath.Join("input", "mhfjmp.bin")
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		log.Fatalf("required file mhfjmp.bin not found in input folder")
	}

	outputDir := "output"
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		log.Fatalf("Error creating output directory: %v", err)
	}

	if err := processCSV(outputDir, "menu_entries", []string{"ID", "Title", "Description", "JumpID", "Unk0C", "AreaID", "AreaID2", "AreaID3", "Unk18", "PosX", "PosY", "PosZ", "Unk28", "Unk29", "Unk30", "PosX1", "PosY1", "PosZ1", "Unk38", "Unk39", "Unk40"}); err != nil {
		log.Fatalf("Error processing CSV: %v", err)
	}
}

func MenuEntryData(writer *csv.Writer, br *BinaryReader) error {
	var menuEntries []MenuEntry
	var soMenuEntry uint32 = 0x0007A0
	var endMenuEntry uint32 = 0x000CD7

	// Seek to menu entries position
	_, err := br.BaseStream.Seek(int64(soMenuEntry), 0)
	if err != nil {
		return fmt.Errorf("failed to seek to menu entries position: %w", err)
	}

	// Calculate number of entries
	count := (endMenuEntry - soMenuEntry) / 56 // Each entry is 56 bytes

	// Read entries
	for i := uint32(0); i < count; i++ {
		entry := MenuEntry{}

		entry.JumpID, err = br.ReadUInt32()
		if err != nil {
			return fmt.Errorf("failed to read JumpID: %w", err)
		}
		entry.Unk0C, err = br.ReadUInt32()
		if err != nil {
			return fmt.Errorf("failed to read Unk0C: %w", err)
		}
		entry.AreaID, err = br.ReadUInt16()
		if err != nil {
			return fmt.Errorf("failed to read AreaID: %w", err)
		}
		entry.AreaID2, err = br.ReadUInt16()
		if err != nil {
			return fmt.Errorf("failed to read AreaID2: %w", err)
		}
		entry.AreaID3, err = br.ReadUInt16()
		if err != nil {
			return fmt.Errorf("failed to read AreaID3: %w", err)
		}
		entry.Unk18, err = br.ReadUInt16()
		if err != nil {
			return fmt.Errorf("failed to read Unk18: %w", err)
		}
		entry.PosX, err = br.ReadFloat32()
		if err != nil {
			return fmt.Errorf("failed to read PosX: %w", err)
		}
		entry.PosY, err = br.ReadFloat32()
		if err != nil {
			return fmt.Errorf("failed to read PosY: %w", err)
		}
		entry.PosZ, err = br.ReadFloat32()
		if err != nil {
			return fmt.Errorf("failed to read PosZ: %w", err)
		}
		entry.Unk28, err = br.ReadUInt8()
		if err != nil {
			return fmt.Errorf("failed to read Unk28: %w", err)
		}
		entry.Unk29, err = br.ReadUInt8()
		if err != nil {
			return fmt.Errorf("failed to read Unk29: %w", err)
		}
		entry.Unk30, err = br.ReadUInt16()
		if err != nil {
			return fmt.Errorf("failed to read Unk30: %w", err)
		}
		entry.PosX1, err = br.ReadFloat32()
		if err != nil {
			return fmt.Errorf("failed to read PosX1: %w", err)
		}
		entry.PosY1, err = br.ReadFloat32()
		if err != nil {
			return fmt.Errorf("failed to read PosY1: %w", err)
		}
		entry.PosZ1, err = br.ReadFloat32()
		if err != nil {
			return fmt.Errorf("failed to read PosZ1: %w", err)
		}
		entry.Unk38, err = br.ReadUInt8()
		if err != nil {
			return fmt.Errorf("failed to read Unk38: %w", err)
		}
		entry.Unk39, err = br.ReadUInt8()
		if err != nil {
			return fmt.Errorf("failed to read Unk39: %w", err)
		}
		entry.Unk40, err = br.ReadUInt16()
		if err != nil {
			return fmt.Errorf("failed to read Unk40: %w", err)
		}

		entry.Title, err = StringFromPointer(br)
		if err != nil {
			return fmt.Errorf("failed to read Title: %w", err)
		}
		entry.Description, err = StringFromPointer(br)
		if err != nil {
			return fmt.Errorf("failed to read Description: %w", err)
		}
		menuEntries = append(menuEntries, entry)
	}

	// Write data
	for i, entry := range menuEntries {
		record := []string{
			fmt.Sprint(i),
			fmt.Sprint(entry.Title),
			fmt.Sprint(entry.Description),
			fmt.Sprint(entry.JumpID),
			fmt.Sprint(entry.Unk0C),
			fmt.Sprint(entry.AreaID),
			fmt.Sprint(entry.AreaID2),
			fmt.Sprint(entry.AreaID3),
			fmt.Sprint(entry.Unk18),
			fmt.Sprint(entry.PosX),
			fmt.Sprint(entry.PosY),
			fmt.Sprint(entry.PosZ),
			fmt.Sprint(entry.Unk28),
			fmt.Sprint(entry.Unk29),
			fmt.Sprint(entry.Unk30),
			fmt.Sprint(entry.PosX1),
			fmt.Sprint(entry.PosY1),
			fmt.Sprint(entry.PosZ1),
			fmt.Sprint(entry.Unk38),
			fmt.Sprint(entry.Unk39),
			fmt.Sprint(entry.Unk40),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("error writing record to CSV: %w", err)
		}
	}

	fmt.Println("Data extraction to CSV completed successfully.")
	return nil
}

type BinaryReader struct {
	BaseStream *os.File
}

func (br *BinaryReader) ReadByte() (byte, error) {
	var b [1]byte
	_, err := br.BaseStream.Read(b[:])
	return b[0], err
}

func (br *BinaryReader) ReadInt32() (int32, error) {
	var b [4]byte
	_, err := br.BaseStream.Read(b[:])
	if err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(b[:])), nil
}

func (br *BinaryReader) ReadInt16() (int16, error) {
	var b [2]byte
	_, err := br.BaseStream.Read(b[:])
	if err != nil {
		return 0, err
	}
	return int16(binary.LittleEndian.Uint16(b[:])), nil
}

func (br *BinaryReader) ReadUInt16() (uint16, error) {
	var b [2]byte
	_, err := br.BaseStream.Read(b[:])
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint16(b[:]), nil
}

func (br *BinaryReader) ReadUInt32() (uint32, error) {
	var b [4]byte
	_, err := br.BaseStream.Read(b[:])
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(b[:]), nil
}

func (br *BinaryReader) ReadFloat32() (float32, error) {
	var b [4]byte
	_, err := br.BaseStream.Read(b[:])
	if err != nil {
		return 0, err
	}
	bits := binary.LittleEndian.Uint32(b[:])
	return math.Float32frombits(bits), nil
}

func (br *BinaryReader) ReadUInt8() (uint8, error) {
	var b [1]byte
	_, err := br.BaseStream.Read(b[:])
	return b[0], err
}

func (br *BinaryReader) Close() error {
	return br.BaseStream.Close()
}

func getBinaryReader(filePath string) (*BinaryReader, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	return &BinaryReader{BaseStream: file}, nil
}

func StringFromPointer(br *BinaryReader) (string, error) {
	offset, err := br.ReadUInt32()
	if err != nil {
		return "", err
	}

	currentPos, err := br.BaseStream.Seek(0, 1)
	if err != nil {
		return "", err
	}

	_, err = br.BaseStream.Seek(int64(offset), 0)
	if err != nil {
		return "", err
	}

	var bytes []byte
	for {
		b, err := br.ReadByte()
		if err != nil || b == 0 {
			break
		}
		bytes = append(bytes, b)
	}

	_, err = br.BaseStream.Seek(currentPos, 0)
	if err != nil {
		return "", err
	}

	// Try to convert from Shift-JIS to UTF-8
	decoder := japanese.ShiftJIS.NewDecoder()
	utf8Bytes, err := decoder.Bytes(bytes)
	if err != nil {
		// If conversion fails, return the raw string
		return string(bytes), nil
	}

	return string(utf8Bytes), nil
}

func processCSV(path, fileName string, header []string) error {
	if err := os.MkdirAll(path, 0777); err != nil {
		return fmt.Errorf("error creating directory %s: %w", path, err)
	}

	filepath := fmt.Sprintf("%s/%s.csv", path, fileName)
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	if err := os.Chmod(filepath, 0777); err != nil {
		return fmt.Errorf("error setting permissions for file: %w", err)
	}

	// Use UTF-8 encoding instead of Shift-JIS
	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write(header); err != nil {
		return fmt.Errorf("error writing header: %w", err)
	}

	switch fileName {
	case "menu_entries":
		brInput, err := getBinaryReader("input/mhfjmp.bin")
		if err != nil {
			return fmt.Errorf("error obtaining binary reader for menu entries: %w", err)
		}
		defer brInput.Close()
		err = MenuEntryData(writer, brInput)
		if err != nil {
			return fmt.Errorf("error extracting menu entry data: %w", err)
		}
	}

	return nil
}
