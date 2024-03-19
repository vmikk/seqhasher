package main

import (
	"bufio"
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

func main() {
	nameFlag := flag.String("name", "", "Optional text to replace input file name in the header")
	headersOnlyFlag := flag.Bool("headersonly", false, "If enabled, output only sequence headers")
	hashTypeFlag := flag.String("hashtype", "sha1", "Defines the hash type: sha1 (default), md5, xxhash, cityhash, murmur3")
	noFileNameFlag := flag.Bool("nofilename", false, "If enabled, disables adding a file name to the sequence header")
	flag.Parse()

	headersOnly = *headersOnlyFlag // Set variables based on the flag values
	hashType = *hashTypeFlag
	noFileName = *noFileNameFlag
	nonFlagArgs := flag.Args() // Get non-flag arguments

	if len(nonFlagArgs) < 1 || len(nonFlagArgs) > 2 || (*nameFlag == "" && len(nonFlagArgs) == 1) {
		fmt.Println("Usage: rechimizer [--name user_text] input_file [output_file]")
		fmt.Println("\nDescription:")
		fmt.Println("  Processes DNA sequences from a FASTA file, calculates a SHA1 hash for each,")
		fmt.Println("  and outputs the modified sequences. To use standard input or output, specify '-' for the file name.")
		fmt.Println("\nParameters:")
		fmt.Println("  input_file    - Path to the input FASTA file or '-' for stdin.")
		fmt.Println("  output_file   - Path to the output file or '-' for stdout. Optional; defaults to stdout if not provided.")
		fmt.Println("  --name        - Optional. Replaces the input file name in the header with the specified text.")
		fmt.Println("  --noFilename      - Optional. Disables adding a file name to the sequence header.")
		fmt.Println("  --headersonly - Optional. Outputs only sequence headers.")
		fmt.Println("  --hashtype    - Optional. The hash type: sha1 (default), md5, xxhash, cityhash (as in VSEARCH), murmur3 (as in Sourmash).")
		fmt.Println("\nExamples:")
		fmt.Println("  rechimizer input.fasta output.fasta")
		fmt.Println("  rechimizer --name 'Sample' - - < input.fasta > output.fasta")
		fmt.Println("\nUse -h for more information.")
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

	processSequences(input, output, inputFileName)
}

func processSequences(input io.Reader, output io.Writer, inputFileName string) {
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
