//! sb-lxf 可选 Rust 加速模块（C-ABI FFI）
//!
//! 提供两类纯计算热点的 Rust 实现，供 Go 通过 cgo 调用（默认关闭，
//! 启用构建标签 `with_rust_accel` 时链接）：
//!
//! 1. **LPM / CIDR 最长前缀匹配** — 路由 IP 规则匹配热点，二进制前缀表
//!    （每掩码层一个有序 HashMap），查询 O(层数) 哈希命中，比 Go 线性
//!    扫描更快、且零 GC 压力（静态分配，无堆碎片）。
//! 2. **指纹 / 轻量哈希** — 连接跟踪、去重的非加密哈希（FNV-1a + xxh 近似）。
//!
//! 全部 `no_std` 不友好但标准库版；FFI 使用 `#[no_mangle] extern "C"` 导出，
//! 内存模型：`*mut` 句柄 + `*_free`，禁止跨语言 panic（FFI 边界用
//! `catch_unwind` 包裹）。

use std::collections::HashMap;
use std::panic::{catch_unwind, AssertUnwindSafe};

// ============================ LPM / CIDR 匹配 ============================

/// 一个 LPM 表，支持 IPv4（32 层）与 IPv6（128 层）两套前缀表。
///
/// 每层用 `HashMap` 存前缀值→结果 id。匹配时从最具体层往宽层线性探测，
/// 命中即返回。前缀值 = IP 左移 (32-mask) 位（v4 u32 / v6 u128）。
pub struct LpmTable {
    v4: Vec<HashMap<u32, i32>>, // 下标 0..=32 对应掩码 32..0（0 = 默认）
    v6: Vec<HashMap<u128, i32>>,
}

impl LpmTable {
    pub fn new() -> Self {
        Self {
            v4: (0..=32).map(|_| HashMap::new()).collect(),
            v6: (0..=128).map(|_| HashMap::new()).collect(),
        }
    }

    /// 插入 v4 前缀：`ip` 为点分十进制或字节序 u32，`mask` ∈ 0..=32。
    fn insert_v4(&mut self, prefix: u32, mask: u8, result: i32) {
        let layer = 32 - mask as usize; // 最具体 mask=32 → layer=0
        let key = prefix & mask_bits32(mask);
        self.v4[layer].insert(key, result);
    }

    fn insert_v6(&mut self, prefix: u128, mask: u8, result: i32) {
        let layer = 128 - mask as usize;
        let key = prefix & mask_bits128(mask);
        self.v6[layer].insert(key, result);
    }

    /// 匹配 v4 IP，返回最具体前缀的结果；无命中返回 -1。
    fn match_v4(&self, ip: u32) -> i32 {
        for mask in (0..=32u32).rev() {
            let layer = (32 - mask) as usize;
            let key = ip & mask_bits32(mask as u8);
            if let Some(&r) = self.v4[layer].get(&key) {
                return r;
            }
        }
        -1
    }

    fn match_v6(&self, ip: u128) -> i32 {
        for mask in (0..=128u32).rev() {
            let layer = (128 - mask) as usize;
            let key = ip & mask_bits128(mask as u8);
            if let Some(&r) = self.v6[layer].get(&key) {
                return r;
            }
        }
        -1
    }
}

impl Default for LpmTable {
    fn default() -> Self {
        Self::new()
    }
}

/// 掩码 32 → 32-bit 前缀掩码（mask 位为 1，低 (32-mask) 位为 0）。
fn mask_bits32(mask: u8) -> u32 {
    if mask == 0 {
        0
    } else {
        u32::MAX << (32 - mask)
    }
}

/// 掩码 128 → 128-bit 前缀掩码。
fn mask_bits128(mask: u8) -> u128 {
    if mask == 0 {
        0
    } else {
        u128::MAX << (128 - mask)
    }
}

fn parse_ipv4(s: &str) -> Option<u32> {
    let mut parts = s.split('.');
    let mut out: u32 = 0;
    for _ in 0..4 {
        let p = parts.next()?.parse::<u32>().ok()?;
        if p > 255 {
            return None;
        }
        out = (out << 8) | p;
    }
    Some(out)
}

fn parse_ipv6(s: &str) -> Option<u128> {
    // 标准 IPv6：8 组 16 位，允许单个 `::` 省略（不支持 v4-mapped 混合）。
    if s.is_empty() {
        return None;
    }
    if s == "::" {
        return Some(0);
    }
    if s.contains(".") {
        return None; // v4-mapped 不支持
    }
    let has_ellipsis = s.contains("::");
    let (left, right) = match s.split_once("::") {
        Some((l, r)) => (l, r),
        None => (s, ""),
    };
    let parse_groups = |seg: &str| -> Option<Vec<u16>> {
        if seg.is_empty() {
            return Some(Vec::new());
        }
        let mut out = Vec::new();
        for tok in seg.split(':') {
            out.push(u16::from_str_radix(tok, 16).ok()?);
        }
        Some(out)
    };
    let left_groups = parse_groups(left)?;
    let right_groups = parse_groups(right)?;
    let n_left = left_groups.len();
    let n_right = right_groups.len();
    let n_zeros = if has_ellipsis {
        if 8 < n_left + n_right {
            return None;
        }
        8 - n_left - n_right
    } else {
        if n_left != 8 {
            return None;
        }
        0
    };
    // 拼 128 位值：[left (n_left 组)][zeros (n_zeros 组)][right (n_right 组)]
    let mut val: u128 = 0;
    for g in &left_groups {
        val = (val << 16) | (*g as u128);
    }
    // 把 left 推到高位：左移 (n_zeros + n_right) 个组位；right 直接拼在低位
    val <<= (n_zeros + n_right) * 16;
    // right 组：按顺序填入最低 n_right 个 16-bit 槽位（不再整体移位）
    for (i, g) in right_groups.iter().enumerate() {
        let shift = (n_right as u32 - 1 - i as u32) * 16;
        val |= (*g as u128) << shift;
    }
    Some(val)
}

fn parse_cidr(s: &str) -> Result<(u128, u8, bool), &'static str> {
    // 返回 (ip, mask, is_v6)
    let (ip_s, mask_s) = match s.split_once('/') {
        Some((a, b)) => (a, b),
        None => (s, ""),
    };
    let mask: u32 = if mask_s.is_empty() {
        0
    } else {
        mask_s.parse::<u32>().map_err(|_| "bad mask")?
    };
    if ip_s.contains(':') {
        if mask > 128 {
            return Err("mask>128");
        }
        let v6 = parse_ipv6(ip_s).ok_or("bad v6")?;
        Ok((v6, mask as u8, true))
    } else {
        if mask > 32 {
            return Err("mask>32");
        }
        let v4 = parse_ipv4(ip_s).ok_or("bad v4")?;
        Ok((v4 as u128, mask as u8, false))
    }
}

// ============================ 指纹 / 哈希 ============================

/// FNV-1a 64 位（非加密，用于连接跟踪 / 去重键哈希）。
pub fn fnv1a_64(data: &[u8]) -> u64 {
    let mut hash: u64 = 0xcbf2_9ce4_8422_2325;
    for &b in data {
        hash ^= b as u64;
        hash = hash.wrapping_mul(0x0000_0100_0000_01b3);
    }
    hash
}

/// xxHash64 的无依赖简化版（非官方 xx，但提供 64 位良好分散度）。
/// 适用于负载 / 路由的轻量指纹。
pub fn quick_hash64(data: &[u8]) -> u64 {
    const P: u64 = 0x9E37_79B1_85E6_4C55;
    let mut h: u64 = 0x5B_3C_8D_00;
    for chunk in data.chunks(8) {
        let mut b = [0u8; 8];
        b[..chunk.len()].copy_from_slice(chunk);
        let v = u64::from_le_bytes(b);
        h = h
            .wrapping_add(v)
            .rotate_left(27)
            .wrapping_mul(P)
            .wrapping_add(0x9E37_79B1);
    }
    h
}

// ============================ C-ABI FFI 导出 ============================

use std::ffi::{c_char, CStr};

/// 返回 C 字符串（调用方用 `sb_rust_free_cstr` 释放）。
#[no_mangle]
pub extern "C" fn sb_rust_version() -> *const c_char {
    let cs = std::ffi::CString::new("sb-lxf-rust-accel 0.1.0").expect("no interior NUL");
    let raw = cs.into_raw();
    // 转移所有权给调用方；调用方须通过 sb_rust_free_cstr 释放
    raw as *const c_char
}

/// 释放 `sb_rust_version` 返回的 C 字符串。
#[no_mangle]
pub extern "C" fn sb_rust_free_cstr(s: *const c_char) {
    if !s.is_null() {
        unsafe { drop(std::ffi::CString::from_raw(s as *mut c_char)) };
    }
}

/// 创建 LPM 表，返回句柄。
#[no_mangle]
pub extern "C" fn sb_rust_lpm_new() -> *mut LpmTable {
    Box::into_raw(Box::new(LpmTable::new()))
}

/// 释放 LPM 表。
#[no_mangle]
pub extern "C" fn sb_rust_lpm_free(table: *mut LpmTable) {
    if !table.is_null() {
        unsafe { drop(Box::from_raw(table)) };
    }
}

/// 插入一条 CIDR 规则。`cidr` 为 NUL 结尾 C 字符串（如 "10.0.0.0/8"）。
/// 返回 0 成功；-1 参数非法。
#[no_mangle]
pub extern "C" fn sb_rust_lpm_insert_cidr(
    table: *mut LpmTable,
    cidr: *const c_char,
    result: i32,
) -> i32 {
    if table.is_null() || cidr.is_null() {
        return -1;
    }
    let s = match unsafe { CStr::from_ptr(cidr) }.to_str() {
        Ok(s) => s,
        Err(_) => return -1,
    };
    match catch_unwind(AssertUnwindSafe(|| {
        let (ip, mask, is_v6) = parse_cidr(s).expect("parse");
        let t = unsafe { &mut *table };
        if is_v6 {
            t.insert_v6(ip, mask, result);
        } else {
            t.insert_v4(ip as u32, mask, result);
        }
    })) {
        Ok(()) => 0,
        Err(_) => -1,
    }
}

/// 匹配 IPv4。`ip` 为 NUL 结尾 C 字符串。命中返回结果 id，否则 -1。
#[no_mangle]
pub extern "C" fn sb_rust_lpm_match_v4(table: *const LpmTable, ip: *const c_char) -> i32 {
    if table.is_null() || ip.is_null() {
        return -1;
    }
    let s = match unsafe { CStr::from_ptr(ip) }.to_str() {
        Ok(s) => s,
        Err(_) => return -1,
    };
    let t = unsafe { &*table };
    match catch_unwind(AssertUnwindSafe(|| {
        let v4 = parse_ipv4(s).unwrap_or(u32::MAX);
        t.match_v4(v4)
    })) {
        Ok(r) => r,
        Err(_) => -1,
    }
}

/// 匹配 IPv6。
#[no_mangle]
pub extern "C" fn sb_rust_lpm_match_v6(table: *const LpmTable, ip: *const c_char) -> i32 {
    if table.is_null() || ip.is_null() {
        return -1;
    }
    let s = match unsafe { CStr::from_ptr(ip) }.to_str() {
        Ok(s) => s,
        Err(_) => return -1,
    };
    let t = unsafe { &*table };
    match catch_unwind(AssertUnwindSafe(|| {
        let v6 = parse_ipv6(s).unwrap_or(u128::MAX);
        t.match_v6(v6)
    })) {
        Ok(r) => r,
        Err(_) => -1,
    }
}

/// FNV-1a 64 位。`data`/`len` 指向字节缓冲区。
#[no_mangle]
pub extern "C" fn sb_rust_fnv1a_64(data: *const u8, len: usize) -> u64 {
    if data.is_null() {
        return 0;
    }
    let slice = unsafe { std::slice::from_raw_parts(data, len) };
    fnv1a_64(slice)
}

/// 轻量 64 位哈希（分散度好，非加密）。
#[no_mangle]
pub extern "C" fn sb_rust_quick_hash64(data: *const u8, len: usize) -> u64 {
    if data.is_null() {
        return 0;
    }
    let slice = unsafe { std::slice::from_raw_parts(data, len) };
    quick_hash64(slice)
}

// 保证链接器可见（防止符号被裁）
#[no_mangle]
pub extern "C" fn sb_rust_symbol_version() -> u32 {
    0x0000_0100
}

// ============================ 单元测试 ============================

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn lpm_v4_basic() {
        let mut t = LpmTable::new();
        t.insert_v4(parse_ipv4("10.0.0.0").unwrap(), 8, 1);
        t.insert_v4(parse_ipv4("10.1.0.0").unwrap(), 16, 2);
        t.insert_v4(parse_ipv4("8.8.8.8").unwrap(), 32, 3);
        assert_eq!(t.match_v4(parse_ipv4("10.1.2.3").unwrap()), 2);
        assert_eq!(t.match_v4(parse_ipv4("10.9.9.9").unwrap()), 1);
        assert_eq!(t.match_v4(parse_ipv4("8.8.8.8").unwrap()), 3);
        assert_eq!(t.match_v4(parse_ipv4("8.8.8.9").unwrap()), -1);
    }

    #[test]
    fn lpm_cidr_parse() {
        let (_ip, m, is6) = parse_cidr("192.168.0.0/16").unwrap();
        assert!(!is6 && m == 16);
        let (_, m2, is62) = parse_cidr("2001:db8::/32").unwrap();
        assert!(is62 && m2 == 32);
    }

    #[test]
    fn fnv_deterministic() {
        assert_eq!(fnv1a_64(b"abc"), fnv1a_64(b"abc"));
        assert_ne!(fnv1a_64(b"abc"), fnv1a_64(b"abd"));
        assert_eq!(fnv1a_64(b""), 0xcbf2_9ce4_8422_2325);
    }

    #[test]
    fn parse_v6() {
        assert_eq!(parse_ipv6("::").unwrap(), 0);
        assert_eq!(
            parse_ipv6("2001:db8::1").unwrap(),
            0x2001_0db8_0000_0000_0000_0000_0000_0001
        );
        assert_eq!(
            parse_ipv6("fe80::ff").unwrap(),
            0xfe80_0000_0000_0000_0000_0000_0000_00ff
        );
        assert!(parse_ipv6("1.2.3.4").is_none());
    }
}

