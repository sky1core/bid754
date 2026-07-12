package testgen

import "testing"

// TestModeDiscOperandEncoders pins the mode-discriminant component encoders
// against independently known canonical BID bit patterns (the +1/+0 encodings
// documented in the corpus comments and the emitted Rust ZERO/ONE constants),
// so a bias or field-width slip cannot silently emit operands that decode as
// different values than the discriminant table intends.
func TestModeDiscOperandEncoders(t *testing.T) {
	// +1 at each width.
	if bits, err := encodeModeDiscOperand32(mdo(1, 0)); err != nil || bits != 0x32800001 {
		t.Errorf("encode32(+1E0) = %#x, %v; want 0x32800001", bits, err)
	}
	if bits, err := encodeModeDiscOperand64(mdo(1, 0)); err != nil || bits != 0x31c0000000000001 {
		t.Errorf("encode64(+1E0) = %#x, %v; want 0x31c0000000000001", bits, err)
	}
	want128 := [16]byte{0x01, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x40, 0x30}
	if raw, err := encodeModeDiscOperand128(mdo(1, 0)); err != nil || raw != want128 {
		t.Errorf("encode128(+1E0) = %x, %v; want %x", raw, err, want128)
	}

	// Zero and sign.
	if bits, err := encodeModeDiscOperand32(mdo(0, 0)); err != nil || bits != 0x32800000 {
		t.Errorf("encode32(+0E0) = %#x, %v; want 0x32800000", bits, err)
	}
	if bits, err := encodeModeDiscOperand32(mdoNeg(1, 0)); err != nil || bits != 0xb2800001 {
		t.Errorf("encode32(-1E0) = %#x, %v; want 0xb2800001", bits, err)
	}
	if bits, err := encodeModeDiscOperand64(mdoNeg(1, 0)); err != nil || bits != 0xb1c0000000000001 {
		t.Errorf("encode64(-1E0) = %#x, %v; want 0xb1c0000000000001", bits, err)
	}

	// A negative exponent: 5E-7 at width 32 (exponent field 94 = 0x5e).
	if bits, err := encodeModeDiscOperand32(mdo(5, -7)); err != nil || bits != 0x2f000005 {
		t.Errorf("encode32(5E-7) = %#x, %v; want 0x2f000005", bits, err)
	}

	// Range rejections reject unresolved input.
	if _, err := encodeModeDiscOperand32(mdo(1<<23, 0)); err == nil {
		t.Error("encode32(coefficient 2^23) must be rejected")
	}
	if _, err := encodeModeDiscOperand64(mdo(1<<53, 0)); err == nil {
		t.Error("encode64(coefficient 2^53) must be rejected")
	}
	if _, err := encodeModeDiscOperand64(mdo(1, 400)); err == nil {
		t.Error("encode64(exponent 400) must be rejected (steering-bit region)")
	}
	if _, err := encodeModeDiscOperand128(modeDiscOperand{CoeffHi: 1 << 49, CoeffLo: 0, Exp: 0}); err == nil {
		t.Error("encode128(coefficient high word 2^49) must be rejected")
	}
}

// TestModeBinaryDiscriminantTableHygiene requires every registered binary
// mode-shape operation to carry at least modeDiscMinPairs encodable operand
// pairs at every width, so a table edit cannot shrink the discriminating
// corpus below usefulness or leave an operand that fails to encode until
// generation time.
func TestModeBinaryDiscriminantTableHygiene(t *testing.T) {
	for op := range modeBinaryDiscriminants {
		for _, width := range []int{32, 64, 128} {
			pairs, err := modeBinaryDiscriminantOperands(op, width)
			if err != nil {
				t.Errorf("op %s width %d: %v", op, width, err)
				continue
			}
			if len(pairs) < modeDiscMinPairs {
				t.Errorf("op %s width %d: %d pairs, want >= %d", op, width, len(pairs), modeDiscMinPairs)
			}
			for i, dp := range pairs {
				for j, operand := range dp {
					var err error
					switch width {
					case 32:
						_, err = encodeModeDiscOperand32(operand)
					case 64:
						_, err = encodeModeDiscOperand64(operand)
					default:
						_, err = encodeModeDiscOperand128(operand)
					}
					if err != nil {
						t.Errorf("op %s width %d pair %d operand %d: %v", op, width, i, j, err)
					}
				}
			}
		}
	}
}
