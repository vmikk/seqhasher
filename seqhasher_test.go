package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

const (
	testFastaPath = "./test/test.fasta"
	testSequences = ">seq1\nACTG\n>seq1_lowercase\nactg\n>seq2\nTGCA\n"

	// ANSI color codes
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorReset  = "\033[0m"
)

func colorize(color, message string) string {
	return color + message + colorReset
}

type testLogger struct {
	t *testing.T
}

func (l *testLogger) Logf(format string, args ...interface{}) {
	l.t.Logf(format, args...)
	l.t.Helper()
}

func (l *testLogger) Errorf(format string, args ...interface{}) {
	l.t.Helper()
	l.t.Errorf(colorize(colorRed, fmt.Sprintf(format, args...)))
}

func (l *testLogger) Fatalf(format string, args ...interface{}) {
	l.t.Helper()
	l.t.Fatalf(colorize(colorRed, fmt.Sprintf(format, args...)))
}

func runTest(t *testing.T, name string, testFunc func(*testing.T)) {
	t.Run(name, func(t *testing.T) {
		logger := &testLogger{t}
		logger.Logf(colorize(colorYellow, "Running test case: %s"), name)
		testFunc(t)
		if t.Failed() {
			logger.Logf(colorize(colorRed, "Test case failed: %s"), name)
			// Print the test name to stderr to ensure it's captured
			fmt.Fprintf(os.Stderr, colorize(colorRed, "FAIL: %s\n"), name)
		} else {
			logger.Logf(colorize(colorGreen, "Test case passed: %s"), name)
		}
	})
}

func TestParseFlags(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	oldFlagCommandLine := flag.CommandLine
	defer func() { flag.CommandLine = oldFlagCommandLine }()

	tests := []struct {
		name     string
		args     []string
		expected config
	}{
		{
			name: "Default settings",
			args: []string{"cmd", "input.fasta"},
			expected: config{
				headersOnly:   false,
				hashType:      "sha1",
				noFileName:    false,
				caseSensitive: false,
				inputFileName: "input.fasta",
			},
		},
		{
			name: "Custom settings",
			args: []string{"cmd", "-headersonly", "-hash", "md5", "-nofilename", "-casesensitive", "input.fasta", "output.fasta"},
			expected: config{
				headersOnly:    true,
				hashType:       "md5",
				noFileName:     true,
				caseSensitive:  true,
				inputFileName:  "input.fasta",
				outputFileName: "output.fasta",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
			os.Args = tt.args
			cfg := parseFlags()
			if !reflect.DeepEqual(cfg, tt.expected) {
				t.Errorf("parseFlags() = %v, want %v", cfg, tt.expected)
				failedTests = append(failedTests, "ParseFlags/"+tt.name)
			}
		})
	}
}

// Check if the hash type validation works correctly
func TestIsValidHashType(t *testing.T) {
	logger := &testLogger{t}
	tests := []struct {
		hashType string
		want     bool
	}{
		{"sha1", true},
		{"md5", true},
		{"xxhash", true},
		{"cityhash", true},
		{"murmur3", true},
		{"invalid", false},
	}

	for _, tt := range tests {
		runTest(t, tt.hashType, func(t *testing.T) {
			logger.Logf(colorize(colorYellow, "Testing hash type: %s"), tt.hashType)
			if got := isValidHashType(tt.hashType); got != tt.want {
				t.Errorf("isValidHashType(%q) = %v, want %v", tt.hashType, got, tt.want)
			}
		})
	}
}

// Test if the input file is correctly handled
func TestGetInput(t *testing.T) {
	logger := &testLogger{t}
	tests := []struct {
		name     string
		fileName string
		wantErr  bool
	}{
		{"Stdin", "-", false},
		{"Existing file", testFastaPath, false},
		{"Non-existent file", "nonexistent.fasta", true},
	}

	for _, tt := range tests {
		runTest(t, tt.name, func(t *testing.T) {
			logger.Logf(colorize(colorYellow, "Testing input: %s"), tt.name)
			input, err := getInput(tt.fileName)
			if (err != nil) != tt.wantErr {
				t.Errorf("getInput() error = %v, wantErr %v", err, tt.wantErr)
			} else if !tt.wantErr && input == nil {
				t.Errorf("getInput() returned nil, want non-nil")
			}
			if input != nil {
				input.Close()
			}
		})
	}
}

// Test if the output file is correctly handled
func TestGetOutput(t *testing.T) {
	logger := &testLogger{t}
	tests := []struct {
		name     string
		fileName string
		wantErr  bool
	}{
		{"Stdout", "-", false},
		{"New file", "test_output.fasta", false},
	}

	for _, tt := range tests {
		runTest(t, tt.name, func(t *testing.T) {
			logger.Logf(colorize(colorYellow, "Testing output: %s"), tt.name)
			output, err := getOutput(tt.fileName)
			if (err != nil) != tt.wantErr {
				t.Errorf("getOutput() error = %v, wantErr %v", err, tt.wantErr)
			} else if !tt.wantErr && output == nil {
				t.Errorf("getOutput() returned nil, want non-nil")
			}
			if output != nil {
				output.Close()
			}
			if tt.fileName != "-" {
				os.Remove(tt.fileName)
			}
		})
	}
}

// Test if the sequence processing works correctly
func TestProcessSequences(t *testing.T) {
	logger := &testLogger{t}
	tests := []struct {
		name     string
		cfg      config
		expected string
	}{
		{
			name: "Default settings",
			cfg: config{
				hashType:      "sha1",
				noFileName:    false,
				caseSensitive: false,
				inputFileName: "test.fasta",
			},
			expected: ">test.fasta;65c89f59d38cdbf90dfaf0b0a6884829df8396b0;seq1\nACTG\n" +
				">test.fasta;65c89f59d38cdbf90dfaf0b0a6884829df8396b0;seq1_lowercase\nACTG\n" +
				">test.fasta;e3da52abc8fbdb38b113a187ed0ac763fa86d1d4;seq2\nTGCA\n",
		},
		{
			name: "Headers only",
			cfg: config{
				headersOnly:   true,
				hashType:      "md5",
				noFileName:    true,
				caseSensitive: false,
				inputFileName: "test.fasta",
			},
			expected: "86bfb9f78dd8b6cd35962bb7324fdbf8;seq1\n" +
				"86bfb9f78dd8b6cd35962bb7324fdbf8;seq1_lowercase\n" +
				"5c15f97a88433c48f8bf76745d9da437;seq2\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger.Logf(colorize(colorYellow, "Testing processSequences: %s"), tt.name)
			input := strings.NewReader(testSequences)
			output := &bytes.Buffer{}
			processSequences(input, output, tt.cfg)
			got := output.String()
			if got != tt.expected {
				t.Errorf("\nProcessSequences failed for %s\nConfig: %+v\nGot:\n%s\nWant:\n%s",
					tt.name, tt.cfg, got, tt.expected)
				failedTests = append(failedTests, "ProcessSequences/"+tt.name)
			}
		})
	}
}

// Verify that each hash function produces the expected output
func TestGetHashFunc(t *testing.T) {
	logger := &testLogger{t}
	testData := []byte("ACTG")
	tests := []struct {
		hashType string
		expected string
	}{
		{"sha1", "65c89f59d38cdbf90dfaf0b0a6884829df8396b0"},
		{"md5", "86bfb9f78dd8b6cd35962bb7324fdbf8"},
		{"xxhash", "704b34bf20faedf2"},
		{"cityhash", "7ee08b0605f909cf400644ddb3b8b80b"},
		{"murmur3", "da48f168029d0eff17c81eff7624a72f"},
	}

	for _, tt := range tests {
		runTest(t, tt.hashType, func(t *testing.T) {
			logger.Logf(colorize(colorYellow, "Testing hash function: %s"), tt.hashType)
			hashFunc := getHashFunc(tt.hashType)
			got := hashFunc(testData)
			if got != tt.expected {
				t.Errorf("\nHash function %s failed\nInput: %s\nGot:  %s\nWant: %s",
					tt.hashType, testData, got, tt.expected)
			}
		})
	}
}

// Run the tests
// + set up a test FASTA file if it doesn't exist
func TestMain(m *testing.M) {
	fmt.Println(colorize(colorYellow, "Setting up test environment..."))

	if _, err := os.Stat(testFastaPath); os.IsNotExist(err) {
		err := os.WriteFile(testFastaPath, []byte(testSequences), 0644)
		if err != nil {
			fmt.Println(colorize(colorRed, "Failed to create test FASTA file"))
			os.Exit(1)
		}
		fmt.Printf(colorize(colorGreen, "Created test FASTA file: %s\n"), testFastaPath)
	} else {
		fmt.Printf(colorize(colorGreen, "Using existing test FASTA file: %s\n"), testFastaPath)
	}

	fmt.Println(colorize(colorYellow, "Running tests..."))

	// Run tests with a custom test function
	exitCode := m.Run()

	// Check for test failures
	if exitCode != 0 {
		fmt.Println(colorize(colorRed, "Some tests failed. Failed tests:"))
		for _, test := range failedTests {
			fmt.Println(colorize(colorRed, "- "+test))
		}
	} else {
		fmt.Println(colorize(colorGreen, "All tests passed!"))
	}

	os.Exit(exitCode)
}

// Global slice to store names of failed tests
var failedTests []string

// TestAll runs all tests and captures failures
func TestAll(t *testing.T) {
	// Run all test functions
	t.Run("ParseFlags", TestParseFlags)
	t.Run("IsValidHashType", TestIsValidHashType)
	t.Run("GetInput", TestGetInput)
	t.Run("GetOutput", TestGetOutput)
	t.Run("ProcessSequences", TestProcessSequences)
	t.Run("GetHashFunc", TestGetHashFunc)

	// Check for test failures
	if t.Failed() {
		failedTests = append(failedTests, t.Name())
	}
}
