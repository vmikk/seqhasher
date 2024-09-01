// This file is part of SeqHasher program (by Vladimir Mikryukov)
// and is licensed under GNU GPL-3.0-or-later.
// See the LICENSE file in the root of the source tree
// or <http://www.gnu.org/licenses/gpl-3.0.html>.

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
	"strings"

	"github.com/shenwei356/bio/seq"
	"github.com/shenwei356/bio/seqio/fastx"

	"github.com/cespare/xxhash/v2"
	"github.com/go-faster/city"
	"github.com/spaolacci/murmur3"

	"github.com/fatih/color"
)

const (
	version         = "1.0.1" // Version of the program
	defaultHashType = "sha1"  // Default hash type
)

var supportedHashTypes = []string{"sha1", "md5", "xxhash", "cityhash", "murmur3"}

// Configuration structure (flags)
type config struct {
	headersOnly    bool
	hashType       string
	noFileName     bool
	caseSensitive  bool
	inputFileName  string
	outputFileName string
	nameOverride   string
}

func main() {
	cfg := parseFlags()

	if cfg.inputFileName == "" {
		printUsage()
		return
	}

	input, err := getInput(cfg.inputFileName)
	if err != nil {
		log.Fatalf("Error opening input: %v", err)
	}
	defer input.Close()

	output, err := getOutput(cfg.outputFileName)
	if err != nil {
		log.Fatalf("Error opening output: %v", err)
	}
	defer output.Close()

	processSequences(input, output, cfg)
}

func parseFlags() config {
	cfg := config{}

	flag.BoolVar(&cfg.headersOnly, "headersonly", false, "Output only headers")
	flag.StringVar(&cfg.hashType, "hash", defaultHashType, "Hash type (sha1, md5, xxhash, cityhash, murmur3)")
	flag.BoolVar(&cfg.noFileName, "nofilename", false, "Do not include file name in output")
	flag.BoolVar(&cfg.caseSensitive, "casesensitive", false, "Case-sensitive hashing")
	flag.StringVar(&cfg.nameOverride, "name", "", "Override input file name in output")

	flag.Usage = printUsage
	flag.Parse()

	args := flag.Args()
	if len(args) > 0 {
		cfg.inputFileName = args[0]
	}
	if len(args) > 1 {
		cfg.outputFileName = args[1]
	}

	// Validate hash type
	if !isValidHashType(cfg.hashType) {
		log.Fatalf("Invalid hash type: %s. Supported types are: %s", cfg.hashType, strings.Join(supportedHashTypes, ", "))
	}

	return cfg
}

func isValidHashType(hashType string) bool {
	for _, supported := range supportedHashTypes {
		if hashType == supported {
			return true
		}
	}
	return false
}

func getInput(fileName string) (io.ReadCloser, error) {
	if fileName == "" || fileName == "-" {
		return os.Stdin, nil
	}
	return os.Open(fileName)
}

func getOutput(fileName string) (io.WriteCloser, error) {
	if fileName == "" || fileName == "-" {
		return os.Stdout, nil
	}
	return os.Create(fileName)
}

func printUsage() {
	if len(os.Args) > 1 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		fmt.Printf("\n%s%s%s\n",
			color.HiGreenString("SeqHasher"),
			color.WhiteString(" : "),
			color.HiMagentaString("DNA Sequence Hashing Tool"))
		fmt.Printf("%s  %s\n", color.HiCyanString("version:"), color.WhiteString(version))
		fmt.Println(color.WhiteString("====================================="))
		fmt.Println(color.HiCyanString("Usage:"))
		fmt.Printf("  %s\n", color.WhiteString("seqhasher [options] <input_file> [output_file]"))
		fmt.Println(color.HiCyanString("\nOverview:"))
		fmt.Println(color.WhiteString("  SeqHasher takes DNA sequences from a FASTA file, computes a hash digest for each sequence,"))
		fmt.Println(color.WhiteString("  and generates an output file with modified sequences."))
		fmt.Println(color.WhiteString("  For input/output via stdin/stdout, use '-' instead of the file name."))
		fmt.Println(color.HiCyanString("\nOptions:"))
		fmt.Printf("  %s %s\n", color.HiMagentaString("--nofilename"), color.WhiteString("    Omit the file name from the sequence header"))
		fmt.Printf("  %s %s\n", color.HiMagentaString("--name <text>"), color.WhiteString("   Replace the input file's name in the header with <text>"))
		fmt.Printf("  %s %s\n", color.HiMagentaString("--headersonly"), color.WhiteString("   Only output sequence headers, excluding the sequences themselves"))
		fmt.Printf("  %s %s\n", color.HiMagentaString("--hash <type>"), color.WhiteString("   Hash algorithm: sha1 (default), md5, xxhash, cityhash, or murmur3"))
		fmt.Printf("  %s %s\n", color.HiMagentaString("--casesensitive"), color.WhiteString(" Take into account sequence case. By default, sequences are converted to uppercase"))
		fmt.Printf("  %s %s\n", color.HiMagentaString("--version"), color.WhiteString("       Prints the version of the program and exits"))
		fmt.Println(color.HiCyanString("\nArguments:"))
		fmt.Printf("  %s %s\n", color.HiMagentaString("<input_file>"), color.WhiteString("    The path to the input FASTA file or '-' for standard input (stdin)"))
		fmt.Printf("  %s %s\n", color.HiMagentaString("[output_file]"), color.WhiteString("   The path to the output file or '-' for standard output (stdout)"))
		fmt.Println(color.WhiteString("                   If omitted, defaults to stdout."))
		fmt.Println(color.HiCyanString("\nExamples:"))
		fmt.Println(color.WhiteString("  seqhasher input.fasta output.fasta"))
		fmt.Println(color.WhiteString("  cat input.fasta | seqhasher --name 'Sample' - - > output.fasta"))
	} else {
		fmt.Fprintf(os.Stderr, "SeqHasher v%s\n", version)
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <input_file> [output_file]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nSupported hash types: %s\n", strings.Join(supportedHashTypes, ", "))
		fmt.Fprintf(os.Stderr, "If input_file is '-' or omitted, reads from stdin.\n")
		fmt.Fprintf(os.Stderr, "If output_file is '-' or omitted, writes to stdout.\n")
		fmt.Fprintf(os.Stderr, "\nFor more detailed help, use -h or --help.\n")
	}
}

func processSequences(input io.Reader, output io.Writer, cfg config) {
	writer := bufio.NewWriter(output)

	reader, err := fastx.NewReaderFromIO(seq.DNA, bufio.NewReader(input), fastx.DefaultIDRegexp)
	if err != nil {
		log.Fatalf("Failed to create reader: %v", err)
	}
	defer reader.Close() // Close the reader after processing

	inputFileName := cfg.inputFileName
	if cfg.nameOverride != "" {
		inputFileName = cfg.nameOverride
	} else if inputFileName != "-" {
		inputFileName = filepath.Base(inputFileName)
	}

	hashFunc := getHashFunc(cfg.hashType)

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
		if !cfg.caseSensitive {
			// Convert sequence to uppercase if case-insensitive hashing is enabled
			seq = bytes.ToUpper(seq)
		}

		hashedSeq := hashFunc(seq)

		// Prepare the new sequence header
		var modifiedHeader string
		if cfg.noFileName {
			modifiedHeader = fmt.Sprintf("%s;%s", hashedSeq, record.Name)
		} else {
			modifiedHeader = fmt.Sprintf("%s;%s;%s", inputFileName, hashedSeq, record.Name)
		}

		if cfg.headersOnly {
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
