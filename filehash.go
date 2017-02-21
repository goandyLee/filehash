package main

import (
        "path/filepath"
        "crypto/sha1"
        "bytes"
        "time"
        "fmt"
        "os"
        "io"
)

var (
    pattern string
    count = 0
    buffer bytes.Buffer
)

func check(err error) {
    if err != nil {
        panic(err)
    }
}

func makerHash(filepath string) error {
    tempFile, err1 := os.Open(filepath)
    check(err1)
    defer tempFile.Close()

    fileHash := sha1.New()
    _, err2 := io.Copy(fileHash, tempFile)
    check(err2)

    tempFileInfo, err3 := tempFile.Stat()
    check(err3)

    tempStr := fmt.Sprintf("%s,%X,%d\n", tempFileInfo.Name(), fileHash.Sum(nil), tempFileInfo.Size())
    // 海量文件另议
    buffer.WriteString(tempStr)
    count++

    return nil
}

/**
 * walkFunc:  
 * Desc:
 *     在指定目录下递归统计文件
 */
func walkFunc(path1 string, info os.FileInfo, err error) error {
    ok, err := filepath.Match(pattern, info.Name())
    if ok {
        if dirFlag := info.IsDir(); dirFlag {
            // 遇到匹配目录，跳过目录下文件及子目录
            //fmt.Println("skip dir:",path1)
            return filepath.SkipDir
        }
        // 遇到匹配文件，跳过当前文件
        //fmt.Println("skip file:",path1)
        return nil
    }
    if dirFlag := info.IsDir(); dirFlag {
       //fmt.Println("skip dir:",path1)
       return nil
    }

    err = makerHash(path1)
    check(err)

    return nil
}

func main() {
    if argc := len(os.Args); argc != 3{
        fmt.Println("usage: filehash DIR PATTERN")
        return
    } 

    pattern = os.Args[2]
    dir := os.Args[1]

    t1 := time.Now()

    filepath.Walk(dir, walkFunc)

    hashmapfile, err0 := os.OpenFile("./hashmap.txt", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
    check(err0)
    defer hashmapfile.Close()

    _, err0 = hashmapfile.WriteString(buffer.String())
    check(err0)

    t2 := time.Now()
    fmt.Printf("filecount:%d, cost:%d\n", count, t2.Sub(t1)/1000000000)
}