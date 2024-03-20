# SeqHasher

## Overview
`seqhasher` is a command-line tool designed to calculate a hash (digest or fingerprint) 
for each sequence in a FASTA file and add it to a sequence header.  

## Usage

```plaintext
seqhasher [--options] input_file [output_file]

Parameters:
  --name: An optional parameter that replaces the input file name in the header of the output with the specified text.
  --nofilename: Optional. Disables adding a file name to the sequence header.
  --headersonly: Optional. Outputs only sequence headers.
  --hashtype: Optional. The hash type: sha1 (default), md5, xxhash, cityhash, murmur3.

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
- SHA1 (default), 160-bit hash value
- MD5, 128-bit hash value
- xxHash ([extremely fast](https://xxhash.com/)), 64-bit hash value
- CityHash (e.g., used in [VSEARCH](https://github.com/torognes/vsearch/)), 128-bit hash value
- Murmur3 (e.g., used in [Sourmash](https://github.com/sourmash-bio/sourmash), but 64-bit), 128-bit hash value

### Examples

To process a FASTA file and output to another file:
```bash
rechimizer input.fasta output.fasta
```

To process a FASTA file from standard input and output to standard output, while replacing the file name in the header with 'Sample':
```bash
cat input.fasta | rechimizer --name 'Sample' - - > output.fasta
# OR
rechimizer --name 'Sample' - - < input.fasta > output.fasta
```

## Benchmark

To evaluate the performance of two solutions for processing DNA sequences, we utilized [`hyperfine`](https://github.com/sharkdp/hyperfine) to compare an AWK-based solution against the `rechimizer` binary.

### Test data

First, let's create the test data: a FASTA file containing 500,000 sequences, each 30 to 3000 nucleotides long.

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

