package lucene46

import (
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
	SI_VERSION_CHECKSUM = 1
	SI_VERSION_CURRENT  = SI_VERSION_CHECKSUM
)

type Lucene46SegmentInfoFormat struct {
}

func NewLucene46SegmentInfoFormat() *Lucene46SegmentInfoFormat {
	return &Lucene46SegmentInfoFormat{}
}

func (f *Lucene46SegmentInfoFormat) SegmentInfoReader() SegmentInfoReader {
	panic("not implemented yet")
}

func (f *Lucene46SegmentInfoFormat) SegmentInfoWriter() SegmentInfoWriter {
	return f
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
			_ = si.Dir.DeleteFile(filename) // ignore error
		} else {
			output.Close()
		}
	}()

	if err = codec.WriteHeader(output, SI_CODEC_NAME, SI_VERSION_CURRENT); err == nil {
		// write the Lucene version that created this segment, since 3.1
		if err = output.WriteString(si.Version()); err == nil {
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
