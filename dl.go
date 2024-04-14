
package main

import (
	"fmt"
	"os"
	fp "path/filepath"
	"io/fs"
	"log"
	"sort"
	"flag"
)

// Rounding with -p and -c
// */=>@|  :  Socket=, Fifo|, Door> (Solaris, unused)
// golang.org/x/term

/*
	fs.DirEntry { also as os.DirEntry 
		Name()
		IsDir()
		Type() FileMode uint32 member fcts: IsDir, IsRegular
		Info() FileInfo
	}

	fs.FileInfo { also as os.FileInfo
		Name()
		Size()
		Mode()
		Time()
		IsDir()
		Sys()
	}

	os.Stat( string ) FileInfo
*/

var MAXDEPTH =  16

type Options struct {
	abspath string
	topn, percent, cumul int
	divisor int64
	suffix bool
}

type Entry struct {
	abspath, suffix string
	bytes, files, dirs, skipped int64	
}

func addTo( e1, e2 Entry ) Entry {
	e1.bytes += e2.bytes
	e1.files += e2.files
	e1.dirs += e2.dirs
	e1.skipped += e2.skipped
	return e1
}

func parse_args() Options {
	// Setup options
	bytes := flag.Bool( "b", false,
		"Display file sizes in bytes" )
	kilos := flag.Bool( "k", false,
		"Display file sizes in multiples of 1000 bytes" )
	kBs := flag.Bool( "x", false,
		"Display file sizes in multiples of 1024 bytes (default)" )
	
	topn := flag.Int( "n", -1,
		"Display only top <n> entries; show all if negative" )
	percent := flag.Int( "p", -1,
		"Display only entries greater than <p>%; show all if negative" )
	cumul := flag.Int( "c", -1, 
		"Display only entries up to <c>% cumulative; show all if negative" )

	nosuffix := flag.Bool( "F", false, "Suppress file type suffix (@/=)" )

	depth := flag.Int( "R", 16, "Maximum recursion depth" )
	
	help := flag.Bool( "h", false, "Display usage" )

	// Parse and usage
	flag.Parse()	
	
	if *help || flag.NArg() > 1 {
		fmt.Fprintf( flag.CommandLine.Output(),
			"Usage: %s [flags] [path]\n", os.Args[0] )
		flag.PrintDefaults()
		fmt.Fprintf( flag.CommandLine.Output(),
			"Current directory is used if no path is specified\n" )
		
		os.Exit(0)
	}

	// Handle path argument
	path := flag.Args()  // []string

	root := ""
	if len(path) > 0 {
		root = path[0]		
	} else {
		tmp, err := os.Getwd()
		if err != nil { log.Fatalln( err ) }
		root = tmp
	}
	
	fi, err := os.Stat( root )
	if err != nil { log.Fatalln( err ) }
	
	if !fi.IsDir() {
		fmt.Printf( "%s is not a directory", root )
		os.Exit(0)
	}

	root, err = fp.Abs( root )
	if err != nil { log.Fatalln( err ) }
	
	// Create Options object
	opts := Options{ abspath: root, divisor: int64(1024), suffix: true }

	if *kilos { opts.divisor = int64(1000) }
	if *bytes { opts.divisor = int64(1) }   // options prevail in this order...
	if *kBs { opts.divisor = int64(1024) }

	if topn != nil { opts.topn = *topn }
	if percent != nil { opts.percent = *percent }
	if cumul != nil { opts.cumul = *cumul }

	if depth != nil { MAXDEPTH = *depth }
	
	if *nosuffix { opts.suffix = false }
	
	return opts
}

func main() {
	log.SetFlags( 0 ) // disable prefix for log statements

	opts := parse_args()

	out := process_root( opts.abspath )

	output( out, opts )
}

func process_root( abs string ) []Entry {
	// Input is dir: examine entries
	items, err := os.ReadDir( abs )
	if err != nil { log.Fatalln( err ) }

	out := []Entry{}
	for _, item := range items {
		entry := Entry{ abspath: fp.Join( abs, item.Name() ) }
		
//		switch mode := item.Type() {
//		switch mode := item.Type(); mode {
		
		switch mode := item.Type(); {
		case mode&fs.ModeSymlink != 0:
			entry.suffix = "@"
			entry.skipped += 1
			
		case mode.IsRegular():
			info, err := item.Info()
			if err != nil { fmt.Fprintln( os.Stderr, err ) }

			entry.files += 1
			entry.bytes += info.Size()
			
		case mode.IsDir():
			entry.suffix = "/"
			
			tmp, err := visit( fp.Join(abs, item.Name()), 0 )
			if err != nil { fmt.Fprintln( os.Stderr, err ) }

			entry = addTo( entry, tmp )
	
		default:
			entry.suffix = "="
			entry.skipped += 1
		}

		out = append( out, entry )
	}

	sort.Slice( out, func(i, j int) bool {
		return out[i].bytes > out[j].bytes } )

	return out
}

func visit( abspath string, depth int ) ( Entry, error ) {
	// abspath MUST be a dir: panic if not
	
	if depth > MAXDEPTH { return Entry{}, nil }
	
	items, err := os.ReadDir( abspath )  // returns fs.DirEntry
	if err != nil { return Entry{}, err }
	
	cntr := Entry{}
	for _, item := range items {
		switch mode := item.Type(); {       // FileMode, uint32

		case mode&fs.ModeSymlink != 0:
			cntr.skipped += 1
			
		case mode.IsRegular():
			info, err := item.Info()
			if err != nil { return Entry{}, err }

			cntr.files += 1
			cntr.bytes += info.Size()
			
		case mode.IsDir():
			tmp, err := visit( fp.Join(abspath, item.Name()), depth+1 )
			if err != nil { return Entry{}, err }
			
			cntr = addTo( cntr, tmp )
			
		default:
			cntr.skipped += 1
		} 		
	}

	return cntr, nil
}

func output( out []Entry, opts Options ) {
	var ttl int64
	for _, entry := range out {
		ttl += entry.bytes
	}
	norm := float64(ttl)/100.0

	f := func(x int64) string { return format(x, opts.divisor ) }
	
	var cum int64
	hidden := len(out)
	for i, entry := range out {
		cum += entry.bytes

		percent := float64(entry.bytes)/norm
		cumpercent := float64(cum)/norm
		
		if ( opts.topn >= 0 && opts.topn <= i ) ||
			percent < float64(opts.percent) ||
			( opts.cumul >=0 && float64(opts.cumul) < cumpercent ) {
			if hidden > 1 {
				cum -= entry.bytes
				break
			}
		}

		suffix := ""
		if opts.suffix == true {
			suffix = entry.suffix
		}
		
		fmt.Printf( "%s\t%3.f%%\t%s\t%3.f%%\t%s%s\n", f(entry.bytes),
			float64(entry.bytes)/norm, f(cum), float64(cum)/norm,
			fp.Base(entry.abspath), suffix )

		hidden -= 1		
	}

	if hidden > 0 {
		fmt.Printf(
			"Omitting %d lines representing %s bytes, %.f%% of total bytes\n",
			hidden, f(ttl-cum), float64(ttl-cum)/norm )
	}
}

func format( x, divisor int64 ) string {
	units := [...]string{ "", "k", "M", "G", "T", "P", "E", "Z", "Y" }

	if x == 0 {
		return "0.0"
	}

	if divisor == 1 {
		return fmt.Sprintf( "%3d", x )
	}

	i, y := 0, x%divisor
	for ; x>divisor; {
		y = x%divisor
		x /= divisor
		i += 1
	}

	unit := ""
	if i > len(units) {
		unit = fmt.Sprintf( "E%d", 3*i )
	} else {
		unit = units[i]
	}
	
	if x < 10 {
		return fmt.Sprintf( "%3.1f%s",
			float64(x) + float64(y)/float64(divisor) + 0.05, unit ) // round up
	} else {
		return fmt.Sprintf( "%3d%s", x, unit )
	}
}

