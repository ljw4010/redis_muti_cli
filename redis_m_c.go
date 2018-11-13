// redisCli.go
package main

import (
    "runtime"

    "flag"
    "sync"
    //  "errors"
    "fmt"
    "time"

    "github.com/garyburd/redigo/redis"
)

var (
    addr, pwd, op                   string
    maxidle, maxactive, idletimeout int
    count                           int64

    wait, h bool
)

func init() {

    flag.StringVar(&addr, "a", "", "redis addr")
    flag.StringVar(&pwd, "p", "", "redis pwd")
    flag.StringVar(&op, "o", "", "redis opration[insert,get,del]")
    flag.IntVar(&maxidle, "i", 2, "redis maxidle")
    flag.IntVar(&maxactive, "x", 4, "redis maxactive")
    flag.IntVar(&idletimeout, "t", 10, "redis idletimeout")
    flag.Int64Var(&count, "c", 1000, "redis insert key count")
    flag.BoolVar(&h, "h", false, "help")
    flag.BoolVar(&wait, "w", false, "redis wait")
    flag.Usage = usage
}

func main() {
    runtime.GOMAXPROCS(runtime.NumCPU())
    flag.Parse()
    if h {
        flag.Usage()
    }
    err := initCache(addr, pwd, maxidle, maxactive, idletimeout, wait)
    if err != nil {
        fmt.Println("init redis failed:", err.Error())
        return
    } else {
        fmt.Println("init redis success")
    }
    start := time.Now().Unix()
    switch op {
    case "insert":
        InsertInfo(count)
        end := time.Now().Unix()
        fmt.Println("start", start, "end:", end, "use:", end-start, "V:", count/(end-start))
    case "get":

        GetKey(count)
        end := time.Now().Unix()
        fmt.Println("start", start, "end:", end, "use:", end-start, "V:", count/(end-start))
    case "del":

        DelKey(count)
        end := time.Now().Unix()
        fmt.Println("start", start, "end:", end, "use:", end-start, "V:", count/(end-start))
    default:
        flag.Usage()
    }

}

func usage() {
    fmt.Println(`version:1.0.0
Usage：./redis_m_c [-a|-p|-o|-i|-x|-t|-c]
Options:
`)
    flag.PrintDefaults()
}

type Cache struct {
    pool *redis.Pool
}

var cache *Cache

var w sync.WaitGroup

func initCache(addr, pwd string, maxidle, maxactive, idletimeout int, wait bool) error {
    cache = &Cache{}

    cache.pool = &redis.Pool{
        Wait:        wait,
        MaxIdle:     maxidle,
        MaxActive:   maxactive,
        IdleTimeout: time.Duration(idletimeout) * time.Second,
        Dial: func() (redis.Conn, error) {
            option := redis.DialConnectTimeout(5 * time.Second)
            c, err := redis.Dial("tcp", addr, option)
            if err != nil {
                fmt.Println("Redis Connect Failed", err.Error())
                return nil, err
            }

            if pwd != "" {
                _, err = c.Do("AUTH", pwd)
                if err != nil {
                    c.Close()
                    return nil, err
                }
            }
            return c, nil
        },
    }

    return nil
}

func InsertInfo(count int64) /*(map[string]string, error)*/ {

    var i int64
    for i = 1; i <= count; i++ {
        w.Add(1)
        conn := cache.pool.Get()
        key := fmt.Sprintf("test%d", i)
        value := fmt.Sprintf("test%d:%d", i, i)
        go func() {
            _, err := conn.Do("SET", key, value)
            if err != nil {
                fmt.Println("insert failed,key:", key, err.Error())
            } else {
                fmt.Println("insert success,key:", key)
            }
            w.Done()
            conn.Close()

        }()

    }

    w.Wait()

}

func GetKey(count int64) {

    var i int64
    for i = 1; i <= count; i++ {
        w.Add(1)
        key := fmt.Sprintf("test%d", i)
        conn := cache.pool.Get()
        go func() {
            value, err := redis.String(conn.Do("GET", key))
            if err != nil {
                fmt.Println("get key  failed:", err.Error())
            } else {
                fmt.Println("get", "key:", key, "value:", value)
            }
            w.Done()
            conn.Close()
        }()
    }

    w.Wait()

}

func DelKey(count int64) {

    var i int64

    for i = 1; i <= count; i++ {
        w.Add(1)
        conn := cache.pool.Get()
        key := fmt.Sprintf("test%d", i)
        go func() {
            _, err := conn.Do("DEL", key)
            if err != nil {
                fmt.Println("del key  failed:", err.Error())
            } else {
                fmt.Println("del success,key", key)
            }
            w.Done()
            conn.Close()
        }()
    }

    w.Wait()

}

