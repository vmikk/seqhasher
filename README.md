# Rechimizer

## Overview
Rechimizer is a command-line tool designed to calculate a SHA1 hash for each sequence in a FASTA file and add it to a sequence header.

## Usage

```plaintext
rechimizer [--name user_text] input_file [output_file]
Parameters:
    input_file: Specifies the path to the input FASTA file or '-' to use standard input (stdin).
    output_file: Specifies the path to the output file or '-' to use standard output (stdout). This parameter is optional; if not provided, the output will be directed to stdout by default.
    --name: An optional parameter that replaces the input file name in the header of the output with the specified text.
```

### Description

The tool can either read the input from a specified file or from standard input (`stdin`), and similarly, it can write the output to a specified file or standard output (`stdout`). The `--name` option allows to customize the header of the output by specifying a text to replace the input file name.

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
