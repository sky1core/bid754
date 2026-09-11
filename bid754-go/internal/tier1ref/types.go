package tier1ref

import "github.com/sky1core/bid754/bid754-go/internal/decimalref"

type Case struct {
	Width    int      `json:"width"`
	Op       string   `json:"op"`
	Mode     string   `json:"mode"`
	Operands []string `json:"operands"`
	Param    string   `json:"param"`
	Target   int      `json:"target"`
}

type Observation struct {
	Path     string `json:"path"`
	Kind     string `json:"kind"`
	Width    int    `json:"width"`
	Value    string `json:"value"`
	Flags    uint32 `json:"flags"`
	HasFlags bool   `json:"has_flags"`
}

type Result struct {
	Kind        string             `json:"kind"`
	Width       int                `json:"width"`
	Decimal     decimalref.Decimal `json:"decimal"`
	Value       string             `json:"value"`
	Flags       uint32             `json:"flags"`
	Quantum     bool               `json:"quantum"`
	AnyZeroSign bool               `json:"any_zero_sign"`
}

const (
	Invalid   uint32 = 0x01
	Overflow  uint32 = 0x08
	Underflow uint32 = 0x10
	Inexact   uint32 = 0x20
)
