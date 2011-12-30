package main

import (
	"io"
	"bufio"
	"fmt"
	"bytes"
	"os"
)

/*
bytes	low	high
1	2^0	2^7 -1
2	2^7	2^14 -1
3	2^14	2^21 -1
4	2^21	2^28 -1
5	2^28	2^35 -1
6	2^35	2^42 -1
7	2^42	2^49 -1
8	2^49	2^56 -1
9	2^56	2^63 -1
10	2^63	2^70 -1
*/
func WriteIntAsVInt(x uint64, w io.Writer) {
	var bytes [10]byte
	pos := 0
	for x != 0 {
		bytes[pos] = byte(x) & mask
		x = x >> 7
		if x != 0 {
			bytes[pos] |= flag
		}
		//fmt.Printf("%08b ", bytes[pos])
		pos++

	}
	//fmt.Println()
	if pos == 0 {
		pos = 1
	}
	w.Write(bytes[:pos])
}

const (
	mask = byte(127)
	flag = byte(128)
	zero = byte(0)
)

func ReadVIntAsInt(r io.Reader) uint64 {
	rr := makeByteReader(r)
	var out uint64
	offset := uint(0)
	for b, err := rr.ReadByte(); err == nil; b, err = rr.ReadByte() {
		//fmt.Printf("%08b ", b)
		out = out | (uint64(b&mask) << offset)
		if b&flag == zero {
			break
		}
		offset += 7
	}
	//fmt.Println()
	return out
}

type myReader interface {
	io.Reader
	io.ByteReader
}

func makeByteReader(r io.Reader) myReader {
	var rr myReader
	var ok bool
	if rr, ok = r.(myReader); !ok {
		rr = bufio.NewReader(r)
	}
	return rr
}

func WriteString(s string, w io.Writer) {
	WriteIntAsVInt(uint64(len(s)), w)
	bytes := []byte(s)
	w.Write(bytes)
}

func ReadString(r io.Reader) string {
	rr := makeByteReader(r)
	length := uint(ReadVIntAsInt(rr))
	//read the next length bytes
	bytes := make([]byte, length)
	for i := 0; i < len(bytes); i++ {
		var err os.Error
		bytes[i], err = rr.ReadByte()
		if err != nil {
			panic("Unexpected End of Buffer")
		}
	}
	return string(bytes)
}

func WriteMap(m map[string]string, w io.Writer) {
	WriteIntAsVInt(uint64(len(m)), w)
	for k, v := range m {
		WriteString(k, w)
		WriteString(v, w)
	}
}

func ReadMap(r io.Reader) map[string]string {
	rr := makeByteReader(r)
	outMap := make(map[string]string)
	length := int(ReadVIntAsInt(rr))
	for i := 0; i < length; i++ {
		k := ReadString(rr)
		v := ReadString(rr)
		outMap[k] = v
	}
	return outMap
}

func main() {
	max := uint64(1<<64 - 1)
	fmt.Println(os.Args)
	if len(os.Args) > 1 {
		fmt.Println("reading max from args")
		count, err := fmt.Sscanf(os.Args[1], "%d", &max)
		if err != nil {
			panic(err)
		}
		fmt.Println("read ", count, ": ", max)
	}
	data := make([]byte, 100)
	buf := bytes.NewBuffer(data)
	fmt.Println("count to ", max)
	for i := uint64(0); i < max; i++ {
		buf.Reset()
		//fmt.Println("writing ", i)
		WriteIntAsVInt(i, buf)
		fmt.Println(buf.Bytes())
		j := ReadVIntAsInt(buf)
		//fmt.Println("read ", j)
		if i != j {
			panic(fmt.Sprintf("%d != %d", i, j))
		}
	}
	//for good measure, check the max value, make sure it works
	buf.Reset()
	WriteIntAsVInt(uint64(1<<64-1), buf)
	fmt.Println(buf.Bytes())
	j := ReadVIntAsInt(buf)
	if j != 1<<64-1 {
		panic(fmt.Sprintf("failed to write max uint64:", j))
	}
	buf.Reset()
	WriteString("This is a test", buf)
	fmt.Println(buf.Bytes())
	result := ReadString(buf)
	fmt.Println(result)
	buf.Reset()
	WriteString("", buf)
	fmt.Println(buf.Bytes())
	result = ReadString(buf)
	fmt.Println(result)
	m := make(map[string]string)
	m["abc"]="def"
	m["123"]="456"
	buf.Reset()
	WriteMap(m,buf)
	fmt.Println(buf.Bytes())
	om := ReadMap(buf)
	fmt.Println(om)
	m2 := make(map[string]string)
	buf.Reset()
	WriteMap(m2,buf)
	fmt.Println(buf.Bytes())
	om2 := ReadMap(buf)
	fmt.Println(om2)
	fmt.Println("done")
}
