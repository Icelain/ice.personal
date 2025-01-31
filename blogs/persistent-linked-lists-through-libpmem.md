---

date: 2025/01/31

---

# Understanding libpmem for Building Persistent Linked Lists

## Core Functions Overview

Based on [the official libpmem documentation](https://pmem.io/pmdk/libpmem)

### 1. pmem_map_file()
- Maps a persistent memory file into memory
- Used in our linked list's `init_pmem()` function
- Advantages:
  - Finds optimal address for large page mappings
  - Returns `is_pmem` flag to identify true persistent memory
  - Handles file creation and mapping in one call

### 2. pmem_persist()
- Ensures data is stored durably in persistent memory
- Used in our linked list for:
  - Node creation
  - List metadata updates
  - Pointer updates
- Most optimal way to flush changes for true pmem
- Performs flush directly from user space when possible

### 3. pmem_msync()
- Wrapper around standard msync()
- Used as fallback when memory isn't true pmem(say we provide the pmem code a regular linux file instead of an nvram device file <-> which is what we're doing)
- Ensures argument alignment per POSIX requirements
- Used in our list's non-pmem code path
- Slower than pmem_persist()

---

## Memory Operations

### 1. pmem_memcpy_persist()
- Optimized version of memcpy for persistent memory
- Uses non-temporal store instructions on Intel platforms
- Bypasses processor caches
- Could be used to optimize our linked list's data copying

### 2. pmem_is_pmem()
- Checks if memory range is true persistent memory
- High overhead - should cache result
- Used in our list's initialization to determine flush strategy

---

## Advanced Flush Control

### 1. pmem_flush()
- First step of persistence: flushes processor caches
- Can be used for fine-grained control
- May be empty function on platforms with eADR(no clue what this is just copied from the docs)

### 2. pmem_drain()
- Second step: ensures hardware buffers are drained
- System-wide operation
- Can be deferred when doing multiple flushes

---

## Best Practices from Our Linked List

1. **Initialization**
```c
pmem_addr = pmem_map_file(PMEM_PATH, PMEM_SIZE,
                         PMEM_FILE_CREATE, 0666,
                         &mapped_len, &is_pmem);
```

2. **Data Persistence**
```c
pmem_persist(list_meta, sizeof(ListMeta));
```

3. **Cleanup**
```c
pmem_unmap(pmem_addr, mapped_len);
```

---

## Memory Ordering Considerations

1. **Persistence is Not Visibility**
- CPU barriers (SFENCE) handle thread visibility
- Only pmem functions guarantee persistence

2. **Write Ordering**
- No guaranteed order without proper flushing
- Must use pmem functions for durability

3. **Atomicity**
- Individual writes up to 8 bytes may be atomic
- Larger operations need explicit handling

---

## Optimization Tips

1. **Batch Operations**
- Group multiple updates before calling pmem_persist()
- Reduces flush overhead but increases time delay for consistency guarantee

2. **Memory Layout**
- Align data structures for optimal performance
- Consider access patterns

3. **Error Handling**
- Always check return values
- Implement proper cleanup on failures

---

## Common Pitfalls

1. **Assuming All Memory is Persistent**
- Always check is_pmem flag
- Use appropriate flush strategy

2. **Unnecessary Flushes**
- Don't flush read-only data
- Batch updates when possible

3. **Incomplete Persistence**
- Ensure all critical data is flushed
- Don't forget pointer updates
