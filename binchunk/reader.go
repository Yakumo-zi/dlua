package binchunk

import (
	"encoding/binary"
	"math"
)

type reader struct {
	data []byte
}

func (r *reader) readByte() byte {
	b := r.data[0]
	r.data = r.data[1:]
	return b
}

func (r *reader) readUint32() uint32 {
	i := binary.LittleEndian.Uint32(r.data)
	r.data = r.data[4:]
	return i
}

func (r *reader) readUint64() uint64 {
	i := binary.LittleEndian.Uint64(r.data)
	r.data = r.data[8:]
	return i
}

func (r *reader) readLuaInterger() int64 {
	return int64(r.readUint64())
}

func (r *reader) readLuaNumber() float64 {
	return math.Float64frombits(r.readUint64())
}

func (r *reader) readBytes(nums uint) []byte {
	bytes := r.data[:nums]
	r.data = r.data[nums:]
	return bytes
}

func (r *reader) readString() string {
	size := uint(r.readByte())
	if size == 0 {
		return ""
	}
	if size == 0xFF {
		size = uint(r.readUint64())
	}
	bytes := r.readBytes(size - 1)
	return string(bytes)
}

func (r *reader) checkHeader() {
	if string(r.readBytes(4)) != LUA_SIGNATURE {
		panic("not a precompiled chunk!")
	} else if r.readByte() != LUAC_VERSION {
		panic("version mismatch!")
	} else if r.readByte() != LUAC_FORMAT {
		panic("format mismatch!")
	} else if string(r.readBytes(6)) != LUAC_DATA {
		panic("corrupted!")
	} else if r.readByte() != CINT_SIZE {
		panic("int size mismatch!")
	} else if r.readByte() != CSIZET_SIZE {
		panic("size_t size mismatch!")
	} else if r.readByte() != INSTRUCTION_SIZE {
		panic("instruction size mismatch!")
	} else if r.readByte() != LUA_INTEGER_SIZE {
		panic("lua_Integer size mismatch!")
	} else if r.readByte() != LUA_NUMBER_SIZE {
		panic("lua_Number size mismatch!")
	} else if r.readLuaInterger() != LUAC_INT {
		panic("endianness mismatch!")
	} else if r.readLuaNumber() != LUAC_NUM {
		panic("float format mismatch!")
	}
}
func (r *reader) readCode() []uint32 {
	code := make([]uint32, r.readUint32())
	for i := range code {
		code[i] = r.readUint32()
	}
	return code
}

func (r *reader) readConstants() []interface{} {
	constants := make([]interface{}, r.readUint32())
	for i := range constants {
		switch r.readByte() {
		case TAG_NIL:
			constants[i] = nil
		case TAG_BOOLEAN:
			constants[i] = r.readByte() != 0
		case TAG_NUMBER:
			constants[i] = r.readLuaNumber()
		case TAG_INTEGER:
			constants[i] = r.readLuaInterger()
		case TAG_SHORT_STR, TAG_LONG_STR:
			constants[i] = r.readString()
		default:
			panic("corrupted!")
		}
	}
	return constants
}

func (r *reader) readUpvalues() []Upvalue {
	upvalues := make([]Upvalue, r.readUint32())
	for i := range upvalues {
		upvalues[i] = Upvalue{
			Instack: r.readByte(),
			Idx:     r.readByte(),
		}
	}
	return upvalues
}

func (r *reader) readProtos(parentSource string) []*Prototype {
	protos := make([]*Prototype, r.readUint32())
	for i := range protos {
		protos[i] = r.readProto(parentSource)
	}
	return protos
}

func (r *reader) readLineInfo() []uint32 {
	lineInfo := make([]uint32, r.readUint32())
	for i := range lineInfo {
		lineInfo[i] = r.readUint32()
	}
	return lineInfo
}

func (r *reader) readLocVars() []LocVar {
	locVars := make([]LocVar, r.readUint32())
	for i := range locVars {
		locVars[i] = LocVar{
			VarName: r.readString(),
			StartPC: r.readUint32(),
			EndPC:   r.readUint32(),
		}
	}
	return locVars
}

func (r *reader) readUpvalueNames() []string {
	upvalueNames := make([]string, r.readUint32())
	for i := range upvalueNames {
		upvalueNames[i] = r.readString()
	}
	return upvalueNames
}

func (r *reader) readProto(parentSource string) *Prototype {
	source := r.readString()
	if source == "" {
		source = parentSource
	}
	return &Prototype{
		Source:          source,
		LineDefined:     r.readUint32(),
		LastLineDefined: r.readUint32(),
		NumParams:       r.readByte(),
		IsVararg:        r.readByte(),
		MaxStackSize:    r.readByte(),
		Code:            r.readCode(),
		Constants:       r.readConstants(),
		Upvalues:        r.readUpvalues(),
		Protos:          r.readProtos(source),
		LineInfo:        r.readLineInfo(),
		LocVars:         r.readLocVars(),
		UpvalueNames:    r.readUpvalueNames(),
	}
}
