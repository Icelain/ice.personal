---

date: 2025/01/31

---

# Understanding libpmem for Persistent Memory Programming

Persistent memory (PMEM) represents a fundamental shift in how we approach data persistence. Sitting at the intersection of storage and memory, PMEM offers byte-addressable access with near-DRAM speeds while maintaining data across power cycles. This transformative technology demands new programming models since traditional paradigms for volatile memory or block storage don't fully capture its unique characteristics.

The **libpmem** library provides a foundation for effectively utilizing persistent memory in applications. This post explores the essential functions and patterns needed to develop robust persistent memory applications, highlighting key considerations when transitioning from traditional memory management to persistent memory programming.

This blog post provides a practical guide to libpmem functions through **C** with examples drawn from real-world persistent memory patterns.

## Table of Contents

1. [Core Functions Overview](#core-functions-overview)
2. [Memory Operations](#memory-operations)
3. [Advanced Flush Control](#advanced-flush-control)
4. [Best Practices for Persistent Memory](#best-practices-for-persistent-memory)
5. [Memory Ordering Considerations](#memory-ordering-considerations)
6. [Optimization Tips](#optimization-tips)
7. [Common Pitfalls](#common-pitfalls)

## Core Functions Overview

Based on [the official libpmem documentation](https://pmem.io/pmdk/libpmem), let's explore the essential functions:

### 1. pmem_map_file()

```c
void *pmem_map_file(const char *path, size_t len, int flags,
                    mode_t mode, size_t *mapped_lenp, int *is_pmemp);
```

This function maps a persistent memory file into memory and serves as the entry point for most PMEM applications. It offers several advantages:

- Finds optimal address for large page mappings
- Returns `is_pmem` flag to identify true persistent memory
- Handles file creation and mapping in one call

### 2. pmem_persist()

```c
void pmem_persist(const void *addr, size_t len);
```

This function ensures data is stored durably in persistent memory. In any PMEM application, you'll use it for:

- Critical data structures
- Metadata updates
- Pointer modifications

It's the most optimal way to flush changes for true pmem and performs flush directly from user space when possible.

### 3. pmem_msync()

```c
int pmem_msync(const void *addr, size_t len);
```

This is a wrapper around standard msync() and serves as a fallback when memory isn't true pmem (such as when testing with regular files instead of actual PMEM devices). It ensures argument alignment per POSIX requirements and is used for portability, though it's slower than pmem_persist().

## Memory Operations

### 1. pmem_memcpy_persist()

```c
void *pmem_memcpy_persist(void *pmemdest, const void *src, size_t len);
```

This function provides an optimized version of memcpy for persistent memory. It uses non-temporal store instructions on Intel platforms and bypasses processor caches, combining the copy and persistence operations in one efficient call.

### 2. pmem_is_pmem()

```c
int pmem_is_pmem(const void *addr, size_t len);
```

This function checks if a memory range is true persistent memory. It has high overhead, so you should cache the result. Applications typically call this during initialization to determine the appropriate flush strategy for the runtime environment.

## Advanced Flush Control

### 1. pmem_flush()

```c
void pmem_flush(const void *addr, size_t len);
```

This is the first step of persistence: flushing processor caches. It can be used for fine-grained control and may be an empty function on platforms with eADR (Extended ADR, a hardware feature that guarantees cache persistence).

### 2. pmem_drain()

```c
void pmem_drain(void);
```

This is the second step: ensuring hardware buffers are drained. It's a system-wide operation that can be deferred when doing multiple flushes, allowing for optimization in batch operations.

## Best Practices for Persistent Memory

1. **Initialization Pattern**
```c
pmem_addr = pmem_map_file(PMEM_PATH, PMEM_SIZE,
                         PMEM_FILE_CREATE, 0666,
                         &mapped_len, &is_pmem);
```

2. **Data Persistence Pattern**
```c
pmem_persist(critical_data, sizeof(critical_data));
```

3. **Clean Shutdown**
```c
pmem_unmap(pmem_addr, mapped_len);
```

## Memory Ordering Considerations

1. **Persistence is Not Visibility**
   - CPU barriers (SFENCE) handle thread visibility
   - Only pmem functions guarantee persistence

2. **Write Ordering**
   - No guaranteed order without proper flushing
   - Must use pmem functions for durability

3. **Atomicity**
   - Individual writes up to 8 bytes may be atomic
   - Larger operations need explicit handling through techniques like logging

## Optimization Tips

1. **Batch Operations**
   - Group multiple updates before calling pmem_persist()
   - Reduces flush overhead but increases time delay for consistency guarantee

2. **Memory Layout**
   - Align data structures for optimal performance
   - Consider cache line boundaries (typically 64 bytes)
   - Keep related data together to minimize flush operations

3. **Error Handling**
   - Always check return values
   - Implement proper cleanup on failures

## Common Pitfalls

1. **Assuming All Memory is Persistent**
   - Always check is_pmem flag
   - Use appropriate flush strategy for the detected memory type

2. **Unnecessary Flushes**
   - Don't flush read-only data
   - Batch updates when possible
   - Be aware of automatic flushes from library calls

3. **Incomplete Persistence**
   - Ensure all critical data is flushed
   - Don't forget metadata and pointer updates
   - Consider transactions for complex data structure modifications

This overview provides the foundation for developing with persistent memory using libpmem. The programming model requires careful consideration of persistence boundaries, ordering, and failure recovery - but the performance and architectural benefits make it worthwhile for many applications.
