# BID Codec Cross-Language Test Vectors

## 사용법

### 벡터 생성
```bash
make generate-testspec
```

`vectors.json`은 `testgen_manifest.json`의 `bid_codec_vectors` 설정을 통해 생성된다. 이 파일은 직접 편집하지 않는다.

### 각 언어에서 검증
`vectors.json`을 읽어서:
1. `hex` → decode → Components와 벡터의 `sign/coefficient/exponent/kind` 비교
2. decode 결과 → `toString` → `decimal_string`과 비교
3. `decimal_string` → `fromString` → encode → `encoded_hex`/`encoded_hi`와 비교
4. canonical인 경우: Components → encode → `encoded_hex`와 비교 (bit roundtrip)
5. invalid byte length 와 malformed string 이 언어별 error mechanism 으로 실패하는지 확인

## 벡터 형식

```json
{
  "type": "bid64",
  "hex": "31c0000000000001",
  "sign": false,
  "coefficient": "1",
  "exponent": 0,
  "kind": "normal",
  "decimal_string": "+1E+0",
  "canonical": true,
  "encoded_hex": "31c0000000000001"
}
```

| 필드 | 설명 |
|------|------|
| type | "bid32", "bid64", "bid128" |
| hex | 입력 BID 값 (hex string, little-endian word order for bid128) |
| hex_hi | bid128 high word (bid128만) |
| sign | true=음수 |
| coefficient | 유효숫자 (decimal string, 특수값은 "") |
| exponent | 10의 지수 |
| kind | "normal", "zero", "inf", "qnan", "snan" |
| payload | NaN payload (있는 경우) |
| decimal_string | 공통 BID codec 문자열 표현. 모든 언어가 `toString`/`fromString`으로 검증한다 |
| canonical | true면 roundtrip 검증 가능 |
| encoded_hex | encode 결과 (canonical이면 hex와 동일) |
| encoded_hi | bid128 encode 결과 high word |

언어별 consumer 는 같은 row 로 bit decode/encode, string render/parse, 그리고
little-endian bytes decode/encode 를 검증한다. BID32/64/128 bytes decode
입력은 각각 정확히 4/8/16바이트여야 한다.

언어별 consumer 는 공통 failure semantics 도 검증한다. empty input, malformed
NaN payload, malformed exponent, multiple decimal points, signed 32-bit 범위
밖 exponent 는 `fromString` 실패여야 한다. 동적 byte buffer 를 받는 언어의
bytes decode API 는 invalid length 를 실패로 보고해야 한다. Rust fixed-array
bytes API 는 타입으로 길이를 강제하며, dynamic slice failure check 는
`try_decode*_bytes` 로 검증한다.

`Encode*` vector 검증은 canonical vectors 와 generated `decimal_string` 에서
파생된 trusted `Components` packing 계약을 검증한다. 이 검증은 out-of-range
또는 malformed `Components` 를 reject 하는 validation suite 가 아니다.

## Payload Scope

현재 standalone BID codec `Components.payload` 는 64-bit payload 필드다.
BID32/BID64 NaN payload 는 이 필드로 완전히 표현되지만, BID128 의 110-bit
NaN payload 중 high payload bits 는 current helper-package schema 에 포함하지
않는다.

따라서 generated BID128 high-payload NaN vector 는 noncanonical/decode-only 로
남아야 하며, full BID128 NaN payload support 로 보고하지 않는다. 이 범위를
넓히려면 Go, Rust, Java, Python, JavaScript/TypeScript, Swift 의 public
`Components` schema 와 vector consumer 를 같은 변경으로 확장해야 한다.

## 진실의 원천

`cmd/testgen`은 `internal/testgen`의 독립 BID bit-layout reference codec과
`testgen_manifest.json`의 seed/count 설정으로 벡터를 생성한다. 프로덕션
`bidcodec` 패키지를 oracle 로 import 하지 않는다.

같은 generator 가 Go, Rust, Java, Python, JavaScript/TypeScript, Swift
consumer 하네스를 생성한다. `make verify-generated` 는 벡터와 consumer
하네스를 함께 재생성 비교한다.
