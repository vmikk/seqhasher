package main

import (
	"bufio"
	"crypto/sha1"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/shenwei356/bio/seq"
	"github.com/shenwei356/bio/seqio/fastx"
)

func main() {
	nameFlag := flag.String("name", "", "Optional text to replace input file name in the header")
	flag.Parse()
	nonFlagArgs := flag.Args() // Get non-flag arguments

	if len(nonFlagArgs) < 1 || len(nonFlagArgs) > 2 || (*nameFlag == "" && len(nonFlagArgs) == 1) {
		fmt.Println("Usage: rechimizer [--name user_text] input_file [output_file]")
		fmt.Println("\nDescription:")
		fmt.Println("  Processes DNA sequences from a FASTA file, calculates a SHA1 hash for each,")
		fmt.Println("  and outputs the modified sequences. To use standard input or output, specify '-' for the file name.")
		fmt.Println("\nParameters:")
		fmt.Println("  input_file   - Path to the input FASTA file or '-' for stdin.")
		fmt.Println("  output_file  - Path to the output file or '-' for stdout. Optional; defaults to stdout if not provided.")
		fmt.Println("  --name       - Optional. Replaces the input file name in the header with the specified text.")
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
		sha1sum := sha1.Sum(seq)
		modifiedHeader := fmt.Sprintf("%s;%x;%s", inputFileName, sha1sum, record.Name)

		if _, err := fmt.Fprintf(writer, ">%s\n%s\n", modifiedHeader, seq); err != nil {
			log.Printf("Error writing record: %v", err)
			continue
		}
	}

	if err := writer.Flush(); err != nil {
		log.Fatalf("Error flushing output: %v", err)
	}
}
