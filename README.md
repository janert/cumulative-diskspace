# Cumulative Diskspace: `dl`

Tool to display directory entries, sorted by size, together with their 
cumulative contribution to the total. 


## Description

This tool displays directory entries, sorted by size, together with the
cumulative contribution they make to the total space consumed. It can
be used to identify space hogs and to assess their impact.

The listing includes directories and files, both regular and hidden.

The tool does not follow symbolic links. The tool skips files and
directories it does not have permissions to read, printing an error
message to STDERR; in this case, the calculated sizes will be
underreported.

The tool reports sizes as reported by the `os.FileInfo.Size()`
function. I believe this is the "apparent size" of the file,
which is the number of valid bytes that can be read from it.
Other tools (such as `du`) may instead report the number of
bytes in the filesystem blocks allocated to the file; this
number will typically be larger.


## Usage 

```
Usage: dl [flags] [path]
  -n int
        Only show top <n> entries; show all if negative (default -1)
  -p int
        Only show entries greater than <p>%;
	show all if negative (default -1)
  -c int
        Only show entries up to <c>% cumulative;
	show all if negative (default -1)
	
  -b    Display file sizes in bytes
  -k    Display file sizes in multiples of 1000 bytes
  -x    Display file sizes in multiples of 1024 bytes (default)

  -F    Suppress file type suffix (@/=)

  -R int
        Set the maximum recursion depth (default 16)

  -h    Display usage
Current directory is used if no path is specified
```

For options that take an argument (`-n`, `-p`, `-c`, and `-R`),
there **must be a space** between the flag and the argument! (In other
words, `dl -n20` is not valid.)**


## Example:

```
shell> dl /usr/
 16G     65%     16G     65%    lib/
5.9G     23%     22G     87%    share/
1.6G      6%     23G     94%    src/
1.4G      5%     25G     99%    bin/
170M      1%     25G     99%    include/
 71M      0%     25G    100%    sbin/
 36M      0%     25G    100%    libexec/
 16M      0%     25G    100%    local/
 13M      0%     25G    100%    lib32/
2.9M      0%     25G    100%    games/
0.0       0%     25G    100%    libx32/
0.0       0%     25G    100%    lib64/
```

In this case, the largest 3 directories occupy almost 95% of the
total; the remaining 9 directories make up the rest.
