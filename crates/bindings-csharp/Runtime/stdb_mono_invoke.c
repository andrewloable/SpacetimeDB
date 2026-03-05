// This file forces mono_runtime_invoke to be pulled from libmonosgen-2.0.a
// during static linking. Without this, bindings.c (compiled as NativeFileReference
// and linked last) would reference the symbol too late for archive extraction.
//
// Only compiled for .NET 10+ builds (guarded by MSBuild target condition).
#include <mono/metadata/object.h>

// Forward declare and reference to force linkage
extern MonoObject* mono_runtime_invoke(MonoMethod *method, void *obj,
                                       void **params, MonoObject **exc);

extern int mono_runtime_run_main(MonoMethod *method, int argc,
                                 char *argv[], MonoObject **exc);

// Wrapper that bindings.c calls — this file is compiled with the main native
// sources and linked before the static archives.
MonoObject* stdb_mono_invoke(MonoMethod *method, void *obj,
                             void **params, MonoObject **exc) {
    return mono_runtime_invoke(method, obj, params, exc);
}

// Wrapper for mono_runtime_run_main — properly handles Main(string[] args)
// and triggers [ModuleInitializer] code in the entry assembly.
int stdb_run_main(MonoMethod *method, int argc, char *argv[], MonoObject **exc) {
    return mono_runtime_run_main(method, argc, argv, exc);
}
