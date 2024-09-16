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
	"github.com/zeebo/blake3"

	"github.com/fatih/color"
	"github.com/will-rowe/nthash"
)

const (
	version         = "1.0.2" // Version of the program
	defaultHashType = "sha1"  // Default hash type
)

var supportedHashTypes = []string{"sha1", "md5", "xxhash", "cityhash", "murmur3", "nthash", "blake3"}

// Configuration structure (flags)
type config struct {
	headersOnly    bool
	hashTypes      []string
	noFileName     bool
	caseSensitive  bool
	inputFileName  string
	outputFileName string
	nameOverride   string
	showVersion    bool
}

func run() string {
	cfg := parseFlags()

	if cfg.showVersion {
		return fmt.Sprintf("SeqHasher %s\n", version)
	}

	if cfg.inputFileName == "" {
		var buf bytes.Buffer
		printUsage()
		return buf.String()
	}

	input, err := getInput(cfg.inputFileName)
	if err != nil {
		return fmt.Sprintf("Error opening input: %v\n", err)
	}
	defer input.Close()

	var outputBuffer bytes.Buffer
	output := bufio.NewWriter(&outputBuffer)

	processSequences(input, output, cfg)

	if err := output.Flush(); err != nil {
		return fmt.Sprintf("Error flushing output: %v\n", err)
	}

	return outputBuffer.String()
}

func main() {
	fmt.Print(run())
}

func parseFlags() config {
	cfg := config{}

	flag.BoolVar(&cfg.headersOnly, "headersonly", false, "Output only headers")
	flag.BoolVar(&cfg.headersOnly, "o", false, "Output only headers (shorthand)")

	var hashTypesString string
	flag.StringVar(&hashTypesString, "hash", defaultHashType, "Hash type(s) (comma-separated: sha1, md5, xxhash, cityhash, murmur3, nthash, blake3)")
	flag.StringVar(&hashTypesString, "H", defaultHashType, "Hash type(s) (shorthand)")

	flag.BoolVar(&cfg.noFileName, "nofilename", false, "Do not include file name in output")
	flag.BoolVar(&cfg.noFileName, "n", false, "Do not include file name in output (shorthand)")

	flag.BoolVar(&cfg.caseSensitive, "casesensitive", false, "Case-sensitive hashing")
	flag.BoolVar(&cfg.caseSensitive, "c", false, "Case-sensitive hashing (shorthand)")

	flag.StringVar(&cfg.nameOverride, "name", "", "Override input file name in output")
	flag.StringVar(&cfg.nameOverride, "f", "", "Override input file name in output (shorthand)")

	flag.BoolVar(&cfg.showVersion, "version", false, "Show version information")
	flag.BoolVar(&cfg.showVersion, "v", false, "Show version information (shorthand)")

	flag.Usage = printUsage
	flag.Parse()

	args := flag.Args()
	if len(args) > 0 {
		cfg.inputFileName = args[0]
	}
	if len(args) > 1 {
		cfg.outputFileName = args[1]
	}

	// Parse hash types
	cfg.hashTypes = strings.Split(hashTypesString, ",")
	for i, ht := range cfg.hashTypes {
		cfg.hashTypes[i] = strings.TrimSpace(ht)
		if !isValidHashType(cfg.hashTypes[i]) {
			log.Fatalf("Invalid hash type: %s. Supported types are: %s", cfg.hashTypes[i], strings.Join(supportedHashTypes, ", "))
		}
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
		fmt.Printf("  %s, %s %s\n", color.HiMagentaString("-o"), color.HiMagentaString("--headersonly"), color.WhiteString("   Only output sequence headers, excluding the sequences themselves"))
		fmt.Printf("  %s, %s %s\n", color.HiMagentaString("-H"), color.HiMagentaString("--hash <type1,type2,...>"), color.WhiteString("   Hash algorithm(s): sha1 (default), md5, xxhash, cityhash, murmur3, nthash, blake3"))
		fmt.Printf("  %s, %s %s\n", color.HiMagentaString("-n"), color.HiMagentaString("--nofilename"), color.WhiteString("    Omit the file name from the sequence header"))
		fmt.Printf("  %s, %s %s\n", color.HiMagentaString("-c"), color.HiMagentaString("--casesensitive"), color.WhiteString(" Take into account sequence case. By default, sequences are converted to uppercase"))
		fmt.Printf("  %s, %s %s\n", color.HiMagentaString("-f"), color.HiMagentaString("--name <text>"), color.WhiteString("   Replace the input file's name in the header with <text>"))
		fmt.Printf("  %s, %s %s\n", color.HiMagentaString("-v"), color.HiMagentaString("--version"), color.WhiteString("       Print the version of the program and exit"))
		fmt.Printf("  %s, %s %s\n", color.HiMagentaString("-h"), color.HiMagentaString("--help"), color.WhiteString("          Show this help message and exit"))
		fmt.Println(color.HiCyanString("\nArguments:"))
		fmt.Printf("  %s %s\n", color.HiMagentaString("<input_file>"), color.WhiteString("    The path to the input FASTA file or '-' for standard input (stdin)"))
		fmt.Printf("  %s %s\n", color.HiMagentaString("[output_file]"), color.WhiteString("   The path to the output file or '-' for standard output (stdout)"))
		fmt.Println(color.WhiteString("                   If omitted, defaults to stdout."))
		fmt.Println(color.HiCyanString("\nExamples:"))
		fmt.Println(color.WhiteString("  seqhasher input.fasta.gz output.fasta"))
		fmt.Println(color.WhiteString("  cat input.fasta | seqhasher --name 'Sample' --hash xxhash - - > output.fasta"))
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

	hashFuncs := make([]func([]byte) string, len(cfg.hashTypes))
	for i, hashType := range cfg.hashTypes {
		hashFuncs[i] = getHashFunc(hashType)
	}

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

		var hashedSeqs []string
		for _, hashFunc := range hashFuncs {
			hashedSeqs = append(hashedSeqs, hashFunc(seq))
		}

		// Join all hashes
		hashedSeq := strings.Join(hashedSeqs, ";")

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
	return func(data []byte) string {
		if len(data) == 0 {
			log.Printf("Error: Empty DNA sequence provided, resulting in an empty hash.")
			return ""
		}

		switch hashType {
		case "md5":
			hash := md5.Sum(data)
			return hex.EncodeToString(hash[:])
		case "xxhash":
			hash := xxhash.Sum64(data)
			return fmt.Sprintf("%016x", hash)
		case "cityhash":
			hash := city.Hash128(data)
			return fmt.Sprintf("%016x%016x", hash.High, hash.Low)
		case "murmur3":
			h1, h2 := murmur3.Sum128(data)
			return fmt.Sprintf("%016x%016x", h1, h2)
		case "nthash":
			hasher, err := nthash.NewHasher(&data, uint(len(data)))
			if err != nil {
				log.Printf("Error creating ntHash hasher: %v", err)
				return ""
			}
			hash, _ := hasher.Next(false) // false for non-canonical hash
			return fmt.Sprintf("%016x", hash)
		case "blake3":
			hash := blake3.Sum256(data)
			return hex.EncodeToString(hash[:])
		default: // Default to SHA1
			hash := sha1.Sum(data)
			return hex.EncodeToString(hash[:])
		}
	}
}
