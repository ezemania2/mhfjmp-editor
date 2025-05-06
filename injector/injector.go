package injector

import (
	"encoding/binary"
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
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

func loadMenuEntriesFromCSV(path string) ([]MenuEntry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var entries []MenuEntry
	for i, rec := range records {
		if i == 0 {
			continue
		}
		if len(rec) < 21 {
			log.Printf("Warning: Line %d does not have enough columns (%d/21)", i, len(rec))
			continue
		}

		entry := MenuEntry{
			Title:       rec[1],
			Description: rec[2],
			JumpID:      parseUint32(rec[3]),
			Unk0C:       parseUint32(rec[4]),
			AreaID:      parseUint16(rec[5]),
			AreaID2:     parseUint16(rec[6]),
			AreaID3:     parseUint16(rec[7]),
			Unk18:       parseUint16(rec[8]),
			PosX:        parseFloat32(rec[9]),
			PosY:        parseFloat32(rec[10]),
			PosZ:        parseFloat32(rec[11]),
			Unk28:       parseUint8(rec[12]),
			Unk29:       parseUint8(rec[13]),
			Unk30:       parseUint16(rec[14]),
			PosX1:       parseFloat32(rec[15]),
			PosY1:       parseFloat32(rec[16]),
			PosZ1:       parseFloat32(rec[17]),
			Unk38:       parseUint8(rec[18]),
			Unk39:       parseUint8(rec[19]),
			Unk40:       parseUint16(rec[20]),
		}
		entryID := parseUint32(rec[0])
		log.Printf("Entry %d loaded (Reference ID: %d): JumpID=%d, AreaID=%d, Pos=(%.2f,%.2f,%.2f)",
			i, entryID, entry.JumpID, entry.AreaID, entry.PosX, entry.PosY, entry.PosZ)
		entries = append(entries, entry)
	}
	return entries, nil
}

func InjectData() {
	entries, err := loadMenuEntriesFromCSV("output/menu_entries.csv")
	if err != nil {
		log.Fatalf("Error loading CSV: %v", err)
	}
	log.Printf("Number of entries loaded from CSV: %d", len(entries))

	data, err := os.ReadFile("input/mhfjmp.bin")
	if err != nil {
		log.Fatalf("Error reading mhfjmp.bin: %v", err)
	}
	log.Printf("Size of mhfjmp.bin file: %d bytes", len(data))

	soMenuEntry := 0xA80C
	const menuEntrySize = 56

	existingEntries := (len(data) - soMenuEntry) / menuEntrySize
	log.Printf("Number of existing entries: %d", existingEntries)

	totalEntriesSize := (existingEntries + len(entries)) * menuEntrySize
	log.Printf("Total size needed for entries: %d bytes", totalEntriesSize)

	textSectionOffset := soMenuEntry + totalEntriesSize
	log.Printf("Text section offset: %d", textSectionOffset)

	totalSize := textSectionOffset + 4096
	output := make([]byte, totalSize)
	copy(output, data)
	log.Printf("Output buffer size: %d", len(output))

	// Replace data at offset 0x00 with F1 0C
	output[0] = 0xF1
	output[1] = 0x0C

	headerSize := 17
	for i := 0; i < 16; i++ {
		output[len(data)+i] = 0x00
	}
	output[len(data)+16] = 0xFF

	soMenuEntry = len(data) + headerSize
	log.Printf("New entry start offset after header: %d", soMenuEntry)

	var stringSection []byte
	var textOffsets []uint32

	for _, entry := range entries {
		titleOffset := uint32(textSectionOffset + len(stringSection))
		stringSection = append(stringSection, encodeShiftJIS(entry.Title)...)
		stringSection = append(stringSection, 0x00)

		descriptionOffset := uint32(textSectionOffset + len(stringSection))
		stringSection = append(stringSection, encodeShiftJIS(entry.Description)...)
		stringSection = append(stringSection, 0x00)

		textOffsets = append(textOffsets, titleOffset, descriptionOffset)
	}
	log.Printf("Text section size: %d", len(stringSection))

	for i, entry := range entries {
		base := soMenuEntry + (i * menuEntrySize)
		log.Printf("Writing entry %d at offset %d", i, base)

		if base+menuEntrySize > len(output) {
			log.Fatalf("Buffer too small for entry %d (offset %d + %d > %d)",
				i, base, menuEntrySize, len(output))
		}

		binary.LittleEndian.PutUint32(output[base:], entry.JumpID)
		binary.LittleEndian.PutUint32(output[base+4:], entry.Unk0C)
		binary.LittleEndian.PutUint16(output[base+8:], entry.AreaID)
		binary.LittleEndian.PutUint16(output[base+10:], entry.AreaID2)
		binary.LittleEndian.PutUint16(output[base+12:], entry.AreaID3)
		binary.LittleEndian.PutUint16(output[base+14:], entry.Unk18)
		writeFloat32(output[base+16:], entry.PosX)
		writeFloat32(output[base+20:], entry.PosY)
		writeFloat32(output[base+24:], entry.PosZ)
		output[base+28] = entry.Unk28
		output[base+29] = entry.Unk29
		binary.LittleEndian.PutUint16(output[base+30:], entry.Unk30)
		writeFloat32(output[base+32:], entry.PosX1)
		writeFloat32(output[base+36:], entry.PosY1)
		writeFloat32(output[base+40:], entry.PosZ1)
		output[base+44] = entry.Unk38
		output[base+45] = entry.Unk39
		binary.LittleEndian.PutUint16(output[base+46:], entry.Unk40)
		binary.LittleEndian.PutUint32(output[base+48:], textOffsets[i*2])
		binary.LittleEndian.PutUint32(output[base+52:], textOffsets[i*2+1])
	}

	copy(output[textSectionOffset:], stringSection)

	err = os.WriteFile("output/mhfjmp_patched.bin", output, 0644)
	if err != nil {
		log.Fatalf("Error writing: %v", err)
	}

	fmt.Println("✅ Injection completed in output/mhfjmp_patched.bin")
}

func encodeShiftJIS(str string) []byte {
	buf := make([]byte, len(str)*2)
	n, _, _ := transform.Bytes(japanese.ShiftJIS.NewEncoder(), []byte(str))
	copy(buf, n)
	return n
}

func parseUint32(s string) uint32 {
	v, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		log.Printf("Error parsing uint32 '%s': %v", s, err)
		return 0
	}
	return uint32(v)
}

func parseUint16(s string) uint16 {
	v, err := strconv.ParseUint(s, 10, 16)
	if err != nil {
		log.Printf("Error parsing uint16 '%s': %v", s, err)
		return 0
	}
	return uint16(v)
}

func parseUint8(s string) uint8 {
	v, err := strconv.ParseUint(s, 10, 8)
	if err != nil {
		log.Printf("Error parsing uint8 '%s': %v", s, err)
		return 0
	}
	return uint8(v)
}

func parseFloat32(s string) float32 {
	v, err := strconv.ParseFloat(s, 32)
	if err != nil {
		log.Printf("Error parsing float32 '%s': %v", s, err)
		return 0
	}
	return float32(v)
}

func writeFloat32(b []byte, f float32) {
	binary.LittleEndian.PutUint32(b, math.Float32bits(f))
}

func Start() {
	InjectData()
}
