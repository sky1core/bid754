package testgen

import (
	"fmt"
	"go/format"
	"strconv"
	"strings"

	"github.com/sky1core/bid754/devtools/internal/genmarker"
	"github.com/sky1core/bid754/devtools/tools/go2rs/apiemit"
)

const (
	tier1PathsGoPath   = "../bid754-go/generated_tier1_public_paths_test.go"
	tier1PathsRustPath = "../bid754-rs/examples/tier1_probe.rs"
)

type tier1PathRow struct {
	width, target, resultWidth, arity    int
	op, kind, goMethod, rustMethod, port string
	mode, flags, defaultPath             bool
}

func tier1PathRows() []tier1PathRow {
	var rows []tier1PathRow
	for _, w := range []int{32, 64, 128} {
		prefix := fmt.Sprintf("Bid%d", w)
		for _, op := range []struct{ name, goName, rustName string }{
			{"rem", "Remainder", "remainder"}, {"fmod", "Fmod", "fmod"},
			{"minnum", "MinNum", "min_num"}, {"maxnum", "MaxNum", "max_num"},
			{"quiet_equal", "QuietEqual", "quiet_eq"}, {"quiet_not_equal", "QuietNotEqual", "quiet_ne"},
			{"quiet_greater", "QuietGreater", "quiet_gt"}, {"quiet_greater_equal", "QuietGreaterEqual", "quiet_ge"},
			{"quiet_greater_unordered", "QuietGreaterUnordered", "quiet_gt_unordered"},
			{"quiet_less", "QuietLess", "quiet_lt"}, {"quiet_less_equal", "QuietLessEqual", "quiet_le"},
			{"quiet_less_unordered", "QuietLessUnordered", "quiet_lt_unordered"},
			{"quiet_not_greater", "QuietNotGreater", "quiet_not_gt"}, {"quiet_not_less", "QuietNotLess", "quiet_not_lt"},
			{"quiet_ordered", "QuietOrdered", "quiet_ordered"}, {"quiet_unordered", "QuietUnordered", "quiet_unordered"},
		} {
			r := tier1PathRow{width: w, resultWidth: w, arity: 2, op: op.name, kind: "decimal", goMethod: op.goName, rustMethod: op.rustName, port: prefix + op.goName, flags: true}
			switch op.name {
			case "rem":
				r.port = prefix + "Rem"
			case "minnum", "maxnum":
				if w == 32 {
					r.port += "WithFlags"
				}
				if w == 128 {
					r.port = prefix + strings.ToUpper(op.name[:1]) + op.name[1:]
				}
			}
			if strings.HasPrefix(op.name, "quiet_") {
				r.kind = "bool"
				r.resultWidth = 0
			}
			rows = append(rows, r)
		}
		scalePort := prefix + "Scalbln"
		if w == 32 {
			scalePort += "WithFlags"
		}
		rows = append(rows, tier1PathRow{width: w, resultWidth: w, arity: 1, op: "scaleb", kind: "decimal", goMethod: "ScaleBWithMode", rustMethod: "scaleb_with_mode", port: scalePort, mode: true, flags: true, defaultPath: true})
		for _, target := range []int{32, 64, 128} {
			if target == w {
				continue
			}
			rows = append(rows, tier1PathRow{width: w, target: target, resultWidth: target, arity: 1, op: "convert", kind: "decimal", goMethod: fmt.Sprintf("ToDecimal%d", target), rustMethod: fmt.Sprintf("to_decimal%d", target), port: fmt.Sprintf("%sToBid%d", prefix, target), mode: target < w, flags: true})
		}
		for _, signed := range []bool{true, false} {
			gtype, rtype, op := "Int", "i", "int"
			if !signed {
				gtype, rtype, op = "Uint", "u", "uint"
			}
			for _, target := range []int{32, 64} {
				mode := w == 32 || w == 64 && target == 64
				rows = append(rows, tier1PathRow{width: w, target: target, resultWidth: w, op: "from_" + op, kind: "decimal", goMethod: fmt.Sprintf("NewDecimal%dFrom%s%d", w, gtype, target), rustMethod: fmt.Sprintf("from_%s%d", rtype, target), port: fmt.Sprintf("%sFrom%s%d", prefix, gtype, target), mode: mode, flags: mode})
			}
			for _, exact := range []bool{false, true} {
				suffix, rsuffix, opsuffix := "", "", ""
				if exact {
					suffix, rsuffix, opsuffix = "Exact", "_exact", "_exact"
				}
				for _, target := range []int{8, 16, 32, 64} {
					rows = append(rows, tier1PathRow{width: w, target: target, resultWidth: target, arity: 1, op: "to_" + op + opsuffix, kind: "integer", goMethod: fmt.Sprintf("ConvertTo%s%d%s", gtype, target, suffix), rustMethod: fmt.Sprintf("to_%s%d%s", rtype, target, rsuffix), port: fmt.Sprintf("%sTo%s%d", prefix, gtype, target), mode: true, flags: true})
				}
			}
		}
	}
	return rows
}

func GenerateTier1PathsOutputs() (map[string][]byte, error) {
	var g, r strings.Builder
	g.WriteString(genmarker.Line("testgen") + "\n\n" + tier1PathsGoSupport)
	r.WriteString(genmarker.Line("testgen") + "\n\n" + tier1PathsRustSupport)
	rows := tier1PathRows()
	g.WriteString("func tier1PublicPaths(c tier1ref.Case) ([]tier1ref.Observation,error) {\nif err:=tier1ref.Validate(c); err!=nil {return nil,err}\nmode,ok:=finitePublicMode(c.Mode)\nif !ok {return nil,fmt.Errorf(\"unsupported public mode %q\",c.Mode)}\nrnd,ok:=finitePortRounding(c.Mode)\nif !ok {return nil,fmt.Errorf(\"unsupported port mode %q\",c.Mode)}\nswitch {\n")
	r.WriteString("fn observe(c: &Case) -> Result<Vec<Observation>,String> {\nvalidate(c)?;\nlet mode=public_mode(&c.mode).ok_or(\"invalid mode\")?;\nlet rnd=port_rounding(&c.mode).ok_or(\"invalid mode\")?;\nmatch (c.width,c.op.as_str(),c.target) {\n")
	for _, row := range rows {
		fmt.Fprintf(&g, "case c.Width==%d && c.Op==%q && c.Target==%d:\n", row.width, row.op, row.target)
		fmt.Fprintf(&r, "(%d,%q,%d) => {\n", row.width, row.op, row.target)
		if err := emitTier1PathRow(&g, &r, row); err != nil {
			return nil, err
		}
		g.WriteString("return obs,nil\n")
		r.WriteString("Ok(obs)\n},\n")
	}
	g.WriteString("default: return nil,fmt.Errorf(\"unsupported Tier1 case: %+v\",c)\n}\n}\n")
	r.WriteString("_ => Err(\"unsupported Tier1 case\".into()),\n}\n}\n")
	emitTier1ExpectedPaths(&g, rows)
	r.WriteString(tier1PathsRustProtocol)
	formatted, err := format.Source([]byte(g.String()))
	if err != nil {
		return nil, fmt.Errorf("format Tier1 Go paths: %w", err)
	}
	return map[string][]byte{tier1PathsGoPath: formatted, tier1PathsRustPath: []byte(r.String())}, nil
}

func tier1GoValue(row tier1PathRow, name string, port bool) string {
	switch row.kind {
	case "integer":
		return "fmt.Sprint(" + name + ")"
	case "bool":
		if port {
			return "strconv.FormatBool(" + name + " != 0)"
		}
		return "strconv.FormatBool(" + name + ")"
	}
	if row.resultWidth == 128 {
		if port {
			return "tier1Port128Hex(" + name + ")"
		}
		return "finiteDec128Hex(" + name + ")"
	}
	if port {
		return fmt.Sprintf("finiteHex%d(%s)", row.resultWidth, name)
	}
	return fmt.Sprintf("finiteHex%d(%s.ToUint%d())", row.resultWidth, name, row.resultWidth)
}

func tier1RustValue(row tier1PathRow, name string, port bool) string {
	switch row.kind {
	case "integer":
		return name + ".to_string()"
	case "bool":
		if port {
			return "(" + name + " != 0).to_string()"
		}
		return name + ".to_string()"
	}
	if row.resultWidth == 128 {
		if port {
			return "hex128(" + name + ".hi," + name + ".lo)"
		}
		return "dec128_hex(" + name + ")"
	}
	if port {
		return fmt.Sprintf("hex%d(%s)", row.resultWidth, name)
	}
	return fmt.Sprintf("hex%d(%s.to_bits())", row.resultWidth, name)
}

func tier1PublicPath(row tier1PathRow) string {
	if row.mode {
		return "public/mode"
	}
	if row.flags {
		return "public/flags"
	}
	return "public/value"
}

func emitTier1PathRow(g, r *strings.Builder, row tier1PathRow) error {
	gp, rp := []string{}, []string{}
	ga, ra := []string{}, []string{}
	if row.arity > 0 {
		fmt.Fprintf(g, "ops,err:=finiteParse%d(c.Operands); if err!=nil {return nil,err}\n", row.width)
		for i := 0; i < row.arity; i++ {
			name := []string{"x", "y"}[i]
			if row.width == 128 {
				fmt.Fprintf(g, "%s:=%s; %sb:=ops[%d].bidgo()\n", name, fmt.Sprintf("ops[%d].Dec", i), name, i)
				fmt.Fprintf(r, "let (%s,%sb)=parse128(&c.operands[%d])?;\n", name, name, i)
			} else {
				fmt.Fprintf(g, "%s:=ops[%d]; %sb:=%s.ToUint%d()\n", name, i, name, name, row.width)
				fmt.Fprintf(r, "let %sb=u%d::from_str_radix(&c.operands[%d],16).map_err(|e|e.to_string())?;\nlet %s=Decimal%d::from_bits(%sb);\n", name, row.width, i, name, row.width, name)
			}
			gp = append(gp, name+"b")
			rp = append(rp, name+"b")
			if i > 0 {
				ga = append(ga, name)
				ra = append(ra, name)
			}
		}
	} else {
		gt, rt, parse := "int", "i", "ParseInt"
		if row.op == "from_uint" {
			gt, rt, parse = "uint", "u", "ParseUint"
		}
		fmt.Fprintf(g, "parsed,err:=strconv.%s(c.Param,10,%d); if err!=nil {return nil,err}; n:=%s%d(parsed)\n", parse, row.target, gt, row.target)
		fmt.Fprintf(r, "let n=c.param.parse::<%s%d>().map_err(|e|e.to_string())?;\n", rt, row.target)
		gp = append(gp, "n")
		rp = append(rp, "n")
		ga = append(ga, "n")
		ra = append(ra, "n")
	}
	if row.op == "scaleb" {
		g.WriteString("n,err:=strconv.ParseInt(c.Param,10,64); if err!=nil {return nil,err}; if int64(int(n))!=n {return nil,fmt.Errorf(\"scaleb exponent does not fit int\")}\n")
		r.WriteString("let n=c.param.parse::<i64>().map_err(|e|e.to_string())?;\n")
		gp = append(gp, "n")
		rp = append(rp, "n")
		ga = append(ga, "int(n)")
		ra = append(ra, "n")
	}
	if row.mode {
		ga = append(ga, "mode")
		ra = append(ra, "mode")
		if row.kind != "integer" {
			gp = append(gp, "rnd")
			rp = append(rp, "rnd")
		}
	}
	gcall := "x." + row.goMethod + "(" + strings.Join(ga, ",") + ")"
	rcall := "x." + row.rustMethod + "(" + strings.Join(ra, ",") + ")"
	if row.arity == 0 {
		gcall = row.goMethod + "(" + strings.Join(ga, ",") + ")"
		if row.flags {
			rcall = fmt.Sprintf("Decimal%d::%s(%s)", row.width, row.rustMethod, strings.Join(ra, ","))
		} else {
			rcall = fmt.Sprintf("Decimal%d::from(n)", row.width)
		}
	}
	g.WriteString("var obs []tier1ref.Observation\n")
	r.WriteString("let mut obs=Vec::new();\n")
	emitTier1PublicObservation(g, r, row, gcall, rcall, tier1PublicPath(row))
	if row.kind == "integer" {
		fmt.Fprintf(g, "var pb %s; var pf uint32\nswitch c.Mode {\n", tier1GoIntType(row))
		r.WriteString("let (pr,praw)=match c.mode.as_str() {\n")
		for _, m := range []struct{ name, suffix string }{{"nearest_even", "Rnint"}, {"nearest_away", "Rninta"}, {"toward_zero", "Int"}, {"toward_positive", "Ceil"}, {"toward_negative", "Floor"}} {
			suffix := m.suffix
			if strings.HasSuffix(row.op, "_exact") {
				suffix = "X" + strings.ToLower(suffix)
			}
			fn := row.port + suffix
			mod, rf, err := resolvePort(fn, "Tier1 "+row.op)
			if err != nil {
				return err
			}
			fmt.Fprintf(g, "case %q: pb,pf=bidgo.%s(%s)\n", m.name, fn, strings.Join(gp, ","))
			fmt.Fprintf(r, "%q => bid754::generated::%s::%s(%s),\n", m.name, mod, rf, strings.Join(rp, ","))
		}
		g.WriteString("default: return nil,fmt.Errorf(\"invalid integer rounding mode\")\n}\n")
		r.WriteString("_ => return Err(\"invalid integer rounding mode\".into()),\n};\n")
	} else {
		mod, rf, err := resolvePort(row.port, "Tier1 "+row.op)
		if err != nil {
			return err
		}
		if !row.flags {
			fmt.Fprintf(g, "pb:=bidgo.%s(%s)\n", row.port, strings.Join(gp, ","))
			fmt.Fprintf(r, "let pr=bid754::generated::%s::%s(%s);\n", mod, rf, strings.Join(rp, ","))
		} else if apiemit.PortPfpsf(row.port) {
			fmt.Fprintf(g, "var pf uint32; pb:=bidgo.%s(%s,&pf)\n", row.port, strings.Join(gp, ","))
			r.WriteString(portCallStmt(row.port, mod, rf, rp) + "\n")
		} else {
			fmt.Fprintf(g, "pb,pf:=bidgo.%s(%s)\n", row.port, strings.Join(gp, ","))
			r.WriteString(portCallStmt(row.port, mod, rf, rp) + "\n")
		}
	}
	gf, rf := "pf", "praw"
	if !row.flags {
		gf, rf = "0", "0"
	}
	fmt.Fprintf(g, "obs=append(obs,tier1ref.Observation{Path:\"go/port\",Kind:%q,Width:%d,Value:%s,Flags:%s,HasFlags:%t})\n", row.kind, row.resultWidth, tier1GoValue(row, "pb", true), gf, row.flags)
	fmt.Fprintf(r, "push(&mut obs,\"rust/port\",%q,%d,%s,%s,%t)?;\n", row.kind, row.resultWidth, tier1RustValue(row, "pr", true), rf, row.flags)
	if row.defaultPath {
		g.WriteString("if c.Mode==\"nearest_even\" {\n")
		r.WriteString("if c.mode==\"nearest_even\" {\n")
		emitTier1PublicObservation(g, r, row, "x.ScaleB(int(n))", "x.scaleb(n)", "public/flags")
		g.WriteString("}\n")
		r.WriteString("}\n")
	}
	return nil
}

func tier1GoIntType(row tier1PathRow) string {
	prefix := "int"
	if strings.HasPrefix(row.op, "to_uint") {
		prefix = "uint"
	}
	return prefix + strconv.Itoa(row.target)
}

func emitTier1PublicObservation(g, r *strings.Builder, row tier1PathRow, gcall, rcall, path string) {
	gf, rf := "0", "0"
	if row.flags {
		fmt.Fprintf(g, "pv,publicFlags:=%s; mapped,err:=finiteMapPublicFlags(publicFlags); if err!=nil {return nil,err}\n", gcall)
		fmt.Fprintf(r, "let (pv,public_flags)=%s; let mapped=map_public_flags(public_flags)?;\n", rcall)
		gf, rf = "mapped", "mapped"
	} else {
		fmt.Fprintf(g, "pv:=%s\n", gcall)
		fmt.Fprintf(r, "let pv=%s;\n", rcall)
	}
	fmt.Fprintf(g, "obs=append(obs,tier1ref.Observation{Path:%q,Kind:%q,Width:%d,Value:%s,Flags:%s,HasFlags:%t})\n", "go/"+path, row.kind, row.resultWidth, tier1GoValue(row, "pv", false), gf, row.flags)
	fmt.Fprintf(r, "push(&mut obs,%q,%q,%d,%s,%s,%t)?;\n", "rust/"+path, row.kind, row.resultWidth, tier1RustValue(row, "pv", false), rf, row.flags)
}

func emitTier1ExpectedPaths(g *strings.Builder, rows []tier1PathRow) {
	g.WriteString("func tier1ExpectedPaths(c tier1ref.Case, language string) []string {\nif language!=\"go\" && language!=\"rust\" {return nil}\nif tier1ref.Validate(c)!=nil {return nil}\nswitch {\n")
	for _, row := range rows {
		fmt.Fprintf(g, `case c.Width==%d && c.Op==%q && c.Target==%d:
paths:=[]string{language+%q,language+"/port"}
`, row.width, row.op, row.target, "/"+tier1PublicPath(row))
		if row.defaultPath {
			g.WriteString("if c.Mode==\"nearest_even\" {paths=append(paths,language+\"/public/flags\")}\n")
		}
		g.WriteString("return paths\n")
	}
	g.WriteString("default: return nil\n}\n}\n")
}

const tier1PathsGoSupport = `package bid754
import (
 "fmt"
 "strconv"
 bidgo "github.com/sky1core/bid754/bid754-go/internal/bidgo"
 "github.com/sky1core/bid754/bid754-go/internal/tier1ref"
)
func tier1Port128Hex(v bidgo.BID_UINT128) string {
 hi,lo:=bidgo.Bid128Words(v)
 return finiteHex128(hi,lo)
}
`

const tier1PathsRustSupport = `use std::io::{self, BufRead, Read, Write};
use bid754::{Decimal32,Decimal64,Decimal128,ExceptionFlags,RoundingMode};
use serde::{Deserialize,Serialize};

#[derive(Deserialize)]
#[serde(deny_unknown_fields)]
struct Case {
 width: u32,
 op: String,
 mode: String,
 #[serde(deserialize_with="deserialize_operands")]
 operands: Vec<String>,
 param: String,
 target: u32,
}
fn deserialize_operands<'de,D:serde::Deserializer<'de>>(d:D)->Result<Vec<String>,D::Error> {
 match Option::<Vec<String>>::deserialize(d)? {
  Some(operands)=>Ok(operands),
  None=>Ok(Vec::new()),
 }
}
#[derive(Deserialize)]
#[serde(deny_unknown_fields)]
struct Request {version: u32, case: Case}
#[derive(Serialize)]
struct Observation {
 path: String,
 kind: String,
 width: u32,
 value: String,
 flags: u32,
 has_flags: bool,
}
#[derive(Serialize)]
struct Response {version: u32, observations: Vec<Observation>}

fn public_mode(mode: &str) -> Option<RoundingMode> {
 Some(match mode {
 "nearest_even"=>RoundingMode::NearestEven,
 "nearest_away"=>RoundingMode::NearestAway,
 "toward_zero"=>RoundingMode::TowardZero,
 "toward_positive"=>RoundingMode::TowardPositive,
 "toward_negative"=>RoundingMode::TowardNegative,
 _=>return None,
 })
}
fn port_rounding(mode: &str) -> Option<i64> {
 Some(match mode {
 "nearest_even"=>0,"toward_negative"=>1,"toward_positive"=>2,"toward_zero"=>3,"nearest_away"=>4,
 _=>return None,
 })
}
fn map_public_flags(flags: ExceptionFlags) -> Result<u32,String> {
 let raw=flags.bits();
 if raw & !0x1f != 0 {return Err(format!("unexpected public flags {raw:#x}"));}
 let mut native=0;
 for (public,port) in [(1,0x20),(2,0x10),(4,8),(8,4),(16,1)] {
  if raw & public != 0 {native |= port;}
 }
 Ok(native)
}
fn push(obs: &mut Vec<Observation>,path: &str,kind: &str,width: u32,value: String,flags: u32,has_flags: bool) -> Result<(),String> {
 if flags & !0x3d != 0 || !has_flags && flags != 0 {return Err(format!("invalid observation flags {flags:#x}"));}
 obs.push(Observation{path:path.into(),kind:kind.into(),width,value,flags,has_flags});
 Ok(())
}
fn hex32(v:u32)->String {format!("{v:08x}")}
fn hex64(v:u64)->String {format!("{v:016x}")}
fn hex128(hi:u64,lo:u64)->String {format!("{hi:016x}:{lo:016x}")}
fn dec128_hex(v:Decimal128)->String {
 let b=v.to_le_bytes();
 let mut hi=[0;8];let mut lo=[0;8];
 hi.copy_from_slice(&b[8..]);lo.copy_from_slice(&b[..8]);
 hex128(u64::from_le_bytes(hi),u64::from_le_bytes(lo))
}
fn parse128(s:&str)->Result<(Decimal128,bid754::gen_types::BID_UINT128),String> {
 let (h,l)=s.split_once(':').ok_or("invalid decimal128")?;
 let hi=u64::from_str_radix(h,16).map_err(|e|e.to_string())?;
 let lo=u64::from_str_radix(l,16).map_err(|e|e.to_string())?;
 let mut b=[0;16];b[..8].copy_from_slice(&lo.to_le_bytes());b[8..].copy_from_slice(&hi.to_le_bytes());
 Ok((Decimal128::from_le_bytes(b),bid754::gen_types::BID_UINT128{lo,hi}))
}
fn canonical_integer(s:&str,signed:bool,bits:u32)->Result<(),String> {
 if signed {
  let n=s.parse::<i64>().map_err(|e|e.to_string())?;
  if n.to_string()!=s || bits==32 && i32::try_from(n).is_err() {return Err("invalid signed integer parameter".into());}
 } else {
  let n=s.parse::<u64>().map_err(|e|e.to_string())?;
  if n.to_string()!=s || bits==32 && u32::try_from(n).is_err() {return Err("invalid unsigned integer parameter".into());}
 }
 Ok(())
}
fn validate(c:&Case)->Result<(),String> {
 let length=match c.width {32=>8,64=>16,128=>33,_=>return Err("invalid decimal width".into())};
 if public_mode(&c.mode).is_none() {return Err("invalid rounding mode".into());}
 let arity=match c.op.as_str() {
 "from_int"|"from_uint"=>{
  if c.target!=32 && c.target!=64 {return Err("invalid integer input width".into());}
  canonical_integer(&c.param,c.op=="from_int",c.target)?;
  0
 },
 "to_int"|"to_uint"|"to_int_exact"|"to_uint_exact"=>{
  if ![8,16,32,64].contains(&c.target) || !c.param.is_empty() {return Err("invalid integer output shape".into());}
  1
 },
 "convert"=>{
  if ![32,64,128].contains(&c.target) || c.target==c.width || !c.param.is_empty() {return Err("invalid decimal conversion shape".into());}
  1
 },
 "scaleb"=>{
  if c.target!=0 {return Err("invalid scale target".into());}
  canonical_integer(&c.param,true,64)?;
  1
 },
 "rem"|"fmod"|"minnum"|"maxnum"|
 "quiet_equal"|"quiet_not_equal"|"quiet_greater"|"quiet_greater_equal"|"quiet_greater_unordered"|
 "quiet_less"|"quiet_less_equal"|"quiet_less_unordered"|"quiet_not_greater"|"quiet_not_less"|"quiet_ordered"|"quiet_unordered"=>{
  if c.target!=0 || !c.param.is_empty() {return Err("invalid binary operation shape".into());}
  2
 },
 _=>return Err("invalid operation".into()),
 };
 if c.operands.len()!=arity {return Err("invalid operand count".into());}
 for raw in &c.operands {
  if raw.len()!=length {return Err("invalid raw width".into());}
  for (i,b) in raw.bytes().enumerate() {
   if c.width==128 && i==16 {
    if b!=b':' {return Err("invalid hi:lo separator".into());}
   } else if !b.is_ascii_digit() && !(b'a'..=b'f').contains(&b) {return Err("invalid lowercase raw hex".into());}
  }
 }
 Ok(())
}
`

const tier1PathsRustProtocol = `
fn run()->Result<(),String> {
 const MAX_RECORD:u64=4096;
 let stdin=io::stdin();let mut input=stdin.lock();
 let stdout=io::stdout();let mut output=stdout.lock();
 loop {
  let mut record=Vec::new();
  let size=(&mut input).take(MAX_RECORD+1).read_until(b'\n',&mut record).map_err(|e|e.to_string())?;
  if size==0 {break;}
  if size as u64>MAX_RECORD {return Err("record exceeds 4096 bytes".into());}
  let request:Request=serde_json::from_slice(&record).map_err(|e|format!("invalid request: {e}"))?;
  if request.version!=1 {return Err("unsupported protocol version".into());}
  let response=Response{version:1,observations:observe(&request.case)?};
  serde_json::to_writer(&mut output,&response).map_err(|e|e.to_string())?;
  output.write_all(b"\n").map_err(|e|e.to_string())?;
  output.flush().map_err(|e|e.to_string())?;
 }
 Ok(())
}
fn main() {
 if let Err(e)=run() {
  eprintln!("tier1_probe: {e}");
  std::process::exit(1);
 }
}
`
