#!/usr/bin/env bash

echo -e "Testing seqhasher binary\n"

# Initialize variables to track test failures and total tests
failed=0
total_tests=0

# Test basic usage
function test_basic_usage {
  result=$(../seqhasher test2.fasta -)
  expected=$(printf ">test2.fasta;e2512172abf8cc9f67fdd49eb6cacf2df71bbad3;seq1\nAAAA\n>test2.fasta;65c89f59d38cdbf90dfaf0b0a6884829df8396b0;seq2\nACTG\n>test2.fasta;e2512172abf8cc9f67fdd49eb6cacf2df71bbad3;seq3\nAAAA\n")
  ((total_tests++))
  if [[ "$result" != "$expected" ]]; then
    echo -e "\e[31m'Basic usage' test failed\e[0m"
    echo -e "\nExpected:"
    echo "$expected"
    echo -e "\nObtained:"
    echo "$result"
    ((failed++))
  else
    echo -e "\e[32m'Basic usage' test passed\e[0m"
  fi
}

# Test custom name
function test_custom_name {
  result=$(../seqhasher --name "custom_name" test2.fasta -)
  expected=$(printf ">custom_name;e2512172abf8cc9f67fdd49eb6cacf2df71bbad3;seq1\nAAAA\n>custom_name;65c89f59d38cdbf90dfaf0b0a6884829df8396b0;seq2\nACTG\n>custom_name;e2512172abf8cc9f67fdd49eb6cacf2df71bbad3;seq3\nAAAA\n")
  ((total_tests++))
  if [[ "$result" != "$expected" ]]; then
    echo -e "\e[31m'Custom name' test failed\e[0m"
    echo -e "\nExpected:"
    echo "$expected"
    echo -e "\nObtained:"
    echo "$result"
    ((failed++))
  else
    echo -e "\e[32m'Custom name' test passed\e[0m"
  fi
}

# Test headers only
function test_headers_only {
  result=$(../seqhasher --headersonly test2.fasta -)
  expected=$(printf "test2.fasta;e2512172abf8cc9f67fdd49eb6cacf2df71bbad3;seq1\ntest2.fasta;65c89f59d38cdbf90dfaf0b0a6884829df8396b0;seq2\ntest2.fasta;e2512172abf8cc9f67fdd49eb6cacf2df71bbad3;seq3\n")
  ((total_tests++))
  if [[ "$result" != "$expected" ]]; then
    echo -e "\e[31m'Headers only' test failed\e[0m"
    echo -e "\nExpected:"
    echo "$expected"
    echo -e "\nObtained:"
    echo "$result"
    ((failed++))
  else
    echo -e "\e[32m'Headers only' test passed\e[0m"
  fi  
}

# Test no filename
function test_no_filename {
  result=$(../seqhasher --headersonly --nofilename test2.fasta -)
  expected=$(printf "e2512172abf8cc9f67fdd49eb6cacf2df71bbad3;seq1\n65c89f59d38cdbf90dfaf0b0a6884829df8396b0;seq2\ne2512172abf8cc9f67fdd49eb6cacf2df71bbad3;seq3\n")
  ((total_tests++))
  if [[ "$result" != "$expected" ]]; then
    echo -e "\e[31m'No filename' test failed\e[0m"
    echo -e "\nExpected:"
    echo "$expected"
    echo -e "\nObtained:"
    echo "$result"
    ((failed++))
  else
    echo -e "\e[32m'No filename' test passed\e[0m"
  fi
}

# Test xxHash and case-sensitive
function test_xxhash_case_sensitive {
  result=$(../seqhasher --headersonly --nofilename --hash xxhash --casesensitive test2.fasta -)
  expected=$(printf "cf40b5b72bc43e77;seq1\n704b34bf20faedf2;seq2\n42a70d1abf84bf32;seq3\n")
  ((total_tests++))
  if [[ "$result" != "$expected" ]]; then
    echo -e "\e[31m'xxHash and case-sensitive' test failed\e[0m"
    echo -e "\nExpected:"
    echo "$expected"
    echo -e "\nObtained:"
    echo "$result"
    ((failed++))
  else
    echo -e "\e[32m'xxHash and case-sensitive' test passed\e[0m"
  fi
}

# Test multiple hashes
function test_multiple_hashes {
  result=$(../seqhasher --headersonly --nofilename --hash sha1,xxhash --casesensitive test2.fasta -)
  expected=$(printf "e2512172abf8cc9f67fdd49eb6cacf2df71bbad3;cf40b5b72bc43e77;seq1\n65c89f59d38cdbf90dfaf0b0a6884829df8396b0;704b34bf20faedf2;seq2\n70c881d4a26984ddce795f6f71817c9cf4480e79;42a70d1abf84bf32;seq3\n")
  ((total_tests++))
  if [[ "$result" != "$expected" ]]; then
    echo -e "\e[31m'Multiple hashes' test failed\e[0m"
    echo -e "\nExpected:"
    echo "$expected"
    echo -e "\nObtained:"
    echo "$result"
    ((failed++))
  else
    echo -e "\e[32m'Multiple hashes' test passed\e[0m"
  fi
}

# Test compressed files
function test_compressed_files {
  for ext in bz2 gz xz zst; do
    result=$(../seqhasher --headersonly --nofilename test2.fasta.$ext -)
    expected=$(printf "e2512172abf8cc9f67fdd49eb6cacf2df71bbad3;seq1\n65c89f59d38cdbf90dfaf0b0a6884829df8396b0;seq2\ne2512172abf8cc9f67fdd49eb6cacf2df71bbad3;seq3\n")
    ((total_tests++))
    if [[ "$result" != "$expected" ]]; then
      echo -e "\e[31m'Compressed file' test failed for .$ext\e[0m"
      echo -e "\nExpected:"
      echo "$expected"
      echo -e "\nObtained:"
      echo "$result"
      ((failed++))
    else
      echo -e "\e[32m'Compressed file' test passed for .$ext\e[0m"
    fi
  done
}

# Test whitespace stripping
function test_whitespace_stripping {
  result=$(../seqhasher --hash sha1 --nofilename seqs_with_spaces.fasta -)
  expected=$(printf ">65c89f59d38cdbf90dfaf0b0a6884829df8396b0;seq1_uppercase\nACTG\n>65c89f59d38cdbf90dfaf0b0a6884829df8396b0;seq2_lowercase\nACTG\n>65c89f59d38cdbf90dfaf0b0a6884829df8396b0;seq3_mixed_case\nACTG\n>65c89f59d38cdbf90dfaf0b0a6884829df8396b0;seq4_with_spaces\nACTG\n>65c89f59d38cdbf90dfaf0b0a6884829df8396b0;seq5_with_tabs\nACTG\n>65c89f59d38cdbf90dfaf0b0a6884829df8396b0;seq6_with_newline\nACTG\n")
  ((total_tests++))
  if [[ "$result" != "$expected" ]]; then
    echo -e "\e[31m'Whitespace stripping' test failed\e[0m"
    echo -e "\nExpected:"
    echo "$expected"
    echo -e "\nObtained:"
    echo "$result"
    ((failed++))
  else
    echo -e "\e[32m'Whitespace stripping' test passed\e[0m"
  fi
}

# Test long sequence (100k bases)
function test_long_sequence {
  # Generate a long sequence of repeating ACTG (50k bases)
  long_seq1=$(printf 'ACTG%.0s' {1..25000})
  long_seq2=$(printf 'A%.0s' {1..100000})
  echo -e ">seq1\n$long_seq1\n>seq2\n$long_seq2" > test_long.fasta
  
  result=$(../seqhasher --headersonly --nofilename test_long.fasta -)
  expected=$(printf "3200abf5034212fb5d90fa242fc1cea5e0f3dd00;seq1\ne27b41eb7f7370b36f1b2fa63f05652f52423277;seq2\n")
  ((total_tests++))
  if [[ "$result" != "$expected" ]]; then
    echo -e "\e[31m'Long sequence' test failed\e[0m"
    echo -e "\nExpected:"
    echo "$expected"
    echo -e "\nObtained:"
    echo "$result"
    ((failed++))
  else
    echo -e "\e[32m'Long sequence' test passed\e[0m"
  fi
  
  rm test_long.fasta
}

# Test FASTQ file
function test_fastq_file {
  result=$(cat test3.fastq | ../seqhasher - -)
  expected=$(printf "@65c89f59d38cdbf90dfaf0b0a6884829df8396b0;seq1\nACTG\n+\nDFGH\n@e2512172abf8cc9f67fdd49eb6cacf2df71bbad3;seq2\nAAAA\n+\nBBBB\n@e3da52abc8fbdb38b113a187ed0ac763fa86d1d4;seq3\nTGCA\n+\nEEEE\n")
  ((total_tests++))
  if [[ "$result" != "$expected" ]]; then
    echo -e "\e[31m'FASTQ file' test failed\e[0m"
    echo -e "\nExpected:"
    echo "$expected"
    echo -e "\nObtained:"
    echo "$result"
    ((failed++))
  else
    echo -e "\e[32m'FASTQ file' test passed\e[0m"
  fi
}

test_basic_usage
test_custom_name
test_headers_only
test_no_filename
test_xxhash_case_sensitive
test_multiple_hashes
test_compressed_files
test_whitespace_stripping
test_long_sequence
test_fastq_file

if [[ $failed -eq 0 ]]; then
  echo -e "\n\e[32mAll $total_tests tests passed\e[0m"
  exit 0
else
  echo -e "\n\e[31m$failed out of $total_tests tests failed\e[0m"
  exit 1
fi
