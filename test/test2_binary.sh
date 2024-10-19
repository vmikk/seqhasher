#!/usr/bin/env bash

set -euo pipefail

# Test basic usage
function test_basic_usage {
  result=$(../seqhasher test2.fasta -)
  expected=$(cat <<EOF
>test2.fasta;e2512172abf8cc9f67fdd49eb6cacf2df71bbad3;seq1
AAAA
>test2.fasta;f74673f038f3657adfa522aa370b5cd161dec321;seq2
ACTG
>test2.fasta;e2512172abf8cc9f67fdd49eb6cacf2df71bbad3;seq3
AAAA
EOF
)
  if [[ "$result" != "$expected" ]]; then

# Test headers only
function test_headers_only {
  result=$(../seqhasher --headersonly test2.fasta -)
  expected=$(cat <<EOF
test2.fasta;e2512172abf8cc9f67fdd49eb6cacf2df71bbad3;seq1
test2.fasta;65c89f59d38cdbf90dfaf0b0a6884829df8396b0;seq2
test2.fasta;e2512172abf8cc9f67fdd49eb6cacf2df71bbad3;seq3
EOF
)
  if [[ "$result" != "$expected" ]]; then
    echo "Headers only test failed"
    exit 1
  fi  
}

    exit 1
  fi
}

test_basic_usage
test_headers_only
