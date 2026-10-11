/* sb-lxf Rust 加速模块 C ABI 头（由 Rust cdylib 导出，cgo 链接） */
#ifndef SB_LXF_RUST_ACCEL_H
#define SB_LXF_RUST_ACCEL_H

#include <stdint.h>
#include <stddef.h>

/* 版本字符串（调用方用 sb_rust_free_cstr 释放） */
const char* sb_rust_version(void);
void sb_rust_free_cstr(const char* s);

/* LPM 表句柄 */
void* sb_rust_lpm_new(void);
void sb_rust_lpm_free(void* table);
int32_t sb_rust_lpm_insert_cidr(void* table, const char* cidr, int32_t result);
int32_t sb_rust_lpm_match_v4(const void* table, const char* ip);
int32_t sb_rust_lpm_match_v6(const void* table, const char* ip);

/* 轻量哈希 */
uint64_t sb_rust_fnv1a_64(const uint8_t* data, size_t len);
uint64_t sb_rust_quick_hash64(const uint8_t* data, size_t len);

/* 模块 ABI 版本（用于运行时探测 / 降级） */
uint32_t sb_rust_symbol_version(void);

#endif
