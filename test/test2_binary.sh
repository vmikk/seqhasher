#!/usr/bin/env bash

set -euo pipefail

# Initialize variables to track test failures and total tests
failed=0
total_tests=0

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
  ((total_tests++))
  if [[ "$result" != "$expected" ]]; then
    echo -e "\e[31m'Basic usage' test failed\e[0m"
    ((failed++))
  else
    echo -e "\e[32m'Basic usage' test passed\e[0m"
  fi
}

# Test custom name
function test_custom_name {
  result=$(../seqhasher --name "custom_name" test2.fasta -)
  expected=$(cat <<EOF
>custom_name;e2512172abf8cc9f67fdd49eb6cacf2df71bbad3;seq1
AAAA
>custom_name;65c89f59d38cdbf90dfaf0b0a6884829df8396b0;seq2
ACTG
>custom_name;e2512172abf8cc9f67fdd49eb6cacf2df71bbad3;seq3
AAAA
EOF
)
  ((total_tests++))
  if [[ "$result" != "$expected" ]]; then
    echo -e "\e[31m'Custom name' test failed\e[0m"
    ((failed++))
  else
    echo -e "\e[32m'Custom name' test passed\e[0m"
  fi
}

# Test headers only
function test_headers_only {
  result=$(../seqhasher --headersonly test2.fasta -)
  expected=$(cat <<EOF
test2.fasta;e2512172abf8cc9f67fdd49eb6cacf2df71bbad3;seq1
test2.fasta;65c89f59d38cdbf90dfaf0b0a6884829df8396b0;seq2
test2.fasta;e2512172abf8cc9f67fdd49eb6cacf2df71bbad3;seq3
EOF
)
  ((total_tests++))
  if [[ "$result" != "$expected" ]]; then
    echo -e "\e[31m'Headers only' test failed\e[0m"
    ((failed++))
  else
    echo -e "\e[32m'Headers only' test passed\e[0m"
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
  ((total_tests++))
  if [[ "$result" != "$expected" ]]; then
    echo -e "\e[31m'No filename' test failed\e[0m"
    ((failed++))
  else
    echo -e "\e[32m'No filename' test passed\e[0m"
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
  ((total_tests++))
  if [[ "$result" != "$expected" ]]; then
    echo -e "\e[31m'xxHash and case-sensitive' test failed\e[0m"
    ((failed++))
  else
    echo -e "\e[32m'xxHash and case-sensitive' test passed\e[0m"
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
  ((total_tests++))
  if [[ "$result" != "$expected" ]]; then
    echo -e "\e[31m'Multiple hashes' test failed\e[0m"
    ((failed++))
  else
    echo -e "\e[32m'Multiple hashes' test passed\e[0m"
  fi
}

# Test compressed files
function test_compressed_files {
  for ext in bz2 gz xz zst; do
    result=$(../seqhasher --headersonly --nofilename test2.fasta.$ext -)
    expected=$(cat <<EOF
e2512172abf8cc9f67fdd49eb6cacf2df71bbad3;seq1
65c89f59d38cdbf90dfaf0b0a6884829df8396b0;seq2
e2512172abf8cc9f67fdd49eb6cacf2df71bbad3;seq3
EOF
)
    ((total_tests++))
    if [[ "$result" != "$expected" ]]; then
      echo -e "\e[31m'Compressed file' test failed for .$ext\e[0m"
      ((failed++))
    else
      echo -e "\e[32m'Compressed file' test passed for .$ext\e[0m"
    fi
  done
}

test_basic_usage
test_custom_name
test_headers_only
test_no_filename
test_xxhash_case_sensitive
test_multiple_hashes
test_compressed_files

if [[ $failed -eq 0 ]]; then
  echo -e "\e[32mAll $total_tests tests passed\e[0m"
  exit 0
else
  echo -e "\e[31m$failed out of $total_tests tests failed\e[0m"
  exit 1
fi
