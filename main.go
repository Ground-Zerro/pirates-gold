package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"pirates-gold/bip39"
	"pirates-gold/checker"
	"pirates-gold/results"
	"pirates-gold/stats"
	"pirates-gold/wallet"
)

const (
	version        = "1.2"
	author         = "Ground_Zerro"
	defaultWorkers = 2
	defaultRate    = 1.0
)

func main() {
	workers   := flag.Int("workers", defaultWorkers, "number of parallel workers for blockchain.info")
	rate      := flag.Float64("rate", defaultRate, "max requests/sec for blockchain.info (blockstream.info is fixed at 0.15 req/s)")
	outDir    := flag.String("out", ".", "output directory for result files")
	count     := flag.Int64("count", 0, "stop after N checks (0 = infinite)")
	showVer   := flag.Bool("version", false, "show version and exit")
	showStats := flag.Bool("stats", false, "show total statistics across all sessions")
	svcFlag   := flag.Bool("service", false, "install or remove systemd service")

	flag.Usage = printHelp
	flag.Parse()

	switch {
	case *showVer:
		fmt.Printf("Pirates Gold v.%s — author: %s\n", version, author)
		return
	case *showStats:
		printTotalStats(*outDir)
		return
	case *svcFlag:
		handleService(*outDir)
		return
	}

	fmt.Printf("Pirates Gold v.%s started. Author: %s\n\n", version, author)

	st, err := stats.New(*outDir)
	if err != nil {
		log.Fatal(err)
	}
	defer st.Close()

	writer, err := results.NewWriter(*outDir)
	if err != nil {
		log.Fatal(err)
	}
	defer writer.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	go func() {
		<-ctx.Done()
		fmt.Println("\nShutting down, please wait...")
	}()

	spawnWorker := func(wg *sync.WaitGroup, c *checker.Checker) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}

				if *count > 0 && st.Checked.Load() >= *count {
					cancel()
					return
				}

				mnemonic, err := bip39.GenerateMnemonic()
				if err != nil {
					log.Printf("mnemonic error: %v", err)
					continue
				}

				address, err := wallet.MnemonicToAddress(mnemonic)
				if err != nil {
					log.Printf("derive error: %v", err)
					continue
				}

				result, err := c.Check(ctx, address)
				if err != nil {
					if ctx.Err() == nil {
						log.Printf("check error [%s]: %v", address, err)
					}
					continue
				}

				writer.Write(mnemonic, address, result.BalanceSat, result.TxCount)
				st.Checked.Add(1)

				switch {
				case result.BalanceSat > 0:
					st.Found.Add(1)
					fmt.Printf("*** JACKPOT *** %s | %s | %d sat\n",
						mnemonic, address, result.BalanceSat)
				case result.TxCount > 0:
					st.Used.Add(1)
					fmt.Printf("[USED] %s | %s | %d tx\n",
						mnemonic, address, result.TxCount)
				}
			}
		}()
	}

	var wg sync.WaitGroup
	spawnWorker(&wg, checker.NewBlockstreamChecker())
	bciChecker := checker.NewBlockchainInfoChecker(*rate)
	for i := 0; i < *workers; i++ {
		spawnWorker(&wg, bciChecker)
	}

	wg.Wait()
	fmt.Printf("\nChecked: %d | Used: %d | Found: %d\n",
		st.Checked.Load(), st.Used.Load(), st.Found.Load())
}

func printTotalStats(dir string) {
	total, err := stats.ParseTotal(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading stats: %v\n", err)
		os.Exit(1)
	}
	sep := strings.Repeat("═", 42)
	fmt.Printf("Pirates Gold v.%s — Total Statistics\n", version)
	fmt.Println(sep)
	fmt.Printf("  Sessions : %d\n", total.Sessions)
	fmt.Printf("  Runtime  : %s\n", formatDuration(total.Seconds))
	fmt.Printf("  Checked  : %s\n", formatNumber(total.Checked))
	fmt.Printf("  Used     : %d\n", total.Used)
	fmt.Printf("  Found    : %d\n", total.Found)
	fmt.Println(sep)
}

func printHelp() {
	fmt.Printf("Pirates Gold v.%s — author: %s\n\n", version, author)
	fmt.Println("Generates random BIP39 seed phrases, derives Bitcoin wallets")
	fmt.Println("(P2PKH, m/44'/0'/0'/0/0) and checks balance via two public APIs:\n")
	fmt.Println("  blockstream.info  — fixed rate 0.15 req/s (respects 700 req/hour limit)")
	fmt.Println("  blockchain.info   — configurable via -rate flag\n")
	fmt.Println("Usage:")
	fmt.Println("  pirates-gold [flags]\n")
	fmt.Println("Flags:")
	fmt.Println("  -workers int    workers for blockchain.info (default: 2)")
	fmt.Println("  -rate float     max req/s for blockchain.info (default: 1.0)")
	fmt.Println("  -out string     output directory for result files (default: .)")
	fmt.Println("  -count int      stop after N checks, 0 = infinite (default: 0)")
	fmt.Println("  -service        install or remove systemd service")
	fmt.Println("  -stats          show total statistics across all sessions")
	fmt.Println("  -version        show version and exit")
	fmt.Println("  -h, --help      show this help\n")
	fmt.Println("Output files:")
	fmt.Println("  used.txt        wallets with transaction history (zero balance)")
	fmt.Println("  found.txt       wallets with positive balance")
	fmt.Println("  stats.txt       per-session statistics\n")
	fmt.Println("Examples:")
	fmt.Println("  pirates-gold -workers 4 -rate 2.0 -out /data/results")
	fmt.Println("  pirates-gold -service")
	fmt.Println("  pirates-gold -stats")
}

func formatDuration(seconds int64) string {
	if seconds == 0 {
		return "0s"
	}
	d := seconds / 86400
	h := (seconds % 86400) / 3600
	m := (seconds % 3600) / 60
	s := seconds % 60
	if d > 0 {
		return fmt.Sprintf("%dd %dh %02dm %02ds", d, h, m, s)
	}
	return fmt.Sprintf("%dh %02dm %02ds", h, m, s)
}

func formatNumber(n int64) string {
	s := fmt.Sprintf("%d", n)
	var b strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(' ')
		}
		b.WriteRune(c)
	}
	return b.String()
}
