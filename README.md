## Geecache
分布式缓存

From https://geektutu.com/post/geecache-day1.html

### 简介
- 采用LRU 缓存淘汰算法
- 支持单机并发的缓存
- 分布式多节点一致性hash
- 缓存节点之间通过 pb 通信
- 支持高防缓存击穿

创建一个缓存Group命名为 `MyCache`, 缓存大小2字节, 如果缓存不存在, 在`geecache.GetterFunc()`中定义如何获取待缓存的数据，这里就直接在字典中获取
``` go
func createGroup() *geecache.Group {
	return geecache.NewGroup("MyCache", 2<<10, geecache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}
```
可以通过 `./server -port=8001 &` 启动缓存节点,默认的peer缓存节点为 
``` bash
8001
8002
8003
```
如果参数中加上 `-api 1`,可以启动一个api server。

缓存结果获取
``` bash
# http://localhost:9999/api?key=Tom
```
如果缓存中有就直接返回,如果缓存中没有。从slow db中取，如果slow db没有就返回空。

### 待完成
缓存过期
