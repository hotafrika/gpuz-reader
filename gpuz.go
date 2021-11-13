package gpuz

import (
	"encoding/binary"
	"fmt"
	"github.com/hotafrika/shm"
	"math"
)

// SharedMemory is used for connecting to GPU-Zs shared memory
type SharedMemory struct {
	name string
	memory *shm.Memory
}

// NewSharedMemory opens shared memory by name
func NewSharedMemory(name string) *SharedMemory {
	return &SharedMemory{name: name}
}

// DefaultSharedMemory opens shared memory for GPU-Z by standard name
func DefaultSharedMemory() *SharedMemory {
	return NewSharedMemory("GPUZShMem")
}

// GetStat returns information in Stat
func (r *SharedMemory) GetStat() (Stat, error) {
	stat := Stat{
		Records:       make(map[string]string),
		SensorRecords: make(map[string]SensorRecord),
	}
	// size is fixed size of GPU-Z shared memory
	memory, err := shm.Open(r.name, 4+4+4+1024*128+624*128)
	if err != nil {
		return stat, err
	}
	r.memory = memory
	defer r.memory.Close()

	version, err := r.readInt32()
	if err != nil {
		return stat, err
	}
	stat.Version = version

	busy, err := r.readInt32()
	if err != nil {
		return stat, err
	}
	stat.Busy = busy
	lastUpdate, err := r.readInt32()
	if err != nil {
		return stat, err
	}
	stat.LastUpdate = lastUpdate

	for i := 0; i < 128; i++ {
		key, err := r.readString(512)
		if err != nil {
			return stat, err
		}
		value, err := r.readString(512)
		if err != nil {
			return stat, err
		}
		if key != "" {
			stat.Records[key] = value
			stat.AvailableRecords = append(stat.AvailableRecords, key)
		}
	}

	for i := 0; i < 128; i++ {
		name, err := r.readString(512)
		if err != nil {
			return stat, err
		}
		unit, err := r.readString(16)
		if err != nil {
			return stat, err
		}
		digits, err := r.readInt32()
		if err != nil {
			return stat, err
		}
		value, err := r.readFloat64()
		if err != nil {
			return stat, err
		}
		sr := SensorRecord{
			Name:   name,
			Unit:   unit,
			Digits: digits,
			Value:  value,
		}
		if sr.Name != "" {
			stat.SensorRecords[sr.Name] = sr
			stat.AvailableSensors = append(stat.AvailableSensors, sr.Name)
		}
	}

	if len(stat.Records) == 0 && len(stat.SensorRecords) == 0 {
		return stat, fmt.Errorf("initialization failed: empty data and sensors. Possibly GPU-Z is not running")
	}
	return stat, nil
}

func (r *SharedMemory) readString(size int) (string, error) {
	b := make([]byte, size)
	_, err := r.memory.Read(b)
	if err != nil {
		return "", err
	}
	utf := make([]byte, 0)
	for i := 0; i+(2-1) < len(b); i += 2 {
		if b[i] == 0 {
			break
		}
		utf = append(utf, b[i])
	}
	return string(utf), nil
}

func (r *SharedMemory) readInt32() (uint32, error) {
	b := make([]byte, 4)
	_, err := r.memory.Read(b)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(b), nil
}

func (r *SharedMemory) readFloat64() (float64, error) {
	b := make([]byte, 8)
	_, err := r.memory.Read(b)
	if err != nil {
		return 0, err
	}
	bits := binary.LittleEndian.Uint64(b)
	fl := math.Float64frombits(bits)
	return fl, err
}

// Stat represents data obtained from GPU-Z through shared memory
type Stat struct {
	Records       map[string]string       // size 512 + 512, count 128
	SensorRecords map[string]SensorRecord // size 624, count 128
	Version       uint32                  // size 4
	Busy          uint32                  // size 4
	LastUpdate    uint32                  // size 4
	AvailableRecords []string
	AvailableSensors []string
}

// SensorRecord contains information about sensor metrics
type SensorRecord struct { // size 624
	Name   string  // size 256 * 2
	Unit   string  // size 8 * 2
	Value  float64 // size 64
	Digits uint32  // size 32
}

// GetRecord returns characteristic by its name
// True if characteristic with this title exists
func (s Stat) GetRecord(title string) (string, bool) {
	val, ok := s.Records[title]
	return val, ok
}

// GetSensorValue returns sensor value by its name
// True if sensor with this title exists
func (s Stat) GetSensorValue(title string) (float64, bool) {
	sr, ok := s.GetSensor(title)
	if !ok {
		return 0, false
	}
	return sr.Value, true
}

// GetSensor returns SensorRecord by its name
// True if sensor with this title exists
func (s Stat) GetSensor(title string) (SensorRecord, bool) {
	sr, ok := s.SensorRecords[title]
	return sr, ok
}

// GetAvailableSensors returns all available sensors titles
func (s Stat) GetAvailableSensors() []string {
	return s.AvailableSensors
}

// GetAvailableRecords returns all available records
func (s Stat) GetAvailableRecords() []string {
	return s.AvailableRecords
}

