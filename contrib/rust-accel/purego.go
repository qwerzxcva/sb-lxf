//go:build !with_rust_accel

// 纯 Go 版本：默认构建（未启用 with_rust_accel）下的回退实现。
// 提供与 cgo_rust.go 相同的公开 API，保证调用方代码无需改动；
// 用 Go 标准库的 net/netip 做 LPM，用 FNV-1a 做哈希。
package rustaccel

import (
	"fmt"
	"hash/fnv"
	"net/netip"
)

// UsesRust 报告当前构建是否链接了 Rust 加速模块（纯 Go 版本为 false）。
func UsesRust() bool { return false }

// Version 返回模块版本字符串。
func Version() string { return "sb-lxf-rust-accel (pure-go fallback)" }

// cidrEntry 存一个 CIDR 前缀 + 结果。
type cidrEntry struct {
	prefix netip.Prefix
	result int32
}

// LpmTable 是一个纯 Go 的 LPM / CIDR 匹配表（线性扫描，规则少时够用）。
type LpmTable struct {
	rules []cidrEntry
}

// NewLpmTable 创建一个空的 LPM 表。
func NewLpmTable() *LpmTable { return &LpmTable{} }

// InsertCIDR 插入一条 CIDR 规则。
func (t *LpmTable) InsertCIDR(cidr string, result int32) error {
	pfx, err := netip.ParsePrefix(cidr)
	if err != nil {
		return fmt.Errorf("bad cidr %q: %w", cidr, err)
	}
	t.rules = append(t.rules, cidrEntry{prefix: pfx, result: result})
	return nil
}

// match 返回最长前缀（最具体）命中的结果 id；无命中 -1。
func (t *LpmTable) match(addr netip.Addr) int32 {
	bestLen := -1
	best := int32(-1)
	for _, e := range t.rules {
		if e.prefix.Contains(addr) && int(e.prefix.Bits()) > bestLen {
			bestLen = int(e.prefix.Bits())
			best = e.result
		}
	}
	return best
}

// MatchV4 匹配 IPv4。
func (t *LpmTable) MatchV4(ip string) int32 {
	a, err := netip.ParseAddr(ip)
	if err != nil {
		return -1
	}
	return t.match(a)
}

// MatchV6 匹配 IPv6。
func (t *LpmTable) MatchV6(ip string) int32 {
	a, err := netip.ParseAddr(ip)
	if err != nil {
		return -1
	}
	return t.match(a)
}

// Free 释放（纯 Go 版无 C 内存，空实现）。
func (t *LpmTable) Free() {}

// Fnv1a64 计算 FNV-1a 64 位哈希（Go 标准库）。
func Fnv1a64(b []byte) uint64 {
	h := fnv.New64a()
	h.Write(b)
	return h.Sum64()
}

// QuickHash64 轻量 64 位哈希（纯 Go，无加密，用于去重/负载键）。
func QuickHash64(b []byte) uint64 {
	const p = 0x9E37_79B1_85E6_4C55
	var h uint64 = 0x5B_3C_8D_00
	for i := 0; i < len(b); i += 8 {
		var v uint64
		n := 0
		for j := i; j < i+8 && j < len(b); j, n = j+1, n+1 {
			v = (v << 8) | uint64(b[j])
		}
		// 补齐剩余字节的高位
		_ = n
		v <<= uint(8 * (8 - n))
		h = (h + v).RotateLeft(27) * p
	}
	return h
}
