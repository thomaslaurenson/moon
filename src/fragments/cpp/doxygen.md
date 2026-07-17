# C++ API documentation

Doxygen conventions for a library's public API. Assumes the C++ style fragment.

A library's headers under `include/<lib>/` are a contract with people who cannot read your implementation and will not ask you what a parameter means. Everything declared there is documented. Internal headers under `src/` are read by people who can see the code next to them, so they are documented where the code is not self-evident, not as a matter of course.

## Where the comment goes

- Every function, class, struct, and enum declared in a public header gets a Doxygen comment on its declaration.
- A static or private helper defined only in a `.cpp` has its comment in the `.cpp`.
- Never duplicate the comment in both the header and the implementation; the declaration owns it, and a copy in the `.cpp` is a copy that goes stale.

## Format

Use `///` triple-slash. The first line is a single summary sentence with no full stop. Document every parameter with `@param`, and include `@return` unless the function returns `void`. Document every exception the function can throw with `@throws`, naming the concrete type:

```cpp
/// Opens an archive from the given path
///
/// @param path Path to the archive file.
/// @param mode Whether to open for reading or writing.
/// @return A handle to the opened archive.
/// @throws mylib::ArchiveOpenError If the file does not exist or cannot be read.
Archive OpenArchive(const std::string &path, OpenMode mode);
```

The `@throws` lines are part of the contract, not a nicety: an exception a consumer cannot discover from the header is an exception they will not catch. Keep them in step with the error hierarchy (see the error handling fragment).

## What to document

Document the contract, not the implementation. A caller needs to know what a function promises, what it requires, and how it fails; they do not need a narration of how it works:

```cpp
// Good - states the contract
/// Returns the entries whose names match a glob pattern
///
/// @param pattern A glob pattern; an empty pattern matches every entry.
/// @return Matching entries in archive order, empty if none match.
std::vector<Entry> FindEntries(const std::string &pattern) const;

// Bad - narrates the body and says nothing a caller needs
/// Loops over the entries vector and checks each one against the pattern
std::vector<Entry> FindEntries(const std::string &pattern) const;
```

Say so explicitly where behaviour is easy to guess wrong: what an empty or null argument does, whether a returned reference outlives the object, whether the call is thread-safe, and which argument owns the memory. These are the questions that bring a consumer to the issue tracker.

## Enums and members

Enum values and public data members get a one-line `///` where the name does not fully carry the meaning. A trailing `///<` is acceptable for short cases:

```cpp
/// How an archive is opened
enum class OpenMode {
    READ,       ///< Open an existing archive; fails if it does not exist
    WRITE,      ///< Create a new archive, truncating any existing file
    APPEND,     ///< Open an existing archive for adding entries
};
```

Do not comment a value whose name already says it. `READ, ///< Read` is noise.
