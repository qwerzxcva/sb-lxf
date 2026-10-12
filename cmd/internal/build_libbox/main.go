package main

import (
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	_ "github.com/sagernet/gomobile"
	"github.com/sagernet/sing-box/cmd/internal/build_shared"
	"github.com/sagernet/sing-box/log"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/common/rw"
	"github.com/sagernet/sing/common/shell"
)

var (
	debugEnabled bool
	target       string
	platform     string
	// slim: Android-only power/size profile — drop heavy features sb-AI does not
	// use (usbip / openvpn / openconnect / tailscale + all ts_omit_*). Keeps the
	// core proxy stack (quic / wireguard / utls / naive / clash_api).
	slimEnabled bool
	// rustAccel: link the optional Rust accelerator (contrib/rust-accel) via cgo,
	// adding the `with_rust_accel` build tag. Off by default — the default build
	// uses the pure-Go fallback in the same package.
	rustAccel bool
	// withTailscale bool
)

func init() {
	flag.BoolVar(&debugEnabled, "debug", false, "enable debug")
	flag.StringVar(&target, "target", "android", "target platform")
	flag.StringVar(&platform, "platform", "", "specify platform")
	flag.BoolVar(&slimEnabled, "slim", false, "android-only slim profile: drop usbip/openvpn/openconnect/tailscale for lower power & smaller AAR")
	flag.BoolVar(&rustAccel, "rust-accel", false, "link the optional Rust accelerator (with_rust_accel tag + contrib/rust-accel static lib)")
	// flag.BoolVar(&withTailscale, "with-tailscale", false, "build tailscale for iOS and tvOS")
}

func main() {
	flag.Parse()

	build_shared.FindMobile()

	switch target {
	case "android":
		buildAndroid()
	case "apple":
		buildApple()
	}
}

var (
	sharedFlags []string
	debugFlags  []string
	sharedTags  []string
	darwinTags  []string
	// memcTags    []string
	notMemcTags []string
	debugTags   []string
)

func init() {
	sharedFlags = append(sharedFlags, "-trimpath")
	sharedFlags = append(sharedFlags, "-buildvcs=false")
	currentTag, err := build_shared.ReadTag()
	if err != nil {
		currentTag = "unknown"
	}
	sharedFlags = append(sharedFlags, "-ldflags", build_shared.LinkerFlags(currentTag, false))
	debugFlags = append(debugFlags, "-ldflags", build_shared.LinkerFlags(currentTag, true))

	sharedTags = append(sharedTags, "with_quic", "with_wireguard", "with_utls", "with_naive_outbound", "with_clash_api", "with_usbip", "with_openvpn", "with_openconnect", "badlinkname", "tfogo_checklinkname0")
	// Optional protocol families — kept in the default (full) build for
	// backward compatibility; the Android-only slim profile excludes them.
	sharedTags = append(sharedTags, "with_anytls", "with_shadowtls", "with_snell", "with_tor", "with_ssh", "with_masque")
	darwinTags = append(darwinTags, "with_dhcp", "grpcnotrace")
	// memcTags = append(memcTags, "with_tailscale")
	sharedTags = append(sharedTags, "with_tailscale", "ts_omit_logtail", "ts_omit_ssh", "ts_omit_drive", "ts_omit_taildrop", "ts_omit_webclient", "ts_omit_doctor", "ts_omit_capture", "ts_omit_kube", "ts_omit_aws", "ts_omit_synology", "ts_omit_bird")
	notMemcTags = append(notMemcTags, "with_low_memory")
	debugTags = append(debugTags, "debug")
}

type AndroidBuildConfig struct {
	AndroidAPI int
	OutputName string
	Tags       []string
}

func filterTags(tags []string, exclude ...string) []string {
	excludeMap := make(map[string]bool)
	for _, tag := range exclude {
		excludeMap[tag] = true
	}
	var result []string
	for _, tag := range tags {
		if !excludeMap[tag] {
			result = append(result, tag)
		}
	}
	return result
}

func checkJavaVersion() {
	var javaPath string
	javaHome := os.Getenv("JAVA_HOME")
	if javaHome == "" {
		javaPath = "java"
	} else {
		javaPath = filepath.Join(javaHome, "bin", "java")
	}

	javaVersion, err := shell.Exec(javaPath, "--version").ReadOutput()
	if err != nil {
		log.Fatal(E.Cause(err, "check java version"))
	}
	if !strings.Contains(javaVersion, "openjdk 17") {
		log.Fatal("java version should be openjdk 17")
	}
}

func getAndroidBindTarget() string {
	if platform != "" {
		return platform
	} else if debugEnabled {
		return "android/arm64"
	}
	return "android"
}

func buildAndroidVariant(config AndroidBuildConfig, bindTarget string) {
	args := []string{
		"bind",
		"-v",
		"-o", config.OutputName,
		"-target", bindTarget,
		"-androidapi", strconv.Itoa(config.AndroidAPI),
		"-javapkg=io.nekohasekai",
		"-libname=box",
	}

	if !debugEnabled {
		args = append(args, sharedFlags...)
	} else {
		args = append(args, debugFlags...)
	}

	args = append(args, "-tags", strings.Join(config.Tags, ","))
	args = append(args, "./experimental/libbox")

	command := exec.Command(build_shared.GoBinPath+"/gomobile", args...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	err := command.Run()
	if err != nil {
		log.Fatal(err)
	}

	copyPath := filepath.Join("..", "sing-box-for-android", "app", "libs")
	if rw.IsDir(copyPath) {
		copyPath, _ = filepath.Abs(copyPath)
		err = rw.CopyFile(config.OutputName, filepath.Join(copyPath, config.OutputName))
		if err != nil {
			log.Fatal(err)
		}
		log.Info("copied ", config.OutputName, " to ", copyPath)
	}
}

func buildAndroid() {
	build_shared.FindSDK()
	checkJavaVersion()

	bindTarget := getAndroidBindTarget()

	// Android-only slim profile: drop heavy features sb-AI does not use, so the
	// AAR is smaller and the kernel uses less CPU/battery at runtime.
	// The full feature set is preserved for non-slim (default) builds.
	// Reference: sb-AI only consumes quic/wireguard/utls/clash_api/tailscale.
	// usbip / openvpn / openconnect / tailscale are unused on Android and add
	// ~tens of MB + background-service overhead.
	var baseTags []string
	if slimEnabled {
		// Drop only the features sb-AI has no reference to. NOTE: tailscale is KEPT
		// (with all ts_omit_*) because sb-AI exposes "tailscale" as a first-class
		// endpoint type (NodeEditorScreen / SingBoxConfigGenerator ENDPOINT_TYPES);
		// removing it would break that feature. usbip / openvpn / openconnect are
		// unreferenced anywhere in sb-AI.
		// Also drop naive_outbound — sb-AI's OUTBOUND_TYPES has NO "naive" entry;
		// naive (HTTP CONNECT tunnel) is not a node type sb-AI supports.
		// Restore tor+ssh — sb-AI's OUTBOUND_TYPES includes both.
		heavyExcluded := []string{
			"with_usbip", "with_openvpn", "with_openconnect", "with_naive_outbound",
			// Android-only (sb-AI): drop protocol families sb-AI never emits.
			// sb-AI OUTBOUND_TYPES = vless/vmess/trojan/ss/hysteria2/hysteria/
			//   http/socks/wireguard/tailscale/tor/ssh/direct/block.
			// Cut: anytls, shadowtls, snell, masque (tor+ssh RESTORED).
			"with_anytls", "with_shadowtls", "with_snell", "with_masque",
		}
		baseTags = append(filterTags(sharedTags, heavyExcluded...))
		// Idle-suspend (SPEC 020): idle WireGuard/Tailscale endpoints suspend after
		// the configured window, cutting background work to zero for unused tunnels.
		// Enabled in the slim profile (power-optimized). The reachability seed in
		// route/reachability_lx.go now uses an optional DNSTransportWithOutboundTag
		// sub-interface, so this tag compiles cleanly without touching the base
		// DNSTransport interface or the ~10 DNS transport implementations.
		baseTags = append(baseTags, "with_lx_idle_suspend")
	} else {
		baseTags = append([]string{}, sharedTags...)
	}

	// Build main variant (SDK 24)
	mainTags := append([]string{}, baseTags...)
	// mainTags = append(mainTags, memcTags...)
	// Optional Rust accelerator (contrib/rust-accel) — cgo links the static lib
	// and enables the Rust LPM/hash path. Off by default.
	if rustAccel {
		mainTags = append(mainTags, "with_rust_accel")
	}
	if debugEnabled {
		mainTags = append(mainTags, debugTags...)
	}
	buildAndroidVariant(AndroidBuildConfig{
		AndroidAPI: 24,
		OutputName: "libbox.aar",
		Tags:       mainTags,
	}, bindTarget)

	// Build legacy variant (SDK 21, no naive outbound)
	legacyTags := filterTags(baseTags, "with_naive_outbound")
	// legacyTags = append(legacyTags, memcTags...)
	if debugEnabled {
		legacyTags = append(legacyTags, debugTags...)
	}
	buildAndroidVariant(AndroidBuildConfig{
		AndroidAPI: 21,
		OutputName: "libbox-legacy.aar",
		Tags:       legacyTags,
	}, bindTarget)
}

func buildApple() {
	var bindTarget string
	if platform != "" {
		bindTarget = platform
	} else if debugEnabled {
		bindTarget = "ios"
	} else {
		bindTarget = "ios,iossimulator,tvos,tvossimulator,macos"
	}

	args := []string{
		"bind",
		"-v",
		"-target", bindTarget,
		"-libname=box",
		"-tags-not-macos=with_low_memory",
		"-iosversion=15.0",
		"-macosversion=13.0",
		"-tvosversion=17.0",
	}
	//if !withTailscale {
	//	args = append(args, "-tags-macos="+strings.Join(memcTags, ","))
	//}

	if !debugEnabled {
		args = append(args, sharedFlags...)
	} else {
		args = append(args, debugFlags...)
	}

	tags := append(sharedTags, darwinTags...)
	//if withTailscale {
	//	tags = append(tags, memcTags...)
	//}
	if debugEnabled {
		tags = append(tags, debugTags...)
	}

	args = append(args, "-tags", strings.Join(tags, ","))
	args = append(args, "./experimental/libbox")

	command := exec.Command(build_shared.GoBinPath+"/gomobile", args...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	err := command.Run()
	if err != nil {
		log.Fatal(err)
	}

	copyPath := filepath.Join("..", "sing-box-for-apple")
	if rw.IsDir(copyPath) {
		targetDir := filepath.Join(copyPath, "Libbox.xcframework")
		targetDir, _ = filepath.Abs(targetDir)
		os.RemoveAll(targetDir)
		os.Rename("Libbox.xcframework", targetDir)
		log.Info("copied to ", targetDir)
	}
}
