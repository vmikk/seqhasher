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

	"github.com/shenwei356/bio/seq"
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
		name           string
		args           []string
		expected       config
		expectedErrMsg string
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
		{
			name:           "Invalid hash type",
			args:           []string{"cmd", "-hash", "invalid,sha1", "input.fasta"},
			expectedErrMsg: "Invalid hash type: invalid. Supported types are: sha1, sha3, md5, xxhash, cityhash, murmur3, nthash, blake3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
			os.Args = tt.args

			cfg, err := parseFlags()

			if tt.expectedErrMsg != "" {
				if err == nil {
					t.Errorf("Expected error, got nil")
				} else if err.Error() != tt.expectedErrMsg {
					t.Errorf("Expected error message %q, got %q", tt.expectedErrMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if !reflect.DeepEqual(cfg, tt.expected) {
					t.Errorf("parseFlags() = %v, want %v", cfg, tt.expected)
					failedTests = append(failedTests, "ParseFlags/"+tt.name)
				}
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
		{"sha3", true},
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
	// defer os.Remove("nonexistent.fasta")
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
		{"sha3", "01eb915e4d8b6d44d0432c12dfdb949c1da1f37c295a653b8761a1e46ed2d76cb0c297d612af809b9691d341cad536df912cbba6e95a93380cdc9f545d9bfdcc"},
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

// Global variables
var (
	failedTests []string // Global slice to store names of failed tests
	silentMode  bool     // Flag to control output in tests
)

// Initialize global variables
func init() {
	failedTests = make([]string, 0)
	silentMode = false
}

// TestAll runs all tests and captures failures
func TestAll(t *testing.T) {
	// Set silent mode for the duration of TestAll
	silentMode = true
	defer func() { silentMode = false }()

	// Save original stdout
	oldStdout := os.Stdout

	// Create a pipe to capture stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create a channel to coordinate output restoration
	done := make(chan bool)

	// Start a goroutine to handle the piped output
	go func() {
		_, _ = io.Copy(io.Discard, r)
		done <- true
	}()

	tests := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{"ParseFlags", TestParseFlags},
		{"IsValidHashType", TestIsValidHashType},
		{"GetInput", TestGetInput},
		{"GetOutput", TestGetOutput},
		{"ProcessSequences", TestProcessSequences},
		{"GetHashFunc", TestGetHashFunc},
		{"CompressedInput", TestCompressedInput},
		{"MainFunction", TestMainFunction},
		{"GetInputError", TestGetInputError},
		{"GetOutputError", TestGetOutputError},
		{"PrintUsage", TestPrintUsage},
		{"ProcessSequencesReaderCreationFailure", TestProcessSequencesReaderCreationFailure},
		{"ProcessSequencesInvalidSequence", TestProcessSequencesInvalidSequence},
		{"ProcessFASTQSequences", TestProcessFASTQSequences},
	}

	// Write directly to the original stdout for our test output
	fmt.Fprintf(oldStdout, "%s\n", colorize(colorYellow, "Running all test suites:"))

	// Run each test
	for _, tt := range tests {
		testName := tt.name
		t.Run(testName, func(t *testing.T) {
			fmt.Fprintf(oldStdout, "%s %s\n",
				colorize(colorYellow, "▶"),
				colorize(colorYellow, "Running test suite: "+testName))

			// Temporarily restore stdout for the test's own output capture
			os.Stdout = w

			// Run the actual test
			tt.testFunc(t)

			if !t.Failed() {
				fmt.Fprintf(oldStdout, "%s %s\n",
					colorize(colorGreen, "✓"),
					colorize(colorGreen, "Test suite passed: "+testName))
			} else {
				fmt.Fprintf(oldStdout, "%s %s\n",
					colorize(colorRed, "✗"),
					colorize(colorRed, "Test suite failed: "+testName))
				failedTests = append(failedTests, testName)
			}
		})
	}

	// Restore stdout
	w.Close()
	<-done
	os.Stdout = oldStdout
	r.Close()

	// Print final summary
	if len(failedTests) > 0 {
		fmt.Fprintf(oldStdout, "\n%s\n", colorize(colorRed, "Some test suites failed:"))
		for _, test := range failedTests {
			fmt.Fprintf(oldStdout, "%s\n", colorize(colorRed, "- "+test))
		}
	} else {
		fmt.Fprintf(oldStdout, "\n%s\n", colorize(colorGreen, "All test suites passed!"))
	}
}

func TestMainFunction(t *testing.T) {
	// Capture and discard output if in silent mode
	var stdout *os.File
	var w, r *os.File
	if silentMode {
		var err error
		stdout = os.Stdout
		r, w, err = os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		os.Stdout = w
	}

	tests := []struct {
		name           string
		args           []string
		expectedOutput string
		expectedError  string
	}{
		{
			name:           "Version flag",
			args:           []string{"cmd", "-version"},
			expectedOutput: fmt.Sprintf("SeqHasher %s\n", version),
		},
		{
			name:           "No input file",
			args:           []string{"cmd"},
			expectedOutput: "Usage:", // Just check for the beginning of usage info
		},
		{
			name:          "Non-existent input file",
			args:          []string{"cmd", "nonexistent_file.fasta"},
			expectedError: "Error opening input: open nonexistent_file.fasta: no such file or directory",
		},
		{
			name:           "Valid input file",
			args:           []string{"cmd", testFastaPath},
			expectedOutput: ";seq1\n", // Check for processed sequence
		},
		{
			name:           "Output to file",
			args:           []string{"cmd", testFastaPath, "test_output.fasta"},
			expectedOutput: "", // Output will be written to file instead of buffer
		},
		{
			name:          "Output to invalid directory",
			args:          []string{"cmd", testFastaPath, "/nonexistent/directory/output.fasta"},
			expectedError: "Error opening output: open /nonexistent/directory/output.fasta: no such file or directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags before each test
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			// Set up arguments
			oldArgs := os.Args
			os.Args = tt.args

			// Prepare a buffer to capture output
			var buf bytes.Buffer

			// Call run() with our buffer
			err := run(&buf)

			// Restore arguments
			os.Args = oldArgs

			// Check error
			if tt.expectedError != "" {
				if err == nil || err.Error() != tt.expectedError {
					t.Errorf("Expected error %q, got %v", tt.expectedError, err)
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// For the "Output to file" test, verify the file contents
			if tt.name == "Output to file" {
				// Read the output file
				content, err := os.ReadFile("test_output.fasta")
				if err != nil {
					t.Errorf("Failed to read output file: %v", err)
				} else {
					// Verify file contains expected content
					if !strings.Contains(string(content), ";seq1\n") {
						t.Errorf("Output file doesn't contain expected content")
					}
				}
				// Clean up the output file
				os.Remove("test_output.fasta")
			}

			// Check buffer output for non-file output tests
			if tt.expectedOutput != "" {
				output := buf.String()
				if !strings.Contains(output, tt.expectedOutput) {
					t.Errorf("Expected output to contain %q, got %q", tt.expectedOutput, output)
				}
			}
		})
	}

	// Restore stdout if in silent mode
	if silentMode {
		w.Close()
		os.Stdout = stdout
		io.Copy(io.Discard, r)
		r.Close()
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
	// Capture and discard output if in silent mode
	var stdout *os.File
	var w, r *os.File
	if silentMode {
		var err error
		stdout = os.Stdout
		r, w, err = os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		os.Stdout = w
	}

	runTest(t, "PrintUsage", func(t *testing.T) {
		logger := &testLogger{t}
		logger.Logf(colorize(colorYellow, "Testing printUsage function"))

		// Test regular usage
		t.Run("RegularUsage", func(t *testing.T) {
			// Save old args and restore after test
			oldArgs := os.Args
			os.Args = []string{"seqhasher"}
			defer func() { os.Args = oldArgs }()

			var buf bytes.Buffer
			printUsage(&buf)
			output := buf.String()

			// Check for expected content in regular usage
			expectedStrings := []string{
				"SeqHasher v",
				"Usage: seqhasher [options]",
				"Options:",
				"Supported hash types:",
				"If input_file is '-' or omitted, reads from stdin",
			}

			for _, str := range expectedStrings {
				if !strings.Contains(output, str) {
					t.Errorf("Expected usage output to contain '%s', but it was not found\nGot:\n%s", str, output)
				}
			}
		})

		// Test detailed help
		t.Run("DetailedHelp", func(t *testing.T) {
			// Save old args and restore after test
			oldArgs := os.Args
			os.Args = []string{"seqhasher", "--help"}
			defer func() { os.Args = oldArgs }()

			var buf bytes.Buffer
			printUsage(&buf)
			output := buf.String()

			// Check for expected content in detailed help
			expectedStrings := []string{
				"SeqHasher",
				"DNA Sequence Hashing Tool",
				"version:",
				"Usage:",
				"Overview:",
				"Options:",
				"Arguments:",
				"Examples:",
				"https://github.com/vmikk/seqhasher",
			}

			for _, str := range expectedStrings {
				if !strings.Contains(output, str) {
					t.Errorf("Expected detailed help to contain '%s', but it was not found\nGot:\n%s", str, output)
				}
			}
		})
	})

	// Restore stdout if in silent mode
	if silentMode {
		w.Close()
		os.Stdout = stdout
		io.Copy(io.Discard, r)
		r.Close()
	}
}

// failingReader is a custom io.Reader that always returns a simple string
type failingReader struct{}

func (fr failingReader) Read(p []byte) (n int, err error) {
	copy(p, "invalid input")
	return len("invalid input"), io.EOF
}

func TestProcessSequencesReaderCreationFailure(t *testing.T) {
	runTest(t, "ProcessSequencesReaderCreationFailure", func(t *testing.T) {
		logger := &testLogger{t}
		logger.Logf(colorize(colorYellow, "Testing processSequences with reader creation failure"))

		input := failingReader{}
		output := &bytes.Buffer{}
		cfg := config{
			hashTypes:     []string{"sha1"},
			noFileName:    false,
			caseSensitive: false,
			inputFileName: "test.fasta",
		}

		err := processSequences(input, output, cfg)

		if err == nil {
			t.Error("Expected an error, but got nil")
		} else if !strings.Contains(err.Error(), "fastx: invalid FASTA/Q format") {
			t.Errorf("Expected error message to contain 'fastx: invalid FASTA/Q format', but got: %v", err)
		}
	})
}

func TestProcessSequencesInvalidSequence(t *testing.T) {
	// Create a temporary test directory
	tmpDir, err := os.MkdirTemp("", "seqhasher_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	runTest(t, "ProcessSequencesInvalidSequence", func(t *testing.T) {
		logger := &testLogger{t}
		logger.Logf(colorize(colorYellow, "Testing processSequences with invalid sequence"))

		// Disable sequence validation
		seq.ValidateSeq = false

		// Create an input with an "invalid" DNA sequence
		invalidInput := strings.NewReader(">seq1\nACTGINVALID\n")

		output := &bytes.Buffer{}
		cfg := config{
			hashTypes:     []string{"sha1"},
			noFileName:    false,
			caseSensitive: false,
			inputFileName: "test.fasta",
		}

		err := processSequences(invalidInput, output, cfg)

		// The sequence should be processed successfully since ValidateSeq is false
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		expected := ">test.fasta;3e06752b63358c1a14c4c364513cbc2250674bb8;seq1\nACTGINVALID\n"
		if got := output.String(); got != expected {
			t.Errorf("Expected output:\n%s\nGot:\n%s", expected, got)
		}
	})
}

func TestProcessFASTQSequences(t *testing.T) {
	logger := &testLogger{t}
	tests := []struct {
		name     string
		cfg      config
		input    string
		expected string
	}{
		{
			name: "Basic FASTQ processing",
			cfg: config{
				hashTypes:     []string{"sha1"},
				noFileName:    false,
				caseSensitive: false,
				inputFileName: "test.fastq",
			},
			input: "@seq1\nACTG\n+\nDFGH\n@seq2\nAAAA\n+\nBBBB\n",
			expected: "@test.fastq;65c89f59d38cdbf90dfaf0b0a6884829df8396b0;seq1\nACTG\n+\nDFGH\n" +
				"@test.fastq;e2512172abf8cc9f67fdd49eb6cacf2df71bbad3;seq2\nAAAA\n+\nBBBB\n",
		},
		{
			name: "FASTQ with headers only",
			cfg: config{
				headersOnly:   true,
				hashTypes:     []string{"sha1"},
				noFileName:    true,
				caseSensitive: false,
				inputFileName: "test.fastq",
			},
			input: "@seq1\nACTG\n+\nDFGH\n@seq2\nAAAA\n+\nBBBB\n",
			expected: "65c89f59d38cdbf90dfaf0b0a6884829df8396b0;seq1\n" +
				"e2512172abf8cc9f67fdd49eb6cacf2df71bbad3;seq2\n",
		},
	}

	for _, tt := range tests {
		runTest(t, tt.name, func(t *testing.T) {
			logger.Logf(colorize(colorYellow, "Testing FASTQ processing: %s"), tt.name)
			input := strings.NewReader(tt.input)
			output := &bytes.Buffer{}
			err := processSequences(input, output, tt.cfg)
			if err != nil {
				t.Errorf("processSequences() error = %v", err)
				return
			}
			got := output.String()
			if got != tt.expected {
				t.Errorf("\nProcessSequences failed for %s\nConfig: %+v\nGot:\n%s\nWant:\n%s",
					tt.name, tt.cfg, got, tt.expected)
				failedTests = append(failedTests, "ProcessFASTQSequences/"+tt.name)
			}
		})
	}
}

func TestFlagUsage(t *testing.T) {
	runTest(t, "FlagUsage", func(t *testing.T) {
		// Save original stderr and create a pipe
		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w

		// Save original flag.CommandLine and args
		oldFlagCommandLine := flag.CommandLine
		oldArgs := os.Args

		// Create new FlagSet and set up flags
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		os.Args = []string{"seqhasher"} // Reset args to avoid interference

		// Set up the Usage function as done in parseFlags
		flag.Usage = func() {
			printUsage(os.Stderr)
		}

		// Call flag.Usage() which should trigger our custom printUsage
		flag.Usage()

		// Close writer and restore stderr
		w.Close()
		os.Stderr = oldStderr

		// Read the output
		var buf bytes.Buffer
		io.Copy(&buf, r)
		r.Close()

		// Restore original flag.CommandLine and args
		flag.CommandLine = oldFlagCommandLine
		os.Args = oldArgs

		// Verify the output contains expected content
		output := buf.String()
		expectedStrings := []string{
			"SeqHasher v",
			"Usage:",
			"Options:",
			"Supported hash types:",
		}

		for _, str := range expectedStrings {
			if !strings.Contains(output, str) {
				t.Errorf("Expected flag.Usage output to contain '%s', but it was not found\nGot:\n%s",
					str, output)
			}
		}
	})
}
