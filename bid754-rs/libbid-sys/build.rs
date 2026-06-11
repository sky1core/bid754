use std::env;
use std::path::PathBuf;

fn main() {
    let manifest_dir =
        PathBuf::from(env::var_os("CARGO_MANIFEST_DIR").expect("CARGO_MANIFEST_DIR not set"));
    let lib_dir = manifest_dir.join("../../devtools/third_party/intel_dfp/lib");
    let lib_path = lib_dir.join("libbid.a");
    if !lib_path.exists() {
        panic!(
            "missing Intel BID static library at {}; \
             run `make setup-native` or `bash ./scripts/setup_c_libs.sh` from the repository root",
            lib_path.display()
        );
    }
    let link_dir = lib_dir
        .canonicalize()
        .unwrap_or_else(|_| lib_dir.clone());

    println!("cargo:rerun-if-changed={}", lib_path.display());
    println!("cargo:rustc-link-search=native={}", link_dir.display());
    println!("cargo:rustc-link-lib=static=bid");
}
