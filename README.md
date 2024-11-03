# SeqHasher

[![Go Test](https://github.com/vmikk/seqhasher/actions/workflows/go-test.yml/badge.svg)](https://github.com/vmikk/seqhasher/actions/workflows/go-test.yml)
[![Integration Tests](https://github.com/vmikk/seqhasher/actions/workflows/bash.yml/badge.svg)](https://github.com/vmikk/seqhasher/actions/workflows/bash.yml)
[![codecov](https://codecov.io/gh/vmikk/seqhasher/branch/main/graph/badge.svg)](https://codecov.io/gh/vmikk/seqhasher)

## Overview
`seqhasher` is a high-performance command-line tool designed to calculate a hash (digest or fingerprint) for each sequence in a FASTA file and add it to the sequence header. It supports multiple hashing algorithms and offers various output options.

## Features

- Fast processing of FASTA files (thanks to [shenwei356/bio](https://github.com/shenwei356/bio) package)
- Support for multiple hash algorithms: SHA-1, SHA-3, MD5, xxHash, CityHash, MurmurHash3, ntHash, and BLAKE3
- Automatic support for compressed input files (`gzip`, `zstd`, `xz`, and `bzip2`)
- Supports reading from STDIN and writing to STDOUT
- Option to output only headers or full sequences
- Case-sensitive hashing option
- Customizable output format (e.g., include filename or a custom text string in the header)

## Quick start

Input data (e.g., `input.fasta`):
```
>seq1
AAAA
>seq2
ACTG
>seq3
aaaa
``` 

Basic usage (default SHA1 hash):
`seqhasher input.fasta -`
```
>input.fasta;e2512172abf8cc9f67fdd49eb6cacf2df71bbad3;seq1
AAAA
>input.fasta;65c89f59d38cdbf90dfaf0b0a6884829df8396b0;seq2
ACTG
>input.fasta;e2512172abf8cc9f67fdd49eb6cacf2df71bbad3;seq3
AAAA
```

Custom name instead of input filename (e.g., useful when processing stdin):
`seqhasher --name "test_file" input.fasta -`
```
>test_file;e2512172abf8cc9f67fdd49eb6cacf2df71bbad3;seq1
AAAA
>test_file;65c89f59d38cdbf90dfaf0b0a6884829df8396b0;seq2
ACTG
>test_file;e2512172abf8cc9f67fdd49eb6cacf2df71bbad3;seq3
AAAA
```

Output only headers:
`seqhasher --headersonly input.fasta -`
```
input.fasta;e2512172abf8cc9f67fdd49eb6cacf2df71bbad3;seq1
input.fasta;65c89f59d38cdbf90dfaf0b0a6884829df8396b0;seq2
input.fasta;e2512172abf8cc9f67fdd49eb6cacf2df71bbad3;seq3
```

Omit filename from output:
`seqhasher --headersonly --nofilename input.fasta -`
```
e2512172abf8cc9f67fdd49eb6cacf2df71bbad3;seq1
65c89f59d38cdbf90dfaf0b0a6884829df8396b0;seq2
e2512172abf8cc9f67fdd49eb6cacf2df71bbad3;seq3
```

Use different hash functions (xxHash) and case-sensitive mode:
`seqhasher --headersonly --nofilename --hash xxhash --casesensitive input.fasta -`
```
cf40b5b72bc43e77;seq1
704b34bf20faedf2;seq2
42a70d1abf84bf32;seq3
```

Multiple hashes (useful to ensure absence of collisions):
`seqhasher --headersonly --nofilename --hash sha1,xxhash --casesensitive input.fasta -`
```
e2512172abf8cc9f67fdd49eb6cacf2df71bbad3;cf40b5b72bc43e77;seq1
65c89f59d38cdbf90dfaf0b0a6884829df8396b0;704b34bf20faedf2;seq2
70c881d4a26984ddce795f6f71817c9cf4480e79;42a70d1abf84bf32;seq3
```

## Usage

```plaintext
seqhasher [--options] <input_file> [output_file]

Options:
  -n, --nofilename: Omit the file name from the sequence header
  -f, --name: Replace the input file name in the header of the output with the specified text
  -o, --headersonly: Only output sequence headers, excluding the sequences themselves
  -H, --hash: Hash algorithm(s), multiple comma-separated values supported: sha1 (default), sha3, md5, xxhash, cityhash, murmur3, nthash, or blake3
  -c, --casesensitive: Take into account sequence case. By default, sequences are converted to uppercase
  -v, --version: Print the version of the program and exit
  -h, --help: Show help message

Arguments:
  input_file: The path to the input FASTA file or '-' for standard input (stdin).
  output_file: The path to the output file or '-' for standard output (stdout). This parameter is optional; if not provided, the output will be directed to stdout by default.
```

### Description

The tool can either read the input from a specified file or from standard input (`stdin`), 
and similarly, it can write the output to a specified file or standard output (`stdout`).  

The `--name` option allows to customize the header of the output by specifying 
a text to replace the input file name.

The `--hash` option allows to specify which hash function to use 
(multiple coma-separated values allowed, e.g., `--hash sha1,nthash`). 
Currently, the following hash functions are supported:  
- `sha1`: [SHA-1](https://en.wikipedia.org/wiki/SHA-1) (default), 160-bit hash value
- `sha3`: [SHA-3](https://en.wikipedia.org/wiki/SHA-3), Keccak-based secure cryptographic hash standard, 512-bit hash value
- `md5`: [MD5](https://en.wikipedia.org/wiki/MD5), 128-bit hash value
- `xxhash`: [xxHash](https://xxhash.com/), extremely fast algorithm, 64-bit hash value
- `cityhash`: [CityHash](https://opensource.googleblog.com/2011/04/introducing-cityhash.html) (e.g., used in [VSEARCH](https://github.com/torognes/vsearch/)), 128-bit hash value
- `murmur3`: [Murmur3](https://en.wikipedia.org/wiki/MurmurHash) (e.g., used in [Sourmash](https://github.com/sourmash-bio/sourmash), but 64-bit), 128-bit hash value
- `nthash`: [ntHash](https://github.com/bcgsc/ntHash) (designed for DNA sequences), 64-bit hash value. This implementation uses the full length of the sequence as the k-mer size, effectively hashing the entire sequence at once using the non-canonical (forward) hash of the sequence
- `blake3`: [BLAKE3](https://github.com/BLAKE3-team/BLAKE3) (fast cryptographic hash function), 256-bit hash value

> [!NOTE]
> The probability of a collision (when different DNA sequences end up with the same hash) 
> is roughly 1 in 2<sup>*nbits*</sup>, where *nbits* is the length of the hash in bits. 
> This means that functions with shorter bit-lengths (e.g., `Murmur3` and `CityHash`) 
> are more likely to have collisions as the dataset grows, 
> while `SHA-3` has a much lower chance of collisions because of its larger bit length. 
> However, shorter hashes are generally faster to compute 
> and take up less space when saved to a file, 
> making them more efficient for some tasks despite the higher collision risk.

### Examples

To process a FASTA file and output to another file:
```bash
seqhasher input.fasta output.fasta
```

To process a FASTA file from standard input and output to standard output, while replacing the file name in the header with 'Sample':
```bash
cat input.fasta | seqhasher --name 'Sample' - - > output.fasta
# OR
seqhasher --name 'Sample' - - < input.fasta > output.fasta
```

## Benchmark

To evaluate the performance of two solutions for processing DNA sequences, 
we utilized [`hyperfine`](https://github.com/sharkdp/hyperfine).

### Test data

First, let's create the test data: 
a FASTA file containing 500,000 sequences, each 30 to 3000 nucleotides long.

```bash
awk -v numSeq=500000 'BEGIN{
    srand();
    for(i=1; i<=numSeq; i++){
        seqLen=int(rand()*(2971))+30;
        printf(">seq_%d\n", i);
        for(j=1; j<=seqLen; j++){
            r=rand();
            if(r < 0.25) nucleotide="A";
            else if(r < 0.5) nucleotide="C";
            else if(r < 0.75) nucleotide="G";
            else nucleotide="T";
            printf("%s", nucleotide);
        }
        printf("\n");
    }
}' > big.fasta
```
The size of the file is ~760MB.


### Hashing functions performance

```bash
hyperfine \
  --runs 10 --warmup 3 \
  --export-markdown hashing_benchmark.md \
  'seqhasher --headersonly --casesensitive --hash md5      big.fasta - > /dev/null' \
  'seqhasher --headersonly --casesensitive --hash sha1     big.fasta - > /dev/null' \
  'seqhasher --headersonly --casesensitive --hash sha3     big.fasta - > /dev/null' \
  'seqhasher --headersonly --casesensitive --hash xxhash   big.fasta - > /dev/null' \
  'seqhasher --headersonly --casesensitive --hash murmur3  big.fasta - > /dev/null' \
  'seqhasher --headersonly --casesensitive --hash cityhash big.fasta - > /dev/null' \
  'seqhasher --headersonly --casesensitive --hash nthash   big.fasta - > /dev/null' \
  'seqhasher --headersonly --casesensitive --hash blake3   big.fasta - > /dev/null' \
  'seqhasher --headersonly --hash sha1,blake3    big.fasta - > /dev/null' \
  'seqhasher --headersonly --hash xxhash,murmur3 big.fasta - > /dev/null'
```

| Command          |      Mean [s] | Min [s] | Max [s] |    Relative |
|:-----------------|--------------:|--------:|--------:|------------:|
| `md5`            | 1.712 ± 0.069 |   1.651 |   1.847 | 1.75 ± 0.10 |
| `sha1`           | 1.614 ± 0.021 |   1.586 |   1.645 | 1.65 ± 0.08 |
| `sha3`           | 4.823 ± 0.135 |   4.707 |   5.090 | 4.93 ± 0.26 |
| `xxhash`         | 0.977 ± 0.043 |   0.941 |   1.079 |        1.00 |
| `murmur3`        | 1.106 ± 0.058 |   1.058 |   1.233 | 1.13 ± 0.08 |
| `cityhash`       | 1.078 ± 0.019 |   1.048 |   1.111 | 1.10 ± 0.05 |
| `nthash`         | 2.138 ± 0.022 |   2.112 |   2.170 | 2.19 ± 0.10 |
| `blake3`         | 1.718 ± 0.066 |   1.645 |   1.864 | 1.76 ± 0.10 |
| `sha1,blake3`    | 3.384 ± 0.096 |   3.290 |   3.640 | 3.46 ± 0.18 |
| `xxhash,murmur3` | 2.234 ± 0.073 |   2.193 |   2.422 | 2.29 ± 0.13 |

`Values are in seconds per 500,000 sequences (756,622,201 bp)`

As shown, xxHash provides the best performance, followed by CityHash and MurmurHash3. 
These hash functions produce relatively short hash fingerprints (64 and 128 bits, respectively). 
In contrast, SHA-3 is the slowest hash function in this benchmark, generating the longest hash (512 bits).  

> [!NOTE]
> However, it's important to note that these values may depend on 
> the instruction set of the CPU being used, as some processors may 
> optimize specific algorithms differently (e.g., via `SIMD` or other hardware acceleration). 
> For example, modern CPUs may use **SHA Extensions** to accelerate SHA-family algorithms. 
> Additionally, the performance reported here is tied to the particular implementations 
> of the hash algorithms used in `seqhasher`. Other implementations may yield different results, 
> and these values should not be interpreted as a definitive ranking of the algorithms themselves.


