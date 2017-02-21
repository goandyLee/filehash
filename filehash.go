package main

import (
	"path/filepath"
        "crypto/sha1"
        "time"
	"fmt"
	"os"
        "io"
)

var (
    pattern = "*go"
    count =0
    hashmapfile *os.File
)

func check(err error) {
    if err != nil {
        panic(err)
    }
}

func makerHash(filepath string) {
    tempFile, err1 := os.Open(filepath)
    check(err1)
    defer tempFile.Close()

    fileHash := sha1.New()
    _, err2 := io.Copy(fileHash, tempFile)
    check(err2)

    tempFileInfo, err3 := tempFile.Stat()
    check(err3)

    tempStr := fmt.Sprintf("%s,%X,%d\n", tempFileInfo.Name(), fileHash.Sum(nil), tempFileInfo.Size())
    //fmt.Println(tempStr)
    hashmapfile.WriteString(tempStr)
    count++
}

/**
 * walkFunc:  
 * Desc:
 * Return:
 */
func walkFunc(path1 string, info os.FileInfo, err error) error {
    ok, err := filepath.Match(pattern, info.Name())
    if ok {
        if dirFlag := info.IsDir(); dirFlag {
            //fmt.Println("skip dir:",path1)
            return filepath.SkipDir
        }
        //fmt.Println("skip file:",path1)
        return nil
    }
    if dirFlag := info.IsDir(); dirFlag {
       //fmt.Println("skip dir:",path1)
       return nil
    }

    makerHash(path1)
    return nil
}

func main() {
    var err0 error

    t1 := time.Now()

    hashmapfile, err0 = os.OpenFile("./hashmap.txt", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
    check(err0)
    defer hashmapfile.Close()

    filepath.Walk("./", walkFunc)

    t2 := time.Now()
    fmt.Printf("filecount:%d, cost:%d\n", count, t2.Sub(t1)/1000000000)
}
