package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
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
				hashTypes:     []string{"sha1"},
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
				hashTypes:      []string{"md5"},
				noFileName:     true,
				caseSensitive:  true,
				inputFileName:  "input.fasta",
				outputFileName: "output.fasta",
			},
		},
		{
			name: "Multiple hash types",
			args: []string{"cmd", "-hash", "sha1,xxhash", "input.fasta"},
			expected: config{
				hashTypes:     []string{"sha1", "xxhash"},
				inputFileName: "input.fasta",
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
		{"nthash", true},
		{"blake3", true},
		{"invalid", false},
		{"", false},
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

	// Test stdout separately
	t.Run("Stdout", func(t *testing.T) {
		output, err := getOutput("-")
		if err != nil {
			t.Errorf("getOutput(\"-\") returned unexpected error: %v", err)
		}
		if output != os.Stdout {
			t.Errorf("getOutput(\"-\") did not return os.Stdout")
		}
		// Don't close os.Stdout
	})
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
				hashTypes:     []string{"sha1"},
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
				hashTypes:     []string{"md5"},
				noFileName:    true,
				caseSensitive: false,
				inputFileName: "test.fasta",
			},
			expected: "86bfb9f78dd8b6cd35962bb7324fdbf8;seq1\n" +
				"86bfb9f78dd8b6cd35962bb7324fdbf8;seq1_lowercase\n" +
				"5c15f97a88433c48f8bf76745d9da437;seq2\n",
		},
		{
			name: "ntHash",
			cfg: config{
				hashTypes:     []string{"nthash"},
				noFileName:    false,
				caseSensitive: false,
				inputFileName: "test.fasta",
			},
			expected: ">test.fasta;508876b331232519;seq1\nACTG\n" +
				">test.fasta;508876b331232519;seq1_lowercase\nACTG\n" +
				">test.fasta;95cecc5106c8fccd;seq2\nTGCA\n",
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
		{"nthash", "508876b331232519"},
		{"blake3", "fe31e49d18b8883e7167198f770b98bba33b533cc12a9bb63ab264e5b70a347a"},
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

// Test if the output of compressed input files matches the output of the non-compressed input
func TestCompressedInput(t *testing.T) {
	logger := &testLogger{t}
	// Define the non-compressed input
	nonCompressedInput := "test.fasta"
	expectedOutput := "65c89f59d38cdbf90dfaf0b0a6884829df8396b0;seq1\n" +
		"65c89f59d38cdbf90dfaf0b0a6884829df8396b0;seq1_lowercase\n" +
		"e3da52abc8fbdb38b113a187ed0ac763fa86d1d4;seq2\n"

	// List of compressed files to test
	compressedFiles := []string{
		"./test/test.fasta.gz",
		"./test/test.fasta.bz2",
		"./test/test.fasta.xz",
		"./test/test.fasta.zst",
	}

	// Test non-compressed input
	t.Run("NonCompressed", func(t *testing.T) {
		logger.Logf(colorize(colorYellow, "Testing non-compressed input: %s"), nonCompressedInput)
		input := strings.NewReader(testSequences)
		output := &bytes.Buffer{}
		processSequences(input, output, config{
			hashTypes:     []string{"sha1"},
			noFileName:    true,
			headersOnly:   true,
			caseSensitive: false,
			inputFileName: nonCompressedInput,
		})
		got := output.String()
		if got != expectedOutput {
			t.Errorf("\nProcessSequences failed for %s\nGot:\n%s\nWant:\n%s",
				nonCompressedInput, got, expectedOutput)
			failedTests = append(failedTests, "ProcessSequences/NonCompressed")
		}
	})

	// Test each compressed input
	for _, fileName := range compressedFiles {
		t.Run(fileName, func(t *testing.T) {
			logger.Logf(colorize(colorYellow, "Testing compressed input: %s"), fileName)
			input, err := getInput(fileName)
			if err != nil {
				t.Errorf("getInput() error = %v", err)
				return
			}
			defer input.Close()

			output := &bytes.Buffer{}
			processSequences(input, output, config{
				hashTypes:     []string{"sha1"},
				noFileName:    true,
				headersOnly:   true,
				caseSensitive: false,
				inputFileName: fileName,
			})
			got := output.String()
			if got != expectedOutput {
				t.Errorf("\nProcessSequences failed for %s\nGot:\n%s\nWant:\n%s",
					fileName, got, expectedOutput)
				failedTests = append(failedTests, "ProcessSequences/"+fileName)
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
	t.Run("ParseFlags", TestParseFlags)
	t.Run("IsValidHashType", TestIsValidHashType)
	t.Run("GetInput", TestGetInput)
	t.Run("GetOutput", TestGetOutput)
	t.Run("ProcessSequences", TestProcessSequences)
	t.Run("GetHashFunc", TestGetHashFunc)
	t.Run("MainFunction", TestMainFunction)

	// Check for test failures
	if t.Failed() {
		failedTests = append(failedTests, t.Name())
	}
}

func TestMainFunction(t *testing.T) {
	// Set up arguments
	oldArgs := os.Args
	os.Args = []string{"cmd", "-version"}

	// Call run() instead of main()
	output := run()

	// Restore arguments
	os.Args = oldArgs

	// Check if version is printed
	expectedOutput := fmt.Sprintf("SeqHasher %s\n", version)
	if output != expectedOutput {
		t.Errorf("Expected output %q, got %q", expectedOutput, output)
	}
}

func TestGetInputError(t *testing.T) {
	_, err := getInput("nonexistent_file.txt")
	if err == nil {
		t.Error("Expected an error for nonexistent file, got nil")
	}
}

func TestGetOutputError(t *testing.T) {
	// Create a directory to test writing to a directory (which should fail)
	err := os.Mkdir("testdir", 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll("testdir")

	_, err = getOutput("testdir")
	if err == nil {
		t.Error("Expected an error when trying to write to a directory, got nil")
	}
}

func TestPrintUsage(t *testing.T) {
	// Redirect stderr to capture output
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	printUsage()

	// Restore stderr
	w.Close()
	os.Stderr = oldStderr

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Check if usage information is printed
	if !strings.Contains(output, "Usage:") || !strings.Contains(output, "Options:") {
		t.Errorf("Expected usage information, got: %s", output)
	}
}
