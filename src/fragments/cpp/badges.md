# C++ badge row

```markdown
![C++ Version](https://img.shields.io/badge/Version-XX-blue?style=flat&logo=cplusplus) ![Code Coverage](https://img.shields.io/badge/Coverage-XX%25-blue?style=flat&logo=cplusplus)
```

Replace the first `XX` with the C++ standard the project compiles against, `CMAKE_CXX_STANDARD` in the root `CMakeLists.txt`, so `20`. It is the language version, as the Go badge shows the Go version, not the project's own version, which the release badge already carries. Replace the second with the coverage percentage, updated on each release. The percentage comes from `make test_coverage`; see the coverage section of cpp/testing.md.
