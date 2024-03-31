# SeqHasher

## Overview
`seqhasher` is a command-line tool designed to calculate a hash (digest or fingerprint) 
for each sequence in a FASTA file and add it to a sequence header.  

## Quick start

Input data (e.g., `inp.fasta`):
```
>seq1
AAAA
>seq2
ACTG
>seq3
AAAA
``` 

`seqhasher input.fasta -`
```
>input.fasta;e2512172abf8cc9f67fdd49eb6cacf2df71bbad3;seq1
AAAA
>input.fasta;65c89f59d38cdbf90dfaf0b0a6884829df8396b0;seq2
ACTG
>input.fasta;e2512172abf8cc9f67fdd49eb6cacf2df71bbad3;seq3
AAAA
```

`seqhasher --name "test_file" input.fasta -`
```
>test_file;e2512172abf8cc9f67fdd49eb6cacf2df71bbad3;seq1
AAAA
>test_file;65c89f59d38cdbf90dfaf0b0a6884829df8396b0;seq2
ACTG
>test_file;e2512172abf8cc9f67fdd49eb6cacf2df71bbad3;seq3
AAAA
```

`seqhasher --headersonly input.fasta -`
```
input.fasta;e2512172abf8cc9f67fdd49eb6cacf2df71bbad3;seq1
input.fasta;65c89f59d38cdbf90dfaf0b0a6884829df8396b0;seq2
input.fasta;e2512172abf8cc9f67fdd49eb6cacf2df71bbad3;seq3
```

`seqhasher --headersonly --nofilename input.fasta -`
```
e2512172abf8cc9f67fdd49eb6cacf2df71bbad3;seq1
65c89f59d38cdbf90dfaf0b0a6884829df8396b0;seq2
e2512172abf8cc9f67fdd49eb6cacf2df71bbad3;seq3
```

`seqhasher --headersonly --nofilename --hashtype xxhash input.fasta -`
```
cf40b5b72bc43e77;seq1
704b34bf20faedf2;seq2
cf40b5b72bc43e77;seq3
```

## Usage

```plaintext
seqhasher [--options] <input_file> [output_file]

Options:
  --name: An optional parameter that replaces the input file name in the header of the output with the specified text.
  --nofilename: Disables adding a file name to the sequence header.
  --headersonly: Outputs only sequence headers.
  --hashtype: The hash type: sha1 (default), md5, xxhash, cityhash, murmur3.
  --casesensitive: Keeps sequences as is without forcing to uppercase.
  --version: Prints the version of the program and exits.

Arguments:
  input_file: Specifies the path to the input FASTA file or '-' to use standard input (stdin).
  output_file: Specifies the path to the output file or '-' to use standard output (stdout). This parameter is optional; if not provided, the output will be directed to stdout by default.
```

### Description

The tool can either read the input from a specified file or from standard input (`stdin`), 
and similarly, it can write the output to a specified file or standard output (`stdout`).  

The `--name` option allows to customize the header of the output by specifying 
a text to replace the input file name.

The `--hashtype` option allows to specify which hash function to use. 
Currently, the following hash functions are supported:  
- `sha1`: [SHA-1](https://en.wikipedia.org/wiki/SHA-1) (default), 160-bit hash value
- `md5`: [MD5](https://en.wikipedia.org/wiki/MD5), 128-bit hash value
- `xxhash`: xxHash ([extremely fast](https://xxhash.com/)), 64-bit hash value
- `cityhash`: [CityHash](https://opensource.googleblog.com/2011/04/introducing-cityhash.html) (e.g., used in [VSEARCH](https://github.com/torognes/vsearch/)), 128-bit hash value
- `murmur3`: [Murmur3](https://en.wikipedia.org/wiki/MurmurHash) (e.g., used in [Sourmash](https://github.com/sourmash-bio/sourmash), but 64-bit), 128-bit hash value

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
  'seqhasher --headersonly --casesensitive --hashtype sha1     big.fasta - > /dev/null' \
  'seqhasher --headersonly --casesensitive --hashtype md5      big.fasta - > /dev/null' \
  'seqhasher --headersonly --casesensitive --hashtype xxhash   big.fasta - > /dev/null' \
  'seqhasher --headersonly --casesensitive --hashtype cityhash big.fasta - > /dev/null' \
  'seqhasher --headersonly --casesensitive --hashtype murmur3  big.fasta - > /dev/null'
```

| Command    |      Mean [s] | Min [s] | Max [s] |    Relative |
|:---------- | -------------:| -------:| -------:| -----------:|
| sha1     | 1.753 ± 0.328 |   1.549 |   2.532 | 1.43 ± 0.41 |
| md5      | 2.120 ± 0.437 |   1.685 |   2.718 | 1.73 ± 0.52 |
| xxhash   | 1.223 ± 0.269 |   0.921 |   1.512 | 1.00        |
| cityhash | 1.288 ± 0.250 |   1.038 |   1.647 | 1.05 ± 0.31 |
| murmur3  | 1.224 ± 0.230 |   1.032 |   1.610 | 1.00 ± 0.29 |


### Processing large file

Compare an `AWK`-based solution against the `seqhasher` binary.

