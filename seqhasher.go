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
	"strings"

	"github.com/shenwei356/bio/seq"
	"github.com/shenwei356/bio/seqio/fastx"

	"github.com/cespare/xxhash/v2"
	"github.com/go-faster/city"
	"github.com/spaolacci/murmur3"
	"github.com/zeebo/blake3"
	"golang.org/x/crypto/sha3"

	"github.com/fatih/color"
	"github.com/will-rowe/nthash"
)

const (
	version         = "1.1.1" // Version of the program
	defaultHashType = "sha1"  // Default hash type
)

var supportedHashTypes = []string{"sha1", "sha3", "md5", "xxhash", "cityhash", "murmur3", "nthash", "blake3"}

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

func main() {
	if err := run(os.Stdout); err != nil {
		log.Fatalf("%v", err)
	}
}

func run(w io.Writer) error {

	// Disable sequence validation
	seq.ValidateSeq = false

	cfg, err := parseFlags()
	if err != nil {
		return err
	}

	if cfg.showVersion {
		fmt.Fprintf(w, "SeqHasher %s\n", version)
		return nil
	}

	if cfg.inputFileName == "" {
		printUsage(w)
		return nil
	}

	input, err := getInput(cfg.inputFileName)
	if err != nil {
		return fmt.Errorf("Error opening input: %v", err)
	}
	defer input.Close()

	output := w
	if cfg.outputFileName != "" && cfg.outputFileName != "-" {
		outputFile, err := getOutput(cfg.outputFileName)
		if err != nil {
			return fmt.Errorf("Error opening output: %v", err)
		}
		defer outputFile.Close()
		output = outputFile
	}

	return processSequences(input, output, cfg)
}

func parseFlags() (config, error) {
	cfg := config{}

	flag.BoolVar(&cfg.headersOnly, "headersonly", false, "Output only headers")
	flag.BoolVar(&cfg.headersOnly, "o", false, "Output only headers (shorthand)")

	var hashTypesString string
	flag.StringVar(&hashTypesString, "hash", defaultHashType, "Hash type(s) (comma-separated: sha1, sha3, md5, xxhash, cityhash, murmur3, nthash, blake3)")
	flag.StringVar(&hashTypesString, "H", defaultHashType, "Hash type(s) (shorthand)")

	flag.BoolVar(&cfg.noFileName, "nofilename", false, "Do not include file name in output")
	flag.BoolVar(&cfg.noFileName, "n", false, "Do not include file name in output (shorthand)")

	flag.BoolVar(&cfg.caseSensitive, "casesensitive", false, "Case-sensitive hashing")
	flag.BoolVar(&cfg.caseSensitive, "c", false, "Case-sensitive hashing (shorthand)")

	flag.StringVar(&cfg.nameOverride, "name", "", "Override input file name in output")
	flag.StringVar(&cfg.nameOverride, "f", "", "Override input file name in output (shorthand)")

	flag.BoolVar(&cfg.showVersion, "version", false, "Show version information")
	flag.BoolVar(&cfg.showVersion, "v", false, "Show version information (shorthand)")

	flag.Usage = func() {
		printUsage(os.Stderr)
	}
	flag.Parse()

	cfg.inputFileName = flag.Arg(0)
	cfg.outputFileName = flag.Arg(1)

	// Parse hash types
	cfg.hashTypes = strings.Split(hashTypesString, ",")
	for _, ht := range cfg.hashTypes {
		if !isValidHashType(strings.TrimSpace(ht)) {
			return config{}, fmt.Errorf("Invalid hash type: %s. Supported types are: %s", ht, strings.Join(supportedHashTypes, ", "))
		}
	}

	return cfg, nil
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

func printUsage(w io.Writer) {
	if len(os.Args) > 1 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		fmt.Fprintf(w, "\n%s%s%s\n",
			color.HiGreenString("SeqHasher"),
			color.WhiteString(" : "),
			color.HiMagentaString("DNA Sequence Hashing Tool"))
		fmt.Fprintf(w, "%s  %s\n", color.HiCyanString("version:"), color.WhiteString(version))
		fmt.Fprintln(w, color.WhiteString("====================================="))
		fmt.Fprintln(w, color.HiCyanString("Usage:"))
		fmt.Fprintf(w, "  %s\n", color.WhiteString("seqhasher [options] <input_file> [output_file]"))
		fmt.Fprintln(w, color.HiCyanString("\nOverview:"))
		fmt.Fprintln(w, color.WhiteString("  SeqHasher takes DNA sequences from a FASTA file, computes a hash digest for each sequence,"))
		fmt.Fprintln(w, color.WhiteString("  and generates an output file with modified headers."))
		fmt.Fprintln(w, color.WhiteString("  For input/output via stdin/stdout, use '-' instead of the file name."))
		fmt.Fprintln(w, color.HiCyanString("\nOptions:"))
		fmt.Fprintf(w, "  %s, %s %s\n", color.HiMagentaString("-o"), color.HiMagentaString("--headersonly"), color.WhiteString("  Output only sequence headers, excluding the sequences themselves"))
		fmt.Fprintf(w, "  %s, %s %s\n", color.HiMagentaString("-H"), color.HiMagentaString("--hash <type1,type2,...>"), color.WhiteString("Hash algorithm(s): sha1 (default), sha3, md5, xxhash, cityhash, murmur3, nthash, blake3"))
		fmt.Fprintf(w, "  %s, %s %s\n", color.HiMagentaString("-c"), color.HiMagentaString("--casesensitive"), color.WhiteString("Take into account sequence case. By default, sequences are converted to uppercase"))
		fmt.Fprintf(w, "  %s, %s %s\n", color.HiMagentaString("-n"), color.HiMagentaString("--nofilename"), color.WhiteString("   Omit the file name from the sequence header"))
		fmt.Fprintf(w, "  %s, %s %s\n", color.HiMagentaString("-f"), color.HiMagentaString("--name <text>"), color.WhiteString("  Replace the input file's name in the header with <text>"))
		fmt.Fprintf(w, "  %s, %s %s\n", color.HiMagentaString("-v"), color.HiMagentaString("--version"), color.WhiteString("      Print the version of the program and exit"))
		fmt.Fprintf(w, "  %s, %s %s\n", color.HiMagentaString("-h"), color.HiMagentaString("--help"), color.WhiteString("         Show this help message and exit"))
		fmt.Fprintln(w, color.HiCyanString("\nArguments:"))
		fmt.Fprintf(w, "  %s %s\n", color.HiMagentaString("<input_file>"), color.WhiteString("    Path to the input FASTA file (supports gzip, zstd, xz, or bzip2 compression)"))
		fmt.Fprintf(w, "  %s\n", color.WhiteString("                 or '-' for standard input (stdin)"))
		fmt.Fprintf(w, "  %s %s\n", color.HiMagentaString("[output_file]"), color.WhiteString("   Path to the output file or '-' for standard output (stdout)"))
		fmt.Fprintln(w, color.WhiteString("                   If omitted, output is sent to stdout."))
		fmt.Fprintln(w, color.HiCyanString("\nExamples:"))
		fmt.Fprintln(w, color.WhiteString("  seqhasher input.fasta.gz output.fasta"))
		fmt.Fprintln(w, color.WhiteString("  cat input.fasta | seqhasher --name 'Sample' --hash xxhash - - > output.fasta"))
		fmt.Fprintln(w, color.WhiteString("  seqhasher --headersonly --nofilename --hash sha1,nthash input.fa.gz - > headers.txt"))
		fmt.Fprintln(w, color.WhiteString("\nFor more information, visit the GitHub repository:"))
		fmt.Fprintln(w, color.WhiteString("https://github.com/vmikk/seqhasher"))
	} else {
		fmt.Fprintf(w, "SeqHasher v%s\n", version)
		fmt.Fprintf(w, "Usage: %s [options] <input_file> [output_file]\n", os.Args[0])
		fmt.Fprintf(w, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(w, "\nSupported hash types: %s\n", strings.Join(supportedHashTypes, ", "))
		fmt.Fprintf(w, "If input_file is '-' or omitted, reads from stdin.\n")
		fmt.Fprintf(w, "If output_file is '-' or omitted, writes to stdout.\n")
		fmt.Fprintf(w, "\nFor more detailed help, use -h or --help.\n")
	}
}

func processSequences(input io.Reader, output io.Writer, cfg config) error {
	writer := bufio.NewWriter(output)
	defer writer.Flush()

	inputFileName := cfg.inputFileName
	if cfg.nameOverride != "" {
		inputFileName = cfg.nameOverride
	} else if inputFileName == "-" {
		cfg.noFileName = true // Skip filename for stdin unless overridden
	}

	reader, err := fastx.NewReaderFromIO(seq.DNA, bufio.NewReader(input), fastx.DefaultIDRegexp)
	if err != nil {
		return fmt.Errorf("Failed to create reader: %v", err)
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
			return fmt.Errorf("Error reading record: %v", err)
		}

		seq := record.Seq.Seq

		// Strip all whitespace characters from sequence before processing
		// (as defined by Unicode's White Space property, which includes
		// '\t', '\n', '\v', '\f', '\r', ' ', U+0085 (NEL), U+00A0 (NBSP)
		seq = bytes.Join(bytes.Fields(seq), nil)

		// Convert sequence to uppercase if case-insensitive hashing is enabled
		if !cfg.caseSensitive {
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
				return fmt.Errorf("Error writing header: %v", err)
			}
		} else {
			// Output the full record
			if _, err := fmt.Fprintf(writer, ">%s\n%s\n", modifiedHeader, seq); err != nil {
				return fmt.Errorf("Error writing record: %v", err)
			}
		}
	}

	return writer.Flush()
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

		case "sha1":
			hash := sha1.Sum(data)
			return hex.EncodeToString(hash[:])
		case "sha3":
			hash := sha3.Sum512(data)
			return hex.EncodeToString(hash[:])
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
