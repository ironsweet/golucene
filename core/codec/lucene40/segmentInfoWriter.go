package lucene40

import (
	"github.com/balzaczyy/golucene/core/codec"
	. "github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
)

const (
	SEGMENT_INFO_NO  = -1
	SEGMENT_INFO_YES = 1
)

// lucene40/Lucne40SegmentInfoWriter.java

// Lucene 4.0 implementation of SegmentInfoWriter

type Lucene40SegmentInfoWriter struct{}

func (w *Lucene40SegmentInfoWriter) Write(dir store.Directory,
	si *SegmentInfo, fis FieldInfos, ctx store.IOContext) (err error) {

	filename := util.SegmentFileName(si.Name, "", LUCENE40_SI_EXTENSION)
	si.AddFile(filename)

	var output store.IndexOutput
	output, err = dir.CreateOutput(filename, ctx)
	if err != nil {
		return err
	}

	var success = false
	defer func() {
		if !success {
			util.CloseWhileSuppressingError(output)
			si.Dir.DeleteFile(filename) // ignore error
		} else {
			err = mergeError(err, output.Close())
		}
	}()

	err = codec.WriteHeader(output, LUCENE40_CODEC_NAME, LUCENE40_VERSION_CURRENT)
	if err != nil {
		return err
	}
	// Write the Lucene version that created this segment, since 3.1
	err = store.Stream(output).WriteString(si.Version().String()).
		WriteInt(int32(si.DocCount())).
		WriteByte(func() byte {
		if si.IsCompoundFile() {
			return SEGMENT_INFO_YES
		}
		return byte((SEGMENT_INFO_NO + 256) % 256) // Go byte is non-negative, unlike Java
	}()).WriteStringStringMap(si.Diagnostics()).
		WriteStringStringMap(map[string]string{}).
		WriteStringSet(si.Files()).Close()
	if err != nil {
		return err
	}

	success = true
	return nil
}
