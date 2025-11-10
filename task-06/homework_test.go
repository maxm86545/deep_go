package main

import (
	"math"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

const (
	bitsX          = 32
	bitsY          = 32
	bitsZ          = 32
	bitsGold       = 31
	bitsMana       = 10
	bitsHealth     = 10
	bitsRespect    = 4
	bitsStrength   = 4
	bitsExperience = 4
	bitsLevel      = 4
	bitsBool       = 1
	bitsType       = 2
	bitsNameLen    = 6
)

const (
	offsetX          = 0
	offsetY          = offsetX + bitsX
	offsetZ          = offsetY + bitsY
	offsetGold       = offsetZ + bitsZ
	offsetMana       = offsetGold + bitsGold
	offsetHealth     = offsetMana + bitsMana
	offsetRespect    = offsetHealth + bitsHealth
	offsetStrength   = offsetRespect + bitsRespect
	offsetExperience = offsetStrength + bitsStrength
	offsetLevel      = offsetExperience + bitsExperience
	offsetHouse      = offsetLevel + bitsLevel
	offsetGun        = offsetHouse + bitsBool
	offsetFamily     = offsetGun + bitsBool
	offsetType       = offsetFamily + bitsBool
	offsetNameLength = offsetType + bitsType
	offsetEnd        = offsetNameLength + bitsNameLen
)

const (
	dataSizeName = 42
	dataSizeData = (offsetEnd + 7) / 8
)

const _ = uint8(1<<bitsNameLen - dataSizeName) // ensures bitsNameLen is sufficient for dataSizeName

type Option func(*GamePerson)

func WithName(name string) func(*GamePerson) {
	return func(p *GamePerson) {
		if len(name) > dataSizeName {
			name = name[:dataSizeName]
		}

		copy(p.name[:], name)
		p.setBits(offsetNameLength, bitsNameLen, uint(len(name)))
	}
}

func WithCoordinates(x, y, z int) func(*GamePerson) {
	return func(p *GamePerson) {
		p.setBits(offsetX, bitsX, uint(x))
		p.setBits(offsetY, bitsY, uint(y))
		p.setBits(offsetZ, bitsZ, uint(z))
	}
}

func WithGold(gold int) func(*GamePerson) {
	return func(p *GamePerson) {
		p.setBits(offsetGold, bitsGold, uint(gold))
	}
}

func WithMana(mana int) func(*GamePerson) {
	return func(p *GamePerson) {
		p.setBits(offsetMana, bitsMana, uint(mana))
	}
}

func WithHealth(health int) func(*GamePerson) {
	return func(p *GamePerson) {
		p.setBits(offsetHealth, bitsHealth, uint(health))
	}
}

func WithRespect(respect int) func(*GamePerson) {
	return func(p *GamePerson) {
		p.setBits(offsetRespect, bitsRespect, uint(respect))
	}
}

func WithStrength(strength int) func(*GamePerson) {
	return func(p *GamePerson) {
		p.setBits(offsetStrength, bitsStrength, uint(strength))
	}
}

func WithExperience(experience int) func(*GamePerson) {
	return func(p *GamePerson) {
		p.setBits(offsetExperience, bitsExperience, uint(experience))
	}
}

func WithLevel(level int) func(*GamePerson) {
	return func(p *GamePerson) {
		p.setBits(offsetLevel, bitsLevel, uint(level))
	}
}

func WithHouse() func(*GamePerson) {
	return func(p *GamePerson) {
		p.setBit(offsetHouse)
	}
}

func WithGun() func(*GamePerson) {
	return func(p *GamePerson) {
		p.setBit(offsetGun)
	}
}

func WithFamily() func(*GamePerson) {
	return func(p *GamePerson) {
		p.setBit(offsetFamily)
	}
}

func WithType(personType int) func(*GamePerson) {
	return func(p *GamePerson) {
		p.setBits(offsetType, bitsType, uint(personType))
	}
}

const (
	BuilderGamePersonType = iota
	BlacksmithGamePersonType
	WarriorGamePersonType
)

// GamePerson is not thread-safe.
type GamePerson struct {
	name [dataSizeName]byte
	data [dataSizeData]byte // Compact bit-shifted storage
}

func NewGamePerson(options ...Option) GamePerson {
	var p GamePerson
	for _, opt := range options {
		opt(&p)
	}

	return p
}

func (p *GamePerson) Name() string {
	length := int(p.getBits(offsetNameLength, bitsNameLen))
	if length > dataSizeName {
		length = dataSizeName
	}

	return string(p.name[:length])
}

func (p *GamePerson) X() int {
	return int(int32(p.getBits(offsetX, bitsX)))
}

func (p *GamePerson) Y() int {
	return int(int32(p.getBits(offsetY, bitsY)))
}

func (p *GamePerson) Z() int {
	return int(int32(p.getBits(offsetZ, bitsZ)))
}

func (p *GamePerson) Gold() int {
	return int(p.getBits(offsetGold, bitsGold))
}

func (p *GamePerson) Mana() int {
	return int(p.getBits(offsetMana, bitsMana))
}

func (p *GamePerson) Health() int {
	return int(p.getBits(offsetHealth, bitsHealth))
}

func (p *GamePerson) Respect() int {
	return int(p.getBits(offsetRespect, bitsRespect))
}

func (p *GamePerson) Strength() int {
	return int(p.getBits(offsetStrength, bitsStrength))
}

func (p *GamePerson) Experience() int {
	return int(p.getBits(offsetExperience, bitsExperience))
}

func (p *GamePerson) Level() int {
	return int(p.getBits(offsetLevel, bitsLevel))
}

func (p *GamePerson) HasHouse() bool {
	return p.getBit(offsetHouse)
}

func (p *GamePerson) HasGun() bool {
	return p.getBit(offsetGun)
}

func (p *GamePerson) HasFamily() bool {
	return p.getBit(offsetFamily)
}

func (p *GamePerson) Type() int {
	return int(p.getBits(offsetType, bitsType))
}

func (p *GamePerson) setBits(offset, bits int, value uint) {
	for i := 0; i < bits; i++ {
		if (value>>i)&1 != 0 {
			p.setBit(offset + i)
		} else {
			p.unsetBit(offset + i)
		}
	}
}

func (p *GamePerson) getBits(offset, bits int) uint {
	var value uint
	for i := 0; i < bits; i++ {
		if p.getBit(offset + i) {
			value |= 1 << i
		}
	}

	return value
}

func (p *GamePerson) setBit(offset int) {
	p.data[offset/8] |= 1 << (offset % 8)
}

func (p *GamePerson) unsetBit(offset int) {
	p.data[offset/8] &^= 1 << (offset % 8)
}

func (p *GamePerson) getBit(offset int) bool {
	return (p.data[offset/8]>>(offset%8))&1 != 0
}

func TestGamePerson(t *testing.T) {
	assert.LessOrEqual(t, unsafe.Sizeof(GamePerson{}), uintptr(64))

	const x, y, z = math.MinInt32, math.MaxInt32, 0
	const name = "aaaaaaaaaaaaa_bbbbbbbbbbbbb_cccccccccccccc"
	const personType = BuilderGamePersonType
	const gold = math.MaxInt32
	const mana = 1000
	const health = 1000
	const respect = 10
	const strength = 10
	const experience = 10
	const level = 10

	options := []Option{
		WithName(name),
		WithCoordinates(x, y, z),
		WithGold(gold),
		WithMana(mana),
		WithHealth(health),
		WithRespect(respect),
		WithStrength(strength),
		WithExperience(experience),
		WithLevel(level),
		WithHouse(),
		WithFamily(),
		WithType(personType),
	}

	person := NewGamePerson(options...)
	assert.Equal(t, name, person.Name())
	assert.Equal(t, x, person.X())
	assert.Equal(t, y, person.Y())
	assert.Equal(t, z, person.Z())
	assert.Equal(t, gold, person.Gold())
	assert.Equal(t, mana, person.Mana())
	assert.Equal(t, health, person.Health())
	assert.Equal(t, respect, person.Respect())
	assert.Equal(t, strength, person.Strength())
	assert.Equal(t, experience, person.Experience())
	assert.Equal(t, level, person.Level())
	assert.True(t, person.HasHouse())
	assert.True(t, person.HasFamily())
	assert.False(t, person.HasGun())
	assert.Equal(t, personType, person.Type())
}

// go test -v homework_test.go -fuzz=FuzzGamePerson
func FuzzGamePerson(f *testing.F) {

	f.Add(
		int32(-2_000_000_000), int32(0), int32(2_000_000_000),
		uint32(0),
		uint16(0), uint16(1000),
		uint8(0), uint8(10), uint8(0), uint8(10),
		false, false, false,
		uint8(BlacksmithGamePersonType),
		"user1",
	)

	f.Add(
		int32(2_000_000_000), int32(2_000_000_000), int32(2_000_000_000),
		uint32(2_000_000_000),
		uint16(1000), uint16(1000),
		uint8(10), uint8(10), uint8(10), uint8(10),
		true, true, true,
		uint8(WarriorGamePersonType),
		"user2",
	)

	f.Fuzz(func(t *testing.T,
		x int32, y int32, z int32,
		gold uint32,
		mana uint16, health uint16,
		respect uint8, strength uint8, experience uint8, level uint8,
		house bool, gun bool, family bool,
		personType uint8,
		name string,
	) {
		if x < -2_000_000_000 || x > 2_000_000_000 {
			return
		}
		if y < -2_000_000_000 || y > 2_000_000_000 {
			return
		}
		if z < -2_000_000_000 || z > 2_000_000_000 {
			return
		}
		if gold > 2_000_000_000 {
			return
		}
		if mana > 1000 || health > 1000 {
			return
		}
		if respect > 10 || strength > 10 || experience > 10 || level > 10 {
			return
		}
		if personType > 2 {
			return
		}
		if len(name) > dataSizeName {
			name = name[:dataSizeName]
		}

		person := NewGamePerson(
			WithName(name),
			WithCoordinates(int(x), int(y), int(z)),
			WithGold(int(gold)),
			WithMana(int(mana)),
			WithHealth(int(health)),
			WithRespect(int(respect)),
			WithStrength(int(strength)),
			WithExperience(int(experience)),
			WithLevel(int(level)),
			func(p *GamePerson) {
				if house {
					WithHouse()(p)
				}
				if gun {
					WithGun()(p)
				}
				if family {
					WithFamily()(p)
				}
			},
			WithType(int(personType)),
		)

		assert.Equal(t, name, person.Name())
		assert.Equal(t, int(x), person.X())
		assert.Equal(t, int(y), person.Y())
		assert.Equal(t, int(z), person.Z())
		assert.Equal(t, int(gold), person.Gold())
		assert.Equal(t, int(mana), person.Mana())
		assert.Equal(t, int(health), person.Health())
		assert.Equal(t, int(respect), person.Respect())
		assert.Equal(t, int(strength), person.Strength())
		assert.Equal(t, int(experience), person.Experience())
		assert.Equal(t, int(level), person.Level())
		assert.Equal(t, house, person.HasHouse())
		assert.Equal(t, gun, person.HasGun())
		assert.Equal(t, family, person.HasFamily())
		assert.Equal(t, int(personType), person.Type())
	})
}
