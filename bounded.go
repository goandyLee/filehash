package main

import (
        "path/filepath"
        "time"  
        "errors"
        "fmt"   
        "os"    
        "io/ioutil"
        "sync"  
        "bytes"  
        "crypto/sha1"
        "runtime"
        "sort"
)

type result struct {
    path string
    filename string
    filesha1 [sha1.Size]byte
    filesize int64
    err  error
}

var (
    pattern string
)

func check(err error) { if err != nil {
        panic(err)
    }
}

func walkFiles(done <-chan struct{}, root string) (<-chan string, <-chan error) {
    paths := make(chan string)
    errc := make(chan error, 1)
    go func() {
        // Close the paths channel after Walk returns.
        defer close(paths)
        // No select needed for this send, since errc is buffered.
        errc <- filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
            if err != nil {
                return err
            }
            if info.Mode().IsDir() {
                return nil
            }
            if ok, _ := filepath.Match(pattern, info.Name()); ok {
               return nil
            }
            select {
            case paths <- path:
            case <-done:
                return errors.New("walk canceled")
            }
            return nil
        })
    }()
    return paths, errc
}

func digester(done <-chan struct{}, paths <-chan string, c chan<- result) {
    for path := range paths {
        data, err := ioutil.ReadFile(path)
        tempFileInfo, _ := os.Stat(path)
        select {
        case c <- result{path,tempFileInfo.Name(), sha1.Sum(data), tempFileInfo.Size(), err}:
        case <-done:
            return
        }
    }
}

func MD5All(root string) (map[string]string, error) {
    // MD5All closes the done channel when it returns; it may do so before
    // receiving all the values from c and errc.
    done := make(chan struct{})
    defer close(done)

    paths, errc := walkFiles(done, root)

    // Start a fixed number of goroutines to read and digest files.
    c := make(chan result)
    var wg sync.WaitGroup
    NCPU := runtime.NumCPU() * 10
    fmt.Printf("NCPU:%d\n",NCPU)
    runtime.GOMAXPROCS(NCPU)

    wg.Add(NCPU)
    for i := 0; i < NCPU; i++ {
        go func() {
            digester(done, paths, c)
            wg.Done()
        }()
    }
    go func() {
        wg.Wait()
        close(c)
    }()

    m := make(map[string]string)
    for r := range c {
        if r.err != nil {
            return nil, r.err
        }
        m[r.path] = fmt.Sprintf("%s, %X, %d\n", r.filename, r.filesha1, r.filesize)
    }
    if err := <-errc; err != nil {
        return nil, err
    }
    return m, nil
}

func main() {
    // Calculate the MD5 sum of all files under the specified directory,
    // then print the results sorted by path name.
    var buffer bytes.Buffer
    if argc := len(os.Args); argc != 3{
        fmt.Println("usage: filehash DIR PATTERN")
        return
    }

    pattern = os.Args[2]

    t1 := time.Now()
    m, err := MD5All(os.Args[1])
    if err != nil {
        fmt.Println(err)
        return
    }
    var paths []string
    for path := range m {
        paths = append(paths, path)
    }
    sort.Strings(paths)
    hashmapfile, err1 := os.OpenFile("./hashmap2.txt", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
    check(err1)
    defer hashmapfile.Close()
    for _, path := range paths {
        buffer.WriteString(m[path])
    }
    _, err6 := hashmapfile.WriteString(buffer.String())
    check(err6)

    t2 := time.Now()
    fmt.Printf("cost:%d\n", t2.Sub(t1))
    fmt.Printf("cost:%d\n", t2.Sub(t1)/1000000000)
}
