package video

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var ffprobeAtomLineRE = regexp.MustCompile(`type:'([^']+)'.* sz:\s+(\d+)\s+(\d+)\s+(\d+)`)

const defaultFastStartOffsetLimit int64 = 32 * 1024 * 1024

type AtomInfo struct {
	Type   string
	Size   int64
	Offset int64
	End    int64
}

type FastStartInfo struct {
	Moov           AtomInfo
	Mdat           *AtomInfo
	FileSize       int64
	NeedsFastStart bool
	Reason         string
}

func IsMP4Family(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".mp4", ".m4v", ".mov", ".3gp", ".3g2", ".mj2", ".m4a":
		return true
	default:
		return false
	}
}

func ProbeTopLevelAtoms(ctx context.Context, path string) ([]AtomInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	fileSize := stat.Size()
	var atoms []AtomInfo
	var offset int64
	hasMoov := false
	hasMdat := false

	for offset+8 <= fileSize {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		atom, err := readAtomAt(file, offset, fileSize)
		if err != nil {
			if hasMoov && hasMdat {
				return atoms, nil
			}
			return nil, err
		}

		atoms = append(atoms, atom)
		switch atom.Type {
		case "moov":
			hasMoov = true
		case "mdat":
			hasMdat = true
		}
		if hasMoov && hasMdat {
			return atoms, nil
		}

		offset = atom.End
	}

	return atoms, nil
}

func CheckFastStart(ctx context.Context, path string) (FastStartInfo, error) {
	atoms, err := ProbeTopLevelAtoms(ctx, path)
	if err != nil {
		return FastStartInfo{}, err
	}

	stat, err := os.Stat(path)
	if err != nil {
		return FastStartInfo{}, err
	}

	var moov *AtomInfo
	var mdat *AtomInfo
	for _, atom := range atoms {
		switch atom.Type {
		case "moov":
			next := atom
			moov = &next
		case "mdat":
			if mdat == nil {
				next := atom
				mdat = &next
			}
		}
	}

	if moov == nil {
		return FastStartInfo{}, fmt.Errorf("moov atom not found in %s", path)
	}

	info := FastStartInfo{
		Moov:     *moov,
		Mdat:     mdat,
		FileSize: stat.Size(),
	}

	if mdat != nil && moov.Offset > mdat.Offset {
		info.NeedsFastStart = true
		info.Reason = "moov atom is after mdat"
		return info, nil
	}

	if moov.Offset > defaultFastStartOffsetLimit {
		info.NeedsFastStart = true
		info.Reason = fmt.Sprintf("moov atom offset is greater than %d bytes", defaultFastStartOffsetLimit)
		return info, nil
	}

	return info, nil
}

func readAtomAt(reader io.ReadSeeker, offset int64, fileSize int64) (AtomInfo, error) {
	if _, err := reader.Seek(offset, io.SeekStart); err != nil {
		return AtomInfo{}, err
	}

	header := make([]byte, 8)
	if _, err := io.ReadFull(reader, header); err != nil {
		return AtomInfo{}, err
	}

	size32 := binary.BigEndian.Uint32(header[0:4])
	atomType := string(header[4:8])
	headerSize := int64(8)
	var atomSize int64

	switch size32 {
	case 0:
		atomSize = fileSize - offset
	case 1:
		largeSizeBytes := make([]byte, 8)
		if _, err := io.ReadFull(reader, largeSizeBytes); err != nil {
			return AtomInfo{}, err
		}
		atomSize = int64(binary.BigEndian.Uint64(largeSizeBytes))
		headerSize = 16
	default:
		atomSize = int64(size32)
	}

	if atomSize < headerSize {
		return AtomInfo{}, fmt.Errorf("invalid atom %q at offset %d", atomType, offset)
	}

	end := offset + atomSize
	if end > fileSize {
		return AtomInfo{}, fmt.Errorf("atom %q at offset %d exceeds file size", atomType, offset)
	}

	return AtomInfo{
		Type:   atomType,
		Size:   atomSize,
		Offset: offset,
		End:    end,
	}, nil
}
func FixFastStart(ctx context.Context, inputPath string) error {
	if IsMP4Family(inputPath) {

		//fastStartInfo, err := CheckFastStart(ctx, inputPath)
		//if err != nil {
		//	return err
		//} else if fastStartInfo.NeedsFastStart {
		//	dir := filepath.Dir(inputPath)
		//	ext := filepath.Ext(inputPath)
		//	base := inputPath[:len(inputPath)-len(ext)]
		//
		//	tmpPath := base + ".faststart.tmp" + ext
		//	backupPath := base + ".backup" + ext
		//
		//	cmd := exec.CommandContext(
		//		ctx,
		//		"ffmpeg",
		//		"-y",
		//		"-i", inputPath,
		//		"-c", "copy",
		//		"-movflags", "+faststart",
		//		tmpPath,
		//	)
		//
		//	cmd.Dir = dir
		//
		//	output, err := cmd.CombinedOutput()
		//	if err != nil {
		//		_ = os.Remove(tmpPath)
		//		return fmt.Errorf("ffmpeg faststart failed: %w: %s", err, string(output))
		//	}
		//
		//	if err := os.Rename(inputPath, backupPath); err != nil {
		//		_ = os.Remove(tmpPath)
		//		return err
		//	}
		//
		//	if err := os.Rename(tmpPath, inputPath); err != nil {
		//		_ = os.Rename(backupPath, inputPath)
		//		_ = os.Remove(tmpPath)
		//		return err
		//	}
		//	return os.Remove(backupPath)
		//}

	}

	return nil
}
