package injector

import (
	"encoding/binary"
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

type AreaEntry struct {
	Index uint16
	Flags uint16
}

type Area struct {
	pEntryData   uint32
	lenEntryData uint32
	pStageIds    uint32
	entry        []AreaEntry
	stageIds     []uint16
}

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
	Unk28       uint32
	PosX1       float32
	PosY1       float32
	PosZ1       float32
	Unk38       uint32
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
			continue // Skip header
		}
		if len(rec) < 17 {
			log.Printf("Warning: Line %d does not have enough columns (%d/17)", i, len(rec))
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
			Unk28:       parseUint32(rec[12]),
			PosX1:       parseFloat32(rec[13]),
			PosY1:       parseFloat32(rec[14]),
			PosZ1:       parseFloat32(rec[15]),
			Unk38:       parseUint32(rec[16]),
		}

		log.Printf("Entry %d loaded: JumpID=%d, AreaID=%d, Pos=(%.2f,%.2f,%.2f)",
			i, entry.JumpID, entry.AreaID, entry.PosX, entry.PosY, entry.PosZ)
		entries = append(entries, entry)
	}
	return entries, nil
}

func loadAreaEntriesFromCSV(path string) ([]Area, uint32, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, 0, err
	}

	var areas []Area
	var numAreas uint32 = 0

	for i, rec := range records {
		if i == 0 {
			continue // Skip header
		}
		if len(rec) < 4 {
			log.Printf("Warning: Line %d does not have enough columns (%d/4)", i, len(rec))
			continue
		}

		// Read AreaIndex from first column
		areaIndex := parseUint32(rec[0])
		// Update numAreas with the last AreaIndex found
		numAreas = areaIndex
		log.Printf("Found AreaIndex: %d", areaIndex)

		area := Area{
			lenEntryData: parseUint32(rec[1]),
			entry:        parseAreaEntries(rec[2]),
			stageIds:     parseStageIds(rec[3]),
		}

		log.Printf("Area %d loaded: Entries=%d, StageIds=%d", i, len(area.entry), len(area.stageIds))
		areas = append(areas, area)
	}

	log.Printf("Total areas loaded: %d (using last AreaIndex: %d)", len(areas), numAreas)
	return areas, numAreas, nil
}

func parseAreaEntries(s string) []AreaEntry {
	var entries []AreaEntry
	// Split by spaces only, keeping the [%s,%s] pairs intact
	parts := strings.Fields(s)
	for _, part := range parts {
		var idx, flags uint16
		// Remove brackets and split by comma
		part = strings.Trim(part, "[]")
		values := strings.Split(part, ",")
		if len(values) == 2 {
			idx = parseUint16(values[0])
			flags = parseUint16(values[1])
			entries = append(entries, AreaEntry{Index: idx, Flags: flags})
		} else {
			log.Printf("Warning: Invalid entry format '%s', expected [idx,flags]", part)
		}
	}
	return entries
}

func parseStageIds(s string) []uint16 {
	var ids []uint16
	// Split by both spaces and commas
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == ' ' || r == ','
	})
	for _, part := range parts {
		id, err := strconv.ParseUint(part, 10, 16)
		if err != nil {
			log.Printf("Warning: Error parsing stage ID '%s': %v", part, err)
			continue
		}
		ids = append(ids, uint16(id))
	}
	return ids
}

func InjectData() {
	entries, err := loadMenuEntriesFromCSV("output/menu_entries.csv")
	if err != nil {
		log.Fatalf("Error loading CSV: %v", err)
	}
	log.Printf("Number of entries loaded from CSV: %d", len(entries))

	areas, numAreas, err := loadAreaEntriesFromCSV("output/area_entries.csv")
	if err != nil {
		log.Fatalf("Error loading area entries: %v", err)
	}
	log.Printf("Number of areas loaded from CSV: %d", len(areas))

	data, err := os.ReadFile("input/mhfjmp.bin")
	if err != nil {
		log.Fatalf("Error reading mhfjmp.bin: %v", err)
	}
	log.Printf("Size of mhfjmp.bin file: %d bytes", len(data))

	// Calculate menu entry section offset
	soMenuEntry := len(data) + 17 // Original file size + header size
	const menuEntrySize = 56

	existingEntries := 0 // We're not using existing entries anymore
	log.Printf("Number of existing entries: %d", existingEntries)

	totalEntriesSize := len(entries) * menuEntrySize
	log.Printf("Total size needed for entries: %d bytes", totalEntriesSize)

	// Calculate text section offset after menu entries
	textSectionOffset := soMenuEntry + totalEntriesSize
	log.Printf("Text section offset: %d", textSectionOffset)

	// Calculate total size needed for areas
	var totalAreaSize uint32
	for _, area := range areas {
		areaSize := 12 + uint32(len(area.entry))*4 + uint32(len(area.stageIds))*2 + 2
		totalAreaSize += areaSize
	}

	// Calculate area section offset after text section
	areaSectionOffset := textSectionOffset + 3072 + 6 // Reserve 3KB for text section + 6 bytes padding
	log.Printf("Area section offset: %d", areaSectionOffset)

	// Calculate total size needed for the entire file
	totalSize := areaSectionOffset + int(totalAreaSize)
	log.Printf("Total size needed: %d bytes (Text section: %d, Area section: %d)",
		totalSize, 3072, totalAreaSize)

	output := make([]byte, totalSize)
	copy(output, data)
	log.Printf("Output buffer size: %d", len(output))

	// Calculate dynamic pointer to menu entries
	menuEntryPointer := uint32(soMenuEntry)
	log.Printf("Menu entry pointer: 0x%X", menuEntryPointer)

	// Write dynamic pointer at offset 0x00
	binary.LittleEndian.PutUint16(output[0x00:], uint16(menuEntryPointer))
	binary.LittleEndian.PutUint16(output[0x02:], uint16(menuEntryPointer>>16))

	// Write number of area entries at offset 0x08
	log.Printf("Number of area entries: %d", numAreas)
	binary.LittleEndian.PutUint32(output[0x08:], numAreas)

	for i := 0; i < 16; i++ {
		output[len(data)+i] = 0x00
	}
	output[len(data)+16] = 0xFF

	// Write menu entries
	var stringSection []byte
	var textOffsets []uint32

	// First, collect all text offsets
	for _, entry := range entries {
		titleOffset := uint32(textSectionOffset + len(stringSection))
		stringSection = append(stringSection, encodeShiftJIS(entry.Title)...)
		stringSection = append(stringSection, 0x00)

		descriptionOffset := uint32(textSectionOffset + len(stringSection))
		stringSection = append(stringSection, encodeShiftJIS(entry.Description)...)
		stringSection = append(stringSection, 0x00)

		textOffsets = append(textOffsets, titleOffset, descriptionOffset)
	}

	// Then write menu entries
	for i, entry := range entries {
		base := soMenuEntry + (i * menuEntrySize)
		log.Printf("Writing menu entry %d at offset %d", i, base)

		if base+menuEntrySize > len(output) {
			log.Fatalf("Buffer too small for menu entry %d (offset %d + %d > %d)",
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
		binary.LittleEndian.PutUint32(output[base+28:], entry.Unk28)
		writeFloat32(output[base+32:], entry.PosX1)
		writeFloat32(output[base+36:], entry.PosY1)
		writeFloat32(output[base+40:], entry.PosZ1)
		binary.LittleEndian.PutUint32(output[base+44:], entry.Unk38)
		binary.LittleEndian.PutUint32(output[base+48:], textOffsets[i*2])
		binary.LittleEndian.PutUint32(output[base+52:], textOffsets[i*2+1])

		log.Printf("Menu entry %d written: JumpID=%d, AreaID=%d, Pos=(%.2f,%.2f,%.2f)",
			i, entry.JumpID, entry.AreaID, entry.PosX, entry.PosY, entry.PosZ)
	}

	// Write text section
	copy(output[textSectionOffset:], stringSection)
	log.Printf("Text section written at offset %d (size: %d)", textSectionOffset, len(stringSection))

	// Write areas
	headersOffset := areaSectionOffset
	dataOffset := headersOffset + len(areas)*12 // 12 bytes per header

	// First pass: write all headers
	var cumulativeOffset uint32 = uint32(dataOffset)
	for i, area := range areas {
		headerOffset := headersOffset + i*12
		// Calculate data offset for this area
		dataOffsetForArea := cumulativeOffset
		stageIdsOffset := dataOffsetForArea + uint32(len(area.entry)*4)

		log.Printf("Writing header for area %d at offset %d (data at %d)", i, headerOffset, dataOffsetForArea)
		binary.LittleEndian.PutUint32(output[headerOffset:], dataOffsetForArea)
		binary.LittleEndian.PutUint32(output[headerOffset+4:], uint32(len(area.entry)))
		binary.LittleEndian.PutUint32(output[headerOffset+8:], stageIdsOffset)

		// Calculate next area's offset
		areaSize := uint32(len(area.entry)*4 + len(area.stageIds)*2 + 2) // +2 for terminator
		cumulativeOffset += areaSize
	}

	// Second pass: write all data
	currentDataOffset := dataOffset
	for i, area := range areas {
		log.Printf("Writing data for area %d at offset %d", i, currentDataOffset)

		// Write entries
		for j, entry := range area.entry {
			entryOffset := currentDataOffset + j*4
			binary.LittleEndian.PutUint16(output[entryOffset:], entry.Index)
			binary.LittleEndian.PutUint16(output[entryOffset+2:], entry.Flags)
		}

		// Write stage IDs
		stageIdsOffset := currentDataOffset + len(area.entry)*4
		log.Printf("Writing stage IDs for area %d at offset %d (count: %d)", i, stageIdsOffset, len(area.stageIds))
		for j, stageId := range area.stageIds {
			stageIdOffset := stageIdsOffset + j*2
			if stageIdOffset+2 > len(output) {
				log.Fatalf("Buffer too small for stage ID %d in area %d (offset %d + 2 > %d)",
					j, i, stageIdOffset, len(output))
			}
			binary.LittleEndian.PutUint16(output[stageIdOffset:], stageId)
		}

		// Add terminating uint16 (0) after stageIds
		terminatorOffset := stageIdsOffset + len(area.stageIds)*2
		if terminatorOffset+2 > len(output) {
			log.Fatalf("Buffer too small for terminator in area %d (offset %d + 2 > %d)",
				i, terminatorOffset, len(output))
		}
		log.Printf("Writing terminator for area %d at offset %d", i, terminatorOffset)
		binary.LittleEndian.PutUint16(output[terminatorOffset:], 0)

		// Update current data offset for next area
		currentDataOffset = terminatorOffset + 2 // Add 2 bytes for the terminator
		log.Printf("Next area data will start at offset %d", currentDataOffset)
	}

	// Update header with new offsets
	binary.LittleEndian.PutUint32(output[0x04:], uint32(areaSectionOffset))
	binary.LittleEndian.PutUint32(output[0x08:], numAreas)

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
