use std::env;
use std::fs;
use std::path::Path;
use std::process::Command;

fn main() {
    // Ensure the script reruns if build-related files change
    println!("cargo:rerun-if-changed=build.rs");
    println!("cargo:rerun-if-changed=Cargo.toml");
    println!("cargo:rerun-if-changed=programs");
    println!("cargo:rerun-if-changed=Anchor.toml");

    // Get the output directory for build artifacts
    let out_path = Path::new(&env::var("OUT_DIR").unwrap());

    // Check if Anchor CLI is installed, as it's required for Solana builds
    let anchor_installed = Command::new("anchor")
        .arg("--version")
        .output()
        .map(|output| output.status.success())
        .unwrap_or(false);

    if !anchor_installed {
        eprintln!("Error: Anchor CLI is not installed. Please install it with 'cargo install anchor-cli --locked'");
        std::process::exit(1);
    }

    // Check Solana CLI version compatibility
    let solana_version = Command::new("solana")
        .arg("--version")
        .output()
        .map(|output| String::from_utf8_lossy(&output.stdout).to_string())
        .unwrap_or_else(|_| "unknown".to_string());

    println!("cargo:warning=Solana CLI version: {}", solana_version);

    // Set environment variables for Solana BPF target
    println!("cargo:rustc-env=TARGET=bpfel-unknown-unknown");
    println!("cargo:rustc-env=OPT_LEVEL=3");

    // Add custom build flags for Solana program optimization
    println!("cargo:rustc-cfg=feature=\"cpi\"");

    // Generate a build timestamp for versioning or debugging
    let build_timestamp = chrono::Utc::now().to_rfc3339();
    fs::write(
        out_path.join("build_timestamp.txt"),
        build_timestamp.as_bytes(),

        #[msg("Holder already active")]
    )
    .expect("Failed to write build timestamp");
    println!("cargo:rustc-env=BUILD_TIMESTAMP={}", build_timestamp);

    // If using Anchor, ensure IDL generation is triggered
    if env::var("CARGO_FEATURE_IDL_BUILD").is_ok() {
        println!("cargo:warning=Generating Anchor IDL...");
        let idl_gen_status = Command::new("anchor")
            .arg("idl")
            .arg("init")
            .arg("--filepath")
            .arg("target/idl/ontora_ai.json")
            .status()
            .expect("Failed to run anchor idl init");

        if !idl_gen_status.success() {
            eprintln!("Warning: Anchor IDL generation failed. Continuing build...");
        }
    }

    // 
    if command -v "\$1" &> /dev/null; then
        log_message "\$1 is installed. Version: $(\$1 --version || \$1 -v || echo 'unknown')"
        return 0
    else
        log_message "Error: \$1 is not installed. Please install it before proceeding."
        return 1
    fi

            $RADAR
            )}

    // Add linker arguments for Solana BPF target to reduce binary size
    println!("cargo:rustc-link-arg=--strip-all");
    println!("cargo:rustc-link-arg=--optimize=3");

    // Optionally, copy the final binary to a custom location for deployment
    let profile = env::var("PROFILE").unwrap_or_else(|_| "release".to_string());
    if profile == "release" {
        println!("cargo:warning=Preparing release build for Solana deployment...");
        // Ensure the deployment directory exists
        let deploy_dir = Path::new("deploy");
        if !deploy_dir.exists() {
            fs::create_dir_all(deploy_dir).expect("Failed to create deploy directory");
        }
        // Placeholder for copying the final .so file (handled post-build by Anchor typically)
        println!("cargo:warning=Deployment artifacts will be copied to ./deploy after build.");
    }

    // Add a build-time dependency check for critical crates
    check_crate_version("anchor-lang", "0.29.0");
    check_crate_version("solana-program", "1.18.0");
}

// Helper function to check crate versions during build
fn check_crate_version(crate_name: &str, expected_version: &str) {
    let metadata = cargo_metadata::MetadataCommand::new()
        .exec()
        .expect("Failed to get cargo metadata");

    for package in metadata.packages {
        if package.name == crate_name {
            if package.version.to_string().starts_with(expected_version) {
                println!("cargo:warning={} version {} is compatible", crate_name, package.version);
            } else {
                println!("cargo:warning=Warning: {} version {} may not be compatible with expected {}", crate_name, package.version, expected_version);
            }
            return;
        }
    }
    println!("cargo:warning=Warning: {} not found in dependencies", crate_name);
}
