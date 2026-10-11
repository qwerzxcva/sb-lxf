//go:build with_rust_accel
// +build with_rust_accel

// Rust 加速版本：通过 cgo 调用 sb-lxf Rust LPM/哈希 模块。
// 启用构建标签 with_rust_accel 时，编译并链接 Rust 静态库
//（contrib/rust-accel/target/release/libsb_lxf_rust_accel.a）。
package rustaccel

/*
#cgo LDFLAGS: -L${SRCDIR}/target/aarch64-linux-android/release -lsb_lxf_rust_accel -lm -lpthread -ldl -Wl,--no-as-needed
#include "sb_rust_accel.h"
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// UsesRust 报告当前构建是否链接了 Rust 加速模块。
func UsesRust() bool { return true }

// Version 返回 Rust 模块版本字符串。
func Version() string {
	cs := C.sb_rust_version()
	if cs == nil {
		return "unknown"
	}
	defer C.sb_rust_free_cstr(cs)
	return C.GoString(cs)
}

// LpmTable 是一个 Rust 实现的 LPM / CIDR 匹配表。
type LpmTable struct {
	impl unsafe.Pointer
}

// NewLpmTable 创建一个 LPM 表。
func NewLpmTable() *LpmTable {
	return &LpmTable{impl: C.sb_rust_lpm_new()}
}

// InsertCIDR 插入一条 CIDR 规则（"10.0.0.0/8" 或 "2001:db8::/32"）。
func (t *LpmTable) InsertCIDR(cidr string, result int32) error {
	cs := C.CString(cidr)
	defer C.free(unsafe.Pointer(cs))
	rc := C.sb_rust_lpm_insert_cidr(t.impl, cs, C.int32_t(result))
	if rc != 0 {
		return fmt.Errorf("insert %q: rc=%d", cidr, int(rc))
	}
	return nil
}

// MatchV4 返回最具体前缀的结果 id；无命中 -1。
func (t *LpmTable) MatchV4(ip string) int32 {
	cs := C.CString(ip)
	defer C.free(unsafe.Pointer(cs))
	return int32(C.sb_rust_lpm_match_v4(t.impl, cs))
}

// MatchV6 返回最具体前缀的结果 id；无命中 -1。
func (t *LpmTable) MatchV6(ip string) int32 {
	cs := C.CString(ip)
	defer C.free(unsafe.Pointer(cs))
	return int32(C.sb_rust_lpm_match_v6(t.impl, cs))
}

// Free 释放 LPM 表。
func (t *LpmTable) Free() {
	if t != nil && t.impl != nil {
		C.sb_rust_lpm_free(t.impl)
		t.impl = nil
	}
}

// Fnv1a64 计算 FNV-1a 64 位哈希。
func Fnv1a64(b []byte) uint64 {
	if len(b) == 0 {
		return 0
	}
	return uint64(C.sb_rust_fnv1a_64(&b[0], C.size_t(len(b))))
}

// QuickHash64 计算轻量 64 位哈希（非加密，分散度好）。
func QuickHash64(b []byte) uint64 {
	if len(b) == 0 {
		return 0
	}
	return uint64(C.sb_rust_quick_hash64(&b[0], C.size_t(len(b))))
}
