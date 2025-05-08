# MHFJMP Editor

A tool for extracting and injecting menu and area entries in Monster Hunter Frontier Z (MHFZ) binary files.

## Features

- Extract menu entries from `mhfjmp.bin` to CSV format
- Extract area entries from `mhfjmp.bin` to CSV format
- Inject modified menu and area entries back into the binary file
- Support for Shift-JIS text encoding
- Dynamic entry management
- Detailed logging for debugging
- Support for area stage IDs and entry flags
- Automatic offset calculations for data injection

## Prerequisites

- Go 1.16 or higher
- Required Go packages:
  - `golang.org/x/text/encoding/japanese`
  - `golang.org/x/text/transform`

## Project Structure

```
mhfjmp-editor/
├── extractor/
│   └── extractor.go    # Handles data extraction to CSV
├── injector/
│   └── injector.go     # Handles data injection from CSV
└── main.go            # Main application entry point
```

## Usage

1. Place your `mhfjmp.bin` file in the `input` directory
2. Run the extraction process:
   ```bash
   go run . e
   ```
3. Edit the generated CSV files:
   - `output/menu_entries.csv` for menu entries
   - `output/area_entries.csv` for area entries
4. Run the injection process:
   ```bash
   go run . i
   ```
5. Find the modified binary at `output/mhfjmp_patched.bin`

## CSV Formats

### Menu Entries CSV
The menu entries CSV file contains the following columns:
- ID (Reference only)
- Title
- Description
- JumpID
- Unk0C
- AreaID
- AreaID2
- AreaID3
- Rotation
- PosX
- PosY
- PosZ
- Unk28
- PosX1
- PosY1
- PosZ1
- Rotation1
- Title
- Description

### Area Entries CSV
The area entries CSV file contains the following columns:
- AreaIndex (Used to determine the number of areas)
- EntryDataLength
- EntryData (Format: [Index,Flags] [Index,Flags] ...)
- StageIDs (Format: ID1 ID2 ID3 ...)

## Data Structure

### Menu Entry Structure
```go
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
    Rotation       uint32
    PosX1       float32
    PosY1       float32
    PosZ1       float32
    Rotation1       uint32
    Title       string
    Description string
}
```

### Area Structure
```go
type Area struct {
    pEntryData   uint32
    lenEntryData uint32
    pStageIds    uint32
    entry        []AreaEntry
    stageIds     []uint16
}

type AreaEntry struct {
    Index uint16
    Flags uint16
}
```

## Notes

- The tool automatically handles text encoding conversion between Shift-JIS and UTF-8
- Area entries are injected after menu entries in the binary file
- Stage IDs are terminated with a uint16(0) after each list
- The number of areas is determined by the last AreaIndex value in the CSV
- All offsets are calculated dynamically based on the data size
