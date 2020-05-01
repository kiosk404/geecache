/**
* @Author: kiosk
* @Mail: weijiaxiang007@foxmail.com
* @Date: 2020/5/1
**/
package geecache

import (
	"fmt"
	pb "geecache/geecachepb"
	"geecache/singleflight"
	"log"
	"sync"
)

// 缓存不存在时 用户的回调函数

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// Group 缓存命名空间，每个Group 拥有唯一的一个name
type Group struct {
	name 		string
	getter 		Getter  // mainCache 取不到的回调函数
	mainCache 	cache   // cache 对象
	peers       PeerPicker // 其余节点
	// 通过 singleflight 模块防止收敛所有调度，防止缓存被击穿
	loader 		*singleflight.Group
}

var (
	mu  sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()

	g := &Group{
		name:      	name,
		getter:    	getter,
		mainCache: 	cache{cacheBytes: cacheBytes},
		loader: 	&singleflight.Group{},
	}

	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// 获取缓存
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		log.Println("[GeeCache hit]")
		return v, nil
	}

	return g.load(key)   // 若缓存不存在从其他节点（如本地, 或其他节点）获取
}

func (g *Group) load(key string) (value ByteView, err error) {   // 如果 远端没有
	//if g.peers != nil {
	//	if peer, ok := g.peers.PickPeer(key); ok {
	//		if value, err = g.getFromPeer(peer, key); err != nil {
	//			return value, nil
	//		}
	//		log.Println("[GeeCache] Failed to get from peer", err)
	//	}
	//}
	//return g.getLocally(key)

	viewi, err := g.loader.Do(key, func() (i interface{}, err error) {  // 使用 loader.Do 封装，防止缓存被击穿
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok{
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, err
				}
			}
		}
		return g.getLocally(key)
	})
	if err == nil {
		return viewi.(ByteView), nil
	}
	return
}

// RegisterPeers registers a PeerPicker for choosing remote peer
func (g *Group) RegisterPeers(peers PeerPicker) {   // 注册节点
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

// 获取缓存实际操作
func (g *Group) getFromPeer(peer PeerGetter,key string) (ByteView, error) {  // 从远处节点获取缓存
	req := &pb.Request{
		Group: g.name,
		Key:   key,
	}
	res := &pb.Response{}
	err := peer.Get(req, res)
	if err != nil {
		return ByteView{}, err
	}
	g.populateCache(key, ByteView{b: res.Value})    // 拿到数据，更新缓存
	return ByteView{b: res.Value}, nil
}

func (g *Group) getLocally(key string) (ByteView, error) {   // 从本地自定义方法中获取缓存
	bytes, err := g.getter.Get(key)   // 这里是从 用户提供的 自定义函数中取的数据
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b : cloneBytes(bytes)}
	g.populateCache(key, value)        // 拿到数据，更新缓存
	return value, nil
}

func(g *Group) populateCache(key string, value ByteView) {  // 更新主缓存
	g.mainCache.add(key, value)
}

