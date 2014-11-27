package lucene46

import (
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/core/codec"
	. "github.com/balzaczyy/golucene/core/codec/spi"
	. "github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
)

// lucene46/Lucene46SegmentInfoFormat.java

const (
	// File extension used to store SegmentInfo.
	SI_EXTENSION        = "si"
	SI_CODEC_NAME       = "Lucene46SegmentInfo"
	SI_VERSION_START    = 0
	SI_VERSION_CHECKSUM = 1
	SI_VERSION_CURRENT  = SI_VERSION_CHECKSUM
)

type Lucene46SegmentInfoFormat struct {
}

func NewLucene46SegmentInfoFormat() *Lucene46SegmentInfoFormat {
	return &Lucene46SegmentInfoFormat{}
}

func (f *Lucene46SegmentInfoFormat) SegmentInfoReader() SegmentInfoReader {
	return f
}

func (f *Lucene46SegmentInfoFormat) SegmentInfoWriter() SegmentInfoWriter {
	return f
}

// codecs/lucene46/Lucene46SegmentInfoReader.java

func (f *Lucene46SegmentInfoFormat) Read(dir store.Directory,
	segName string, ctx store.IOContext) (si *SegmentInfo, err error) {

	filename := util.SegmentFileName(segName, "", SI_EXTENSION)
	var input store.ChecksumIndexInput
	if input, err = dir.OpenChecksumInput(filename, ctx); err != nil {
		return
	}

	var success = false
	defer func() {
		if !success {
			util.CloseWhileSuppressingError(input)
		} else {
			err = input.Close()
		}
	}()

	var codecVersion int
	if codecVersion, err = asInt(codec.CheckHeader(input, SI_CODEC_NAME, SI_VERSION_START, SI_VERSION_CURRENT)); err != nil {
		return
	}
	var versionStr string
	if versionStr, err = input.ReadString(); err != nil {
		return
	}
	var version util.Version
	if version, err = util.ParseVersion(versionStr); err != nil {
		err = errors.New(fmt.Sprintf("unable to parse version string (resource=%v): %v", input, err))
		return
	}
	var docCount int
	if docCount, err = asInt(input.ReadInt()); err != nil {
		return
	} else if docCount < 0 {
		return nil, errors.New(fmt.Sprintf("invalid docCount: %v (resource=%v)", docCount, input))
	}
	var b byte
	if b, err = input.ReadByte(); err != nil {
		return
	}
	var isCompoundFile = b == YES
	var diagnostics map[string]string
	if diagnostics, err = input.ReadStringStringMap(); err != nil {
		return
	}
	var files map[string]bool
	if files, err = input.ReadStringSet(); err != nil {
		return
	}

	if codecVersion >= SI_VERSION_CHECKSUM {
		_, err = codec.CheckFooter(input)
	} else {
		err = codec.CheckEOF(input)
	}
	if err != nil {
		return
	}

	si = NewSegmentInfo(dir, version, segName, docCount, isCompoundFile, nil, diagnostics)
	si.SetFiles(files)

	success = true
	return si, nil
}

func asInt(n int32, err error) (int, error) {
	if err != nil {
		return 0, err
	}
	return int(n), nil
}

func (f *Lucene46SegmentInfoFormat) Write(dir store.Directory, si *SegmentInfo, fis FieldInfos, ctx store.IOContext) error {
	filename := util.SegmentFileName(si.Name, "", SI_EXTENSION)
	si.AddFile(filename)

	output, err := dir.CreateOutput(filename, ctx)
	if err != nil {
		return err
	}

	var success = false
	defer func() {
		if !success {
			util.CloseWhileSuppressingError(output)
			si.Dir.DeleteFile(filename) // ignore error
		} else {
			output.Close()
		}
	}()

	if err = codec.WriteHeader(output, SI_CODEC_NAME, SI_VERSION_CURRENT); err == nil {
		version := si.Version()
		assert2(version[0] == 3 || version[0] == 4,
			"invalid major version: should be 3 or 4 but got: %v", version[0])
		// write the Lucene version that created this segment, since 3.1
		if err = output.WriteString(version.String()); err == nil {
			if err = output.WriteInt(int32(si.DocCount())); err == nil {

				flag := NO
				if si.IsCompoundFile() {
					flag = YES
				}
				if err = output.WriteByte(byte(flag)); err == nil {
					if err = output.WriteStringStringMap(si.Diagnostics()); err == nil {
						if err = output.WriteStringSet(si.Files()); err == nil {
							if err = codec.WriteFooter(output); err == nil {
								success = true
							}
						}
					}
				}
			}
		}
	}
	if err != nil {
		return err
	}
	success = true
	return nil
}
