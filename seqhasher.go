package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/shenwei356/bio/seq"
	"github.com/shenwei356/bio/seqio/fastx"

	"github.com/cespare/xxhash/v2"
	"github.com/go-faster/city"
	"github.com/spaolacci/murmur3"
)

// Define variables
var headersOnly bool
var hashType string
var noFileName bool
var caseSensitive bool

func main() {
	nameFlag := flag.String("name", "", "Optional text to replace input file name in the header")
	headersOnlyFlag := flag.Bool("headersonly", false, "If enabled, output only sequence headers")
	hashTypeFlag := flag.String("hashtype", "sha1", "Defines the hash type: sha1 (default), md5, xxhash, cityhash, murmur3")
	noFileNameFlag := flag.Bool("nofilename", false, "If enabled, disables adding a file name to the sequence header")
	caseSensitiveFlag := flag.Bool("casesensitive", false, "If enabled, keeps sequences as is without converting to uppercase")
	flag.Parse()

	// Set variables based on the flag values
	headersOnly = *headersOnlyFlag
	hashType = *hashTypeFlag
	noFileName = *noFileNameFlag
	caseSensitive = *caseSensitiveFlag

	// Get non-flag arguments
	nonFlagArgs := flag.Args()

	// Validate hashType
	validHashTypes := map[string]bool{
		"sha1":     true,
		"md5":      true,
		"xxhash":   true,
		"cityhash": true,
		"murmur3":  true,
	}

	if _, valid := validHashTypes[hashType]; !valid {
		fmt.Printf("Warning: Unsupported hash type '%s'. Using 'sha1' as default.\n", hashType)
		hashType = "sha1" // Set to default
	}

	if len(nonFlagArgs) < 1 || len(nonFlagArgs) > 2 || (*nameFlag == "" && len(nonFlagArgs) == 1) {
		fmt.Println("SeqHasher: DNA Sequence Hashing Tool")
		fmt.Println("=====================================")
		fmt.Println("Usage:")
		fmt.Println("  seqhasher [options] <input_file> [output_file]")
		fmt.Println("\nOverview:")
		fmt.Println("  SeqHasher takes DNA sequences from a FASTA file, computes a hash digest for each sequence,")
		fmt.Println("  and generates an output file with modified sequences.")
		fmt.Println("  For input/output via stdin/stdout, use '-' instead of the file name.")
		fmt.Println("\nOptions:")
		fmt.Println("  --name <text>         Replace the input file's name in the header with <text>.")
		fmt.Println("  --nofilename          Omit the file name from the sequence header.")
		fmt.Println("  --headersonly         Only output sequence headers, excluding the sequences themselves.")
		fmt.Println("  --hashtype <type>     Specify the hash algorithm: sha1 (default), md5, xxhash, cityhash, or murmur3.")
		fmt.Println("  --casesensitive       Take into account sequence case. By default, sequences are converted to uppercase.")
		fmt.Println("\nArguments:")
		fmt.Println("  <input_file>          The path to the input FASTA file or '-' for standard input (stdin).")
		fmt.Println("  [output_file]         The path to the output file or '-' for standard output (stdout).")
		fmt.Println("                        If omitted, defaults to stdout.")
		fmt.Println("\nExamples:")
		fmt.Println("  seqhasher input.fasta output.fasta")
		fmt.Println("  cat input.fasta | seqhasher --name 'Sample' - - > output.fasta")
		fmt.Println("\nFor additional information, use the -h flag.")
		return
	}

	var input io.Reader
	var output io.Writer
	var inputFileName string

	inputFileName = nonFlagArgs[0]
	if inputFileName == "-" {
		input = os.Stdin
		if *nameFlag == "" {
			inputFileName = "stdin"
		} else {
			inputFileName = *nameFlag
		}
	} else {
		inputFile, err := os.Open(inputFileName)
		if err != nil {
			log.Fatalf("Error opening input file %s: %v", inputFileName, err)
		}
		defer inputFile.Close()
		input = inputFile
		if *nameFlag != "" {
			inputFileName = *nameFlag
		} else {
			inputFileName = filepath.Base(inputFileName)
		}
	}

	if len(nonFlagArgs) == 2 {
		if nonFlagArgs[1] == "-" {
			output = os.Stdout
		} else {
			outputFile, err := os.Create(nonFlagArgs[1])
			if err != nil {
				log.Fatalf("Error creating output file %s: %v", nonFlagArgs[1], err)
			}
			defer outputFile.Close()
			output = outputFile
		}
	} else {
		output = os.Stdout
	}

	processSequences(input, output, inputFileName, caseSensitive)
}

func processSequences(input io.Reader, output io.Writer, inputFileName string, caseSensitive bool) {
	writer := bufio.NewWriter(output)

	reader, err := fastx.NewReaderFromIO(seq.DNA, bufio.NewReader(input), fastx.DefaultIDRegexp)
	if err != nil {
		log.Fatalf("Failed to create reader: %v", err)
	}
	defer reader.Close() // Ensure to close the reader when done

	for {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("Error reading record: %v", err)
			continue
		}

		seq := record.Seq.Seq
		if !caseSensitive {
			// Convert to uppercase if caseSensitive is false
			seq = bytes.ToUpper(seq)
		}

		hashFunc := getHashFunc(hashType)
		hashedSeq := hashFunc(seq)

		// Prepare the new sequence header
		var modifiedHeader string
		if noFileName {
			modifiedHeader = fmt.Sprintf("%s;%s", hashedSeq, record.Name)
		} else {
			modifiedHeader = fmt.Sprintf("%s;%s;%s", inputFileName, hashedSeq, record.Name)
		}

		if headersOnly {
			// Output only the header, without the '>' sign, if `--headersonly` is enabled
			if _, err := fmt.Fprintf(writer, "%s\n", modifiedHeader); err != nil {
				log.Printf("Error writing header: %v", err)
				continue
			}
		} else {
			// Output the full record
			if _, err := fmt.Fprintf(writer, ">%s\n%s\n", modifiedHeader, seq); err != nil {
				log.Printf("Error writing record: %v", err)
				continue
			}
		}
	}

	if err := writer.Flush(); err != nil {
		log.Fatalf("Error flushing output: %v", err)
	}
}

// getHashFunc returns a function that takes a byte slice and returns a hex string
// of the hash based on the specified hash type.
func getHashFunc(hashType string) func([]byte) string {
	switch hashType {
	case "md5":
		return func(data []byte) string {
			hash := md5.Sum(data)
			return hex.EncodeToString(hash[:])
		}
	case "xxhash":
		return func(data []byte) string {
			hash := xxhash.Sum64(data)
			return fmt.Sprintf("%016x", hash)
		}
	case "cityhash":
		return func(data []byte) string {
			hash := city.Hash128(data)
			return fmt.Sprintf("%016x%016x", hash.High, hash.Low)
		}
	case "murmur3":
		return func(data []byte) string {
			h1, h2 := murmur3.Sum128(data)
			return fmt.Sprintf("%016x%016x", h1, h2)
		}
	default: // Default to SHA1
		return func(data []byte) string {
			hash := sha1.Sum(data)
			return hex.EncodeToString(hash[:])
		}
	}
}
