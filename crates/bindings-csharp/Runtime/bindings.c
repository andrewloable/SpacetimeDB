#include <assert.h>
// #include <mono/metadata/appdomain.h>
// #include <mono/metadata/object.h>
#include <stdbool.h>
#include <stdint.h>
#include <unistd.h>

#if !defined(EXPERIMENTAL_WASM_AOT) && !defined(DOTNET_WASI_P2)
#include "driver.h"
#elif defined(DOTNET_WASI_P2)
// .NET 10+: Mono headers are available but mono_wasm_invoke_method_ref is gone.
// Use mono_runtime_invoke instead.
#include <mono/metadata/assembly.h>
#include <mono/metadata/loader.h>
#include <mono/metadata/object.h>

// These functions are still available in .NET 10 wasi-wasm runtime.
MonoAssembly *mono_wasm_assembly_load(const char *name);
MonoClass *mono_wasm_assembly_find_class(MonoAssembly *assembly,
                                         const char *namespace_name,
                                         const char *name);
MonoMethod *mono_wasm_assembly_find_method(MonoClass *klass, const char *name,
                                           int arguments);
#endif

#define OPAQUE_TYPEDEF(name, T) \
  typedef struct name {         \
    T inner;                    \
  } name

OPAQUE_TYPEDEF(Status, uint16_t);
OPAQUE_TYPEDEF(TableId, uint32_t);
OPAQUE_TYPEDEF(IndexId, uint32_t);
OPAQUE_TYPEDEF(ColId, uint16_t);
OPAQUE_TYPEDEF(IndexType, uint8_t);
OPAQUE_TYPEDEF(LogLevel, uint8_t);
OPAQUE_TYPEDEF(BytesSink, uint32_t);
OPAQUE_TYPEDEF(BytesSource, uint32_t);
OPAQUE_TYPEDEF(RowIter, uint32_t);
OPAQUE_TYPEDEF(ConsoleTimerId, uint32_t);

#define CSTR(s) (uint8_t*)s, sizeof(s) - 1

#define STDB_EXTERN(name) \
  __attribute__((import_module(SPACETIME_MODULE_VERSION), import_name(#name))) extern

#if !defined(EXPERIMENTAL_WASM_AOT)
// Create wrapper functions so the linker sees C-level usage of the imports
// and doesn't garbage-collect them. The _imp suffix is the actual WASM import,
// and the wrapper is what P/Invoke resolves to.
// export_name prevents wasm-opt from removing these "unused" wrappers.
// They are called from C# P/Invoke at runtime (invisible to static analysis).
#define IMPORT(ret, name, params, args)                        \
  STDB_EXTERN(name) ret name##_imp params;                     \
  __attribute__((export_name(#name))) ret name params {        \
    return name##_imp args;                                    \
  }
#else
#define IMPORT(ret, name, params, args) STDB_EXTERN(name) ret name params;
#endif

#define SPACETIME_MODULE_VERSION "spacetime_10.0"
IMPORT(Status, table_id_from_name,
       (const uint8_t* name, uint32_t name_len, TableId* id),
       (name, name_len, id));
IMPORT(Status, index_id_from_name,
       (const uint8_t* name, uint32_t name_len, IndexId* id),
       (name, name_len, id));
IMPORT(Status, datastore_table_row_count,
       (TableId table_id, uint64_t* count),
       (table_id, count));
IMPORT(Status, datastore_table_scan_bsatn,
       (TableId table_id, RowIter* iter),
       (table_id, iter));
IMPORT(Status, datastore_index_scan_range_bsatn,
       (IndexId index_id, const uint8_t* prefix, uint32_t prefix_len, ColId prefix_elems,
        const uint8_t* rstart, uint32_t rstart_len, const uint8_t* rend, uint32_t rend_len, RowIter* iter),
       (index_id, prefix, prefix_len, prefix_elems, rstart, rstart_len, rend, rend_len, iter));
IMPORT(Status, datastore_btree_scan_bsatn,
       (IndexId index_id, const uint8_t* prefix, uint32_t prefix_len, ColId prefix_elems,
        const uint8_t* rstart, uint32_t rstart_len, const uint8_t* rend, uint32_t rend_len, RowIter* iter),
       (index_id, prefix, prefix_len, prefix_elems, rstart, rstart_len, rend, rend_len, iter));
IMPORT(int16_t, row_iter_bsatn_advance,
       (RowIter iter, uint8_t* buffer_ptr, size_t* buffer_len_ptr),
       (iter, buffer_ptr, buffer_len_ptr));
IMPORT(uint16_t, row_iter_bsatn_close, (RowIter iter), (iter));
IMPORT(Status, datastore_insert_bsatn, (TableId table_id, uint8_t* row_ptr, size_t* row_len_ptr),
       (table_id, row_ptr, row_len_ptr));
IMPORT(Status, datastore_update_bsatn, (TableId table_id, IndexId index_id, uint8_t* row_ptr, size_t* row_len_ptr),
       (table_id, index_id, row_ptr, row_len_ptr));
IMPORT(Status, datastore_delete_by_index_scan_range_bsatn,
       (IndexId index_id, const uint8_t* prefix, uint32_t prefix_len, ColId prefix_elems,
        const uint8_t* rstart, uint32_t rstart_len, const uint8_t* rend, uint32_t rend_len, uint32_t* num_deleted),
       (index_id, prefix, prefix_len, prefix_elems, rstart, rstart_len, rend, rend_len, num_deleted));
IMPORT(Status, datastore_delete_by_btree_scan_bsatn,
       (IndexId index_id, const uint8_t* prefix, uint32_t prefix_len, ColId prefix_elems,
        const uint8_t* rstart, uint32_t rstart_len, const uint8_t* rend, uint32_t rend_len, uint32_t* num_deleted),
       (index_id, prefix, prefix_len, prefix_elems, rstart, rstart_len, rend, rend_len, num_deleted));
IMPORT(Status, datastore_delete_all_by_eq_bsatn,
       (TableId table_id, const uint8_t* rel_ptr, uint32_t rel_len,
        uint32_t* num_deleted),
       (table_id, rel_ptr, rel_len, num_deleted));
IMPORT(int16_t, bytes_source_read, (BytesSource source, uint8_t* buffer_ptr, size_t* buffer_len_ptr),
       (source, buffer_ptr, buffer_len_ptr));
IMPORT(uint16_t, bytes_sink_write, (BytesSink sink, const uint8_t* buffer_ptr, size_t* buffer_len_ptr),
       (sink, buffer_ptr, buffer_len_ptr));
IMPORT(void, console_log,
       (LogLevel level, const uint8_t* target_ptr, uint32_t target_len,
        const uint8_t* filename_ptr, uint32_t filename_len, uint32_t line_number,
        const uint8_t* message_ptr, uint32_t message_len),
       (level, target_ptr, target_len, filename_ptr, filename_len, line_number,
        message_ptr, message_len));
IMPORT(ConsoleTimerId, console_timer_start,
       (const uint8_t* name, size_t name_len),
       (name, name_len));
IMPORT(Status, console_timer_end,
       (ConsoleTimerId stopwatch_id),
       (stopwatch_id));
IMPORT(void, volatile_nonatomic_schedule_immediate,
       (const uint8_t* name, size_t name_len, const uint8_t* args, size_t args_len),
       (name, name_len, args, args_len));
IMPORT(void, identity, (void* id_ptr), (id_ptr));
#undef SPACETIME_MODULE_VERSION

#define SPACETIME_MODULE_VERSION "spacetime_10.4"
IMPORT(Status, datastore_index_scan_point_bsatn,
       (IndexId index_id, const uint8_t* point, uint32_t point_len, RowIter* iter),
       (index_id, point, point_len, iter));
IMPORT(Status, datastore_delete_by_index_scan_point_bsatn,
       (IndexId index_id, const uint8_t* point, uint32_t point_len, uint32_t* num_deleted),
       (index_id, point, point_len, num_deleted));
#undef SPACETIME_MODULE_VERSION

#define SPACETIME_MODULE_VERSION "spacetime_10.1"
IMPORT(int16_t, bytes_source_remaining_length, (BytesSource source, uint32_t* out), (source, out));
#undef SPACETIME_MODULE_VERSION

#define SPACETIME_MODULE_VERSION "spacetime_10.2"
IMPORT(int16_t, get_jwt, (const uint8_t* connection_id_ptr, BytesSource* bytes_ptr), (connection_id_ptr, bytes_ptr));
#undef SPACETIME_MODULE_VERSION

#define SPACETIME_MODULE_VERSION "spacetime_10.3"
IMPORT(uint16_t, procedure_start_mut_tx, (int64_t* micros), (micros));
IMPORT(uint16_t, procedure_commit_mut_tx, (void), ());
IMPORT(uint16_t, procedure_abort_mut_tx, (void), ());
IMPORT(uint16_t, procedure_http_request,
       (const uint8_t* request_ptr, uint32_t request_len,
        const uint8_t* body_ptr, uint32_t body_len,
        BytesSource* out),
       (request_ptr, request_len, body_ptr, body_len, out));
#undef SPACETIME_MODULE_VERSION

#if !defined(EXPERIMENTAL_WASM_AOT)
static MonoClass* ffi_class;

#define CEXPORT(name) __attribute__((export_name(#name))) name

#define PREINIT(priority, name) void CEXPORT(__preinit__##priority##_##name)()

PREINIT(10, startup) {
#if !defined(DOTNET_WASI_P2)
  // .NET 8: mono_wasm_load_runtime alone isn't enough because it doesn't
  // reach the assembly with Main, so module descriptor remains unpopulated.
  // Invoke actual _start instead.
  extern void _start();
  _start();
#else
  // .NET 10+: _start() calls main() which runs to completion and calls exit().
  // We need to:
  // 1. Initialize the Mono runtime
  // 2. Load and run the entry assembly (to trigger static initializers)
  //    but without letting main() exit the process.
  extern int initialize_runtime();
  initialize_runtime();

  // Load the entry point assembly and run its Main method.
  // This triggers the [ModuleInitializer] generated code that registers
  // tables and reducers with the SpacetimeDB runtime.
  extern const char* dotnet_wasi_getentrypointassemblyname();
  extern MonoMethod* mono_wasi_assembly_get_entry_point(MonoAssembly *assembly);
  // stdb_run_main wraps mono_runtime_run_main, which properly handles
  // Main(string[] args) and triggers module initializers.
  extern int stdb_run_main(MonoMethod*, int, char*[], MonoObject**);

  const char* entry_name = dotnet_wasi_getentrypointassemblyname();
  MonoAssembly* entry_asm = mono_wasm_assembly_load(entry_name);
  if (!entry_asm) {
    entry_asm = mono_assembly_open(entry_name, NULL);
  }
  assert(entry_asm && "Failed to load entry assembly");

  // Run Main() via mono_runtime_run_main — this properly constructs
  // the string[] args parameter, triggers [ModuleInitializer] in the
  // user assembly, and returns without calling exit().
  MonoMethod* entry_method = mono_wasi_assembly_get_entry_point(entry_asm);
  if (entry_method) {
    MonoObject* exc = NULL;
    char* empty_argv[] = { (char*)entry_name, NULL };
    stdb_run_main(entry_method, 0, empty_argv, &exc);
    // Ignore any exception - Main() for SpacetimeDB modules is typically empty
  }
#endif

  MonoAssembly* assembly = mono_wasm_assembly_load("SpacetimeDB.Runtime");
  if (!assembly) {
    assembly = mono_wasm_assembly_load("SpacetimeDB.Runtime.dll");
  }
  ffi_class = assembly ? mono_wasm_assembly_find_class(
      assembly, "SpacetimeDB.Internal", "Module") : NULL;
  assert(ffi_class &&
         "FFI export class (SpacetimeDB.Internal.Module) not found");
}

#if defined(DOTNET_WASI_P2)
// .NET 10+: mono_wasm_invoke_method_ref was removed. Use stdb_mono_invoke
// which is a thin wrapper around mono_runtime_invoke, compiled from
// stdb_mono_invoke.c — that file is linked before the static archives so
// mono_runtime_invoke gets extracted from libmonosgen-2.0.a.
extern MonoObject* stdb_mono_invoke(MonoMethod *method, void *obj,
                                    void **params, MonoObject **exc);

#define EXPORT_WITH_MONO_RES(ret, res_code, name, params, args...)            \
  static MonoMethod* ffi_method_##name;                                       \
  PREINIT(20, find_##name) {                                                  \
    ffi_method_##name = mono_wasm_assembly_find_method(ffi_class, #name, -1); \
    assert(ffi_method_##name && "FFI export method not found");               \
  }                                                                           \
  ret CEXPORT(name) params {                                                  \
    MonoObject* exc = NULL;                                                   \
    MonoObject* res = stdb_mono_invoke(ffi_method_##name, NULL,               \
                                       (void*[]){args}, &exc);                \
    res_code                                                                  \
  }
#else
// .NET 8: Use mono_wasm_invoke_method_ref
#define EXPORT_WITH_MONO_RES(ret, res_code, name, params, args...)            \
  static MonoMethod* ffi_method_##name;                                       \
  PREINIT(20, find_##name) {                                                  \
    ffi_method_##name = mono_wasm_assembly_find_method(ffi_class, #name, -1); \
    assert(ffi_method_##name && "FFI export method not found");               \
  }                                                                           \
  ret CEXPORT(name) params {                                                  \
    MonoObject* res;                                                          \
    mono_wasm_invoke_method_ref(ffi_method_##name, NULL, (void*[]){args},     \
                                NULL, &res);                                  \
    res_code                                                                  \
  }
#endif

#define EXPORT(ret, name, params, args...)                                             \
  EXPORT_WITH_MONO_RES(ret, return *(ret*)mono_object_unbox(res);, name, params, args) \

#define EXPORT_VOID(name, params, args...)                                    \
  EXPORT_WITH_MONO_RES(void, return;, name, params, args)                      \

EXPORT_VOID(__describe_module__, (BytesSink description), &description);

EXPORT(int16_t, __call_reducer__,
       (uint32_t id,
        uint64_t sender_0, uint64_t sender_1, uint64_t sender_2, uint64_t sender_3,
        uint64_t conn_id_0, uint64_t conn_id_1,
        uint64_t timestamp, BytesSource args, BytesSink error),
       &id,
       &sender_0, &sender_1, &sender_2, &sender_3,
       &conn_id_0, &conn_id_1,
       &timestamp, &args, &error);

EXPORT(int16_t, __call_procedure__,
       (uint32_t id,
        uint64_t sender_0, uint64_t sender_1, uint64_t sender_2, uint64_t sender_3,
        uint64_t conn_id_0, uint64_t conn_id_1,
        uint64_t timestamp, BytesSource args, BytesSink result_sink),
       &id,
       &sender_0, &sender_1, &sender_2, &sender_3,
       &conn_id_0, &conn_id_1,
       &timestamp, &args, &result_sink);

EXPORT(int16_t, __call_view__,
       (uint32_t id,
        uint64_t sender_0, uint64_t sender_1, uint64_t sender_2, uint64_t sender_3,
        BytesSource args, BytesSink rows),
       &id,
       &sender_0, &sender_1, &sender_2, &sender_3,
       &args, &rows);

EXPORT(int16_t, __call_view_anon__,
       (uint32_t id, BytesSource args, BytesSink rows),
       &id, &args, &rows);
#endif

// Shims to avoid dependency on WASI in the generated Wasm file.

#include <stdlib.h>
#include <wasi/api.h>

// Ignore warnings about anonymous parameters, this is to avoid having
// to write `int arg0`, `int arg1`, etc. for every function.
#pragma clang diagnostic ignored "-Wc2x-extensions"

// Based on
// https://github.com/WebAssembly/wasi-libc/blob/main/libc-bottom-half/sources/__wasilibc_real.c,

#define WASI_NAME(name) __imported_wasi_snapshot_preview1_##name

// Shim for WASI calls that always unconditionally succeeds.
// This is suitable for most (but not all) WASI functions used by .NET.
#define WASI_SHIM(name, params) \
  int32_t WASI_NAME(name) params { return 0; }

WASI_SHIM(environ_get, (int32_t, int32_t));
WASI_SHIM(environ_sizes_get, (int32_t, int32_t));
WASI_SHIM(clock_time_get, (int32_t, int64_t, int32_t));
WASI_SHIM(fd_advise, (int32_t, int64_t, int64_t, int32_t));
WASI_SHIM(fd_allocate, (int32_t, int64_t, int64_t));
WASI_SHIM(fd_close, (int32_t));
WASI_SHIM(fd_datasync, (int32_t));
WASI_SHIM(fd_fdstat_get, (int32_t, int32_t));
WASI_SHIM(fd_fdstat_set_flags, (int32_t, int32_t));
WASI_SHIM(fd_fdstat_set_rights, (int32_t, int64_t, int64_t));
WASI_SHIM(fd_filestat_get, (int32_t, int32_t));
WASI_SHIM(fd_filestat_set_size, (int32_t, int64_t));
WASI_SHIM(fd_filestat_set_times, (int32_t, int64_t, int64_t, int32_t));
WASI_SHIM(fd_pread, (int32_t, int32_t, int32_t, int64_t, int32_t));
WASI_SHIM(fd_prestat_dir_name, (int32_t, int32_t, int32_t));
WASI_SHIM(fd_pwrite, (int32_t, int32_t, int32_t, int64_t, int32_t));
WASI_SHIM(fd_read, (int32_t, int32_t, int32_t, int32_t));
WASI_SHIM(fd_readdir, (int32_t, int32_t, int32_t, int64_t, int32_t));
WASI_SHIM(fd_renumber, (int32_t, int32_t));
WASI_SHIM(fd_seek, (int32_t, int64_t, int32_t, int32_t));
WASI_SHIM(fd_sync, (int32_t));
WASI_SHIM(fd_tell, (int32_t, int32_t));
WASI_SHIM(path_create_directory, (int32_t, int32_t, int32_t));
WASI_SHIM(path_filestat_get, (int32_t, int32_t, int32_t, int32_t, int32_t));
WASI_SHIM(path_filestat_set_times,
          (int32_t, int32_t, int32_t, int32_t, int64_t, int64_t, int32_t));
WASI_SHIM(path_link,
          (int32_t, int32_t, int32_t, int32_t, int32_t, int32_t, int32_t));
WASI_SHIM(path_open, (int32_t, int32_t, int32_t, int32_t, int32_t, int64_t,
                      int64_t, int32_t, int32_t));
WASI_SHIM(path_readlink,
          (int32_t, int32_t, int32_t, int32_t, int32_t, int32_t));
WASI_SHIM(path_remove_directory, (int32_t, int32_t, int32_t));
WASI_SHIM(path_rename, (int32_t, int32_t, int32_t, int32_t, int32_t, int32_t));
WASI_SHIM(path_symlink, (int32_t, int32_t, int32_t, int32_t, int32_t));
WASI_SHIM(path_unlink_file, (int32_t, int32_t, int32_t));
WASI_SHIM(poll_oneoff, (int32_t, int32_t, int32_t, int32_t));
WASI_SHIM(sched_yield, ());
WASI_SHIM(random_get, (int32_t, int32_t));
WASI_SHIM(sock_accept, (int32_t, int32_t, int32_t));
WASI_SHIM(sock_recv, (int32_t, int32_t, int32_t, int32_t, int32_t, int32_t));
WASI_SHIM(sock_send, (int32_t, int32_t, int32_t, int32_t, int32_t));
WASI_SHIM(sock_shutdown, (int32_t, int32_t));

// Mono retrieves executable name via argv[0], so we need to shim it with
// some dummy name instead of returning an empty argv[] array to avoid
// assertion failures.
const char executable_name[] = "stdb.wasm";

int32_t WASI_NAME(args_sizes_get)(__wasi_size_t* argc,
                                  __wasi_size_t* argv_buf_size) {
  *argc = 1;
  *argv_buf_size = sizeof(executable_name);
  return 0;
}

int32_t WASI_NAME(args_get)(uint8_t** argv, uint8_t* argv_buf) {
  argv[0] = argv_buf;
  __builtin_memcpy(argv_buf, executable_name, sizeof(executable_name));
  return 0;
}

// Clock resolution should be non-zero.
int32_t WASI_NAME(clock_res_get)(int32_t, uint64_t* timestamp) {
  *timestamp = 1;
  return 0;
}

// For `fd_write`, we need to at least collect and report sum of sizes.
// If we report size 0, the caller will assume that the write failed and will
// try again, which will result in an infinite loop.
int32_t WASI_NAME(fd_write)(__wasi_fd_t fd, const __wasi_ciovec_t* iovs,
                            size_t iovs_len, __wasi_size_t* retptr0) {
  for (size_t i = 0; i < iovs_len; i++) {
    // Note: this will produce ugly broken output, but there's not much we can
    // do about it until we have proper line-buffered WASI writer in the core.
    // It's better than nothing though.
    console_log((LogLevel){fd == STDERR_FILENO ? /*WARN*/ 1 : /*INFO*/
                                2},
                 CSTR("wasi"), CSTR(__FILE__), __LINE__, iovs[i].buf,
                 iovs[i].buf_len);
    *retptr0 += iovs[i].buf_len;
  }
  return 0;
}

// BADF indicates end of iteration for preopens; we must return it instead of
// "success" to prevent infinite loop.
int32_t WASI_NAME(fd_prestat_get)(int32_t, int32_t) {
  return __WASI_ERRNO_BADF;
}

// Actually exit runtime on `proc_exit`.
_Noreturn void WASI_NAME(proc_exit)(int32_t code) { exit(code); }

// There is another rogue import of sock_accept somewhere in .NET that doesn't
// match the scheme above.
// Maybe this one?
// https://github.com/dotnet/runtime/blob/085ddb7f9b26f01ae1b6842db7eacb6b4042e031/src/mono/mono/component/mini-wasi-debugger.c#L12-L14

int32_t sock_accept(int32_t, int32_t, int32_t) { return 0; }
