# MHFJMP Editor

A tool for extracting and injecting menu entries in Monster Hunter Frontier Z (MHFZ) binary files.

(Disclaimer you still have to edit the original pointer to make it works)

## Features

- Extract menu entries from `mhfjmp.bin` to CSV format
- Inject modified menu entries back into the binary file
- Support for Shift-JIS text encoding
- Dynamic entry management
- Detailed logging for debugging

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
3. Edit the generated CSV file in `output/menu_entries.csv`
4. Run the injection process:
   ```bash
   go run . i
   ```
5. Find the modified binary at `output/mhfjmp_patched.bin`

## CSV Format

The CSV file contains the following columns:
- ID (Reference only)
- Title
- Description
- JumpID
- Unk0C
- AreaID
- AreaID2
- AreaID3
- Unk18
- PosX
- PosY
- PosZ
- Unk28
- Unk29
- Unk30
- PosX1
- PosY1
- PosZ1
- Unk38
- Unk39
- Unk40
