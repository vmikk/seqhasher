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

# Test no filename
function test_no_filename {
  result=$(../seqhasher --headersonly --nofilename test2.fasta -)
  expected=$(cat <<EOF
e2512172abf8cc9f67fdd49eb6cacf2df71bbad3;seq1
65c89f59d38cdbf90dfaf0b0a6884829df8396b0;seq2
e2512172abf8cc9f67fdd49eb6cacf2df71bbad3;seq3
EOF
)
  if [[ "$result" != "$expected" ]]; then
    echo "No filename test failed"
    exit 1
  fi
}

# Test xxHash and case-sensitive
function test_xxhash_case_sensitive {
  result=$(../seqhasher --headersonly --nofilename --hash xxhash --casesensitive test2.fasta -)
  expected=$(cat <<EOF  
cf40b5b72bc43e77;seq1
704b34bf20faedf2;seq2
42a70d1abf84bf32;seq3
EOF
)
  if [[ "$result" != "$expected" ]]; then
    echo "xxHash and case-sensitive test failed"
    exit 1
  fi
}

# Test multiple hashes
function test_multiple_hashes {
  result=$(../seqhasher --headersonly --nofilename --hash sha1,xxhash --casesensitive test2.fasta -)
  expected=$(cat <<EOF
e2512172abf8cc9f67fdd49eb6cacf2df71bbad3;cf40b5b72bc43e77;seq1
65c89f59d38cdbf90dfaf0b0a6884829df8396b0;704b34bf20faedf2;seq2
70c881d4a26984ddce795f6f71817c9cf4480e79;42a70d1abf84bf32;seq3  
EOF
)
  if [[ "$result" != "$expected" ]]; then
    echo "Multiple hashes test failed"
    exit 1
  fi
}

test_basic_usage
test_headers_only
test_no_filename
test_xxhash_case_sensitive
test_multiple_hashes
