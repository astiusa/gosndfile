package sndfile

import "os"
import "testing"
import "fmt"
import "strings"
import "unsafe"

func TestGetLibVersion(t *testing.T) {
	s, _ := GetLibVersion()
	fmt.Println(s)
	if !strings.HasPrefix(s, "libsndfile") {
		t.Errorf("version string \"%s\" had unexpected prefix", s)
	}
}

func TestGetLogInfo(t *testing.T) {
	var i Info
	f, err := Open("fwerefrg", Read, &i)
//	fmt.Println(f)
//	fmt.Println(err)
	s, err := f.GetLogInfo()
	fmt.Println("TestGetLogInfo output: ", s)
	if err != nil {
		t.Error("TestGetLogInfo err: ", err)
	}
}

func TestFileCommands(t *testing.T) {
	var i Info
	f, err := Open("funky.aiff", ReadWrite, &i)
	if err != nil {
		t.Fatalf("open file failed %s", err)
	}
	
	max, err := f.CalcSignalMax()
	if err != nil {
		t.Fatalf("signal max failed %s", err)
	}
	
	fmt.Printf("max signal %f\n", max)

	max, err = f.CalcNormSignalMax()
	if err != nil {
		t.Fatalf("norm signal max failed %s", err)
	}
	
	fmt.Printf("norm max signal %f\n", max)

	maxarr, err := f.CalcMaxAllChannels()
	if err != nil {
		t.Fatalf("max all chans failed %s", err)
	}
	
	fmt.Printf("max all chans signal %v\n", maxarr)
	
	maxarr, err = f.CalcNormMaxAllChannels()
	if err != nil {
		t.Fatalf("max all chans failed %s", err)
	}
	
	fmt.Printf("norm max all chans signal %v\n", maxarr)


	max, ok := f.GetSignalMax()
	if !ok {
		t.Error("got unexpected failure from GetSignalMax with val ", max)
	}

	maxarr, ok = f.GetMaxAllChannels()
	if !ok {
		t.Error("got unexpected failure from GetMaxAllChannels with vals", maxarr)
	}
	
	f.Close()
	
}

func TestFormats(t *testing.T) {
	simpleformats := GetSimpleFormatCount()
	fmt.Println("--- Supported simple formats")
	for i := 0; i < simpleformats; i++ {
		f, name, ext, ok := GetSimpleFormat(i)
		fmt.Printf("%08x %s %s\n", f, name, ext)
		if !ok {
			t.Error("error from GetSimpleFormat()")
		}
	}
	
	fmt.Println("--- Supported formats")
	// following is straight from examples in libsndfile distribution
	majorcount := GetMajorFormatCount()
	subcount := GetSubFormatCount()
	for m := 0; m < majorcount; m++ {
		f, name, ext, ok := GetMajorFormatInfo(m)
		if ok {
			fmt.Printf("--- MAJOR 0x%08x %v Extension: .%v\n", f, name, ext)
			for s := 0; s < subcount; s++ {
				t, sname, sok := GetSubFormatInfo(s)
				var i Info
				i.Channels = 1
				i.Format = Format(f|t)
				if sok && FormatCheck(i) {
					fmt.Printf("   0x%08x %v %v\n", f|t, name, sname)
				} else {
//					fmt.Printf("no format pair 0x%x\n", f|t)
				}
			}
		} else {
			fmt.Printf("no format for number %v\n", m)
		}
	}
}

func isLittleEndian() bool {
	var i int32 = 0x01020304
	u := unsafe.Pointer(&i)
	pb := (*byte)(u)
	b := *pb
	return (b == 0x04)
}

func TestRawSwap(t *testing.T) {
	// set up file to be checked
	i := &Info{0, 44100, 1, SF_FORMAT_WAV|SF_FORMAT_PCM_16,0,0}
	f, err := Open("leout.wav", Write, i)
	if err != nil {
		t.Fatalf("couldn't open file for writing: %v", err)
	}
	if isLittleEndian() && f.RawNeedsEndianSwap() {
		t.Errorf("little endian file and little endian cpu shuld not report needing swap but does!")
	} else if !isLittleEndian() && !f.RawNeedsEndianSwap() {
		t.Errorf("little endian file and big endian machine should report needing swap, but doesn't!")
	}
	f.Close()
}

func TestGenericCmd(t *testing.T) {
	i := GenericCmd(nil, 0x1000, nil, 0)
	c := make([]byte, i)
	GenericCmd(nil, 0x1000, unsafe.Pointer(&c[0]), i)
	if !strings.HasPrefix(string(c), "libsndfile") {
		t.Errorf("version string \"%s\" had unexpected prefix", string(c))
	}	
}

func TestTruncate(t *testing.T) {
	// first write 100 samples to a file
	var i Info
	i.Samplerate = 44100
	i.Channels = 1
	i.Format = SF_FORMAT_AIFF|SF_FORMAT_PCM_24
	os.Remove("truncout.aiff")
	f, err := Open("truncout.aiff", ReadWrite, &i)
	if err != nil {
		t.Fatalf("couldn't open file for output! %v", err)
	}
	
	var junk [100]int32
	written, err := f.WriteItems(junk[0:100])
	if written != 100 {
		t.Errorf("wrong written count %d", written)
	}
	
	f.WriteSync()
	
	f.Truncate(20)

	f.WriteSync()
	
	seek, err := f.Seek(0, Current)
	if seek != 20 {
		t.Errorf("wrong seek %v", seek)
	}
	if err != nil {
		t.Errorf("error! %v", err)
	}
	f.Close()
}

func TestMax(t *testing.T) {
	// open file with no peak chunk
	var i Info
	i.Samplerate = 44100
	i.Channels = 4
	i.Format = SF_FORMAT_AIFF|SF_FORMAT_PCM_24
	
	f, err := Open("addpeakchunk1.aiff", Write, &i)
	if err != nil {
		t.Fatalf("couldn't open file %v", err)
	}
	
	f.SetAddPeakChunk(false)
	err = f.WriteItems([]int32{1,2,1,2,-1,-2,-1,-2,2,4,2,4,-2,-4,-2,-4})
	if err != nil {
		t.Error("write err:",err)
	}
	f.Close()

	f, err = Open("addpeakchunk1.aiff", Read, &i)
	if err != nil {
		t.Fatalf("couldn't open file %v", err)
	}
	
	// calc signals
	// make sure peak chunk returns false
	f.Close()
	
	// repeat for peak chunk, making sure peak chunk returns same value
}