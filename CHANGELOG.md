# Changelog

- [SeqHasher v1.2.0](https://github.com/vmikk/seqhasher/releases/tag/1.2.0) - 2026-05-22
[![GitHub downloads for SeqHasher v1.2.0](https://img.shields.io/github/downloads/vmikk/seqhasher/1.2.0/total.svg)](https://github.com/vmikk/seqhasher/releases/tag/1.2.0)
    - `seqhasher`:
        - Added `xxh3`, `k12`, `rapidhash`, and `rapidhash32` hash algorithms  
        - Fixed FASTA output after FASTQ processing when pooled `fastx` records retain stale quality data (related to `shenwei356/bio` updates to v0.14.0)  
        - Updated dependencies  

- [SeqHasher v1.1.2](https://github.com/vmikk/seqhasher/releases/tag/1.1.2) - 2025-05-17
[![GitHub downloads for SeqHasher v1.1.2](https://img.shields.io/github/downloads/vmikk/seqhasher/1.1.2/total.svg)](https://github.com/vmikk/seqhasher/releases/tag/1.1.2)
    - `seqhasher`:
        - Updated dependencies
    - CI:
        - Updated release artifact handling

- [SeqHasher v1.1.1](https://github.com/vmikk/seqhasher/releases/tag/1.1.1) - 2024-12-08
[![GitHub downloads for SeqHasher v1.1.1](https://img.shields.io/github/downloads/vmikk/seqhasher/1.1.1/total.svg)](https://github.com/vmikk/seqhasher/releases/tag/1.1.1)
    - `seqhasher`:
        - Added support for FASTQ files
        - Added the SHA-3 hash algorithm (`--hash sha3`)
        - Strip whitespace from sequences before hashing
        - Disabled DNA sequence validation to support non-DNA characters
        - Disabled filename output for stdin unless overridden
        - Refactored sequence-processing code
    - Tests:
        - Added binary-level integration tests
        - Improved CI test coverage
    - Project:
        - Added a Zenodo DOI

- [SeqHasher v1.1.0](https://github.com/vmikk/seqhasher/releases/tag/1.1.0) - 2024-09-22
[![GitHub downloads for SeqHasher v1.1.0](https://img.shields.io/github/downloads/vmikk/seqhasher/1.1.0/total.svg)](https://github.com/vmikk/seqhasher/releases/tag/1.1.0)
    - `seqhasher`:
        - Added support for multiple hash algorithms, for example `--hash sha1,xxhash`
        - Added `nthash` and `blake3` hash algorithms
        - Added short flags, such as `-H` for `--hash`
        - Improved the `seqhasher --help` output
    - Tests:
        - Added continuous integration tests

- [SeqHasher v1.0.0](https://github.com/vmikk/seqhasher/releases/tag/1.0.0) - 2024-03-30
[![GitHub downloads for SeqHasher v1.0.0](https://img.shields.io/github/downloads/vmikk/seqhasher/1.0.0/total.svg)](https://github.com/vmikk/seqhasher/releases/tag/1.0.0)
    - `seqhasher`:
        - First stable release
        - Added the `--casesensitive` flag
        - Fixed hash name validation

- [SeqHasher v0.4](https://github.com/vmikk/seqhasher/releases/tag/0.4) - 2024-03-19
[![GitHub downloads for SeqHasher v0.4](https://img.shields.io/github/downloads/vmikk/seqhasher/0.4/total.svg)](https://github.com/vmikk/seqhasher/releases/tag/0.4)
    - `seqhasher`:
        - Added the `--nofilename` flag
        - Fixed zero padding of hashes

- [SeqHasher v0.3](https://github.com/vmikk/seqhasher/releases/tag/0.3) - 2024-03-18
[![GitHub downloads for SeqHasher v0.3](https://img.shields.io/github/downloads/vmikk/seqhasher/0.3/total.svg)](https://github.com/vmikk/seqhasher/releases/tag/0.3)
    - `seqhasher`:
        - Added the `--hashtype` argument
        - Added SHA-1 as the default hash algorithm
        - Added MD5, xxHash, CityHash, and MurmurHash3 hash algorithms

- [SeqHasher v0.2](https://github.com/vmikk/seqhasher/releases/tag/0.2) - 2024-03-17
[![GitHub downloads for SeqHasher v0.2](https://img.shields.io/github/downloads/vmikk/seqhasher/0.2/total.svg)](https://github.com/vmikk/seqhasher/releases/tag/0.2)
    - `seqhasher`:
        - Added the `--headersonly` flag

- [SeqHasher v0.1](https://github.com/vmikk/seqhasher/releases/tag/0.1) - 2024-03-16
[![GitHub downloads for SeqHasher v0.1](https://img.shields.io/github/downloads/vmikk/seqhasher/0.1/total.svg)](https://github.com/vmikk/seqhasher/releases/tag/0.1)
    - `seqhasher`:
        - Initial pre-release
