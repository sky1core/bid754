package bidgo

// bid64_to_string 관련 상수 및 테이블
// Intel bid128_2_str_tables.c에서 기계적 포팅

const (
	bid_Twoto60_m_10to18 = 152921504606846976
	bid_Twoto60          = 0x1000000000000000
	bid_Inv_Tento9       = 2305843009 // floor(2^61/10^9)
	bid_Twoto30_m_10to9  = 73741824
	bid_Tento9           = 1000000000
	bid_Tento6           = 1000000
	bid_Tento3           = 1000
)
